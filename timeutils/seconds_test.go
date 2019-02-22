package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AssertSeconds(t *testing.T) {
	var seconds int64 = 1550837382
	result, err := AssertSeconds(seconds)
	assert.Equal(t, seconds, result)
	assert.NoError(t, err)
}

func Test_MillisecondsShouldBeConvertedToSeconds(t *testing.T) {
	var ms int64 = 1550837382666
	result, err := AssertSeconds(ms)
	assert.Equal(t, ms/int64(time.Second/time.Millisecond), result)
	assert.Error(t, err)
}

func Test_MicrosecondsShouldBeConvertedToSeconds(t *testing.T) {
	var microseconds int64 = 1550837382666000
	result, err := AssertSeconds(microseconds)
	assert.Equal(t, microseconds/int64(time.Second/time.Microsecond), result)
	assert.Error(t, err)
}

func Test_NanosecondsShouldBeConvertedToSeconds(t *testing.T) {
	var nanoseconds int64 = 1550837382666000123
	result, err := AssertSeconds(nanoseconds)
	assert.Equal(t, nanoseconds/int64(time.Second/time.Nanosecond), result)
	assert.Error(t, err)
}
