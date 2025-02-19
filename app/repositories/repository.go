package repositories

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"cheeseburger/app/models"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrNotFound = errors.New("record not found")
)

type Repository struct {
	db                 *badger.DB
	mutex              sync.RWMutex
	dbPath             string
	isTestDB           bool
	testPostCounter    int
	testCommentCounter int
}

func NewRepository(path string) (*Repository, error) {
	isTest := false
	if path == "" || path == "test_db" {
		// If no path is provided or if "test_db" is explicitly used,
		// create a unique temporary directory for testing to ensure isolation.
		tempPath, err := os.MkdirTemp("", "cheeseburger_test_db_")
		if err != nil {
			return nil, fmt.Errorf("Error creating temp dir: %v", err)
		}
		path = tempPath
		isTest = true
	}
	opts := badger.DefaultOptions(path).
		WithLogger(nil).
		WithSyncWrites(false).
		WithNumVersionsToKeep(1).
		WithNumGoroutines(1)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	// For testing, ensure the database is clean by dropping all keys.
	if isTest {
		if err := db.DropAll(); err != nil {
			return nil, fmt.Errorf("failed to drop all keys: %v", err)
		}
	}
	return &Repository{
		db:       db,
		mutex:    sync.RWMutex{},
		dbPath:   path,
		isTestDB: isTest,
	}, nil
}

func (r *Repository) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	err := r.db.Close()
	if err != nil {
		return err
	}

	// Clean up test database
	if r.isTestDB {
		err = os.RemoveAll(r.dbPath)
		if err != nil {
			return fmt.Errorf("failed to cleanup test database: %v", err)
		}
	}
	return nil
}

// Post methods

func (r *Repository) CreatePost(post *models.Post) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var maxID int
	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("post:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var id int
			_, err := fmt.Sscanf(string(k), "post:%d", &id)
			if err == nil && id > maxID {
				maxID = id
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	post.ID = maxID + 1
	post.BeforeCreate()

	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("post:%d", post.ID))
		return txn.Set(key, data)
	})
}

func (r *Repository) GetPost(id int) (*models.Post, error) {
	var post models.Post

	// First get the post data
	r.mutex.RLock()
	err := r.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("post:%d", id))
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &post)
		})
	})
	r.mutex.RUnlock()

	if err != nil {
		return nil, err
	}

	// Then load comments in a separate lock
	comments, err := r.ListCommentsByPost(id)
	if err != nil {
		return nil, err
	}
	post.Comments = comments

	return &post, nil
}

func (r *Repository) UpdatePost(post *models.Post) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("post:%d", post.ID))
		return txn.Set(key, data)
	})
}

func (r *Repository) DeletePost(id int) error {
	// Delete all comments first
	comments, err := r.ListCommentsByPost(id)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		err = r.DeleteComment(comment.ID)
		if err != nil {
			return err
		}
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("post:%d", id))
		return txn.Delete(key)
	})
}

func (r *Repository) ListPosts() ([]*models.Post, error) {
	var posts []*models.Post
	var postIDs []int

	// First get all posts without comments
	r.mutex.RLock()
	err := r.db.View(func(txn *badger.Txn) error {
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
			postIDs = append(postIDs, post.ID)
		}
		return nil
	})
	r.mutex.RUnlock()

	if err != nil {
		return nil, err
	}

	// Then load comments for each post
	for i, post := range posts {
		comments, err := r.ListCommentsByPost(postIDs[i])
		if err != nil {
			return nil, err
		}
		post.Comments = comments
	}

	return posts, nil
}

// Comment methods

func (r *Repository) CreateComment(comment *models.Comment) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var maxID int
	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("comment:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var id int
			_, err := fmt.Sscanf(string(item.Key()), "comment:%d", &id)
			if err == nil && id > maxID {
				maxID = id
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	comment.ID = maxID + 1
	comment.BeforeCreate()

	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("comment:%d", comment.ID))
		return txn.Set(key, data)
	})
}

func (r *Repository) GetComment(id int) (*models.Comment, error) {
	var comment models.Comment

	// First get the comment data
	r.mutex.RLock()
	err := r.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("comment:%d", id))
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &comment)
		})
	})
	r.mutex.RUnlock()

	if err != nil {
		return nil, err
	}

	// Then load associated post in a separate lock
	post, err := r.GetPost(comment.PostID)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	comment.Post = post

	return &comment, nil
}

func (r *Repository) UpdateComment(comment *models.Comment) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("comment:%d", comment.ID))
		return txn.Set(key, data)
	})
}

func (r *Repository) DeleteComment(id int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("comment:%d", id))
		return txn.Delete(key)
	})
}

func (r *Repository) ListCommentsByPost(postID int) ([]*models.Comment, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var comments []*models.Comment

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("comment:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var comment models.Comment
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &comment)
			})
			if err != nil {
				return err
			}

			if comment.PostID == postID {
				comments = append(comments, &comment)
			}
		}
		return nil
	})

	return comments, err
}

func (r *Repository) Clear() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	err := r.db.DropAll()
	if err != nil {
		return err
	}
	if r.isTestDB {
		r.testPostCounter = 0
		r.testCommentCounter = 0
	}
	return nil
}
