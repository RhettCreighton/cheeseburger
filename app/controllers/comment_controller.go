package controllers

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"cheeseburger/app/services"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

// CommentController handles HTTP requests for comments
type CommentController struct {
	commentService *services.CommentService
	templates      map[string]*template.Template
}

// SetService sets the comment service for testing
func (cc *CommentController) SetService(service *services.CommentService) {
	cc.commentService = service
}

// NewCommentController creates a new CommentController
func NewCommentController() *CommentController {
	return &CommentController{
		templates: loadCommentTemplates(""),
	}
}

// NewCommentControllerWithDB creates a new CommentController with a DB instance
func NewCommentControllerWithDB(db *badger.DB) *CommentController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	commentService := services.NewCommentService(commentRepo, postRepo)

	return &CommentController{
		commentService: commentService,
		templates:      loadCommentTemplates(""),
	}
}

// NewCommentControllerWithPath creates a new CommentController with a custom base path
func NewCommentControllerWithPath(basePath string) *CommentController {
	return &CommentController{
		templates: loadCommentTemplates(basePath),
	}
}

// NewCommentControllerWithDBAndPath creates a new CommentController with a DB instance and custom base path
func NewCommentControllerWithDBAndPath(db *badger.DB, basePath string) *CommentController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	commentService := services.NewCommentService(commentRepo, postRepo)

	return &CommentController{
		commentService: commentService,
		templates:      loadCommentTemplates(basePath),
	}
}

// loadCommentTemplates loads and parses all comment-related templates
func loadCommentTemplates(basePath string) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["new"] = template.Must(template.ParseFiles(
		filepath.Join(basePath, "app/views/layout.html"),
		filepath.Join(basePath, "app/views/comments/new.html"),
	))
	templates["list"] = template.Must(template.ParseFiles(
		filepath.Join(basePath, "app/views/layout.html"),
		filepath.Join(basePath, "app/views/comments/list.html"),
		filepath.Join(basePath, "app/views/shared/comments.html"),
	))
	return templates
}

// New displays the form for creating a new comment
func (cc *CommentController) New(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		cc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	data := struct {
		PostID int
	}{
		PostID: postID,
	}

	if err := cc.templates["new"].ExecuteTemplate(w, "layout", data); err != nil {
		cc.sendError(w, r, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Index handles listing all comments for a post
func (cc *CommentController) Index(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		cc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := cc.commentService.ListPostComments(postID)
	if err != nil {
		cc.sendError(w, r, "Failed to fetch comments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an API request
	accept := r.Header.Get("Accept")
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		cc.sendJSON(w, comments)
	} else {
		data := struct {
			PostID   int
			Comments []*models.Comment
		}{
			PostID:   postID,
			Comments: comments,
		}

		if err := cc.templates["list"].ExecuteTemplate(w, "layout", data); err != nil {
			cc.sendError(w, r, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Create handles creating a new comment
func (cc *CommentController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		cc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		cc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	accept := r.Header.Get("Accept")
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			cc.sendError(w, r, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			cc.sendError(w, r, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		comment.Author = r.FormValue("author")
		comment.Content = r.FormValue("content")
	}
	comment.PostID = postID

	if err := cc.commentService.CreateComment(&comment); err != nil {
		cc.sendError(w, r, "Failed to create comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond based on request type
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		cc.sendJSON(w, comment)
	} else {
		http.Redirect(w, r, "/posts/"+strconv.Itoa(postID), http.StatusSeeOther)
	}
}

// Edit handles editing an existing comment
func (cc *CommentController) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		cc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		cc.sendError(w, r, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		cc.sendError(w, r, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	comment.ID = id

	if err := cc.commentService.UpdateComment(&comment); err != nil {
		cc.sendError(w, r, "Failed to update comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cc.sendJSON(w, comment)
}

// Delete handles deleting a comment
func (cc *CommentController) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		cc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		cc.sendError(w, r, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	if err := cc.commentService.DeleteComment(id); err != nil {
		cc.sendError(w, r, "Failed to delete comment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods for consistent response handling

func (cc *CommentController) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (cc *CommentController) sendError(w http.ResponseWriter, r *http.Request, message string, status int) {
	accept := r.Header.Get("Accept")
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	} else {
		http.Error(w, message, status)
	}
}
