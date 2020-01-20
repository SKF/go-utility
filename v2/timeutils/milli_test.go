package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	millisecondsFor20180306 = 1528030261000
	millisecondsFor22861120 = 9999999999999
)

func Test_MillisecondsShouldBeWithinRealisticInterval(t *testing.T) {
	assert.True(t, MillisecondsNow() > millisecondsFor20180306)
	assert.True(t, MillisecondsNow() < millisecondsFor22861120)
}

func Test_MillisecondsUnix(t *testing.T) {
	now := time.Now()
	assert.Equal(t, MillisecondsUnix(now.Add(time.Second)), MillisecondsUnix(now)+1000) // nolint: gomnd
}

func Test_MillisecondsConversion(t *testing.T) {
	ms := MillisecondsNow()
	assert.Equal(t, ms, MillisecondsUnix(MillisecondsTime(ms)))
}

func Test_AssertMilliseconds(t *testing.T) {
	var ms int64 = 1550837382666
	result, err := AssertMilliseconds(ms)
	assert.Equal(t, ms, result)
	assert.NoError(t, err)
}

func Test_SecondsShouldBeConvertedIntoMilliseconds(t *testing.T) {
	var seconds int64 = 1550837382
	result, err := AssertMilliseconds(seconds)
	assert.Equal(t, seconds*int64(time.Second/time.Millisecond), result)
	assert.Error(t, err)
}

func Test_MicrosecondsShouldBeConvertedIntoMilliseconds(t *testing.T) {
	var microseconds int64 = 1550837382666000
	result, err := AssertMilliseconds(microseconds)
	assert.Equal(t, microseconds/int64(time.Millisecond/time.Microsecond), result)
	assert.Error(t, err)
}

func Test_NanosecondsShouldBeConvertedIntoMilliseconds(t *testing.T) {
	var nanoseconds int64 = 1550837382666000456
	result, err := AssertMilliseconds(nanoseconds)
	assert.Equal(t, nanoseconds/int64(time.Millisecond/time.Nanosecond), result)
	assert.Error(t, err)
}
