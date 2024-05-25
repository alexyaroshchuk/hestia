package models

import (
	"time"
)

// EmailToken contains the state of a token that was sent via email.
type EmailToken struct {
	ID         string
	TokenHash  string
	UserID     string
	Email      string
	Purpose    TokenPurpose
	CreatedAt  time.Time
	ConsumedAt *time.Time
}

// TokenPurpose is the purpose of an email token.
type TokenPurpose string

const (
	// TokenPurposeActivate indicates a token should be used to activate an user.
	TokenPurposeActivate TokenPurpose = "activate"
	// TokenPurposePasswordReset indicates a token should be used to reset a password.
	TokenPurposePasswordReset TokenPurpose = "password_reset"
)

// EmailTokenRaw is the raw data that will be send to the user via email.
type EmailTokenRaw struct {
	ID    string
	Token string
}
