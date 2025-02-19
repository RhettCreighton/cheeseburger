package repositories

import (
	"fmt"

	"cheeseburger/app/models"

	"github.com/dgraph-io/badger/v4"
)

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
			return ErrNotFound
		}
		if err != nil {
			return err
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

// List retrieves a paginated list of posts
func (r *BadgerPostRepository) List(limit, offset int) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		// Skip offset items
		count := 0
		prefix := []byte(PostKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if count < offset {
				count++
				continue
			}
			if count >= offset+limit {
				break
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
		key := []byte(fmt.Sprintf("%s%d", PostKeyPrefix, post.ID))

		// Verify post exists
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
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

		// Verify post exists
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		return txn.Delete(key)
	})
}
