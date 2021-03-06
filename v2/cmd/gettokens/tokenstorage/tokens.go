package tokenstorage

import (
	"fmt"

	"github.com/SKF/go-utility/v2/auth"
)

type Storage struct {
}

var ErrNotFound = fmt.Errorf("tokens not found")

func (s Storage) GetTokens() (auth.Tokens, error) {
	return auth.Tokens{}, ErrNotFound
}

func New() Storage {
	return Storage{}
}
