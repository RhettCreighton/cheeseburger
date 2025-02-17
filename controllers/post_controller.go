package controllers

import (
	"net/http"

	"github.com/dgraph-io/badger/v4"
)

// PostController handles HTTP requests for blog posts.
type PostController struct {
	DB *badger.DB
}

// NewPostController creates and returns a new PostController without a DB.
func NewPostController() *PostController {
	return &PostController{}
}

// NewPostControllerWithDB creates and returns a new PostController with a DB instance.
func NewPostControllerWithDB(db *badger.DB) *PostController {
	return &PostController{DB: db}
}

// Index handles listing all posts.
func (pc *PostController) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement post listing logic.
	w.Write([]byte("Listing posts"))
}

// Show handles displaying a single post.
func (pc *PostController) Show(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logic to show a single post.
	w.Write([]byte("Showing post"))
}

// Create handles creating a new post.
func (pc *PostController) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement new post creation logic.
	w.Write([]byte("Creating post"))
}

// Edit handles editing an existing post.
func (pc *PostController) Edit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement post update logic.
	w.Write([]byte("Editing post"))
}
