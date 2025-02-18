package models

import "time"

// Comment represents a comment on a blog post.
type Comment struct {
	ID        int
	PostID    int
	Author    string
	Content   string
	CreatedAt time.Time
}
