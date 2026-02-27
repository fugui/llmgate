package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"llmgate/internal/config"
)

// RateLimiter handles request rate limiting
type RateLimiter struct {
	config    config.RateLimitConfig
	limiters  map[string]*rate.Limiter
	mu        sync.RWMutex
	cleanupInterval time.Duration
}

// New creates a new RateLimiter instance
func New(cfg config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:    cfg,
		limiters:  make(map[string]*rate.Limiter),
		cleanupInterval: 10 * time.Minute,
	}

	// Start cleanup goroutine
	if cfg.PerIP {
		go rl.cleanupLoop()
	}

	return rl
}

// Allow checks if a request should be allowed based on rate limiting rules
func (rl *RateLimiter) Allow(r *http.Request) bool {
	if !rl.config.Enabled {
		return true
	}

	key := rl.getKey(r)
	limiter := rl.getLimiter(key)

	return limiter.Allow()
}

// getKey returns the rate limit key for a request
func (rl *RateLimiter) getKey(r *http.Request) string {
	if !rl.config.PerIP {
		return "global"
	}

	// Try to get the real client IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

// getLimiter returns or creates a rate limiter for the given key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}

	// Create new limiter
	limiter = rate.NewLimiter(
		rate.Limit(rl.config.RequestsPerSecond),
		rl.config.BurstSize,
	)
	rl.limiters[key] = limiter

	return limiter
}

// cleanupLoop periodically removes stale limiters
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		// In a production system, we'd track last access time
		// and remove entries that haven't been used recently
		// For simplicity, we'll just keep them
	}
}

// GetRateLimitInfo returns the current rate limit status for a key
func (rl *RateLimiter) GetRateLimitInfo(key string) (limit rate.Limit, burst int, tokens float64) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		return rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize, float64(rl.config.BurstSize)
	}

	return limiter.Limit(), limiter.Burst(), limiter.Tokens()
}

// Middleware returns HTTP middleware that applies rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(r) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
