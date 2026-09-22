package model

import "time"

type Article struct {
	ID            int64
	ProjectID     int64
	CreatorUserID int64
	Title         string
	Content       string
	CreatedAt     time.Time
}
