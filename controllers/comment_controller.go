package controllers

import (
	"html/template"
	"net/http"

	"github.com/dgraph-io/badger/v4"
)

// CommentController handles HTTP requests for comments.
type CommentController struct {
	DB *badger.DB
}

// New displays the form for creating a new comment.
func (cc *CommentController) New(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/layout.html", "views/comments/new.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", nil)
}

// NewCommentController creates and returns a new CommentController without a DB.
func NewCommentController() *CommentController {
	return &CommentController{}
}

// NewCommentControllerWithDB creates and returns a new CommentController with a DB instance.
func NewCommentControllerWithDB(db *badger.DB) *CommentController {
	return &CommentController{DB: db}
}

// Index handles listing all comments.
func (cc *CommentController) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement comment listing logic.
	w.Write([]byte("Listing comments"))
}

// Create handles creating a new comment.
func (cc *CommentController) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement new comment creation logic.
	w.Write([]byte("Creating comment"))
}

// Edit handles editing an existing comment.
func (cc *CommentController) Edit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement comment update logic.
	w.Write([]byte("Editing comment"))
}
