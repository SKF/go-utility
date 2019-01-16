package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Errors_Fail(t *testing.T) {
	err := New("Test new with args '%s'", "msg")
	assert.Equal(t, err.Error(), "Test new with args 'msg'")
}
