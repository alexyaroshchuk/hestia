package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"hestia/pkg/auth"
	"hestia/pkg/models"
	"hestia/pkg/repos"
)

type EmailInterface interface {
	Login(w http.ResponseWriter, r *http.Request) error
	Register(w http.ResponseWriter, r *http.Request) error
}

type AuthService struct {
	repo       *repos.Store
	jwtManager *auth.JWTConfig
	errHandler ErrFunc
}

func NewAuthServer(db *sql.DB, jwtManager *auth.JWTConfig, errHandler ErrFunc) *AuthService {
	return &AuthService{
		repo:       repos.New(db),
		jwtManager: jwtManager,
		errHandler: errHandler,
	}
}

func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) error {
	var c models.Credentials
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		s.errHandler(err)
		return err
	}

	users, err := s.repo.FindUsers(r.Context(), models.UserFilter{
		Emails:   []string{c.Email},
		IsActive: true,
	})
	if err != nil {
		s.errHandler(err)
		return err
	}
	if len(users) == 0 {
		s.errHandler(err)
		return ErrInvalidCredentials
	}
	match := models.Match(users[0].PasswordHash, c.Password)
	if !match {
		s.errHandler(err)
		return ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(&models.User{
		ID:    users[0].ID,
		Email: users[0].Email,
		Role:  users[0].Role,
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	j, err := json.Marshal(token)
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

func (s *AuthService) Register(w http.ResponseWriter, r *http.Request) error {
	var user models.Register
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		s.errHandler(err)
		return err
	}
	ps, err := models.ParsePassword(user.Password)
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
		err = tx.CreateUser(models.User{
			Email:        user.Email,
			PasswordHash: string(pwdHash),
			Role:         defaultRole,
			IsActive:     true,
		})
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

func (s *AuthService) inTx(ctx context.Context, f func(tx models.Tx) error) error {
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
