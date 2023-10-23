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

	ts := time.Now().Add(tokenExpireDurationDiff)

	for _, claim := range []*jwt.NumericDate{
		claims.ExpiresAt,
		claims.IssuedAt,
		claims.NotBefore,
	} {
		if claim == nil {
			continue
		}

		if claim.Before(ts) {
			return false
		}
	}

	return true
}
