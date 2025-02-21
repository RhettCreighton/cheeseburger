package repositories

import (
	"testing"

	"cheeseburger/app/models"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetNextID(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(tmpDir))
	assert.NoError(t, err)
	defer db.Close()

	t.Run("first ID", func(t *testing.T) {
		err := db.Update(func(txn *badger.Txn) error {
			id, err := getNextID(txn, PostSeqKey)
			assert.NoError(t, err)
			assert.Equal(t, 1, id)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("sequential IDs", func(t *testing.T) {
		err := db.Update(func(txn *badger.Txn) error {
			// Get multiple IDs and verify they are sequential
			for i := 2; i <= 5; i++ {
				id, err := getNextID(txn, PostSeqKey)
				assert.NoError(t, err)
				assert.Equal(t, i, id)
			}
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("different sequence keys", func(t *testing.T) {
		err := db.Update(func(txn *badger.Txn) error {
			// Test that different sequence keys maintain separate counters
			_, err := getNextID(txn, PostSeqKey)
			assert.NoError(t, err)

			commentID, err := getNextID(txn, CommentSeqKey)
			assert.NoError(t, err)
			assert.Equal(t, 1, commentID, "Comment sequence should start from 1")

			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("persistence", func(t *testing.T) {
		// First transaction
		err := db.Update(func(txn *badger.Txn) error {
			id, err := getNextID(txn, "test:seq")
			assert.NoError(t, err)
			assert.Equal(t, 1, id)
			return nil
		})
		assert.NoError(t, err)

		// Second transaction should continue from last ID
		err = db.Update(func(txn *badger.Txn) error {
			id, err := getNextID(txn, "test:seq")
			assert.NoError(t, err)
			assert.Equal(t, 2, id)
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestMarshalEntity(t *testing.T) {
	t.Run("marshal post", func(t *testing.T) {
		post := &models.Post{
			ID:      1,
			Title:   "Test Post",
			Content: "Test Content",
		}

		data, err := marshalEntity(post)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify data can be unmarshaled back
		var unmarshaled models.Post
		err = unmarshalEntity(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, post.ID, unmarshaled.ID)
		assert.Equal(t, post.Title, unmarshaled.Title)
		assert.Equal(t, post.Content, unmarshaled.Content)
	})

	t.Run("marshal comment", func(t *testing.T) {
		comment := &models.Comment{
			ID:      1,
			PostID:  2,
			Author:  "Test Author",
			Content: "Test Content",
		}

		data, err := marshalEntity(comment)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify data can be unmarshaled back
		var unmarshaled models.Comment
		err = unmarshalEntity(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, comment.ID, unmarshaled.ID)
		assert.Equal(t, comment.PostID, unmarshaled.PostID)
		assert.Equal(t, comment.Author, unmarshaled.Author)
		assert.Equal(t, comment.Content, unmarshaled.Content)
	})

	t.Run("marshal invalid entity", func(t *testing.T) {
		// Create a struct with an invalid field for JSON marshaling
		invalidEntity := struct {
			Ch chan int
		}{
			Ch: make(chan int),
		}

		_, err := marshalEntity(invalidEntity)
		assert.Error(t, err)
	})
}

func TestUnmarshalEntity(t *testing.T) {
	t.Run("unmarshal post", func(t *testing.T) {
		data := []byte(`{"id":1,"title":"Test Post","content":"Test Content"}`)
		var post models.Post
		err := unmarshalEntity(data, &post)
		assert.NoError(t, err)
		assert.Equal(t, 1, post.ID)
		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, "Test Content", post.Content)
	})

	t.Run("unmarshal comment", func(t *testing.T) {
		data := []byte(`{"id":1,"postId":2,"author":"Test Author","content":"Test Content"}`)
		var comment models.Comment
		err := unmarshalEntity(data, &comment)
		assert.NoError(t, err)
		assert.Equal(t, 1, comment.ID)
		assert.Equal(t, 2, comment.PostID)
		assert.Equal(t, "Test Author", comment.Author)
		assert.Equal(t, "Test Content", comment.Content)
	})

	t.Run("unmarshal invalid JSON", func(t *testing.T) {
		data := []byte(`{"id":1,invalid json}`)
		var post models.Post
		err := unmarshalEntity(data, &post)
		assert.Error(t, err)
	})

	t.Run("unmarshal into nil", func(t *testing.T) {
		data := []byte(`{"id":1}`)
		err := unmarshalEntity(data, nil)
		assert.Error(t, err)
	})
}
