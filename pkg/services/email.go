package services

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"hestia/pkg/custerrors"
	"hestia/pkg/models"
	"hestia/pkg/repos"
	"hestia/pkg/utils"
)

// Mailer is used to send templated email.
type Mailer interface {
	Send(ctx context.Context, template string, to string, data interface{}) error
}

// EmailService is the type that provides the main rules for authentication.
type EmailService struct {
	repo       *repos.Store
	wg         *sync.WaitGroup
	errHandler ErrFunc
	mailer     Mailer

	// NowFunc is used to get the current time.
	// Exposed for testing purposes.
	NowFunc func() time.Time
}

// NewEmailService creates a new Service.
func NewEmailService(db *sql.DB, errHandler ErrFunc) *EmailService {
	svc := &EmailService{
		repo:       repos.New(db),
		wg:         &sync.WaitGroup{},
		errHandler: errHandler,

		NowFunc: time.Now,
	}

	return svc
}

// RequestPasswordReset requests a password reset for the user with the provided email address.
func (s *EmailService) RequestPasswordReset(ctx context.Context, addr string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		wCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		err := s.passwordReset(wCtx, addr)
		if err != nil {
			s.errHandler(err)
			return
		}
	}()
}

func (s *EmailService) passwordReset(ctx context.Context, addr string) error {
	now := s.NowFunc()

	token, err := utils.TokenGenerator()
	if err != nil {
		return err
	}

	tokenID := utils.RandStringBytes()

	emailToken := models.EmailToken{
		ID:         tokenID,
		TokenHash:  token,
		UserID:     "",
		Email:      addr,
		Purpose:    models.TokenPurposePasswordReset,
		CreatedAt:  now,
		ConsumedAt: nil,
	}

	err = s.inTx(ctx, func(tx models.Tx) error {
		// Find the user with the provided email address.

		user, txErr := s.findUserByEmail(ctx, models.UserFilter{
			Emails:   []string{addr},
			IsActive: true,
		})
		if txErr != nil {
			return txErr
		}

		// Create the new password reset token.
		emailToken.UserID = user.ID

		txErr = tx.CreateEmailToken(emailToken)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = s.mailer.Send(ctx, "password-reset-request", addr, models.EmailTokenRaw{
		ID:    emailToken.ID,
		Token: token,
	})
	if err != nil {
		return err
	}

	return nil

}

func (s *EmailService) findUserByEmail(ctx context.Context, filter models.UserFilter) (models.User, error) {
	users, err := s.repo.FindUsers(ctx, filter)
	if err != nil {
		return models.User{}, err
	}

	if len(users) != 1 {
		return models.User{}, custerrors.ErrNotFound
	}

	return users[0], nil
}

func (s *EmailService) inTx(ctx context.Context, f func(tx models.Tx) error) error {
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
