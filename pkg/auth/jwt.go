package auth

import (
	"fmt"
	"time"

	"hestia/pkg/models"

	"github.com/dgrijalva/jwt-go"
)

// JWTConfig is the configuration for the Service.
type JWTConfig struct {
	SecretKey string
	// TokenDuration is the duration a token is valid.
	TokenDuration time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func NewServiceConfig(secretKey string, tokenDuration time.Duration) *JWTConfig {
	return &JWTConfig{secretKey, tokenDuration}
}

func (manager *JWTConfig) Generate(user *models.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.TokenDuration).Unix(),
		},
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.SecretKey))
}

func (manager *JWTConfig) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected token signing method")
		}

		return []byte(manager.SecretKey), nil
	},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
