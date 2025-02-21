package controllers

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"cheeseburger/app/services"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/mux"
)

// PostController handles HTTP requests for blog posts
type PostController struct {
	postService *services.PostService
	templates   map[string]*template.Template
}

// SetService sets the post service for testing
func (pc *PostController) SetService(service *services.PostService) {
	pc.postService = service
}

// NewPostController creates a new PostController
func NewPostController() *PostController {
	// Use a unique temporary directory for tests
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, fmt.Sprintf("test_%d.db", time.Now().UnixNano()))
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		panic(err)
	}
	return NewPostControllerWithDB(db)
}

// NewPostControllerWithDB creates a new PostController with a DB instance
func NewPostControllerWithDB(db *badger.DB) *PostController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	postService := services.NewPostService(postRepo, commentRepo)

	return &PostController{
		postService: postService,
		templates:   loadTemplates(""),
	}
}

// NewPostControllerWithPath creates a new PostController with a custom base path
func NewPostControllerWithPath(basePath string) *PostController {
	// Use a unique temporary directory for tests
	tmpDir := os.TempDir()
	dbPath := filepath.Join(tmpDir, fmt.Sprintf("test_%d.db", time.Now().UnixNano()))
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		panic(err)
	}
	return NewPostControllerWithDBAndPath(db, basePath)
}

// NewPostControllerWithDBAndPath creates a new PostController with a DB instance and custom base path
func NewPostControllerWithDBAndPath(db *badger.DB, basePath string) *PostController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	postService := services.NewPostService(postRepo, commentRepo)

	return &PostController{
		postService: postService,
		templates:   loadTemplates(basePath),
	}
}

// loadTemplates loads and parses all templates
func loadTemplates(basePath string) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles(
		filepath.Join(basePath, "app/views/layout.html"),
		filepath.Join(basePath, "app/views/posts/index.html"),
	))
	templates["show"] = template.Must(template.ParseFiles(
		filepath.Join(basePath, "app/views/layout.html"),
		filepath.Join(basePath, "app/views/posts/show.html"),
		filepath.Join(basePath, "app/views/shared/comments.html"),
	))
	templates["new"] = template.Must(template.ParseFiles(
		filepath.Join(basePath, "app/views/layout.html"),
		filepath.Join(basePath, "app/views/posts/new.html"),
	))
	return templates
}

// New displays the form for creating a new post
func (pc *PostController) New(w http.ResponseWriter, r *http.Request) {
	if err := pc.templates["new"].ExecuteTemplate(w, "layout", nil); err != nil {
		pc.sendError(w, r, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Index handles listing all posts
func (pc *PostController) Index(w http.ResponseWriter, r *http.Request) {
	// Parse page parameter
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse per_page parameter
	perPage := 10
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	posts, err := pc.postService.ListPosts(page, perPage)
	if err != nil {
		pc.sendError(w, r, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an API request
	accept := r.Header.Get("Accept")
	if accept == "application/json" || strings.HasPrefix(r.URL.Path, "/api") {
		pc.sendJSON(w, map[string]interface{}{
			"posts": posts,
			"page":  page,
		})
	} else {
		data := struct {
			Posts []*models.Post
			Page  int
		}{
			Posts: posts,
			Page:  page,
		}

		if err := pc.templates["index"].ExecuteTemplate(w, "layout", data); err != nil {
			pc.sendError(w, r, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Show handles displaying a single post
func (pc *PostController) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		pc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := pc.postService.GetPost(id)
	if err != nil {
		pc.sendError(w, r, "Post not found", http.StatusNotFound)
		return
	}

	// Check if this is an API request
	accept := r.Header.Get("Accept")
	if accept == "application/json" || strings.HasPrefix(r.URL.Path, "/api") {
		pc.sendJSON(w, post)
	} else {
		data := struct {
			*models.Post
			Comments []*models.Comment
		}{
			Post:     post,
			Comments: post.Comments,
		}

		if err := pc.templates["show"].ExecuteTemplate(w, "layout", data); err != nil {
			pc.sendError(w, r, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Create handles creating a new post
func (pc *PostController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		pc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form or JSON depending on request type
	var post models.Post
	accept := r.Header.Get("Accept")
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			pc.sendError(w, r, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			pc.sendError(w, r, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
	}

	if err := pc.postService.CreatePost(&post); err != nil {
		pc.sendError(w, r, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond based on request type
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		pc.sendJSON(w, post)
	} else {
		http.Redirect(w, r, "/posts/"+strconv.Itoa(post.ID), http.StatusSeeOther)
	}
}

// Edit handles editing an existing post
func (pc *PostController) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		pc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		pc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		pc.sendError(w, r, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	post.ID = id

	if err := pc.postService.UpdatePost(&post); err != nil {
		pc.sendError(w, r, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pc.sendJSON(w, post)
}

// Delete handles deleting a post
func (pc *PostController) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		pc.sendError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		pc.sendError(w, r, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if err := pc.postService.DeletePost(id); err != nil {
		pc.sendError(w, r, "Failed to delete post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods for consistent response handling

func (pc *PostController) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (pc *PostController) sendError(w http.ResponseWriter, r *http.Request, message string, status int) {
	accept := r.Header.Get("Accept")
	if accept == "application/json" || r.URL.Path[:4] == "/api" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	} else {
		http.Error(w, message, status)
	}
}
