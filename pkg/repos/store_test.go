package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hestia/pkg/custerrors"
	"hestia/pkg/models"
	"reflect"
	"testing"
	"time"
)

func Test_Tx_CreateUser(t *testing.T) {
	t.Run("ok, create user", inTx(func(t *testing.T, tx models.Tx) {
		user := newUser(t, nil)

		err := tx.CreateUser(user)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}

		assertFindUser(t, tx, user)
	}))

	t.Run("fail, email constraint violated", inTx(func(t *testing.T, tx models.Tx) {
		user1 := newUser(t, nil)
		err := tx.CreateUser(user1)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}

		user2 := newUser(t, nil)
		err = tx.CreateUser(user2)
		if !errors.Is(err, custerrors.ErrConstraintViolated) {
			t.Fatalf("expected errors to be %v got %v (via errors.Is)", custerrors.ErrConstraintViolated, err)
		}
	}))

	t.Run("fail, zero ID", inTx(func(t *testing.T, tx models.Tx) {
		user := newUser(t, func(u *models.User) {
			u.ID = ""
		})

		err := tx.CreateUser(user)
		if !errors.Is(err, custerrors.ErrConstraintViolated) {
			t.Fatalf("expected errors to be %v got %v (via errors.Is)", custerrors.ErrConstraintViolated, err)
		}
	}))
}

func Test_Tx_UpdateUser(t *testing.T) {
	setup := func(t *testing.T, tx models.Tx) models.User {
		user := newUser(t, nil)
		err := tx.CreateUser(user)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}

		return user
	}

	t.Run("ok, update user", inTx(func(t *testing.T, tx models.Tx) {
		user := setup(t, tx)

		// Update all fields that can be modified.
		//user.Email = "jacob@example.com")) //TODO fix it
		user.IsActive = true
		user.CreatedAt = now(t, 1)
		user.UpdatedAt = now(t, 2)

		err := tx.UpdateUser(user)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}

		assertFindUser(t, tx, user)
	}))

	t.Run("fail, not found", inTx(func(t *testing.T, tx models.Tx) {
		setup(t, tx)

		user2 := newUser(t, func(u *models.User) {
			u.ID = "7777"
		})

		err := tx.UpdateUser(user2)
		if !errors.Is(err, custerrors.ErrNotFound) {
			t.Fatalf("expected errors to be %v got %v (via errors.Is)", custerrors.ErrNotFound, err)
		}
	}))

	t.Run("fail, change email to an existing email", inTx(func(t *testing.T, tx models.Tx) {
		user1 := setup(t, tx)

		user2 := newUser(t, func(u *models.User) {
			u.ID = "12345678"
			u.Email = "test@test.com"
		})

		err := tx.CreateUser(user2)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}

		// Attempt to change user1's email to user2's email.
		user1.Email = user2.Email
		err = tx.UpdateUser(user1)
		if !errors.Is(err, custerrors.ErrConstraintViolated) {
			t.Fatalf("expected errors to be %v got %v (via errors.Is)", custerrors.ErrConstraintViolated, err)
		}
	}))
}

