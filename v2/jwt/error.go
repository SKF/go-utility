package jwt

import "errors"

type ErrNotValidNow struct {
	underLyingErr error
}

func (e ErrNotValidNow) Error() string {
	return "token is not valid right now: " + e.underLyingErr.Error()
}

func (e ErrNotValidNow) Unwrap() error {
	return e.underLyingErr
}

func (e ErrNotValidNow) Is(target error) bool {
	switch target.(type) {
	case ErrNotValidNow, *ErrNotValidNow:
		return true
	}

	return errors.Is(e.underLyingErr, target)
}
