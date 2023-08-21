package jwt

import "errors"

var ErrNotValidNow = errors.New("token is not valid right now")
