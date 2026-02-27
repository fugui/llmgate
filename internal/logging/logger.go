package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"llmgate/internal/config"
)

// Logger handles application logging
type Logger struct {
	config    config.LoggingConfig
	output    io.Writer
	mu        sync.Mutex
	file      *os.File
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// RequestLog represents an HTTP request log entry
type RequestLog struct {
	Timestamp   string `json:"timestamp"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Query       string `json:"query,omitempty"`
	ClientIP    string `json:"client_ip"`
	UserAgent   string `json:"user_agent,omitempty"`
	StatusCode  int    `json:"status_code"`
	Duration    string `json:"duration"`
	Provider    string `json:"provider,omitempty"`
	Error       string `json:"error,omitempty"`
}

// New creates a new Logger instance
func New(cfg config.LoggingConfig) *Logger {
	logger := &Logger{
		config: cfg,
		output: os.Stdout,
	}

	// Open log file if specified
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.output = file
			logger.file = file
		} else {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		}
	}

	return logger
}

// Close closes the logger and any open files
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// log writes a log entry with the specified level
func (l *Logger) log(level, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.config.Format == "json" {
		data, _ := json.Marshal(entry)
		fmt.Fprintln(l.output, string(data))
	} else {
		// Text format
		if len(fields) > 0 {
			fmt.Fprintf(l.output, "[%s] %s: %s %v\n", entry.Timestamp, level, message, fields)
		} else {
			fmt.Fprintf(l.output, "[%s] %s: %s\n", entry.Timestamp, level, message)
		}
	}
}

// shouldLog checks if the given log level should be logged
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	configuredLevel := levels[l.config.Level]
	messageLevel := levels[level]

	return messageLevel >= configuredLevel
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log("debug", message, f)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log("info", message, f)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log("warn", message, f)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log("error", message, f)
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(req *RequestLog) {
	if !l.config.RequestLogging {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.config.Format == "json" {
		data, _ := json.Marshal(req)
		fmt.Fprintln(l.output, string(data))
	} else {
		fmt.Fprintf(l.output, "[%s] %s %s %d %s\n",
			req.Timestamp, req.Method, req.Path, req.StatusCode, req.Duration)
	}
}

// RequestLoggingMiddleware returns middleware that logs HTTP requests
func (l *Logger) RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		reqLog := &RequestLog{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Method:     r.Method,
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			ClientIP:   r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			StatusCode: wrapped.statusCode,
			Duration:   duration.String(),
		}

		l.LogRequest(reqLog)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LogBody logs request or response body for debugging
func (l *Logger) LogBody(prefix string, body []byte, isResponse bool) {
	if l.config.Level != "debug" {
		return
	}

	// Limit body size for logging
	maxSize := 1024
	if len(body) > maxSize {
		body = body[:maxSize]
	}

	// Try to format as JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		body = prettyJSON.Bytes()
	}

	typeStr := "request"
	if isResponse {
		typeStr = "response"
	}

	l.Debug(fmt.Sprintf("%s %s body", prefix, typeStr), map[string]interface{}{
		"body": string(body),
	})
}
