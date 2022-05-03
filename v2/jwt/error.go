package jwt

var ErrNotValidNow = errExpiredType{}

func errNotValidNow(underlyingErr error) error {

	return errExpiredType{
		underLyingErr: underlyingErr,
	}
}

type errExpiredType struct {
	underLyingErr error
}

func (e errExpiredType) Error() string {
	return "token is not valid right now: " + e.underLyingErr.Error()
}

func (e errExpiredType) Unwrap() error {
	return e.underLyingErr
}

func (e errExpiredType) Is(_ error) bool {
	return true
}
