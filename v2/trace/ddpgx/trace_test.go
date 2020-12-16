package ddpgx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EscapeString(t *testing.T) {
	testsStrings := map[string]string{
		" \t ":                                  "",
		" \t \n":                                "",
		"base case":                             "base case",
		"newline\r\ntest\n":                     "newline test",
		"newline\n   with \t   \n   whitespace": "newline with whitespace",
		"\t trim spaces!\n ":                    "trim spaces!",
	}

	for input, expected := range testsStrings {
		actual, ok := escapeValue(input).(string)
		require.True(t, ok)
		assert.Equal(t, expected, actual)
	}

	testInputInt := 42
	actual, ok := escapeValue(testInputInt).(int)
	require.True(t, ok)
	assert.Equal(t, testInputInt, actual)
}
