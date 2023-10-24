package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/golang-jwt/jwt/v5"
)

func Test_IsTokenValid(t *testing.T) {
	mySigningKey := []byte("test_key")
	ts := time.Now()

	tests := []struct {
		expiresAt, issuedAt time.Time
		name                string
		expireDurationDiff  time.Duration
		expected            bool
	}{
		{
			name:               "valid claims",
			expiresAt:          ts.Add(time.Hour),
			expireDurationDiff: time.Minute * 5,
			issuedAt:           ts.Add(-(time.Minute * 10)),
			expected:           true,
		},
		{
			name:               "issuedAt in future",
			expiresAt:          ts.Add(time.Hour),
			expireDurationDiff: time.Minute * 5,
			issuedAt:           ts.Add(time.Hour),
			expected:           false,
		},
		{
			name:               "tokenexpiration inside diff window",
			expiresAt:          ts.Add(time.Minute * 4),
			expireDurationDiff: time.Minute * 5,
			issuedAt:           ts,
			expected:           false,
		},
		{
			name:               "token expired",
			expiresAt:          ts.Add(-time.Minute),
			expireDurationDiff: 0,
			issuedAt:           ts.Add(-(time.Minute * 10)),
			expected:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := &jwt.RegisteredClaims{
				ExpiresAt: &jwt.NumericDate{Time: test.expiresAt},
				IssuedAt:  &jwt.NumericDate{Time: test.issuedAt},
				Issuer:    "test",
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			ss, err := token.SignedString(mySigningKey)

			require.NoError(t, err)
			require.Equal(t, test.expected, IsTokenValid(ss, test.expireDurationDiff))
		})

	}

}
