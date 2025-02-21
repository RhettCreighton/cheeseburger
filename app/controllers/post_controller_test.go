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

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestPostController(t *testing.T) (*PostController, *services.PostService, *mock.PostRepository) {
	postRepo := mock.NewPostRepository()
	commentRepo := mock.NewCommentRepository()
	postService := services.NewPostService(postRepo, commentRepo)
	controller := &PostController{
		postService: postService,
		templates:   make(map[string]*template.Template),
	}
	return controller, postService, postRepo
}

func setupRouter(controller *PostController) *mux.Router {
	router := mux.NewRouter()

	// Register routes manually since we don't have access to the RegisterRoutes method
	router.HandleFunc("/posts", controller.Create).Methods("POST")
	router.HandleFunc("/posts", controller.Index).Methods("GET")
	router.HandleFunc("/posts/{id:[0-9]+}", controller.Show).Methods("GET")
	router.HandleFunc("/posts/{id:[0-9]+}", controller.Edit).Methods("PUT")
	router.HandleFunc("/posts/{id:[0-9]+}", controller.Delete).Methods("DELETE")

	return router
}

func TestNewPostController(t *testing.T) {
	// Create temporary view directories and files
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "app", "views")

	dirs := []string{
		filepath.Join(viewsDir, "posts"),
		filepath.Join(viewsDir, "shared"),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	files := map[string]string{
		filepath.Join(viewsDir, "layout.html"):             `{{define "layout"}}{{template "content" .}}{{end}}`,
		filepath.Join(viewsDir, "posts", "index.html"):     `{{define "content"}}Index{{end}}`,
		filepath.Join(viewsDir, "posts", "show.html"):      `{{define "content"}}Show{{end}}`,
		filepath.Join(viewsDir, "posts", "new.html"):       `{{define "content"}}New{{end}}`,
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

	t.Run("NewPostController", func(t *testing.T) {
		controller := NewPostController()
		assert.NotNil(t, controller)
		assert.NotNil(t, controller.templates)

		// Verify all templates are loaded
		expectedTemplates := []string{"index", "show", "new"}
		for _, name := range expectedTemplates {
			assert.NotNil(t, controller.templates[name], "Template %s should be loaded", name)
		}
	})

	t.Run("NewPostControllerWithDB", func(t *testing.T) {
		// Create a temporary DB
		dbPath := filepath.Join(t.TempDir(), "test.db")
		db, err := badger.Open(badger.DefaultOptions(dbPath))
		assert.NoError(t, err)
		defer db.Close()

		controller := NewPostControllerWithDB(db)
		assert.NotNil(t, controller)
		assert.NotNil(t, controller.templates)
		assert.NotNil(t, controller.postService)
	})

	t.Run("New action", func(t *testing.T) {
		controller := NewPostController()
		req := httptest.NewRequest(http.MethodGet, "/posts/new", nil)
		w := httptest.NewRecorder()

		controller.New(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "New")
	})
}

func TestPostController(t *testing.T) {
	controller, service, postRepo := setupTestPostController(t)
	router := setupRouter(controller)

	t.Run("create post", func(t *testing.T) {
		payload := `{
			"title": "Test Post",
			"content": "This is a test post content"
		}`

		req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json") // Still need Content-Type for POST/PUT requests
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Post
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotZero(t, response.ID)
		assert.Equal(t, "Test Post", response.Title)
		assert.Equal(t, "This is a test post content", response.Content)
	})

	t.Run("get post", func(t *testing.T) {
		// Create a post first
		post := &models.Post{
			Title:   "Test Post",
			Content: "Test Content",
		}
		err := service.CreatePost(post)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/posts/"+strconv.Itoa(post.ID), nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Post
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, post.Title, response.Title)
		assert.Equal(t, post.Content, response.Content)
	})

	t.Run("update post", func(t *testing.T) {
		payload := `{
			"title": "Updated Title",
			"content": "Updated content"
		}`

		req := httptest.NewRequest(http.MethodPut, "/posts/1", strings.NewReader(payload))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json") // Still need Content-Type for PUT requests
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Post
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", response.Title)
		assert.Equal(t, "Updated content", response.Content)
	})

	t.Run("delete post", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/posts/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify post is deleted
		req = httptest.NewRequest(http.MethodGet, "/posts/1", nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list posts", func(t *testing.T) {
		// Clear the repository first
		postRepo.Clear()

		// Create multiple posts
		for i := 0; i < 3; i++ {
			post := &models.Post{
				Title:   "List Test Post",
				Content: "Content for list test",
			}
			err := service.CreatePost(post)
			assert.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/posts?page=1&per_page=2", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Posts []*models.Post `json:"posts"`
			Page  int            `json:"page"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(response.Posts))
		assert.Equal(t, 1, response.Page)
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Run("empty title", func(t *testing.T) {
			payload := `{
				"title": "",
				"content": "Valid content"
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json") // Still need Content-Type for POST requests
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("empty content", func(t *testing.T) {
			payload := `{
				"title": "Valid Title",
				"content": ""
			}`

			req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json") // Still need Content-Type for POST requests
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
