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

// PostController handles HTTP requests for blog posts
type PostController struct {
	postService *services.PostService
	templates   map[string]*template.Template
}

// NewPostController creates a new PostController
func NewPostController() *PostController {
	return &PostController{
		templates: loadTemplates(),
	}
}

// NewPostControllerWithDB creates a new PostController with a DB instance
func NewPostControllerWithDB(db *badger.DB) *PostController {
	postRepo := repositories.NewBadgerPostRepository(db)
	commentRepo := repositories.NewBadgerCommentRepository(db)
	postService := services.NewPostService(postRepo, commentRepo)

	return &PostController{
		postService: postService,
		templates:   loadTemplates(),
	}
}

// loadTemplates loads and parses all templates
func loadTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("app/views/layout.html", "app/views/posts/index.html"))
	templates["show"] = template.Must(template.ParseFiles(
		"app/views/layout.html",
		"app/views/posts/show.html",
		"app/views/shared/comments.html",
	))
	templates["new"] = template.Must(template.ParseFiles("app/views/layout.html", "app/views/posts/new.html"))
	return templates
}

// New displays the form for creating a new post
func (pc *PostController) New(w http.ResponseWriter, r *http.Request) {
	if err := pc.templates["new"].ExecuteTemplate(w, "layout", nil); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Index handles listing all posts
func (pc *PostController) Index(w http.ResponseWriter, r *http.Request) {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	posts, err := pc.postService.ListPosts(page, 10)
	if err != nil {
		http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts []*models.Post
		Page  int
	}{
		Posts: posts,
		Page:  page,
	}

	if err := pc.templates["index"].ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Show handles displaying a single post
func (pc *PostController) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := pc.postService.GetPost(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	data := struct {
		*models.Post
		Comments []*models.Comment
	}{
		Post:     post,
		Comments: post.Comments,
	}

	if err := pc.templates["show"].ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// Create handles creating a new post
func (pc *PostController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form or JSON depending on content type
	var post models.Post
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
	}

	if err := pc.postService.CreatePost(&post); err != nil {
		http.Error(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond based on content type
	if contentType == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	} else {
		http.Redirect(w, r, "/posts/"+strconv.Itoa(post.ID), http.StatusSeeOther)
	}
}

// Edit handles editing an existing post
func (pc *PostController) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	post.ID = id

	if err := pc.postService.UpdatePost(&post); err != nil {
		http.Error(w, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// Delete handles deleting a post
func (pc *PostController) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if err := pc.postService.DeletePost(id); err != nil {
		http.Error(w, "Failed to delete post: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
