package repos

import (
	"database/sql"

	"hestia/pkg/models"
)

type Tx struct {
	tx    *sql.Tx
	store *Store
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// CreateUser creates a user in the database.
func (t *Tx) FindUsers(filter models.UserFilter) ([]models.User, error) {
	return selectUsers(func(query string, params ...any) (*sql.Rows, error) {
		return t.tx.Query(query, params...)
	}, filter)
}

// CreateUser creates a user in the database.
func (t *Tx) CreateUser(u models.User) error {
	return insertUser(t.tx.Exec, u)
}

// UpdateUser updates a user in the database.
func (t *Tx) UpdateUser(u models.User) error {
	return updateUser(t.tx.Exec, u)
}

// DeleteUser delete users based on the id.
func (t *Tx) DeleteUser(id string) error {
	return deleteUser(t.tx.Exec, id)
}

// CreateEmailToken creates an email token in the database.
func (t *Tx) CreateEmailToken(tok models.EmailToken) error {
	return insertEmailToken(t.tx.Exec, tok)
}

// UpdateEmailToken updates an email token in the database.
func (t *Tx) UpdateEmailToken(tok models.EmailToken) error {
	return updateEmailToken(t.tx.Exec, tok)
}

// CreateFlat creates a flat in the database.
func (t *Tx) CreateFlat(f models.Flat) error {
	return insertFlat(t.tx.Exec, f)
}

// UpdateFlat updates a flat in the database.
func (t *Tx) UpdateFlat(u models.Flat) error {
	return updateFlat(t.tx.Exec, u)
}

// DeleteFlat delete users based on the id.
func (t *Tx) DeleteFlat(id string) error {
	return deleteFlat(t.tx.Exec, id)
}
