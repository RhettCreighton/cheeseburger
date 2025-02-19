package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommentValidation(t *testing.T) {
	tests := []struct {
		name    string
		comment *Comment
		wantErr bool
	}{
		{
			name: "valid comment",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				Author:    "John Doe",
				Content:   "This is a valid comment",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "author too short",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				Author:    "a",
				Content:   "This is a valid comment",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty content",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				Author:    "John Doe",
				Content:   "",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero creation time",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				Author:    "John Doe",
				Content:   "Valid content",
				CreatedAt: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommentBeforeCreate(t *testing.T) {
	comment := &Comment{
		ID:      1,
		PostID:  1,
		Author:  "John Doe",
		Content: "Test Comment",
	}

	assert.True(t, comment.CreatedAt.IsZero())
	comment.BeforeCreate()
	assert.False(t, comment.CreatedAt.IsZero())
}

func TestCommentSetPost(t *testing.T) {
	comment := &Comment{
		ID:      1,
		Author:  "John Doe",
		Content: "Test Comment",
	}

	t.Run("set valid post", func(t *testing.T) {
		post := &Post{
			ID:      1,
			Title:   "Test Post",
			Content: "Test Content",
		}

		err := comment.SetPost(post)
		assert.NoError(t, err)
		assert.Equal(t, post.ID, comment.PostID)
		assert.Equal(t, post, comment.Post)
	})

	t.Run("set nil post", func(t *testing.T) {
		err := comment.SetPost(nil)
		assert.Error(t, err)
	})
}
