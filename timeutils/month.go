package timeutils

import (
	"fmt"
	"time"
)

const earliestYear = 1801
const latestYear = 2999

// GetPeriodsStartAndEndUTC returns the start and end timestamps (in milliseconds) given two months for location UTC (timezone)
func GetPeriodsStartAndEndUTC(firstyyyymm string, lastyyyymm string) (int64, int64, error) {
	start, err := toTime(firstyyyymm)
	if err != nil {
		return 0, 0, err
	}

	end, err := toTime(lastyyyymm)
	end = lastDayOfMonth(end)
	if err != nil {
		return 0, 0, err
	}

	if start.After(end) {
		err = fmt.Errorf("start %s may not be after end %s", firstyyyymm, lastyyyymm)
		return 0, 0, err
	}

	return MillisecondsUnix(start), MillisecondsUnix(end), nil
}

func toTime(input string) (time.Time, error) {
	layout := "200601"
	return time.ParseInLocation(layout, input, time.UTC)
}

func lastDayOfMonth(input time.Time) time.Time {
	return input.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
}
