package models

import (
	"errors"
	"time"
)

// Validate checks if the comment meets all validation requirements
func (c *Comment) Validate() error {
	if err := validate.Struct(c); err != nil {
		return err
	}

	if c.CreatedAt.IsZero() {
		return errors.New("created_at cannot be zero")
	}

	return nil
}

// BeforeCreate sets up any necessary fields before creation
func (c *Comment) BeforeCreate() {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
}

// SetPost sets the parent post and updates the PostID
func (c *Comment) SetPost(post *Post) error {
	if post == nil {
		return errors.New("post cannot be nil")
	}

	c.Post = post
	c.PostID = post.ID
	return nil
}
