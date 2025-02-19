package models

import (
	"errors"
	"time"
)

// Validate checks if the post meets all validation requirements
func (p *Post) Validate() error {
	if err := validate.Struct(p); err != nil {
		return err
	}

	if p.CreatedAt.IsZero() {
		return errors.New("created_at cannot be zero")
	}

	return nil
}

// BeforeCreate sets up any necessary fields before creation
func (p *Post) BeforeCreate() {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
}

// AddComment adds a comment to the post
func (p *Post) AddComment(comment *Comment) error {
	if comment == nil {
		return errors.New("comment cannot be nil")
	}

	comment.PostID = p.ID
	p.Comments = append(p.Comments, comment)
	return nil
}

// RemoveComment removes a comment from the post
func (p *Post) RemoveComment(commentID int) error {
	for i, comment := range p.Comments {
		if comment.ID == commentID {
			p.Comments = append(p.Comments[:i], p.Comments[i+1:]...)
			return nil
		}
	}
	return errors.New("comment not found")
}
