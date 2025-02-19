package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPostValidation(t *testing.T) {
	tests := []struct {
		name    string
		post    *Post
		wantErr bool
	}{
		{
			name: "valid post",
			post: &Post{
				ID:        1,
				Title:     "Valid Title",
				Content:   "This is valid content that meets the minimum length requirement",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "title too short",
			post: &Post{
				ID:        1,
				Title:     "ab",
				Content:   "This is valid content",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "content too short",
			post: &Post{
				ID:        1,
				Title:     "Valid Title",
				Content:   "Too short",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero creation time",
			post: &Post{
				ID:        1,
				Title:     "Valid Title",
				Content:   "This is valid content",
				CreatedAt: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostBeforeCreate(t *testing.T) {
	post := &Post{
		ID:      1,
		Title:   "Test Post",
		Content: "Test Content",
	}

	assert.True(t, post.CreatedAt.IsZero())
	post.BeforeCreate()
	assert.False(t, post.CreatedAt.IsZero())
}

func TestPostCommentManagement(t *testing.T) {
	post := &Post{
		ID:      1,
		Title:   "Test Post",
		Content: "Test Content",
	}

	t.Run("add comment", func(t *testing.T) {
		comment := &Comment{
			ID:      1,
			Author:  "Test Author",
			Content: "Test Comment",
		}

		err := post.AddComment(comment)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(post.Comments))
		assert.Equal(t, post.ID, comment.PostID)
	})

	t.Run("add nil comment", func(t *testing.T) {
		err := post.AddComment(nil)
		assert.Error(t, err)
	})

	t.Run("remove existing comment", func(t *testing.T) {
		err := post.RemoveComment(1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(post.Comments))
	})

	t.Run("remove non-existent comment", func(t *testing.T) {
		err := post.RemoveComment(999)
		assert.Error(t, err)
	})
}
