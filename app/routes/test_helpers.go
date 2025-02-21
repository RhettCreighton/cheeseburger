package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"cheeseburger/app/models"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

const (
	PostKeyPrefix = "post:"
	PostSeqKey    = "seq:post"
)

func setupTestTemplates(t *testing.T) string {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "app", "views")

	// Create directories
	dirs := []string{
		filepath.Join(viewsDir, "posts"),
		filepath.Join(viewsDir, "comments"),
		filepath.Join(viewsDir, "shared"),
		filepath.Join(tmpDir, "static"),
	}
	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}

	// Create template files
	templates := map[string]string{
		filepath.Join(viewsDir, "layout.html"):          `{{define "layout"}}<!DOCTYPE html><html><body>{{template "content" .}}</body></html>{{end}}`,
		filepath.Join(viewsDir, "posts/index.html"):     `{{define "content"}}<div class="posts">{{range .Posts}}<h2>{{.Title}}</h2>{{end}}</div>{{end}}`,
		filepath.Join(viewsDir, "posts/show.html"):      `{{define "content"}}<h1>{{.Title}}</h1><p>{{.Content}}</p>{{end}}`,
		filepath.Join(viewsDir, "posts/new.html"):       `{{define "content"}}<form method="POST"><input name="title"><textarea name="content"></textarea></form>{{end}}`,
		filepath.Join(viewsDir, "comments/list.html"):   `{{define "content"}}<div class="comments">{{range .Comments}}<p>{{.Content}}</p>{{end}}</div>{{end}}`,
		filepath.Join(viewsDir, "comments/new.html"):    `{{define "content"}}<form method="POST"><textarea name="content"></textarea></form>{{end}}`,
		filepath.Join(viewsDir, "shared/comments.html"): `{{define "comments"}}{{template "content" .}}{{end}}`,
	}
	for path, content := range templates {
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	// Create static test file
	cssContent := "body { background: #f0f0f0; }"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "static/style.css"), []byte(cssContent), 0644))

	return tmpDir
}

func setupTestDB(t *testing.T) *badger.DB {
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func setupTestData(t *testing.T, db *badger.DB) {
	// Create a test post
	post := &models.Post{
		Title:     "Test Post",
		Content:   "This is a test post with sufficient content",
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.Update(func(txn *badger.Txn) error {
		id, err := getNextID(txn, PostSeqKey)
		if err != nil {
			return err
		}
		post.ID = id
		data, err := marshalEntity(post)
		if err != nil {
			return err
		}
		return txn.Set([]byte(PostKeyPrefix+strconv.Itoa(id)), data)
	}))
}

func getNextID(txn *badger.Txn, key string) (int, error) {
	item, err := txn.Get([]byte(key))
	var id int
	if err == badger.ErrKeyNotFound {
		id = 1
	} else if err != nil {
		return 0, err
	} else {
		err = item.Value(func(val []byte) error {
			id, _ = strconv.Atoi(string(val))
			return nil
		})
		if err != nil {
			return 0, err
		}
		id++
	}
	err = txn.Set([]byte(key), []byte(strconv.Itoa(id)))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func marshalEntity(entity interface{}) ([]byte, error) {
	return json.Marshal(entity)
}
