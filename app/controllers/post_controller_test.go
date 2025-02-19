package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"cheeseburger/app/models"
	"cheeseburger/app/repositories/mock"
	"cheeseburger/app/services"

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

func TestPostController(t *testing.T) {
	controller, service, postRepo := setupTestPostController(t)
	router := setupRouter(controller)

	t.Run("create post", func(t *testing.T) {
		payload := `{
			"title": "Test Post",
			"content": "This is a test post content"
		}`

		req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
