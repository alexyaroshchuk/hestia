package main

import (
	"database/sql"
	"fmt"

	"hestia/pkg/custerrors"
	"hestia/pkg/db"
	"hestia/pkg/models"

	"github.com/brianvoe/gofakeit/v6"
)

var conn = "host=localhost port=5432 user=test password=password dbname=hestia sslmode=disable"

func main() {
	// Fake flats
	var flat models.Flat
	var dummyFlats []models.Flat
	pgsql, err := db.OpenPGSQL(conn)
	if err != nil {
		return
	}

	for i := 0; i < 20; i++ {
		err := gofakeit.Struct(&flat)
		if err != nil {
			return
		}

		err = insert(flat, pgsql)
		if err != nil {
			fmt.Println("got error from transaction:", err)
			return
		}
	}

	fmt.Println(dummyFlats)
}

func insert(f models.Flat, pgsql *sql.DB) error {
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

	_, err = pgsql.Exec(s, params...)
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	return nil
}
