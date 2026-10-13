package model

import "time"

type Article struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ArticleContributor struct {
	ArticleID int64 `json:"article_id"`
	UserID    int64 `json:"user_id"`
}

type ArticleChange struct {
	ArticleChange int64     `json:"article_id"`
	UserID        int64     `json:"user_id"`
	Message       string    `json:"message"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}
