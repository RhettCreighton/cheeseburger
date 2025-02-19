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
		req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/posts/"+strconv.Itoa(post.ID)+"/comments", nil)
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
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
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
