package jwt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestDoNotLooseUnderlyingError(t *testing.T) {
	err1 := fmt.Errorf("my nice error")
	err2 := fmt.Errorf("my nice error: %w", err1)

	err := errNotValidNowType{underLyingErr: err2}

	require.ErrorIs(t, err, err1)
}

func TestIsNotValidNowError(t *testing.T) {
	err := errNotValidNowType{errors.New("")}

	require.True(t, errors.Is(err, ErrNotValidNow))
}

func TestRandomErrorIsNotNotValidNowErr(t *testing.T) {
	err := errors.New("random Error")
	require.False(t, errors.Is(err, ErrNotValidNow))
}

func TestIsNotValidNowErrorIsNotRandomError(t *testing.T) {
	err := errNotValidNowType{errors.New("")}

	require.False(t, errors.Is(err, errors.New("my random error")))
}

func TestErrorTextIsKept(t *testing.T) {
	internalError := errors.New("my error")
	err := errNotValidNowType{underLyingErr: internalError}

	fmt.Printf("%s\n", err)
	require.Contains(t, err.Error(), internalError.Error())
}

func TestIs(t *testing.T) {
	require.False(t, errors.Is(errors.New(""), ErrNotValidNow))
}

func TestGetUnderlying(t *testing.T) {
	err1 := jwt.NewValidationError("oh no", 2)
	err := errNotValidNowType{
		underLyingErr: err1,
	}

	ve := &jwt.ValidationError{}
	vep := &ve
	require.True(t, errors.As(err, vep))
}
