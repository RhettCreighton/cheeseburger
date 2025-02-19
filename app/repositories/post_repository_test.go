package repositories

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cheeseburger/app/models"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestPostRepository(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	repo, err := NewRepository(tmpDir + "/test.db")
	assert.NoError(t, err)
	t.Cleanup(func() {
		repo.Close()
	})

	t.Run("create and get post", func(t *testing.T) {
		post := &models.Post{
			Title:     "Test Post",
			Content:   "This is a test post content",
			CreatedAt: time.Now(),
		}

		// Create post
		err := repo.CreatePost(post)
		assert.NoError(t, err)
		assert.Greater(t, post.ID, 0)

		// Get post directly from DB to avoid loading comments
		var retrieved models.Post
		err = repo.db.View(func(txn *badger.Txn) error {
			key := []byte(fmt.Sprintf("post:%d", post.ID))
			item, err := txn.Get(key)
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				return json.Unmarshal(val, &retrieved)
			})
		})
		assert.NoError(t, err)
		assert.Equal(t, post.Title, retrieved.Title)
		assert.Equal(t, post.Content, retrieved.Content)
	})

	t.Run("update post", func(t *testing.T) {
		post := &models.Post{
			Title:     "Original Title",
			Content:   "Original content",
			CreatedAt: time.Now(),
		}

		err := repo.CreatePost(post)
		assert.NoError(t, err)

		post.Title = "Updated Title"
		post.Content = "Updated content"

		err = repo.UpdatePost(post)
		assert.NoError(t, err)

		// Get updated post directly from DB
		var updated models.Post
		err = repo.db.View(func(txn *badger.Txn) error {
			key := []byte(fmt.Sprintf("post:%d", post.ID))
			item, err := txn.Get(key)
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				return json.Unmarshal(val, &updated)
			})
		})
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated content", updated.Content)
	})

	t.Run("delete post", func(t *testing.T) {
		post := &models.Post{
			Title:     "Post to Delete",
			Content:   "This post will be deleted",
			CreatedAt: time.Now(),
		}

		err := repo.CreatePost(post)
		assert.NoError(t, err)

		err = repo.DeletePost(post.ID)
		assert.NoError(t, err)

		_, err = repo.GetPost(post.ID)
		assert.Error(t, err)
	})

	t.Run("list posts", func(t *testing.T) {
		// Create multiple posts
		for i := 0; i < 3; i++ {
			post := &models.Post{
				Title:     "List Test Post",
				Content:   "Content for list test",
				CreatedAt: time.Now(),
			}
			err := repo.CreatePost(post)
			assert.NoError(t, err)
		}

		// List posts directly from DB
		var posts []*models.Post
		err = repo.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()

			prefix := []byte("post:")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				var post models.Post
				err := item.Value(func(val []byte) error {
					return json.Unmarshal(val, &post)
				})
				if err != nil {
					return err
				}
				posts = append(posts, &post)
			}
			return nil
		})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 3)
	})

	t.Run("get post with comments", func(t *testing.T) {
		post := &models.Post{
			Title:     "Post with Comments",
			Content:   "This post has comments",
			CreatedAt: time.Now(),
		}

		err := repo.CreatePost(post)
		assert.NoError(t, err)

		comment := &models.Comment{
			PostID:    post.ID,
			Author:    "Test Author",
			Content:   "Test Comment",
			CreatedAt: time.Now(),
		}

		err = repo.CreateComment(comment)
		assert.NoError(t, err)

		retrieved, err := repo.GetPost(post.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, retrieved.Comments)
		assert.Equal(t, comment.Content, retrieved.Comments[0].Content)
	})
}
