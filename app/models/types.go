package models

import "time"

// Post represents a blog post with comments.
type Post struct {
	ID        int        `validate:"required,gte=0"`
	Title     string     `validate:"required,min=3,max=100"`
	Content   string     `validate:"required,min=10"`
	CreatedAt time.Time  `validate:"required"`
	Comments  []*Comment `validate:"-"`
}

// Comment represents a comment on a blog post.
type Comment struct {
	ID        int       `validate:"required,gte=0"`
	PostID    int       `validate:"required,gte=0"`
	Author    string    `validate:"required,min=2,max=50"`
	Content   string    `validate:"required,min=1,max=500"`
	CreatedAt time.Time `validate:"required"`
	Post      *Post     `validate:"-"`
}
