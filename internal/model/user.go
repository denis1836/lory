package model

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	Description  string
	CreatedAt    time.Time
	LastOnline   time.Time
}
