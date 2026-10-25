package model

import "time"

type Article struct {
	ID        uint64    `json:"id"`
	ProjectID uint64    `json:"project_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedBy uint64    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type ArticleRevision struct {
	ID             uint64    `json:"art_rev_id"`
	ArticleID      uint64    `json:"id"`
	UserID         uint64    `json:"user_id"`
	RevisionNumber uint64    `json:"rev_num"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	ContentHash    string    `json:"content_hash"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}
