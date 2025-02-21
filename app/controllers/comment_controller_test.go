package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"cheeseburger/app/models"
	"cheeseburger/app/repositories/mock"
	"cheeseburger/app/services"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestCommentController(t *testing.T) (*CommentController, *services.CommentService, *services.PostService) {
	postRepo := mock.NewPostRepository()
	commentRepo := mock.NewCommentRepository()
	postService := services.NewPostService(postRepo, commentRepo)
	commentService := services.NewCommentService(commentRepo, postRepo)
	controller := &CommentController{
		commentService: commentService,
		templates:      make(map[string]*template.Template),
	}
	return controller, commentService, postService
}

func setupCommentRouter(controller *CommentController) *mux.Router {
	router := mux.NewRouter()

	// Register routes manually
	router.HandleFunc("/posts/{postId:[0-9]+}/comments", controller.Create).Methods("POST")
	router.HandleFunc("/posts/{postId:[0-9]+}/comments", controller.Index).Methods("GET")
	router.HandleFunc("/posts/{postId:[0-9]+}/comments/new", controller.New).Methods("GET")
	router.HandleFunc("/comments/{id:[0-9]+}", controller.Edit).Methods("PUT")
	router.HandleFunc("/comments/{id:[0-9]+}", controller.Delete).Methods("DELETE")

	return router
}

func TestNewCommentController(t *testing.T) {
	// Create temporary view directories and files
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "app", "views")

	dirs := []string{
		filepath.Join(viewsDir, "comments"),
		filepath.Join(viewsDir, "shared"),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	files := map[string]string{
		filepath.Join(viewsDir, "layout.html"):             `{{define "layout"}}{{template "content" .}}{{end}}`,
		filepath.Join(viewsDir, "comments", "new.html"):    `{{define "content"}}New{{end}}`,
		filepath.Join(viewsDir, "comments", "list.html"):   `{{define "content"}}List{{end}}`,
		filepath.Join(viewsDir, "shared", "comments.html"): `{{define "comments"}}Comments{{end}}`,
	}
	for path, content := range files {
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Set working directory to temp dir for template loading
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer os.Chdir(originalWd)

	t.Run("NewCommentController", func(t *testing.T) {
		controller := NewCommentController()
		assert.NotNil(t, controller)
		assert.NotNil(t, controller.templates)

		// Verify all templates are loaded
		expectedTemplates := []string{"new", "list"}
		for _, name := range expectedTemplates {
			assert.NotNil(t, controller.templates[name], "Template %s should be loaded", name)
		}
	})

	t.Run("NewCommentControllerWithDB", func(t *testing.T) {
		// Create a temporary DB
		dbPath := filepath.Join(t.TempDir(), "test.db")
		db, err := badger.Open(badger.DefaultOptions(dbPath))
		assert.NoError(t, err)
		defer db.Close()

		controller := NewCommentControllerWithDB(db)
		assert.NotNil(t, controller)
		assert.NotNil(t, controller.templates)
		assert.NotNil(t, controller.commentService)
	})

	t.Run("New action", func(t *testing.T) {
		controller := NewCommentController()
		req := httptest.NewRequest(http.MethodGet, "/posts/1/comments/new", nil)
		req.Header.Set("Accept", "application/json")

		// Add route parameters
		router := mux.NewRouter()
		router.HandleFunc("/posts/{postId:[0-9]+}/comments/new", controller.New)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "New")
	})
}

func TestCommentController(t *testing.T) {
	controller, _, postService := setupTestCommentController(t)
	router := setupCommentRouter(controller)

	// Create a test post first
	post := &models.Post{
		Title:   "Test Post",
		Content: "Test Content",
	}
	err := postService.CreatePost(post)
	assert.NoError(t, err)

	t.Run("create comment", func(t *testing.T) {
		payload := `{
			"author": "Test Author",
			"content": "This is a test comment"
		}`

		req := httptest.NewRequest(http.MethodPost, "/posts/"+strconv.Itoa(post.ID)+"/comments", strings.NewReader(payload))
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Comment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotZero(t, response.ID)
		assert.Equal(t, "Test Author", response.Author)
		assert.Equal(t, "This is a test comment", response.Content)
		assert.Equal(t, post.ID, response.PostID)
	})

	t.Run("list comments", func(t *testing.T) {
		// Create multiple comments
		for i := 0; i < 3; i++ {
			payload := `{
				"author": "List Test Author",
				"content": "Content for list test"
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts/"+strconv.Itoa(post.ID)+"/comments", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/posts/"+strconv.Itoa(post.ID)+"/comments", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var comments []*models.Comment
		err := json.Unmarshal(w.Body.Bytes(), &comments)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(comments)) // 3 new + 1 from previous test
		for _, comment := range comments {
			assert.Equal(t, post.ID, comment.PostID)
		}
	})

	t.Run("update comment", func(t *testing.T) {
		payload := `{
			"author": "Updated Author",
			"content": "Updated comment",
			"postId": 1
		}`

		req := httptest.NewRequest(http.MethodPut, "/comments/1", strings.NewReader(payload))
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Comment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Author", response.Author)
		assert.Equal(t, "Updated comment", response.Content)
		assert.Equal(t, post.ID, response.PostID)
	})

	t.Run("delete comment", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/comments/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify comment is deleted by checking the comments list
		req = httptest.NewRequest(http.MethodGet, "/posts/"+strconv.Itoa(post.ID)+"/comments", nil)
		req.Header.Set("Accept", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var comments []*models.Comment
		err := json.Unmarshal(w.Body.Bytes(), &comments)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(comments)) // One less than before
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Run("empty author", func(t *testing.T) {
			payload := `{
				"author": "",
				"content": "Valid content"
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts/"+strconv.Itoa(post.ID)+"/comments", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("empty content", func(t *testing.T) {
			payload := `{
				"author": "Valid Author",
				"content": ""
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts/"+strconv.Itoa(post.ID)+"/comments", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("invalid post ID", func(t *testing.T) {
			payload := `{
				"author": "Valid Author",
				"content": "Valid content"
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts/999/comments", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
