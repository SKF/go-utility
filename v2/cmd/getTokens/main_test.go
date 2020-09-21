package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/cmd/getTokens/model"

	"github.com/stretchr/testify/mock"
)

func TestSignInInitiate(t *testing.T) {
	smock := ssoMock{}
	expectedTokens := model.Tokens{
		IdentityToken: "idToken",
		AccessToken:   "AToken",
		RefreshToken:  "Rtoken",
	}

	smock.On("SignInInitiate", mock.Anything).Return(expectedTokens, nil)

	tokens, err := smock.SignInInitiate(model.Config{})
	require.NoError(t, err)
	require.Equal(t, expectedTokens.AccessToken, tokens.AccessToken)
}

type ssoMock struct {
	mock.Mock
}

func (m *ssoMock) SignInInitiate(cfg model.Config) (model.Tokens, error) {
	args := m.Called(cfg)

	return args.Get(0).(model.Tokens), args.Error(1)
}
