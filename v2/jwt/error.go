package jwt

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

func (e errNotValidNowType) Is(_ error) bool {
	return true
}
