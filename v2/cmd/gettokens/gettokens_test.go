package gettokens_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/cmd/gettokens"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage/fakefile"
)

func TestSignIn_New(t *testing.T) {
	gettokens.New(nil, nil)
}

const (
	user         = "user"
	userPassword = "password"
)

func TestSignIn_CorrectPasswordReturnsToken(t *testing.T) {
	output := myWriter{}
	input := myReader{}
	handler := gettokens.New(&input, &output).
		WithSignIn(SignInMock).
		WithStorage(tokenstorage.New(fakefile.New()))

	input.writeString(user + "\n")
	input.writeString(userPassword + "\n")

	tokens, err := handler.SignIn()
	require.NoError(t, err)
	require.NotEmpty(t, tokens.RefreshToken)
	require.NotEmpty(t, tokens.IdentityToken)
	require.NotEmpty(t, tokens.AccessToken)

	expectedMessage := "please enter username\nplease enter password\n"

	require.Equal(t, expectedMessage, output.getOutput())
}

func TestSignIn_InvalidPasswordReturnsError(t *testing.T) {
	output := myWriter{}
	input := myReader{}
	input.writeString(user + "\n")
	input.writeString("badPassword\n")

	handler := gettokens.New(&input, &output).
		WithSignIn(SignInMock).
		WithStorage(tokenstorage.New(fakefile.New()))

	_, err := handler.SignIn()
	require.Error(t, err)
}

func TestSignIn_UseRefreshTokenIfExists(t *testing.T) {
	output := myWriter{}
	input := myReader{}
	const testdata = `sandbox:
  accesstoken: actoken
  identitytoken: idToken
  refreshtoken: old-refresh-token`
	s := tokenstorage.New(fakefile.New([]byte(testdata)...))

	handler := gettokens.New(&input, &output).WithSignInToken(SignInTokenMock).WithStorage(s)

	token, err := handler.SignIn()
	require.NoError(t, err)

	require.Equal(t, "new-access-token", token.AccessToken)
}

type myWriter struct {
	data []byte
}

func (w *myWriter) Write(newData []byte) (int, error) {
	w.data = append(w.data, newData...)

	return len(newData), nil
}

func (w *myWriter) getOutput() string {
	return string(w.data)
}

type myReader struct {
	data []byte
}

func (r *myReader) Read(p []byte) (int, error) {
	for i := range p {
		if i >= len(r.data) {
			r.data = r.data[i:]
			return i, nil
		}

		p[i] = r.data[i]
	}

	r.data = []byte{}
	return len(p), nil
}

func (r *myReader) writeString(s string) {
	r.data = append(r.data, []byte(s)...)
}

func SignInMock(ctx context.Context, username, password string) (auth.Tokens, error) {
	users := map[string]string{
		user: userPassword,
	}

	pw, ok := users[username]
	if !ok {
		return auth.Tokens{}, fmt.Errorf("user: %s not found", username)
	}

	if pw == password {
		return auth.Tokens{
			AccessToken:   "new-access-token",
			IdentityToken: "new-id-token",
			RefreshToken:  "new-refresh-token",
		}, nil
	}

	return auth.Tokens{}, fmt.Errorf("bad password")
}

func SignInTokenMock(ctx context.Context, refreshToken string) (auth.Tokens, error) {
	tokens := map[string]bool{
		"old-refresh-token": true,
	}

	if !tokens[refreshToken] {
		return auth.Tokens{}, fmt.Errorf("invalid token: %s", refreshToken)
	}

	return auth.Tokens{
		AccessToken:   "new-access-token",
		IdentityToken: "new-id-token",
		RefreshToken:  "new-refresh-token",
	}, nil
}
