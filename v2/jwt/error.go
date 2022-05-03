package jwt

import "errors"

var ErrNotValidNow = errNotValidNowType{}

type errNotValidNowType struct {
	underLyingErr error
}

func (e errNotValidNowType) Error() string {
	return "token is not valid right now: " + e.underLyingErr.Error()
}

func (e errNotValidNowType) Unwrap() error {
	return e.underLyingErr
}

func (e errNotValidNowType) Is(target error) bool {
	switch target.(type) {
	case errNotValidNowType:
		return true
	}

	return errors.Is(e.underLyingErr, target)
}
