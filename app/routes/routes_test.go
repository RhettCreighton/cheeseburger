package routes

import (
	"cheeseburger/app/controllers"
	"cheeseburger/app/middleware"
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"cheeseburger/app/services"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T, db *badger.DB) (*mux.Router, *controllers.PostController, *controllers.CommentController) {
	// Use test helpers from test_helpers.go for setting up templates and database.
	tmpDir := setupTestTemplates(t)

	var postController *controllers.PostController
	var commentController *controllers.CommentController
	router := mux.NewRouter()

	// Apply global middleware.
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Create API subrouter with JSON content type middleware.
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.ContentTypeJSON)

	// Create repositories and services.
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)

	postService := services.NewPostService(postRepo, commentRepo)
	commentService := services.NewCommentService(commentRepo, postRepo)

	// Create controllers with DB and template path.
	postController = controllers.NewPostControllerWithDBAndPath(db, tmpDir)
	commentController = controllers.NewCommentControllerWithDBAndPath(db, tmpDir)

	// Set services.
	postController.SetService(postService)
	commentController.SetService(commentService)

	// Create a test post.
	post := &models.Post{
		Title:   "Test Post",
		Content: "This is a test post",
	}
	err := postService.CreatePost(post)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Set up NotFoundHandler for API routes.
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
			return
		}
		http.NotFound(w, r)
	})

	// API routes.
	apiPosts := apiRouter.PathPrefix("/posts").Subrouter()
	apiPosts.HandleFunc("", postController.Index).Methods("GET")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	apiPosts.HandleFunc("", postController.Create).Methods("POST")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// API Comments endpoints.
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	apiRouter.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	apiRouter.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	// Serve static files.
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Web routes.
	router.HandleFunc("/", postController.Index).Methods("GET")

	// Posts web endpoints.
	posts := router.PathPrefix("/posts").Subrouter()
	posts.HandleFunc("", postController.Index).Methods("GET")
	posts.HandleFunc("/new", postController.New).Methods("GET")
	posts.HandleFunc("", postController.Create).Methods("POST")
	posts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	posts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	posts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// Comments web endpoints.
	posts.HandleFunc("/{postId:[0-9]+}/comments/new", commentController.New).Methods("GET")
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	router.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	router.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	// API routes with JSON content type.
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.ContentTypeJSON)

	// Posts API endpoints.
	apiPosts = api.PathPrefix("/posts").Subrouter()
	apiPosts.HandleFunc("", postController.Index).Methods("GET")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	apiPosts.HandleFunc("", postController.Create).Methods("POST")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// Comments API endpoints.
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	return router, postController, commentController
}

func TestSetupRoutes(t *testing.T) {
	// Use helper functions from test_helpers.go.
	setupTestTemplates(t)
	db := setupTestDB(t)
	router, _, _ := setupTestRouter(t, db)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "GET posts",
			method:         "GET",
			path:           "/api/posts",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "GET single post",
			method:         "GET",
			path:           "/api/posts/1",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "GET post comments",
			method:         "GET",
			path:           "/api/posts/1/comments",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "Invalid post ID",
			method:         "GET",
			path:           "/api/posts/invalid",
			expectedStatus: http.StatusNotFound,
			expectedHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if strings.HasPrefix(tt.path, "/api/") {
				req.Header.Set("Accept", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"))
		})
	}
}

func TestSetupMVCRoutes(t *testing.T) {
	setupTestTemplates(t)
	db := setupTestDB(t)
	router, _, _ := setupTestRouter(t, db)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		isAPI          bool
	}{
		{
			name:           "Home page",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			isAPI:          false,
		},
		{
			name:           "New post form",
			method:         "GET",
			path:           "/posts/new",
			expectedStatus: http.StatusOK,
			isAPI:          false,
		},
		{
			name:           "API get posts",
			method:         "GET",
			path:           "/api/posts",
			expectedStatus: http.StatusOK,
			isAPI:          true,
		},
		{
			name:           "New comment form",
			method:         "GET",
			path:           "/posts/1/comments/new",
			expectedStatus: http.StatusOK,
			isAPI:          false,
		},
		{
			name:           "Static file",
			method:         "GET",
			path:           "/static/test.txt",
			expectedStatus: http.StatusNotFound,
			isAPI:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.isAPI || strings.HasPrefix(tt.path, "/api/") {
				req.Header.Set("Accept", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.isAPI {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestStartServer(t *testing.T) {
	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	go func() {
		err := StartServer("localhost:0", router) // Port 0 picks a random available port
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("StartServer failed: %v", err)
		}
	}()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}
