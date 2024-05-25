package models

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 512

	SecretMarker = "<!SECRET_REDACTED!>"
)

var ErrInvalidPassword = fmt.Errorf("invalid password")

type Password struct {
	plain []byte
}

// ParsePassword creates a new Password from a plaintext string.
// It errors if the password is too short or too long.
func ParsePassword(pwd string) (Password, error) {
	if len(pwd) < minPasswordBytes || len(pwd) > maxPasswordBytes {
		return Password{}, ErrInvalidPassword
	}

	return Password{
		plain: []byte(pwd),
	}, nil
}

// Match checks if the plaintext password matches the given hash.
func Match(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Hash hashes the plaintext password.
func (p Password) Hash() ([]byte, error) {
	// Need to invert the call because we don't want to expose p.plain.
	return bcrypt.GenerateFromPassword(p.plain, bcrypt.DefaultCost)
}

func (p Password) Format(f fmt.State, verb rune) {
	f.Write([]byte(SecretMarker))
}

func (p Password) MarshalText() ([]byte, error) {
	return []byte(SecretMarker), nil
}
