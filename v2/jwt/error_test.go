package jwt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoNotLooseUnderlyingError(t *testing.T) {
	internalError := fmt.Errorf("my nice error")
	err := errNotValidNow(internalError)

	require.ErrorIs(t, err, internalError)
}

func TestIsNotValidNowError(t *testing.T) {
	err := errNotValidNow(errors.New(""))

	require.True(t, errors.Is(err, ErrNotValidNow))
}

func TestRandomErrorIsNotNotValidNowErr(t *testing.T) {
	err := errors.New("random Error")
	require.False(t, errors.Is(err, errExpiredType{}))
}

func TestErrorTextIsKept(t *testing.T) {
	internalError := errors.New("my error")
	err := errNotValidNow(internalError)

	fmt.Printf("%s\n", err)
	require.Contains(t, err.Error(), internalError.Error())
}
