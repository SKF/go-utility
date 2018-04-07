package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New_IsValid(t *testing.T) {
	assert.True(t, New().IsValid())
}

func Test_IsValid(t *testing.T) {
	assert.True(t, IsValid("9f220f42-2fa7-46f9-8a25-f3ba17328a13"))
	assert.True(t, IsValid(EmptyUUID.String()))
}

func Test_IsValid_CaseInsensitive(t *testing.T) {
	assert.True(t, IsValid("9f220f42-2fa7-46F9-8a25-f3BA17328A13"))
}

func Test_IsValid_EmptyString(t *testing.T) {
	assert.False(t, IsValid(""))
}

func Test_IsValid_TooLongUUID(t *testing.T) {
	assert.False(t, IsValid("9f220f42-2fa7-46f9-8a25-f3ba17328a13-XX"))
	assert.False(t, IsValid("G041695E-930c-11e7-abc4-cec278b6b50a"))
}
