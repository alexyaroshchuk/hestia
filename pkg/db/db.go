package db

import (
	"database/sql"
)

const ()

// OpenPGSQL TODO read about some configs for POSTGRESQL
func OpenPGSQL(sqlcode string) (*sql.DB, error) {
	db, err := sql.Open("postgres", sqlcode)
	if err != nil {
		return nil, err
	}

	return db, nil
}
