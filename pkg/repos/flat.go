package repos

import (
	"hestia/pkg/custerrors"
	"hestia/pkg/db"
	"hestia/pkg/models"
)

func insertFlat(ef execFunc, u models.Flat) error {
	q := db.Query{}
	count := 0

	q.Unsafe(`INSERT INTO flats (id, email, password_hash, is_active, created_at, updated_at) VALUES (`)
	q.Params(&count, u.ID, u.Email, u.PasswordHash, u.IsActive, u.CreatedAt, u.UpdatedAt)
	q.Unsafe(`)`)

	s, params, err := q.Get()
	if err != nil {
		return err
	}

	_, err = ef(s, params...)
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	return nil
}

func updateFlat(ef execFunc, u models.Flat) error {
	return nil
}

func selectFlat(qf queryFunc, id string) (models.Flat, error) {
	return models.Flat{}, nil
}

func selectFlats(qf queryFunc, f models.FlatFilter) ([]models.Flat, error) {

	return nil, nil
}

func deleteFlat(ef execFunc, uuid string) error {
	return nil
}
