package models

import (
	"time"
)

type FlatFilter struct {
}

// Flat contains the data for a flat.
type Flat struct {
	ID           string
	Email        string
	PasswordHash string
	IsActive     bool
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
