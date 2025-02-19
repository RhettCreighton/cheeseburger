package mock

import (
	"cheeseburger/app/models"
	"cheeseburger/app/repositories"
	"sync"
)

type PostRepository struct {
	posts  map[int]*models.Post
	nextID int
	mutex  sync.RWMutex
}

type CommentRepository struct {
	comments map[int]*models.Comment
	nextID   int
	mutex    sync.RWMutex
}

func NewPostRepository() *PostRepository {
	return &PostRepository{
		posts:  make(map[int]*models.Post),
		nextID: 1,
	}
}

func (m *PostRepository) Clear() {
	m.posts = make(map[int]*models.Post)
	m.nextID = 1
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{
		comments: make(map[int]*models.Comment),
		nextID:   1,
	}
}

// PostRepository implementation
func (m *PostRepository) Create(post *models.Post) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	post.ID = m.nextID
	m.nextID++
	m.posts[post.ID] = post
	return nil
}

func (m *PostRepository) GetByID(id int) (*models.Post, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	post, exists := m.posts[id]
	if !exists {
		return nil, repositories.ErrNotFound
	}
	return post, nil
}

func (m *PostRepository) Update(post *models.Post) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.posts[post.ID]; !exists {
		return repositories.ErrNotFound
	}
	m.posts[post.ID] = post
	return nil
}

func (m *PostRepository) Delete(id int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.posts[id]; !exists {
		return repositories.ErrNotFound
	}
	delete(m.posts, id)
	return nil
}

func (m *PostRepository) List(limit, offset int) ([]*models.Post, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var posts []*models.Post
	count := 0
	for id := 1; id <= m.nextID-1; id++ {
		if post, exists := m.posts[id]; exists {
			if count >= offset && len(posts) < limit {
				posts = append(posts, post)
			}
			count++
		}
	}
	return posts, nil
}

// CommentRepository implementation
func (m *CommentRepository) Create(comment *models.Comment) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	comment.ID = m.nextID
	m.nextID++
	m.comments[comment.ID] = comment
	return nil
}

func (m *CommentRepository) GetByID(id int) (*models.Comment, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	comment, exists := m.comments[id]
	if !exists {
		return nil, repositories.ErrNotFound
	}
	return comment, nil
}

func (m *CommentRepository) Update(comment *models.Comment) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.comments[comment.ID]; !exists {
		return repositories.ErrNotFound
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *CommentRepository) Delete(id int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.comments[id]; !exists {
		return repositories.ErrNotFound
	}
	delete(m.comments, id)
	return nil
}

func (m *CommentRepository) ListByPost(postID int) ([]*models.Comment, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var comments []*models.Comment
	for _, comment := range m.comments {
		if comment.PostID == postID {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}
