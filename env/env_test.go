package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MustGetAsString_Ok(t *testing.T) {
	err := os.Setenv("Node_Table", "NodeTable")
	assert.Nil(t, err)

	assert.Equal(t, "NodeTable", MustGetAsString("Node_Table"))
}

func Test_MustGetAsString_Panic(t *testing.T) {
	err := os.Unsetenv("Node_Table")
	assert.Nil(t, err)

	assert.Panics(t, func() { MustGetAsString("Node_Table") })
}

func Test_GetAsString_Default(t *testing.T) {
	err := os.Unsetenv("notexist")
	assert.Nil(t, err)

	assert.Equal(t,
		"thisisthedefault",
		GetAsString("notexist", "thisisthedefault"),
	)
}

func Test_GetAsString_Env(t *testing.T) {
	err := os.Setenv("notexist", "thisisnotthedefault")
	assert.Nil(t, err)

	assert.Equal(t,
		"thisisnotthedefault",
		GetAsString("notexist", "thisisthedefault"),
	)
}

func Test_GetFloat_Env(t *testing.T) {
	err := os.Setenv("PI", "3.14159")
	assert.Nil(t, err)

	assert.Equal(t, 3.14159, GetAsFloat("PI", 3.0))
}

func Test_GetFloat_Default(t *testing.T) {
	err := os.Unsetenv("PI")
	assert.Nil(t, err)

	assert.Equal(t, 3.0, GetAsFloat("PI", 3.0))
}

func Test_GetFloat_NumberFormatFailure(t *testing.T) {
	err := os.Setenv("PI", "3.1A159")
	assert.Nil(t, err)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	assert.Equal(t, 3.0, GetAsFloat("PI", 3.0))
}

func Test_GetInt_Env(t *testing.T) {
	err := os.Setenv("INT", "3")
	assert.Nil(t, err)

	assert.Equal(t, 3, GetAsInt("INT", 11))
}

func Test_GetInt_Default(t *testing.T) {
	err := os.Unsetenv("INT")
	assert.Nil(t, err)

	assert.Equal(t, 42, GetAsInt("INT", 42))
}

func Test_GetInt_NumberFormatFailure(t *testing.T) {
	err := os.Setenv("INT", "3A")
	assert.Nil(t, err)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	assert.Equal(t, 11, GetAsInt("INT", 11))
}
