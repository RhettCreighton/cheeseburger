package repositories

import (
	"fmt"

	"cheeseburger/app/models"

	"github.com/dgraph-io/badger/v4"
)

// BadgerCommentRepository implements CommentRepository using BadgerDB
type BadgerCommentRepository struct {
	db *badger.DB
}

// NewBadgerCommentRepository creates a new BadgerCommentRepository
func NewBadgerCommentRepository(db *badger.DB) *BadgerCommentRepository {
	return &BadgerCommentRepository{db: db}
}

// Create creates a new comment
func (r *BadgerCommentRepository) Create(comment *models.Comment) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Get next ID
		id, err := getNextID(txn, CommentSeqKey)
		if err != nil {
			return err
		}
		comment.ID = id

		// Marshal comment
		data, err := marshalEntity(comment)
		if err != nil {
			return err
		}

		// Save comment with post ID in key for efficient listing
		key := []byte(fmt.Sprintf("%s%d:%d", CommentKeyPrefix, comment.PostID, comment.ID))
		return txn.Set(key, data)
	})
}

// GetByID retrieves a comment by ID
func (r *BadgerCommentRepository) GetByID(id int) (*models.Comment, error) {
	var comment models.Comment
	var found bool

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(CommentKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				if err := unmarshalEntity(val, &comment); err != nil {
					return err
				}
				if comment.ID == id {
					found = true
					return nil
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal comment: %v", err)
			}
			if found {
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotFound
	}
	return &comment, nil
}

// ListByPost retrieves all comments for a post
func (r *BadgerCommentRepository) ListByPost(postID int) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("%s%d:", CommentKeyPrefix, postID))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var comment models.Comment
			err := item.Value(func(val []byte) error {
				return unmarshalEntity(val, &comment)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal comment: %v", err)
			}
			comments = append(comments, &comment)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// Update updates an existing comment
func (r *BadgerCommentRepository) Update(comment *models.Comment) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Find the comment's key
		var key []byte
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(CommentKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var existingComment models.Comment
			err := item.Value(func(val []byte) error {
				return unmarshalEntity(val, &existingComment)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal comment: %v", err)
			}
			if existingComment.ID == comment.ID {
				key = item.Key()
				break
			}
		}

		if key == nil {
			return ErrNotFound
		}

		// Marshal and save updated comment
		data, err := marshalEntity(comment)
		if err != nil {
			return err
		}
		return txn.Set(key, data)
	})
}

// Delete deletes a comment by ID
func (r *BadgerCommentRepository) Delete(id int) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Find the comment's key
		var key []byte
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(CommentKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var comment models.Comment
			err := item.Value(func(val []byte) error {
				return unmarshalEntity(val, &comment)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal comment: %v", err)
			}
			if comment.ID == id {
				key = item.Key()
				break
			}
		}

		if key == nil {
			return ErrNotFound
		}

		return txn.Delete(key)
	})
}
