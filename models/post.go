package models

import "time"

// Post represents a blog post with comments.
type Post struct {
	ID        int
	Title     string
	Content   string
	CreatedAt time.Time
	Comments  []*Comment
}
