package repositories

import "cheeseburger/app/models"

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id int) (*models.Post, error)
	List(limit, offset int) ([]*models.Post, error)
	Update(post *models.Post) error
	Delete(id int) error
}

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	Create(comment *models.Comment) error
	GetByID(id int) (*models.Comment, error)
	ListByPost(postID int) ([]*models.Comment, error)
	Update(comment *models.Comment) error
	Delete(id int) error
}
