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

	var claims jwt.RegisteredClaims

	_, _, err := jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return false
	}

	ts := time.Now()

	if claims.ExpiresAt != nil && ts.After(claims.ExpiresAt.Time.Add(-tokenExpireDurationDiff)) {
		return false
	}

	if claims.IssuedAt != nil && ts.Before(claims.IssuedAt.Time) {
		return false
	}

	if claims.NotBefore != nil && ts.Before(claims.NotBefore.Time) {
		return false
	}

	return true
}
