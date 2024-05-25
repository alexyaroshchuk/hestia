package auth

import (
	"fmt"
)

type Interceptor struct {
	jwtManager      *JWTConfig
	accessibleRoles map[string][]string
}

func NewAuthInterceptor(jwtManager *JWTConfig, accessibleRoles map[string][]string) *Interceptor {
	return &Interceptor{jwtManager, accessibleRoles}
}

func (interceptor *Interceptor) Authorize(method, accessToken string) (userID string, err error) {
	accessibleRoles, ok := interceptor.accessibleRoles[method]
	if !ok {
		return "", fmt.Errorf("no permission to access this RPC")
	}
	claims, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	for _, role := range accessibleRoles {
		if role == claims.Role {
			return claims.ID, nil
		}
	}

	return "", fmt.Errorf("no permission to access this RPC")
}
