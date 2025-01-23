package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// IsTokenValid checks if the token is still valid
func IsTokenValid(token string, tokenExpireDurationDiff time.Duration) bool {
	if token == "" {
		return false
	}

	parser := jwt.Parser{
		SkipClaimsValidation: true,
	}

	var claims jwt.RegisteredClaims

	_, _, err := parser.ParseUnverified(token, &claims)
	if err != nil {
		return false
	}

	// Verify if token still valid within the current time diff
	// no need to sign in once again
	ts := time.Now().Add(tokenExpireDurationDiff)

	return claims.VerifyExpiresAt(ts, false) &&
		claims.VerifyIssuedAt(ts, false) &&
		claims.VerifyNotBefore(ts, false)
}
