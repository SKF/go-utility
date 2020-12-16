package ddpgx

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
)

func Test_EscapeString(t *testing.T) {
	tests := map[string]interface{}{
		" \t ":                                  " \t ",
		" \t \n":                                " ",
		"base case":                             "base case",
		"newline\r\ntest\n":                     "newline test ",
		"newline\n   with \t   \n   whitespace": "newline with whitespace",
	}

	for input, expected := range tests {
		actual, ok := escapeValue(input).(string)
		require.True(t, ok)
		assert.Equal(t, expected, actual)
	}
}
