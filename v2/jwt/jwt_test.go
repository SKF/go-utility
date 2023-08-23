package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	ljwk "github.com/lestrrat-go/jwx/v2/jwk"
	ljwt "github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests relies on global variables in packages and can't be run in
// parallel.

func createKey(t *testing.T) (ljwk.Key, ljwk.Set) {
	// Create an RSA keypair
	valid, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JWK key from the RSA keypair
	validKey, err := ljwk.FromRaw(valid)
	require.NoError(t, err)

	// Adds fields expected by our packages
	// The "kid" needs to be different in each test to ensure
	// the library refetches the keyset.
	validKey.Set(ljwk.KeyIDKey, t.Name())      //nolint:errcheck
	validKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
	validKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

	// Cast to a private key
	validJWTKey, ok := validKey.(ljwk.RSAPrivateKey)
	require.True(t, ok)

	// Create a JWKS
	validSet := ljwk.NewSet()

	err = validSet.AddKey(validJWTKey)
	require.NoError(t, err)

	return validKey, validSet
}

func createJWKSServer(t *testing.T, set ljwk.Set) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// The server returns the public keyset
		public, err := ljwk.PublicSetOf(set)
		require.NoError(t, err)
		require.NoError(t, json.NewEncoder(w).Encode(public))
	}))

	jwk.KeySetURL = s.URL

	return s
}

func createSignedToken(t *testing.T, key ljwk.Key, values map[string]any) []byte {
	// Create a token and add fields for an access token
	token := ljwt.New()

	for key, value := range values {
		token.Set(key, value) //nolint:errcheck
	}

	// Signed it using the private key
	signed, err := ljwt.Sign(token, ljwt.WithKey(jwa.RS256, key))
	require.NoError(t, err)

	return signed
}

func Test_AccessToken(t *testing.T) {
	validKey, validSet := createKey(t)

	s := createJWKSServer(t, validSet)
	defer s.Close()

	signed := createSignedToken(t, validKey, map[string]any{
		"token_use": jwt.TokenUseAccess,
		"username":  "a.b@example.com",
	})

	// Parse it
	parsedToken, err := jwt.Parse(string(signed))
	require.NoError(t, err)

	claims := parsedToken.GetClaims()

	// Verify that the expected fields are present in the parse token
	assert.Equal(t, jwt.TokenUseAccess, claims.CognitoClaims.TokenUse)
	assert.Equal(t, "a.b@example.com", claims.CognitoClaims.Username)
}

func Test_InvalidKeyShouldFailValidation(t *testing.T) {
	_, validSet := createKey(t)

	fakeKey, _ := createKey(t)

	s := createJWKSServer(t, validSet)
	defer s.Close()

	signed := createSignedToken(t, fakeKey, map[string]any{
		"token_use": jwt.TokenUseAccess,
		"username":  "a.b@example.com",
	})

	_, err := jwt.Parse(string(signed))
	require.Error(t, err)
}

func Test_ExpiredToken(t *testing.T) {
	validKey, validSet := createKey(t)

	s := createJWKSServer(t, validSet)
	defer s.Close()

	signed := createSignedToken(t, validKey, map[string]any{
		ljwt.ExpirationKey: time.Now().UTC().Add(-1 * time.Hour),
		"token_use":        jwt.TokenUseAccess,
		"username":         "a.b@example.com",
	})

	_, err := jwt.Parse(string(signed))
	require.Error(t, err)
}

func Test_ValidInFuture(t *testing.T) {
	validKey, validSet := createKey(t)

	s := createJWKSServer(t, validSet)
	defer s.Close()

	signed := createSignedToken(t, validKey, map[string]any{
		ljwt.NotBeforeKey: time.Now().UTC().Add(1 * time.Hour),
		"token_use":       jwt.TokenUseAccess,
		"username":        "a.b@example.com",
	})

	_, err := jwt.Parse(string(signed))
	require.Error(t, err)
}
