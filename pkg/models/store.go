package models

import (
	"context"
)

// Store provides access to the user store.
type Store interface {
	BeginTx(ctx context.Context) (Tx, error)

	FindUsers(ctx context.Context, filter UserFilter) ([]User, error)

	FindFlats(ctx context.Context, filter FlatFilter) ([]Flat, error)
	GetFlatByID(ctx context.Context, id string) (Flat, error)
}

// Tx is a transaction.
type Tx interface {
	Commit() error
	Rollback() error

	CreateUser(u User) error
	FindUsers(filter UserFilter) ([]User, error)
	DeleteUser(id string) error
	UpdateUser(u User) error

	CreateEmailToken(t EmailToken) error
	UpdateEmailToken(t EmailToken) error

	CreateFlat(u Flat) error
	DeleteFlat(id string) error
	UpdateFlat(u Flat) error
}
