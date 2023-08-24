package jwt

import "errors"

type ErrNotValidNow struct {
	underlyingErr error
}

func (e ErrNotValidNow) Error() string {
	return "token is not valid right now: " + e.underlyingErr.Error()
}

func (e ErrNotValidNow) Unwrap() error {
	return e.underlyingErr
}

func (e ErrNotValidNow) Is(target error) bool {
	switch target.(type) {
	case ErrNotValidNow, *ErrNotValidNow:
		return true
	}

	return errors.Is(e.underlyingErr, target)
}
