package tokenstorage_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage/fakefile"
)

func TestStorage_noTokens(t *testing.T) {
	f := fakefile.New()

	s := tokenstorage.New(f)

	_, err := s.GetTokens("")
	require.Error(t, tokenstorage.ErrNotFound, err)
}

func TestStorage_GetTokens(t *testing.T) {
	f := fakefile.New([]byte(testdata)...)

	s := tokenstorage.New(f)

	tokens, err := s.GetTokens("sandbox")
	require.NoError(t, err)

	require.Equal(t, "actoken", tokens.AccessToken)
	require.Equal(t, "idToken", tokens.IdentityToken)
	require.Equal(t, "refreshToken", tokens.RefreshToken)
}

const testdata = `
sandbox:
  accesstoken: actoken
  identitytoken: idToken
  refreshtoken: refreshToken`

func TestStorage_GetTokensTwice(t *testing.T) {
	f := fakefile.New([]byte(testdata)...)
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
	f := fakefile.New([]byte(testdata)...)
	s := tokenstorage.New(f)
	tokens := auth.Tokens{
		AccessToken:   "atok",
		IdentityToken: "itok",
		RefreshToken:  "rtok",
	}

	stage := "test"
	err := s.SetTokens(stage, tokens)
	require.NoError(t, err)

	fetchedTokens, err := s.GetTokens(stage)
	require.NoError(t, err)

	require.Equal(t, "atok", fetchedTokens.AccessToken)
	require.Equal(t, "itok", fetchedTokens.IdentityToken)
	require.Equal(t, "rtok", fetchedTokens.RefreshToken)
}

func TestStorage_StageNotFound(t *testing.T) {
	f := fakefile.New([]byte(testdata)...)
	s := tokenstorage.New(f)
	stage := "invalidStage"

	_, err := s.GetTokens(stage)
	require.Error(t, err)

}
