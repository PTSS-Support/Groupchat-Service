package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs information about each HTTP request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture the status code
		lrw := newLoggingResponseWriter(w)

		// Record start time
		start := time.Now()

		// Call the next handler in the chain
		next.ServeHTTP(lrw, r)

		// Calculate request duration
		duration := time.Since(start)

		// Log the request details
		log.Printf(
			"Method: %s | Path: %s | Status: %d | Duration: %v | IP: %s | User-Agent: %s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new loggingResponseWriter
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
