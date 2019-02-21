package timeutils

import (
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func Test_MillisecondsShouldBeWithinRealisticInterval(t *testing.T) {
	assert.True(t, Milliseconds() > 1528030261000) // 2018-03-06
	assert.True(t, Milliseconds() < 9999999999999) // 2286-11-20
}

func Test_MillisecondsTime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, MillisecondsTime(now.Add(time.Second)), MillisecondsTime(now)+1000)
}
