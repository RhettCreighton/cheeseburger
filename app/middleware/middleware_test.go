package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Create a channel to signal when handler is done
	done := make(chan bool)

	// Create a channel to pass captured output
	output := make(chan string)

	// Run handler in goroutine
	go func() {
		handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)
		done <- true
	}()

	// Run output capture in goroutine
	go func() {
		<-done // Wait for handler to complete
		w.Close()
		os.Stdout = oldStdout // Restore stdout before reading output

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output <- buf.String()
	}()

	// Wait for output
	logOutput := <-output
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "took")
}

func TestRecoverer(t *testing.T) {
	handler := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error\n", w.Body.String())
}

func TestContentTypeJSON(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedHeader string
	}{
		{
			name:           "API route",
			path:           "/api/test",
			expectedHeader: "application/json",
		},
		{
			name:           "Non-API route",
			path:           "/test",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ContentTypeJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			assert.Equal(t, tt.expectedHeader, contentType)
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Test all middleware working together
	handler := Logger(Recoverer(ContentTypeJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "panic") {
			panic("test panic")
		}
		w.WriteHeader(http.StatusOK)
	}))))

	tests := []struct {
		name           string
		path           string
		expectPanic    bool
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "Normal API request",
			path:           "/api/test",
			expectPanic:    false,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "Panic request",
			path:           "/api/panic",
			expectPanic:    true,
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "text/plain; charset=utf-8", // Error responses use text/plain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
		})
	}
}
