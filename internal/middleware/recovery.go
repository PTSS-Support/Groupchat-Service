package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return a 500 Internal Server Error
				respondWithError(w, http.StatusInternalServerError, "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
