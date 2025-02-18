package services

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"fmt"
	"time"
)

// PostService handles business logic for blog posts
type PostService struct {
	postRepo    repositories.PostRepository
	commentRepo repositories.CommentRepository
}

// NewPostService creates a new PostService
func NewPostService(postRepo repositories.PostRepository, commentRepo repositories.CommentRepository) *PostService {
	return &PostService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

// CreatePost creates a new blog post with validation
func (s *PostService) CreatePost(post *models.Post) error {
	// Validate post
	if err := validatePost(post); err != nil {
		return fmt.Errorf("invalid post: %v", err)
	}

	// Set creation time
	post.CreatedAt = time.Now()

	// Create post
	return s.postRepo.Create(post)
}

// GetPost retrieves a post by ID with its comments
func (s *PostService) GetPost(id int) (*models.Post, error) {
	// Get post
	post, err := s.postRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Get comments for post
	comments, err := s.commentRepo.ListByPost(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	// Attach comments to post
	post.Comments = comments

	return post, nil
}

// ListPosts retrieves a paginated list of posts
func (s *PostService) ListPosts(page, perPage int) ([]*models.Post, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage
	posts, err := s.postRepo.List(perPage, offset)
	if err != nil {
		return nil, err
	}

	// Get comments for each post
	for _, post := range posts {
		comments, err := s.commentRepo.ListByPost(post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get comments for post %d: %v", post.ID, err)
		}
		post.Comments = comments
	}

	return posts, nil
}

// UpdatePost updates an existing post with validation
func (s *PostService) UpdatePost(post *models.Post) error {
	// Validate post
	if err := validatePost(post); err != nil {
		return fmt.Errorf("invalid post: %v", err)
	}

	// Verify post exists
	existing, err := s.postRepo.GetByID(post.ID)
	if err != nil {
		return err
	}

	// Preserve creation time
	post.CreatedAt = existing.CreatedAt

	// Update post
	return s.postRepo.Update(post)
}

// DeletePost deletes a post and all its comments
func (s *PostService) DeletePost(id int) error {
	// Get comments for the post
	comments, err := s.commentRepo.ListByPost(id)
	if err != nil {
		return fmt.Errorf("failed to get comments: %v", err)
	}

	// Delete all comments
	for _, comment := range comments {
		if err := s.commentRepo.Delete(comment.ID); err != nil {
			return fmt.Errorf("failed to delete comment %d: %v", comment.ID, err)
		}
	}

	// Delete the post
	return s.postRepo.Delete(id)
}

// validatePost validates a post's fields
func validatePost(post *models.Post) error {
	if post.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(post.Title) > 200 {
		return fmt.Errorf("title is too long (maximum 200 characters)")
	}
	if post.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}
