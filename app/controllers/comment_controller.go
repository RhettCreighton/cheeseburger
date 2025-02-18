package controllers

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"cheeseburger/app/services"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

// CommentController handles HTTP requests for comments
type CommentController struct {
	commentService *services.CommentService
	templates      map[string]*template.Template
}

// NewCommentController creates a new CommentController
func NewCommentController() *CommentController {
	return &CommentController{
		templates: loadCommentTemplates(),
	}
}

// NewCommentControllerWithDB creates a new CommentController with a DB instance
func NewCommentControllerWithDB(db *badger.DB) *CommentController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	commentService := services.NewCommentService(commentRepo, postRepo)

	return &CommentController{
		commentService: commentService,
		templates:      loadCommentTemplates(),
	}
}

// loadCommentTemplates loads and parses all comment-related templates
func loadCommentTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["new"] = template.Must(template.ParseFiles("app/views/layout.html", "app/views/comments/new.html"))
	templates["list"] = template.Must(template.ParseFiles(
		"app/views/layout.html",
		"app/views/comments/list.html",
		"app/views/shared/comments.html",
	))
	return templates
}

// New displays the form for creating a new comment
func (cc *CommentController) New(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	data := struct {
		PostID int
	}{
		PostID: postID,
	}

	if err := cc.templates["new"].ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Index handles listing all comments for a post
func (cc *CommentController) Index(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := cc.commentService.ListPostComments(postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		PostID   int
		Comments []*models.Comment
	}{
		PostID:   postID,
		Comments: comments,
	}

	// Respond based on content type
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	} else {
		if err := cc.templates["list"].ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Create handles creating a new comment
func (cc *CommentController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		comment.Author = r.FormValue("author")
		comment.Content = r.FormValue("content")
	}
	comment.PostID = postID

	if err := cc.commentService.CreateComment(&comment); err != nil {
		http.Error(w, "Failed to create comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond based on content type
	if contentType == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
	} else {
		http.Redirect(w, r, "/posts/"+strconv.Itoa(postID), http.StatusSeeOther)
	}
}

// Edit handles editing an existing comment
func (cc *CommentController) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	comment.ID = id

	if err := cc.commentService.UpdateComment(&comment); err != nil {
		http.Error(w, "Failed to update comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// Delete handles deleting a comment
func (cc *CommentController) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	if err := cc.commentService.DeleteComment(id); err != nil {
		http.Error(w, "Failed to delete comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
