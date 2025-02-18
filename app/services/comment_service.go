package services

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"fmt"
	"time"
)

// CommentService handles business logic for comments
type CommentService struct {
	commentRepo repositories.CommentRepository
	postRepo    repositories.PostRepository
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo repositories.CommentRepository, postRepo repositories.PostRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

// CreateComment creates a new comment with validation
func (s *CommentService) CreateComment(comment *models.Comment) error {
	// Validate comment
	if err := validateComment(comment); err != nil {
		return fmt.Errorf("invalid comment: %v", err)
	}

	// Verify post exists
	_, err := s.postRepo.GetByID(comment.PostID)
	if err != nil {
		return fmt.Errorf("post not found: %v", err)
	}

	// Set creation time
	comment.CreatedAt = time.Now()

	// Create comment
	return s.commentRepo.Create(comment)
}

// GetComment retrieves a comment by ID
func (s *CommentService) GetComment(id int) (*models.Comment, error) {
	return s.commentRepo.GetByID(id)
}

// ListPostComments retrieves all comments for a post
func (s *CommentService) ListPostComments(postID int) ([]*models.Comment, error) {
	// Verify post exists
	_, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %v", err)
	}

	return s.commentRepo.ListByPost(postID)
}

// UpdateComment updates an existing comment with validation
func (s *CommentService) UpdateComment(comment *models.Comment) error {
	// Validate comment
	if err := validateComment(comment); err != nil {
		return fmt.Errorf("invalid comment: %v", err)
	}

	// Verify comment exists and belongs to the specified post
	existing, err := s.commentRepo.GetByID(comment.ID)
	if err != nil {
		return err
	}
	if existing.PostID != comment.PostID {
		return fmt.Errorf("comment does not belong to specified post")
	}

	// Preserve creation time and post ID
	comment.CreatedAt = existing.CreatedAt
	comment.PostID = existing.PostID

	// Update comment
	return s.commentRepo.Update(comment)
}

// DeleteComment deletes a comment
func (s *CommentService) DeleteComment(id int) error {
	// Verify comment exists
	_, err := s.commentRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.commentRepo.Delete(id)
}

// validateComment validates a comment's fields
func validateComment(comment *models.Comment) error {
	if comment.PostID <= 0 {
		return fmt.Errorf("invalid post ID")
	}
	if comment.Author == "" {
		return fmt.Errorf("author is required")
	}
	if len(comment.Author) > 100 {
		return fmt.Errorf("author name is too long (maximum 100 characters)")
	}
	if comment.Content == "" {
		return fmt.Errorf("content is required")
	}
	if len(comment.Content) > 1000 {
		return fmt.Errorf("content is too long (maximum 1000 characters)")
	}
	return nil
}
