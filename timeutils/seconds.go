package timeutils

import (
	"errors"
)

func AssertSeconds(timestamp int64) (timestampSeconds int64, err error) {
	var timestampYear3000Seconds int64 = 32503680000
	for timestamp > timestampYear3000Seconds {
		// Make sure timestamp is not in nanosecond, microsecond or millisecond format
		timestamp /= 1000
		err = errors.New("got timestamp that was not in seconds, had to convert")
	}

	if timestamp < 0 {
		err = errors.New("got a timestamp that's before 1970, this is probably bad")
	}

	timestampSeconds = timestamp
	return timestampSeconds, err
}
