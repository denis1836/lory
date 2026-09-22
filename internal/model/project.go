package model

import "time"

type Project struct {
	ID          int64
	OwnerUserID int64
	Name        string
	Description string
	CreatedAt   time.Time
}
