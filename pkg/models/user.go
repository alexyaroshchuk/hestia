package models

import (
	"time"
)

// User contains the data for a user.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Register are used to register users.
type Register struct {
	Password string
	Email    string
}

// Credentials are used to authenticate users.
type Credentials struct {
	Password string
	Email    string
}

type UserFilter struct {
	IDs      []string
	Emails   []string
	Password Password
	IsActive bool
}
