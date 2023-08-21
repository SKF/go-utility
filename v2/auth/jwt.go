package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// IsTokenValid checks if the token is still valid
func IsTokenValid(token string, tokenExpireDurationDiff time.Duration) bool {
	if token == "" {
		return false
	}

	parser := jwt.NewParser(jwt.WithLeeway(tokenExpireDurationDiff))

	var claims jwt.RegisteredClaims

	_, _, err := parser.ParseUnverified(token, &claims)
	return err == nil
}
