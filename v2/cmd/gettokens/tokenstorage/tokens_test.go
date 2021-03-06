package tokenstorage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
)

func getFile() (*os.File, error) {
	return os.OpenFile("apa.yaml", os.O_RDWR|os.O_CREATE, 0)
	//return os.Open("apa.yaml")
}

func TestStorage_noTokens(t *testing.T) {
	f, err := getFile()
	require.NoError(t, err)
	defer f.Close()

	s := tokenstorage.New(f)

	_, err = s.GetTokens("")
	require.Error(t, tokenstorage.ErrNotFound, err)
}

func TestStorage_GetTokens(t *testing.T) {
	f, err := getFile()
	require.NoError(t, err)
	defer f.Close()

	s := tokenstorage.New(f)

	tokens, err := s.GetTokens("sandbox")
	require.NoError(t, err)

	require.Equal(t, "actoken", tokens.AccessToken)
	require.Equal(t, "idToken", tokens.IdentityToken)
	require.Equal(t, "refreshToken", tokens.RefreshToken)
}

func TestStorage_GetTokensTwice(t *testing.T) {
	f, err := getFile()
	require.NoError(t, err)
	defer f.Close()

	s := tokenstorage.New(f)

	for i := 0; i < 2; i++ {
		tokens, err := s.GetTokens("sandbox")
		require.NoError(t, err)

		require.Equal(t, "actoken", tokens.AccessToken)
		require.Equal(t, "idToken", tokens.IdentityToken)
		require.Equal(t, "refreshToken", tokens.RefreshToken)
	}
}

func TestStorage_SetTokens(t *testing.T) {
	f, err := getFile()
	require.NoError(t, err)
	defer f.Close()

	s := tokenstorage.New(f)
	tokens := auth.Tokens{
		AccessToken:   "atok",
		IdentityToken: "itok",
		RefreshToken:  "rtok",
	}

	stage := "test"
	err = s.SetTokens(stage, tokens)
	require.NoError(t, err)

	fetchedTokens, err := s.GetTokens(stage)
	require.NoError(t, err)

	require.Equal(t, "atok", fetchedTokens.AccessToken)
	require.Equal(t, "itok", fetchedTokens.IdentityToken)
	require.Equal(t, "rtok", fetchedTokens.RefreshToken)
}

// err no stage not found
