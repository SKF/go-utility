package timeutils

import (
	"errors"
	"time"
)

// Milliseconds returns milliseconds for time now
func Milliseconds() int64 {
	return MillisecondsTime(time.Now())
}

// MillisecondsTime returns milliseconds for given time
func MillisecondsTime(t time.Time) int64 {
	return (t.UnixNano()) / time.Millisecond.Nanoseconds()
}

// MillisecondsToTime returns a Time struct for a given millisecond timestamp
func MillisecondsToTime(milli int64) time.Time {
	return time.Unix(0, milli*1000000)
}

func AssertMilliseconds(timestamp int64) (timestampMilliseconds int64, err error) {
	var timestampYear3000Milliseconds int64 = 32503680000000
	for timestamp > timestampYear3000Milliseconds {
		// Make sure timestamp is not in nanosecond or microsecond format
		timestamp /= 1000
		err = errors.New("got timestamp that was not milliseconds, had to convert")
	}

	var timestampYear3000Seconds int64 = 32503680000
	if timestamp >= 1 && timestamp < timestampYear3000Seconds {
		err = errors.New("got timestamp was in seconds (not milliseconds), had to convert")
		timestamp *= 1000
	}

	if timestamp < 0 {
		err = errors.New("got a timestamp that's before 1970, this is probably bad")
	}

	timestampMilliseconds = timestamp
	return timestampMilliseconds, err
}
