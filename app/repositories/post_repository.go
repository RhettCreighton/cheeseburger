package repositories

import (
	"cheeseburger/app/models"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id int) (*models.Post, error)
	List(limit, offset int) ([]*models.Post, error)
	Update(post *models.Post) error
	Delete(id int) error
}

// BadgerPostRepository implements PostRepository using BadgerDB
type BadgerPostRepository struct {
	db *badger.DB
}

// NewBadgerPostRepository creates a new BadgerPostRepository
func NewBadgerPostRepository(db *badger.DB) *BadgerPostRepository {
	return &BadgerPostRepository{db: db}
}

// Create creates a new post
func (r *BadgerPostRepository) Create(post *models.Post) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Get next ID
		id, err := getNextID(txn, PostSeqKey)
		if err != nil {
			return err
		}
		post.ID = id

		// Marshal post
		data, err := marshalEntity(post)
		if err != nil {
			return err
		}

		// Save post
		key := []byte(fmt.Sprintf("%s%d", PostKeyPrefix, post.ID))
		return txn.Set(key, data)
	})
}

// GetByID retrieves a post by ID
func (r *BadgerPostRepository) GetByID(id int) (*models.Post, error) {
	var post models.Post
	err := r.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("%s%d", PostKeyPrefix, id))
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("post not found: %d", id)
		}
		if err != nil {
			return fmt.Errorf("failed to get post: %v", err)
		}

		return item.Value(func(val []byte) error {
			return unmarshalEntity(val, &post)
		})
	})
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// List retrieves a list of posts with pagination
func (r *BadgerPostRepository) List(limit, offset int) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(PostKeyPrefix)
		skipped := 0
		count := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix) && count < limit; it.Next() {
			if skipped < offset {
				skipped++
				continue
			}

			item := it.Item()
			var post models.Post
			err := item.Value(func(val []byte) error {
				return unmarshalEntity(val, &post)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal post: %v", err)
			}

			posts = append(posts, &post)
			count++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// Update updates an existing post
func (r *BadgerPostRepository) Update(post *models.Post) error {
	return r.db.Update(func(txn *badger.Txn) error {
		// Check if post exists
		key := []byte(fmt.Sprintf("%s%d", PostKeyPrefix, post.ID))
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("post not found: %d", post.ID)
		}
		if err != nil {
			return fmt.Errorf("failed to get post: %v", err)
		}

		// Marshal and save updated post
		data, err := marshalEntity(post)
		if err != nil {
			return err
		}
		return txn.Set(key, data)
	})
}

// Delete deletes a post by ID
func (r *BadgerPostRepository) Delete(id int) error {
	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("%s%d", PostKeyPrefix, id))
		// Check if post exists
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("post not found: %d", id)
		}
		if err != nil {
			return fmt.Errorf("failed to get post: %v", err)
		}

		return txn.Delete(key)
	})
}
