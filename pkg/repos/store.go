package repos

import (
	"context"
	"database/sql"

	"hestia/pkg/models"
)

// Store is responsible for interacting with a database.
type Store struct {
	db *sql.DB
}

func New(pDB *sql.DB) *Store {
	return &Store{
		db: pDB,
	}
}

// BeginTx starts a new transaction.
func (s *Store) BeginTx(ctx context.Context) (models.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx:    tx,
		store: s,
	}, nil
}

func (s *Store) FindUsers(ctx context.Context, filter models.UserFilter) ([]models.User, error) {
	return selectUsers(func(query string, params ...any) (*sql.Rows, error) {
		return s.db.QueryContext(ctx, query, params...)
	}, filter)
}

func (s *Store) FindFlats(ctx context.Context, filter models.FlatFilter) ([]models.Flat, error) {
	return selectFlats(func(query string, params ...any) (*sql.Rows, error) {
		return s.db.QueryContext(ctx, query, params...)
	}, filter)
}

func (s *Store) GetFlatByID(ctx context.Context, id string) (models.Flat, error) {
	return selectFlat(func(query string, params ...any) (*sql.Rows, error) {
		return s.db.QueryContext(ctx, query, params...)
	}, id)
}

type execFunc func(query string, params ...any) (sql.Result, error)
type queryFunc func(query string, params ...any) (*sql.Rows, error)

func anySlice[T any](s []T) []any {
	out := make([]any, 0, len(s))
	for _, v := range s {
		out = append(out, v)
	}
	return out
}
