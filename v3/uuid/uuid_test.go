package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestUUID_StringPtr(t *testing.T) {
	id := New()
	assert.Equal(t, id.String(), *id.StringPtr())
}

func Test_StringList(t *testing.T) {
	assert.Nil(t, StringList())

	id := New()
	id2 := New()
	list := StringList(id, id2)
	require.Len(t, list, 2)
	assert.Equal(t, id.String(), list[0])
	assert.Equal(t, id2.String(), list[1])
}
