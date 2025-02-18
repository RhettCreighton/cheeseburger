package controllers

import (
	"html/template"
	"net/http"

	"github.com/dgraph-io/badger/v4"
)

// PostController handles HTTP requests for blog posts.
type PostController struct {
	DB *badger.DB
}

// New displays the form for creating a new post.
func (p *PostController) New(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/layout.html", "views/posts/new.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", nil)
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
	tmpl, err := template.ParseFiles("views/layout.html", "views/posts/index.html")
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Posts []interface{}
	}{
		Posts: []interface{}{"Test Post"},
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
