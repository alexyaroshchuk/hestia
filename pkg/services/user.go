package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"hestia/pkg/custerrors"
	"hestia/pkg/middlewares"
	"hestia/pkg/models"
	"hestia/pkg/repos"
)

var (
	ErrDuplicateUser      = errors.New("duplicate user")
	UserNotFound          = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const defaultRole = "user"

// ErrFunc is a function that handles errors.
type ErrFunc func(error)

type UserInterface interface {
	Get(w http.ResponseWriter, r *http.Request) error
	GetAll(w http.ResponseWriter, r *http.Request) error
	Put(w http.ResponseWriter, r *http.Request) error
	Post(w http.ResponseWriter, r *http.Request) error
	Delete(w http.ResponseWriter, r *http.Request) error
}

// UserService is the type that provides the main rules for authentication.
type UserService struct {
	repo       *repos.Store
	wg         *sync.WaitGroup
	errHandler ErrFunc

	// NowFunc is used to get the current time.
	// Exposed for testing purposes.
	NowFunc func() time.Time
}

// NewUserService creates a new Service.
func NewUserService(db *sql.DB, errHandler ErrFunc) *UserService {
	svc := &UserService{
		repo:       repos.New(db),
		wg:         &sync.WaitGroup{},
		errHandler: errHandler,

		NowFunc: time.Now,
	}

	return svc
}

func (s *UserService) Get(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	users, err := s.repo.FindUsers(r.Context(), models.UserFilter{
		IDs: []string{id},
	})
	if err != nil {
		s.errHandler(err)
		return err
	}
	if len(users) == 0 {
		s.errHandler(UserNotFound)
		return custerrors.ErrNotFound
	}

	j, err := json.Marshal(users[0])
	if err != nil {
		s.errHandler(err)
		return err
	}
	_, err = w.Write(j)
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *UserService) GetAll(w http.ResponseWriter, r *http.Request) error {
	_, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		return UserNotFound
	}

	users, err := s.repo.FindUsers(r.Context(), models.UserFilter{})
	if err != nil {
		s.errHandler(err)
		return err
	}
	if len(users) == 0 {
		s.errHandler(UserNotFound)
		return custerrors.ErrNotFound
	}

	j, err := json.Marshal(users)
	if err != nil {
		s.errHandler(err)
		return err
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(j)
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *UserService) Put(w http.ResponseWriter, r *http.Request) error {
	now := s.NowFunc()
	id := r.PathValue("id")
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		s.errHandler(err)
		return err
	}
	err = s.inTx(r.Context(), func(tx models.Tx) error {
		user.ID = id
		user.UpdatedAt = now
		err := tx.UpdateUser(user)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *UserService) Post(w http.ResponseWriter, r *http.Request) error {
	now := s.NowFunc()
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		s.errHandler(err)
		return err
	}
	ps, err := models.ParsePassword(user.PasswordHash)
	if err != nil {
		s.errHandler(err)
		return err
	}

	pwdHash, err := ps.Hash()
	if err != nil {
		s.errHandler(err)
		return err
	}
	err = s.inTx(r.Context(), func(tx models.Tx) error {
		user.UpdatedAt = now
		user.CreatedAt = now
		user.PasswordHash = string(pwdHash)
		user.IsActive = true
		err := tx.CreateUser(user)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *UserService) Delete(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	err := s.inTx(r.Context(), func(tx models.Tx) error {
		err := tx.DeleteUser(id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *UserService) inTx(ctx context.Context, f func(tx models.Tx) error) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		rBackErr := tx.Rollback()
		if rBackErr != nil {
			err = errors.Join(err, rBackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
