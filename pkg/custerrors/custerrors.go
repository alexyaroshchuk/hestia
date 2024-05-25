package custerrors

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

const (
	UniqueViolationErr = pq.ErrorCode("23505")
)

var (
	ErrNotFound           = errors.New("not found")
	ErrConstraintViolated = errors.New("already exists")
)

// MapDBErr maps database errors to appropriate custom errors errors.
// If err is nil, MapDBErr returns nil.
func MapDBErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == UniqueViolationErr {
			return ErrConstraintViolated
		}
	}

	return err
}
