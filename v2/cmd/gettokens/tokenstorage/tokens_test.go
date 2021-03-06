package tokenstorage_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/cmd/gettokens/tokenstorage"
)

func TestStorage(t *testing.T) {
	s := tokenstorage.New()

	_, err := s.GetTokens()
	require.Error(t, tokenstorage.ErrNotFound, err)
}
