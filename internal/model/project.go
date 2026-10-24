package model

import "time"

type Project struct {
	ID          int64     `json:"id"`
	OwnerUserID int64     `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
