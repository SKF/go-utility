package timeutils

import (
	"time"

	"github.com/pkg/errors"
)

// MillisecondsNow returns milliseconds for time now
func MillisecondsNow() int64 {
	return MillisecondsUnix(time.Now())
}

// MillisecondsUnix returns milliseconds for given time
func MillisecondsUnix(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// MillisecondsTime returns a Time struct for a given millisecond timestamp
func MillisecondsTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// AssertMilliseconds ensures that a given timestamp is in milliseconds.
// If not, a timestamp converted into milliseconds as well as an error will be returned.
func AssertMilliseconds(ms int64) (timestampMilliseconds int64, err error) {
	if ms == 0 {
		return ms, errors.New("got a timestamp that's before 1970, this is probably bad")
	}

	var timestampYear3000Seconds int64 = 32503680000
	if ms < timestampYear3000Seconds {
		ms *= 1000
		err = errors.New("got timestamp was in seconds (not milliseconds), had to convert")

		return ms, err
	}

	var timestampYear3000Milliseconds int64 = 32503680000000
	for ms > timestampYear3000Milliseconds {
		// Make sure timestamp is not in nanosecond or microsecond format
		ms /= 1000
		err = errors.New("got timestamp that was not milliseconds, had to convert")
	}

	return ms, err
}
