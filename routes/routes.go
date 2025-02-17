package routes

import (
	"net/http"

	"cheeseburger/controllers"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

// SetupRoutes defines the application's routes and returns a router.
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	postController := controllers.NewPostController()
	commentController := controllers.NewCommentController()

	// Routes for posts
	router.HandleFunc("/posts", postController.Index).Methods("GET")
	router.HandleFunc("/posts/{id}", postController.Show).Methods("GET")
	router.HandleFunc("/posts", postController.Create).Methods("POST")
	router.HandleFunc("/posts/{id}", postController.Edit).Methods("PUT")

	// Routes for comments on posts
	router.HandleFunc("/posts/{postId}/comments", commentController.Index).Methods("GET")
	router.HandleFunc("/posts/{postId}/comments", commentController.Create).Methods("POST")
	router.HandleFunc("/comments/{id}", commentController.Edit).Methods("PUT")

	return router
}

// SetupMVCRoutes defines the MVC application's routes and returns a router, using the provided Badger DB.
func SetupMVCRoutes(db *badger.DB) *mux.Router {
	router := mux.NewRouter()

	postController := controllers.NewPostControllerWithDB(db)
	commentController := controllers.NewCommentControllerWithDB(db)

	// Routes for posts
	router.HandleFunc("/posts", postController.Index).Methods("GET")
	router.HandleFunc("/posts/{id}", postController.Show).Methods("GET")
	router.HandleFunc("/posts", postController.Create).Methods("POST")
	router.HandleFunc("/posts/{id}", postController.Edit).Methods("PUT")

	// Routes for comments on posts
	router.HandleFunc("/posts/{postId}/comments", commentController.Index).Methods("GET")
	router.HandleFunc("/posts/{postId}/comments", commentController.Create).Methods("POST")
	router.HandleFunc("/comments/{id}", commentController.Edit).Methods("PUT")

	return router
}

// StartServer starts the HTTP server on the specified address with the given router.
func StartServer(addr string, router http.Handler) error {
	return http.ListenAndServe(addr, router)
}
