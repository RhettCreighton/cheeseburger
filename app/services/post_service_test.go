package services

import (
	"sort"
	"testing"

	"cheeseburger/app/models"
	"cheeseburger/app/repositories"

	"github.com/stretchr/testify/assert"
)

type mockPostRepo struct {
	posts  map[int]*models.Post
	nextID int
}

type mockCommentRepo struct {
	comments map[int]*models.Comment
	nextID   int
}

func newMockPostRepo() *mockPostRepo {
	return &mockPostRepo{
		posts:  make(map[int]*models.Post),
		nextID: 1,
	}
}

func newMockCommentRepo() *mockCommentRepo {
	return &mockCommentRepo{
		comments: make(map[int]*models.Comment),
		nextID:   1,
	}
}

// PostRepository implementation
func (m *mockPostRepo) Create(post *models.Post) error {
	post.ID = m.nextID
	m.nextID++
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepo) GetByID(id int) (*models.Post, error) {
	post, exists := m.posts[id]
	if !exists {
		return nil, repositories.ErrNotFound
	}
	return post, nil
}

func (m *mockPostRepo) Update(post *models.Post) error {
	if _, exists := m.posts[post.ID]; !exists {
		return repositories.ErrNotFound
	}
	m.posts[post.ID] = post
	return nil
}

func (m *mockPostRepo) Delete(id int) error {
	if _, exists := m.posts[id]; !exists {
		return repositories.ErrNotFound
	}
	delete(m.posts, id)
	return nil
}

func (m *mockPostRepo) List(limit, offset int) ([]*models.Post, error) {
	var posts []*models.Post
	// First collect all posts
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	// Sort posts by ID to ensure consistent ordering
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID < posts[j].ID
	})
	// Handle pagination
	if offset >= len(posts) {
		return []*models.Post{}, nil
	}
	end := offset + limit
	if end > len(posts) {
		end = len(posts)
	}
	return posts[offset:end], nil
}

// CommentRepository implementation
func (m *mockCommentRepo) Create(comment *models.Comment) error {
	comment.ID = m.nextID
	m.nextID++
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepo) GetByID(id int) (*models.Comment, error) {
	comment, exists := m.comments[id]
	if !exists {
		return nil, repositories.ErrNotFound
	}
	return comment, nil
}

func (m *mockCommentRepo) Update(comment *models.Comment) error {
	if _, exists := m.comments[comment.ID]; !exists {
		return repositories.ErrNotFound
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepo) Delete(id int) error {
	if _, exists := m.comments[id]; !exists {
		return repositories.ErrNotFound
	}
	delete(m.comments, id)
	return nil
}

func (m *mockCommentRepo) ListByPost(postID int) ([]*models.Comment, error) {
	var comments []*models.Comment
	for _, comment := range m.comments {
		if comment.PostID == postID {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}

func TestPostService(t *testing.T) {
	postRepo := newMockPostRepo()
	commentRepo := newMockCommentRepo()
	service := NewPostService(postRepo, commentRepo)

	t.Run("create post", func(t *testing.T) {
		post := &models.Post{
			Title:   "Test Post",
			Content: "This is a test post content",
		}

		err := service.CreatePost(post)
		assert.NoError(t, err)
		assert.Equal(t, 1, post.ID)
		assert.False(t, post.CreatedAt.IsZero())
	})

	t.Run("get post", func(t *testing.T) {
		post, err := service.GetPost(1)
		assert.NoError(t, err)
		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, "This is a test post content", post.Content)
	})

	t.Run("update post", func(t *testing.T) {
		post := &models.Post{
			ID:      1,
			Title:   "Updated Title",
			Content: "Updated content",
		}

		err := service.UpdatePost(post)
		assert.NoError(t, err)

		updated, err := service.GetPost(1)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated content", updated.Content)
	})

	t.Run("delete post", func(t *testing.T) {
		// Create a post with comments
		post := &models.Post{
			Title:   "Post to Delete",
			Content: "This post will be deleted",
		}
		err := service.CreatePost(post)
		assert.NoError(t, err)

		// Add some comments
		comment := &models.Comment{
			PostID:  post.ID,
			Author:  "Test Author",
			Content: "Test Comment",
		}
		err = commentRepo.Create(comment)
		assert.NoError(t, err)

		// Delete the post
		err = service.DeletePost(post.ID)
		assert.NoError(t, err)

		// Verify post is deleted
		_, err = service.GetPost(post.ID)
		assert.Equal(t, repositories.ErrNotFound, err)

		// Verify comments are deleted
		comments, err := commentRepo.ListByPost(post.ID)
		assert.NoError(t, err)
		assert.Empty(t, comments)
	})

	t.Run("list posts", func(t *testing.T) {
		// Reset the repository
		postRepo = newMockPostRepo()
		commentRepo = newMockCommentRepo()
		service = NewPostService(postRepo, commentRepo)

		// Create multiple posts
		for i := 0; i < 5; i++ {
			post := &models.Post{
				Title:   "List Test Post",
				Content: "Content for list test",
			}
			err := service.CreatePost(post)
			assert.NoError(t, err)
		}

		// Test pagination
		posts, err := service.ListPosts(1, 3) // page 1, 3 per page
		assert.NoError(t, err)
		assert.Equal(t, 3, len(posts))

		posts, err = service.ListPosts(2, 3) // page 2, 3 per page
		assert.NoError(t, err)
		assert.Equal(t, 2, len(posts))
	})

	t.Run("validation errors", func(t *testing.T) {
		t.Run("empty title", func(t *testing.T) {
			post := &models.Post{
				Title:   "", // empty title
				Content: "This is valid content",
			}
			err := service.CreatePost(post)
			assert.Error(t, err)
		})

		t.Run("empty content", func(t *testing.T) {
			post := &models.Post{
				Title:   "Valid Title",
				Content: "", // empty content
			}
			err := service.CreatePost(post)
			assert.Error(t, err)
		})

		t.Run("title too long", func(t *testing.T) {
			longTitle := make([]byte, 201)
			for i := range longTitle {
				longTitle[i] = 'a'
			}
			post := &models.Post{
				Title:   string(longTitle),
				Content: "Valid content",
			}
			err := service.CreatePost(post)
			assert.Error(t, err)
		})
	})
}
