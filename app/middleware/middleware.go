package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// Logger logs information about each request
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		// Log request details using fmt.Printf for proper stdout capture
		fmt.Printf("[ %s ] %s %s took %v\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path, duration)
	})
}

// Recoverer recovers from panics and logs the error
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				println("PANIC:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ContentTypeJSON sets the Content-Type header to application/json for API routes
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set JSON content type for API routes
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			// Create a response wrapper to ensure content type persists
			wrapper := &responseWriter{ResponseWriter: w}
			wrapper.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(wrapper, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// responseWriter wraps http.ResponseWriter to ensure headers can be set after writing
type responseWriter struct {
	http.ResponseWriter
	written bool
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}
