package services

import (
	"testing"

	"cheeseburger/app/models"
	"cheeseburger/app/repositories"

	"github.com/stretchr/testify/assert"
)

func TestCommentService(t *testing.T) {
	postRepo := newMockPostRepo()
	commentRepo := newMockCommentRepo()
	service := NewCommentService(commentRepo, postRepo)

	// Create a test post first
	post := &models.Post{
		Title:   "Test Post",
		Content: "Test Content",
	}
	err := postRepo.Create(post)
	assert.NoError(t, err)

	t.Run("create comment", func(t *testing.T) {
		comment := &models.Comment{
			PostID:  post.ID,
			Author:  "Test Author",
			Content: "Test Comment Content",
		}

		err := service.CreateComment(comment)
		assert.NoError(t, err)
		assert.Equal(t, 1, comment.ID)
		assert.False(t, comment.CreatedAt.IsZero())
	})

	t.Run("get comment", func(t *testing.T) {
		comment, err := service.GetComment(1)
		assert.NoError(t, err)
		assert.Equal(t, "Test Author", comment.Author)
		assert.Equal(t, "Test Comment Content", comment.Content)
		assert.Equal(t, post.ID, comment.PostID)
	})

	t.Run("update comment", func(t *testing.T) {
		comment := &models.Comment{
			ID:      1,
			PostID:  post.ID,
			Author:  "Updated Author",
			Content: "Updated content",
		}

		err := service.UpdateComment(comment)
		assert.NoError(t, err)

		updated, err := service.GetComment(1)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Author", updated.Author)
		assert.Equal(t, "Updated content", updated.Content)
	})

	t.Run("delete comment", func(t *testing.T) {
		err := service.DeleteComment(1)
		assert.NoError(t, err)

		_, err = service.GetComment(1)
		assert.Equal(t, repositories.ErrNotFound, err)
	})

	t.Run("list post comments", func(t *testing.T) {
		// Create multiple comments
		for i := 0; i < 3; i++ {
			comment := &models.Comment{
				PostID:  post.ID,
				Author:  "List Test Author",
				Content: "Content for list test",
			}
			err := service.CreateComment(comment)
			assert.NoError(t, err)
		}

		comments, err := service.ListPostComments(post.ID)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(comments))
		for _, comment := range comments {
			assert.Equal(t, post.ID, comment.PostID)
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Run("empty author", func(t *testing.T) {
			comment := &models.Comment{
				PostID:  post.ID,
				Author:  "", // empty author
				Content: "Valid content",
			}
			err := service.CreateComment(comment)
			assert.Error(t, err)
		})

		t.Run("empty content", func(t *testing.T) {
			comment := &models.Comment{
				PostID:  post.ID,
				Author:  "Valid Author",
				Content: "", // empty content
			}
			err := service.CreateComment(comment)
			assert.Error(t, err)
		})

		t.Run("invalid post ID", func(t *testing.T) {
			comment := &models.Comment{
				PostID:  999, // non-existent post
				Author:  "Valid Author",
				Content: "Valid content",
			}
			err := service.CreateComment(comment)
			assert.Error(t, err)
		})

		t.Run("author too long", func(t *testing.T) {
			longAuthor := make([]byte, 101)
			for i := range longAuthor {
				longAuthor[i] = 'a'
			}
			comment := &models.Comment{
				PostID:  post.ID,
				Author:  string(longAuthor),
				Content: "Valid content",
			}
			err := service.CreateComment(comment)
			assert.Error(t, err)
		})

		t.Run("content too long", func(t *testing.T) {
			longContent := make([]byte, 1001)
			for i := range longContent {
				longContent[i] = 'a'
			}
			comment := &models.Comment{
				PostID:  post.ID,
				Author:  "Valid Author",
				Content: string(longContent),
			}
			err := service.CreateComment(comment)
			assert.Error(t, err)
		})
	})
}
