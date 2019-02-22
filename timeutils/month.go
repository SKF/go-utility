package timeutils

import (
	"fmt"
	"strconv"
	"time"
)

const earliestYear = 1801
const latestYear = 2999

// GetPeriodsStartAndEndUTC returns the start and end timestamps (in milliseconds) given two months for location UTC (timezone)
func GetPeriodsStartAndEndUTC(firstyyyymm string, lastyyyymm string) (start int64, end int64, err error) {
	start, _, err = getMonthStartAndEnd(firstyyyymm, *time.UTC)
	if err == nil {
		_, end, err = getMonthStartAndEnd(lastyyyymm, *time.UTC)
	}

	if start > end {
		start = 0
		end = 0
		err = fmt.Errorf("start %s may not be after end %s", firstyyyymm, lastyyyymm)
	}

	return
}

// getMonthStartAndEnd returns timstamps (in milliseconds) for the month's start and end for the given location
func getMonthStartAndEnd(yyyymm string, location time.Location) (start int64, end int64, err error) {
	syyyy, smm, err := validFormat(yyyymm)
	if err != nil {
		return 0, 0, err
	}

	yyyy, err := strconv.Atoi(syyyy)
	if err != nil {
		return 0, 0, err
	}

	mm, err := strconv.Atoi(smm)
	if err != nil {
		return 0, 0, err
	}

	tt := time.Date(yyyy, time.Month(mm), 1, 0, 0, 0, 0, &location)
	start = MillisecondsUnix(tt)

	tt = tt.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	end = MillisecondsUnix(tt)

	return start, end, nil
}

// validFormat returns err if yyyymm string format invalid - otherwise returns yyyy and mm
func validFormat(yyyymm string) (validatedyyyy string, validatedmm string, err error) {
	// default to current month
	if yyyymm == "" {
		now := time.Now()
		yyyymm = now.Format("200601")
	}

	if len(yyyymm) != 6 {
		err = fmt.Errorf("yyyymm must be 6 characters long - %s", yyyymm)
		return
	}

	if _, err = strconv.ParseInt(yyyymm, 10, 64); err != nil {
		err = fmt.Errorf("yyyymm must be integer - %s", yyyymm)
		return
	}

	syyyy := yyyymm[0:4]
	yyyy, err := strconv.ParseInt(syyyy, 10, 64)
	if err != nil {
		err = fmt.Errorf("yyyy cannot be parsed - %s", syyyy)
		return
	}

	smm := yyyymm[4:]
	mm, err := strconv.ParseInt(smm, 10, 64)
	if err != nil {
		err = fmt.Errorf("mm cannot be parsed - %s", smm)
		return
	}

	if yyyy < earliestYear || yyyy > latestYear {
		err = fmt.Errorf("yyyy must be between 1801 and 2999")
		return
	}

	if mm < 1 || mm > 12 {
		err = fmt.Errorf("mm must be between 01 and 12")
		return
	}

	return syyyy, smm, err
}
