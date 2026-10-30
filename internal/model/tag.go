package model

import "time"

type Tag struct {
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	IsGlobal    bool      `json:"global"`
	CreatedAt   time.Time `json:"created_at"`
}
