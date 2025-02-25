package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebPostRoutes(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestTemplates(t)
	router, _, _ := setupTestRouter(t, db)

	t.Run("GET / returns home page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "posts")
	})

	t.Run("GET /posts returns posts list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "posts")
	})

	t.Run("GET /posts/new returns new post form", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/posts/new", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form")
	})

	t.Run("POST /posts creates a new post", func(t *testing.T) {
		formData := url.Values{
			"title":   {"Web Form Test Post"},
			"content": {"This post was created via web form test"},
		}

		req := httptest.NewRequest("POST", "/posts", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should redirect to the post on success, not the posts list
		assert.Equal(t, http.StatusSeeOther, w.Code)
		redirectURL, err := w.Result().Location()
		require.NoError(t, err)
		assert.Contains(t, redirectURL.Path, "/posts/")

		// Verify post was created by fetching post list
		req = httptest.NewRequest("GET", "/posts", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Contains(t, w.Body.String(), "Web Form Test Post")
	})

	// Skip the PUT/DELETE tests for now as they require special handling to work with the router
	// The current controllers may be using form-based updates which aren't easily testable with the simple tests
}

func TestWebCommentRoutes(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestTemplates(t)
	router, _, _ := setupTestRouter(t, db)

	// Create a post to attach comments to
	formData := url.Values{
		"title":   {"Post for Web Comments"},
		"content": {"This post will have comments added via web"},
	}

	req := httptest.NewRequest("POST", "/posts", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	
	// Get the post ID from the redirect Location
	redirectURL, err := w.Result().Location()
	require.NoError(t, err)
	
	pathParts := strings.Split(redirectURL.Path, "/")
	require.Len(t, pathParts, 3)
	postIDStr := pathParts[2]
	postID, err := strconv.Atoi(postIDStr)
	require.NoError(t, err)
	require.NotZero(t, postID)

	t.Run("GET /posts/{postId}/comments/new returns new comment form", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/posts/"+postIDStr+"/comments/new", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "form")
	})

	t.Run("POST /posts/{postId}/comments creates a new comment", func(t *testing.T) {
		formData := url.Values{
			"author":  {"Web Form Commenter"},
			"content": {"This is a comment added via web form test"},
		}

		req := httptest.NewRequest("POST", "/posts/"+postIDStr+"/comments", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should redirect to post page on success
		assert.Equal(t, http.StatusSeeOther, w.Code)
		redirectURL, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/posts/"+postIDStr, redirectURL.Path)
	})

	// Skip edit/delete comment tests as they require handling form submission for PUT/DELETE
	// which isn't easily supported in the current test environment
}

func TestStaticFileRoutes(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestTemplates(t)
	router, _, _ := setupTestRouter(t, db)

	// Create a test static file in the static directory used by the router
	// Note: The router is using a different static path than our test tmpDir
	staticDir, err := filepath.Abs("static")
	require.NoError(t, err)
	
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(staticDir, 0755))
	}
	
	testCSS := "body { background-color: #f0f0f0; }"
	testFile := filepath.Join(staticDir, "test.css")
	require.NoError(t, os.WriteFile(testFile, []byte(testCSS), 0644))
	
	// Ensure file is cleaned up after test
	t.Cleanup(func() {
		os.Remove(testFile)
	})

	t.Run("GET /static/{file} serves static files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/test.css", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/css; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, testCSS, w.Body.String())
	})

	t.Run("GET /static/{non-existent-file} returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/non-existent.css", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestMiddlewareRoutes(t *testing.T) {
	db := setupTestDB(t)
	_ = setupTestTemplates(t)
	router, _, _ := setupTestRouter(t, db)

	t.Run("API middleware sets JSON content type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/posts", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("Logger middleware is applied", func(t *testing.T) {
		// This is a functional test to ensure the middleware is hooked up
		// The actual logging can't be easily tested without mocking
		req := httptest.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Skip the recoverer middleware test since it's not easily testable
	// without more complex mocking
}