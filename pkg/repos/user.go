package repos

import (
	"fmt"

	"hestia/pkg/custerrors"
	"hestia/pkg/db"
	"hestia/pkg/models"
)

func insertUser(ef execFunc, u models.User) error {
	q := db.Query{}
	count := 0

	q.Unsafe(`INSERT INTO "users" (email, password_hash, role, is_active, created_at, updated_at) VALUES (`)
	q.Params(&count, u.Email, u.PasswordHash, u.Role, u.IsActive, u.CreatedAt, u.UpdatedAt)
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

func updateUser(ef execFunc, u models.User) error {
	q := db.Query{}
	count := 0

	q.Unsafe(`UPDATE users SET `)

	q.Unsafe(`updated_at = `)
	q.Param(&count, u.UpdatedAt)

	if u.Email != "" {
		q.Unsafe(`, email = `)
		q.Param(&count, u.Email)
	}

	if u.Role != "" {
		q.Unsafe(`, role = `)
		q.Param(&count, u.Role)
	}

	if u.PasswordHash != "" {
		q.Unsafe(`, password_hash = `)
		q.Param(&count, u.PasswordHash)
	}

	if u.IsActive != false {
		q.Unsafe(`, is_active = `)
		q.Param(&count, u.IsActive)
	}

	q.Unsafe(` WHERE id = `)
	q.Params(&count, u.ID)

	s, params, err := q.Get()
	if err != nil {
		return err
	}

	result, err := ef(s, params...)
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %w", custerrors.ErrNotFound)
	}

	return nil
}

func selectUsers(qf queryFunc, f models.UserFilter) ([]models.User, error) {
	q := db.Query{}
	count := 0
	q.Unsafe(`SELECT id, email, password_hash, role, is_active, created_at, updated_at FROM users WHERE 1=1 `)

	if len(f.IDs) > 0 {
		q.Unsafe(`AND id IN (`)
		q.Params(&count, anySlice(f.IDs)...)
		q.Unsafe(`)`)
	}

	if len(f.Emails) > 0 {
		q.Unsafe(`AND email IN (`)
		q.Params(&count, anySlice(f.Emails)...)
		q.Unsafe(`)`)
	}

	if f.IsActive != false {
		q.Unsafe("AND is_active = ")
		q.Param(&count, f.IsActive)
	}

	q.Unsafe(` ORDER BY id ASC`)

	s, params, err := q.Get()
	if err != nil {
		return nil, err
	}

	rows, err := qf(s, params...)
	if err != nil {
		return nil, custerrors.MapDBErr(err)
	}

	defer rows.Close()

	out := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, custerrors.MapDBErr(err)
		}

		out = append(out, u)
	}

	if err := rows.Err(); err != nil {
		return nil, custerrors.MapDBErr(err)
	}

	return out, nil
}

func deleteUser(ef execFunc, id string) error {
	q := db.Query{}
	count := 0

	q.Unsafe(`DELETE from users `)
	q.Unsafe(` WHERE id = `)
	q.Params(&count, id)

	s, params, err := q.Get()
	if err != nil {
		return err
	}

	result, err := ef(s, params...)
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return custerrors.MapDBErr(err)
	}

	if rows == 0 {

		return fmt.Errorf("user not found: %w", custerrors.ErrNotFound)
	}

	return nil
}
