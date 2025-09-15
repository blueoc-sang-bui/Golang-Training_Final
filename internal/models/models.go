package models

import (
	"time"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type ActivityLog struct {
	ID       int       `json:"id"`
	Action   string    `json:"action"`
	PostID   int       `json:"post_id"`
	LoggedAt time.Time `json:"logged_at"`
}
