package repos

import (
	"hestia/pkg/custerrors"
	"hestia/pkg/db"
	"hestia/pkg/models"
)

func insertFlat(ef execFunc, f models.Flat) error {
	q := db.Query{}
	count := 0

	q.Unsafe(`INSERT INTO "flats" (title, price, 
                   					address, surface, rooms, 
                   					floor, available_from, rent, 
                   					deposit, description, created_at, updated_at) VALUES (`)
	q.Params(&count,
		f.Title,
		f.Price,
		f.Address,
		f.Surface,
		f.Rooms,
		f.Floor,
		f.AvailableFrom,
		f.Rent,
		f.Deposit,
		f.Description,
		f.CreatedAt,
		f.UpdatedAt)
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