func Test_Tx_FindUser(t *testing.T) {
	setupUsers := func(t *testing.T, tx models.Tx) []models.User {
		users := []models.User{
			newUser(t, nil),
			newUser(t, func(u *models.User) {
				u.ID = "1234567890"
				u.Email = "test@test.com"
				u.IsActive = true
			}),
			newUser(t, func(u *models.User) {
				u.ID = "1234567890"
				u.Email = "test33@test.com"
			}),
		}

		for i := range users {
			err := tx.CreateUser(users[i])
			if err != nil {
				t.Fatalf("failed to save user: %v", err)
			}
		}

		return users
	}

	tests := map[string]struct {
		filter   models.UserFilter
		wantFunc func([]models.User) []models.User
	}{
		"ok, all users, empty slices": {
			filter: models.UserFilter{
				IDs:      []string{},
				Emails:   []string{},
				IsActive: true,
			},
			wantFunc: func(users []models.User) []models.User {
				return users
			},
		},
		"ok, active users": {
			filter: models.UserFilter{
				IsActive: true,
			},
			wantFunc: func(users []models.User) []models.User {
				return users[1:2]
			},
		},
		"ok, one by id": {
			filter: models.UserFilter{
				IDs: []string{"123456789"},
			},
			wantFunc: func(users []models.User) []models.User {
				return []models.User{users[1]}
			},
		},
		"ok, several by id": {
			filter: models.UserFilter{
				IDs: []string{
					"12345678",
					"98765432",
				},
			},
			wantFunc: func(users []models.User) []models.User {
				return []models.User{
					users[0], users[2],
				}
			},
		},
		"ok, one by email": {
			filter: models.UserFilter{
				Emails: []string{
					"test@test.com",
				},
			},
			wantFunc: func(users []models.User) []models.User {
				return []models.User{users[1]}
			},
		},
		"ok, several by email": {
			filter: models.UserFilter{
				Emails: []string{
					"jacob@example.com",
					"eva@example.com",
				},
			},
			wantFunc: func(users []models.User) []models.User {
				return []models.User{
					users[1], users[2],
				}
			},
		},
		"ok, combine filters": {
			filter: models.UserFilter{
				IDs: []string{
					"0e61a06e-bbf6-4b87-aaaa-75fee0f38cca",
					"d622d0b0-465c-4c4d-b084-028c9787e1de",
				},
				Emails: []string{
					"test5@test",
				},
				IsActive: false,
			},
			wantFunc: func(users []models.User) []models.User {
				return users[0:1]
			},
		},
		"ok, no results": {
			filter: models.UserFilter{
				IDs: []string{""},
			},
			wantFunc: func(users []models.User) []models.User {
				return []models.User{}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			store := storeForTest(t)

			tx, err := store.BeginTx(context.Background())
			if err != nil {
				t.Fatalf("failed to begin tx: %v", err)
			}

			users := setupUsers(t, tx)
			want := tc.wantFunc(users)

			// first check if FindUsers works on the tx
			got, err := tx.FindUsers(tc.filter)
			if err != nil {
				t.Fatalf("failed to find users: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got\n%#v\nwant\n%#v\n", got, want)
			}

			err = tx.Commit()
			if err != nil {
				t.Fatalf("failed to commit tx: %v", err)
			}

			// then, check if FindUsers works on the store itself.
			got, err = store.FindUsers(context.Background(), tc.filter)
			if err != nil {
				t.Fatalf("failed to find users: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got\n%#v\nwant\n%#v\n", got, want)
			}
		})
	}
}

func inTx(f func(*testing.T, models.Tx)) func(*testing.T) {
	return func(t *testing.T) {
		store := storeForTest(t)

		tx, err := store.BeginTx(context.Background())
		if err != nil {
			t.Fatalf("failed to begin tx: %v", err)
		}

		f(t, tx)

		err = tx.Commit()
		if err != nil {
			t.Fatalf("failed to commit tx: %v", err)
		}
	}
}

func now(t *testing.T, i int) time.Time {
	t.Helper()

	if i > 9 {
		t.Fatalf("invalid time index: %d", i)
	}

	ts, err := time.Parse(time.RFC3339, fmt.Sprintf("2021-01-01T00:00:0%dZ", i))
	if err != nil {
		t.Fatalf("failed to parse time: %v", err)
	}

	return ts
}

func storeForTest(t *testing.T) *Store {
	t.Helper()

	//testDB := testdb.RunWhile(t, true) //todo add test db implementation
	testDB := &sql.DB{}
	return New(testDB)
}

func newUser(t *testing.T, modFunc func(*models.User)) models.User {
	t.Helper()

	u := models.User{
		ID:           "123456789",
		Email:        "test@test.com",
		PasswordHash: "123456789",
		CreatedAt:    now(t, 0),
		UpdatedAt:    now(t, 0),
	}

	if modFunc != nil {
		modFunc(&u)
	}

	return u
}

func assertFindUser(t *testing.T, tx models.Tx, want models.User) {
	t.Helper()

	got, err := tx.FindUsers(models.UserFilter{IDs: []string{want.ID}})
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 user, got %d", len(got))
	}

	if !reflect.DeepEqual(got[0], want) {
		t.Errorf("got\n%#v\nwant\n%#v\n", got[0], want)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
