package model

import "time"

type Asset struct {
	ID            int64
	ProjectID     int64
	ArticleID     int64
	CreatorUserID int64
	Hash          string
	AssetType     string
	Comment       string
	CreatedAt     time.Time
}
