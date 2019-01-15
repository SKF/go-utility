package errors

import (
	"fmt"
	"runtime/debug"
)

type Error struct {
	Err   error
	Stack []byte
}

func New(format string, args ...interface{}) StackTrace {
	return Wrap(fmt.Errorf(format, args...))
}

func Wrap(err error) StackTrace {
	if stackerr, ok := err.(StackTrace); ok {
		return stackerr
	}

	return &Error{
		Err:   err,
		Stack: debug.Stack(),
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) StackTrace() []byte {
	return e.Stack
}

type StackTrace interface {
	Error() string
	StackTrace() []byte
}
