package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	ljwk "github.com/lestrrat-go/jwx/v2/jwk"
	ljwt "github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AccessToken(t *testing.T) {
	// Create an RSA keypair
	valid, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create a JWK key from the RSA keypair
	validKey, err := ljwk.FromRaw(valid)
	require.NoError(t, err)

	// Adds fields expected by our packages
	validKey.Set(ljwk.KeyIDKey, "thekey")      //nolint:errcheck
	validKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
	validKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

	// Cast to a private key
	validJWTKey, ok := validKey.(ljwk.RSAPrivateKey)
	require.True(t, ok)

	// Create a JWKS
	validSet := ljwk.NewSet()

	err = validSet.AddKey(validJWTKey)
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// The server returns the public keyset
		publicSet, setErr := ljwk.PublicSetOf(validSet)
		require.NoError(t, setErr)
		err = json.NewEncoder(w).Encode(publicSet)
		require.NoError(t, err)
	}))
	defer s.Close()

	jwk.KeySetURL = s.URL

	// Create a token and add fields for an access token
	token := ljwt.New()
	token.Set("token_use", jwt.TokenUseAccess) //nolint:errcheck
	token.Set("username", "a.b@example.com")   //nolint:errcheck

	// Signed it using the private key
	signed, err := ljwt.Sign(token, ljwt.WithKey(jwa.RS256, validKey))
	require.NoError(t, err)

	// Parse it
	parsedToken, err := jwt.Parse(string(signed))
	require.NoError(t, err)

	claims := parsedToken.GetClaims()

	// Verify that the expected fields are present in the parse token
	assert.Equal(t, jwt.TokenUseAccess, claims.CognitoClaims.TokenUse)
	assert.Equal(t, "a.b@example.com", claims.CognitoClaims.Username)
}

func Test_InvalidKeyShouldFailValidation(t *testing.T) {
	valid, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	validKey, err := ljwk.FromRaw(valid)
	require.NoError(t, err)

	validKey.Set(ljwk.KeyIDKey, "thekey")      //nolint:errcheck
	validKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
	validKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

	validJWTKey, ok := validKey.(ljwk.RSAPrivateKey)
	require.True(t, ok)

	validSet := ljwk.NewSet()

	err = validSet.AddKey(validJWTKey)
	require.NoError(t, err)

	fake, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	fakeKey, err := ljwk.FromRaw(fake)
	require.NoError(t, err)

	fakeKey.Set(ljwk.KeyIDKey, "thekey")      //nolint:errcheck
	fakeKey.Set(ljwk.AlgorithmKey, jwa.RS256) //nolint:errcheck
	fakeKey.Set(ljwk.KeyUsageKey, "sig")      //nolint:errcheck

	fakeJWTKey, ok := fakeKey.(ljwk.RSAPrivateKey)
	require.True(t, ok)

	fakeSet := ljwk.NewSet()

	err = fakeSet.AddKey(fakeJWTKey)
	require.NoError(t, err)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		publicSet, setErr := ljwk.PublicSetOf(validSet)
		require.NoError(t, setErr)
		err = json.NewEncoder(w).Encode(publicSet)
		require.NoError(t, err)
	}))
	defer s.Close()

	jwk.KeySetURL = s.URL

	token := ljwt.New()
	token.Set(`token_use`, `access`)         //nolint:errcheck
	token.Set(`username`, `a.b@example.com`) //nolint:errcheck

	signed, err := ljwt.Sign(token, ljwt.WithKey(jwa.RS256, fakeKey))
	require.NoError(t, err)

	_, err = jwt.Parse(string(signed))
	require.Error(t, err)
}
