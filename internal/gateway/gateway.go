package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"llmgate/internal/config"
	"llmgate/internal/logging"
	"llmgate/internal/ratelimit"
)

// Gateway handles LLM API request routing and proxying
type Gateway struct {
	config      *config.Config
	logger      *logging.Logger
	rateLimiter *ratelimit.RateLimiter
	proxies     map[string]*httputil.ReverseProxy
	client      *http.Client
}

// New creates a new Gateway instance
func New(cfg *config.Config, logger *logging.Logger, rateLimiter *ratelimit.RateLimiter) *Gateway {
	g := &Gateway{
		config:      cfg,
		logger:      logger,
		rateLimiter: rateLimiter,
		proxies:     make(map[string]*httputil.ReverseProxy),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	// Initialize reverse proxies for each provider
	g.initializeProxies()

	return g
}

// initializeProxies creates reverse proxies for each configured provider
func (g *Gateway) initializeProxies() {
	for name, provider := range g.config.Providers {
		if provider.BaseURL == "" {
			g.logger.Warn(fmt.Sprintf("Skipping provider %s: no base URL configured", name))
			continue
		}

		targetURL, err := url.Parse(provider.BaseURL)
		if err != nil {
			g.logger.Error(fmt.Sprintf("Invalid base URL for provider %s", name), map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Transport = &loggingTransport{
			base:   g.client.Transport,
			logger: g.logger,
		}
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			g.logger.Error("Proxy error", map[string]interface{}{
				"provider": name,
				"error":    err.Error(),
			})
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "proxy error",
				"message": err.Error(),
			})
		}

		// Modify request to add API key header
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = targetURL.Host

			// Add API key based on provider type
			switch name {
			case "openai":
				if provider.APIKey != "" {
					req.Header.Set("Authorization", "Bearer "+provider.APIKey)
				}
			case "claude":
				if provider.APIKey != "" {
					req.Header.Set("x-api-key", provider.APIKey)
					req.Header.Set("anthropic-version", "2023-06-01")
				}
			case "azure":
				if provider.APIKey != "" {
					req.Header.Set("api-key", provider.APIKey)
				}
			}
		}

		g.proxies[name] = proxy
	}
}

// ServeHTTP implements the http.Handler interface
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if !g.rateLimiter.Allow(r) {
		g.logger.Warn("Rate limit exceeded", map[string]interface{}{
			"client_ip": r.RemoteAddr,
			"path":      r.URL.Path,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "rate limit exceeded",
		})
		return
	}

	// Route the request to the appropriate provider
	g.routeRequest(w, r)
}

// routeRequest determines the target provider and proxies the request
func (g *Gateway) routeRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Find matching routing rule
	rule := g.config.GetRoutingRule(path)
	if rule == nil {
		// Use default provider for unmatched paths
		rule = &config.RoutingRule{
			Path:      path,
			Providers: []string{g.config.Routing.DefaultProvider},
		}
	}

	// Try providers in order until one succeeds
	var lastErr error
	var responseWritten bool
	for _, providerName := range rule.Providers {
		proxy, exists := g.proxies[providerName]
		if !exists {
			g.logger.Warn(fmt.Sprintf("Provider %s not configured", providerName))
			continue
		}

		provider, ok := g.config.GetProvider(providerName)
		if !ok {
			continue
		}

		// Skip provider if BaseURL is empty
		if provider.BaseURL == "" {
			g.logger.Warn(fmt.Sprintf("Provider %s has no base URL configured", providerName))
			continue
		}

		// Clone the request for retry capability
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		// Log request body if configured
		if g.config.Logging.RequestLogging {
			g.logger.LogBody(providerName, body, false)
		}

		// Create new request with body
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))

		// Add provider info to context for logging
		ctx := withProviderContext(r.Context(), providerName)
		r = r.WithContext(ctx)

		g.logger.Info(fmt.Sprintf("Routing request to %s", providerName), map[string]interface{}{
			"path":     path,
			"method":   r.Method,
			"provider": providerName,
		})

		// Create a response writer wrapper to capture the response
		wrapped := &responseCapture{ResponseWriter: w}
		proxy.ServeHTTP(wrapped, r)

		// If successful, return
		if wrapped.statusCode < 500 {
			return
		}

		// Provider returned 5xx error, try next provider
		g.logger.Warn(fmt.Sprintf("Provider %s returned error, trying next", providerName), map[string]interface{}{
			"status_code": wrapped.statusCode,
		})
		lastErr = fmt.Errorf("provider %s returned status %d", providerName, wrapped.statusCode)
		responseWritten = wrapped.written

		// Reset body for next provider
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	// All providers failed
	if lastErr != nil {
		g.logger.Error("All providers failed", map[string]interface{}{
			"error": lastErr.Error(),
		})
		if !responseWritten {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "all providers unavailable",
			})
		}
	}
}

// contextKey is a type for context keys
type contextKey string

const providerContextKey contextKey = "provider"

// withProviderContext adds provider name to context
func withProviderContext(ctx context.Context, provider string) context.Context {
	return context.WithValue(ctx, providerContextKey, provider)
}

// responseCapture wraps ResponseWriter to capture status code
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

type responseWriterFlusher interface {
	http.ResponseWriter
	http.Flusher
}

func (rc *responseCapture) WriteHeader(code int) {
	if !rc.written {
		rc.statusCode = code
		rc.written = true
		rc.ResponseWriter.WriteHeader(code)
	}
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	if !rc.written {
		rc.WriteHeader(http.StatusOK)
	}
	return rc.ResponseWriter.Write(b)
}

func (rc *responseCapture) Flush() {
	if f, ok := rc.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// loggingTransport wraps an http.RoundTripper to log requests
type loggingTransport struct {
	base   http.RoundTripper
	logger *logging.Logger
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if lt.base == nil {
		lt.base = http.DefaultTransport
	}

	start := time.Now()
	resp, err := lt.base.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		lt.logger.Error("Request failed", map[string]interface{}{
			"url":      req.URL.String(),
			"error":    err.Error(),
			"duration": duration.String(),
		})
		return nil, err
	}

	lt.logger.Debug("Request completed", map[string]interface{}{
		"url":          req.URL.String(),
		"status":       resp.StatusCode,
		"duration":     duration.String(),
		"content_type": resp.Header.Get("Content-Type"),
	})

	return resp, err
}

// HealthHandler returns a health check handler
func (g *Gateway) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check provider health
		providers := make(map[string]string)
		for name, cfg := range g.config.Providers {
			if cfg.BaseURL != "" {
				providers[name] = "configured"
			} else {
				providers[name] = "not_configured"
			}
		}

		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"providers": providers,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
}

// AddRequestIDMiddleware adds a request ID to each request
func AddRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), generateRandomString(8))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// CORS Middleware
func CORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// StripPathPrefix strips the given prefix from request URLs
func StripPathPrefix(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		next.ServeHTTP(w, r)
	})
}
