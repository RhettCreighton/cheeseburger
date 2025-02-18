package routes

import (
	"net/http"

	"cheeseburger/controllers"
	"cheeseburger/middleware"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

// SetupRoutes defines the application's routes and returns a router.
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.ContentTypeJSON)

	postController := controllers.NewPostController()
	commentController := controllers.NewCommentController()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Posts API endpoints
	posts := api.PathPrefix("/posts").Subrouter()
	posts.HandleFunc("", postController.Index).Methods("GET")
	posts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	posts.HandleFunc("", postController.Create).Methods("POST")
	posts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	posts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// Comments API endpoints
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	return router
}

// SetupMVCRoutes defines the MVC application's routes and returns a router, using the provided Badger DB.
func SetupMVCRoutes(db *badger.DB) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	postController := controllers.NewPostControllerWithDB(db)
	commentController := controllers.NewCommentControllerWithDB(db)

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Web routes
	router.HandleFunc("/", postController.Index).Methods("GET")

	// Posts web endpoints
	posts := router.PathPrefix("/posts").Subrouter()
	posts.HandleFunc("", postController.Index).Methods("GET")
	posts.HandleFunc("/new", postController.New).Methods("GET")
	posts.HandleFunc("", postController.Create).Methods("POST")
	posts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	posts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	posts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// Comments web endpoints
	posts.HandleFunc("/{postId:[0-9]+}/comments/new", commentController.New).Methods("GET")
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	posts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	router.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	router.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	// API routes with JSON content type
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.ContentTypeJSON)

	// Posts API endpoints
	apiPosts := api.PathPrefix("/posts").Subrouter()
	apiPosts.HandleFunc("", postController.Index).Methods("GET")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Show).Methods("GET")
	apiPosts.HandleFunc("", postController.Create).Methods("POST")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Edit).Methods("PUT")
	apiPosts.HandleFunc("/{id:[0-9]+}", postController.Delete).Methods("DELETE")

	// Comments API endpoints
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Index).Methods("GET")
	apiPosts.HandleFunc("/{postId:[0-9]+}/comments", commentController.Create).Methods("POST")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Edit).Methods("PUT")
	api.HandleFunc("/comments/{id:[0-9]+}", commentController.Delete).Methods("DELETE")

	return router
}

// StartServer starts the HTTP server on the specified address with the given router.
func StartServer(addr string, router http.Handler) error {
	return http.ListenAndServe(addr, router)
}
