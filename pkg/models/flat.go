package models

import (
	"time"
)

type FlatFilter struct {
}

type Url struct {
	Url string
}

// Flat contains the data for a flat.
type Flat struct {
	ID            string
	Title         string
	Price         string
	Address       string
	Surface       string
	Rooms         string
	Floor         string
	AvailableFrom string
	Rent          string
	Deposit       string
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
