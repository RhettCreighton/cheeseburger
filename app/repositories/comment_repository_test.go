package repositories

import (
	"testing"
	"time"

	"cheeseburger/app/models"

	"github.com/stretchr/testify/assert"
)

func TestCommentRepository(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	repo, err := NewRepository(tmpDir + "/test.db")
	assert.NoError(t, err)
	t.Cleanup(func() {
		repo.Close()
	})

	// Create a post first
	post := &models.Post{
		Title:     "Test Post",
		Content:   "Test Content",
		CreatedAt: time.Now(),
	}
	err = repo.CreatePost(post)
	assert.NoError(t, err)

	t.Run("create and get comment", func(t *testing.T) {
		comment := &models.Comment{
			PostID:    post.ID,
			Author:    "Test Author",
			Content:   "Test Comment Content",
			CreatedAt: time.Now(),
		}

		// Create comment
		err := repo.CreateComment(comment)
		assert.NoError(t, err)
		assert.Greater(t, comment.ID, 0)

		// Get comment
		retrieved, err := repo.GetComment(comment.ID)
		assert.NoError(t, err)
		assert.Equal(t, comment.Author, retrieved.Author)
		assert.Equal(t, comment.Content, retrieved.Content)
		assert.Equal(t, comment.PostID, retrieved.PostID)
	})

	t.Run("update comment", func(t *testing.T) {
		comment := &models.Comment{
			PostID:    post.ID,
			Author:    "Original Author",
			Content:   "Original content",
			CreatedAt: time.Now(),
		}

		err := repo.CreateComment(comment)
		assert.NoError(t, err)

		comment.Author = "Updated Author"
		comment.Content = "Updated content"

		err = repo.UpdateComment(comment)
		assert.NoError(t, err)

		updated, err := repo.GetComment(comment.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Author", updated.Author)
		assert.Equal(t, "Updated content", updated.Content)
	})

	t.Run("delete comment", func(t *testing.T) {
		comment := &models.Comment{
			PostID:    post.ID,
			Author:    "Author to Delete",
			Content:   "This comment will be deleted",
			CreatedAt: time.Now(),
		}

		err := repo.CreateComment(comment)
		assert.NoError(t, err)

		err = repo.DeleteComment(comment.ID)
		assert.NoError(t, err)

		_, err = repo.GetComment(comment.ID)
		assert.Error(t, err)
	})

	t.Run("list comments by post", func(t *testing.T) {
		// Create multiple comments
		for i := 0; i < 3; i++ {
			comment := &models.Comment{
				PostID:    post.ID,
				Author:    "List Test Author",
				Content:   "Content for list test",
				CreatedAt: time.Now(),
			}
			err := repo.CreateComment(comment)
			assert.NoError(t, err)
		}

		comments, err := repo.ListCommentsByPost(post.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(comments), 3)
	})

	t.Run("cascade delete with post", func(t *testing.T) {
		// Create a new post
		cascadePost := &models.Post{
			Title:     "Cascade Test Post",
			Content:   "This post will be deleted with its comments",
			CreatedAt: time.Now(),
		}
		err := repo.CreatePost(cascadePost)
		assert.NoError(t, err)

		// Add some comments
		for i := 0; i < 3; i++ {
			comment := &models.Comment{
				PostID:    cascadePost.ID,
				Author:    "Cascade Test Author",
				Content:   "This comment should be deleted with the post",
				CreatedAt: time.Now(),
			}
			err := repo.CreateComment(comment)
			assert.NoError(t, err)
		}

		// Delete the post
		err = repo.DeletePost(cascadePost.ID)
		assert.NoError(t, err)

		// Verify comments are deleted
		comments, err := repo.ListCommentsByPost(cascadePost.ID)
		assert.NoError(t, err)
		assert.Empty(t, comments)
	})
}
