package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getMonthStartAndEnd(t *testing.T) {
	start, end, err := GetPeriodsStartAndEndUTC("201805", "201805")
	assert.NoError(t, err)
	assert.Equal(t, int64(1525132800000), start)
	assert.Equal(t, int64(1527811199999), end)
}

func Test_NotNumericStart(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC("abcdef", "201805")
	assert.Error(t, err)
}

func Test_NotNumericEnd(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC("201805", "abcdef")
	assert.Error(t, err)
}

func Test_OutOfRange(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC("987613", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("299913", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("201813", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("201800", "201805")
	assert.Error(t, err)
}

func Test_WrongLength(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC(" ", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("2018", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("20181", "201805")
	assert.Error(t, err)

	_, _, err = GetPeriodsStartAndEndUTC("2018 12", "201805")
	assert.Error(t, err)
}

func Test_GetPeriodsStartAndEndUTC(t *testing.T) {
	start, end, err := GetPeriodsStartAndEndUTC("201801", "201809")
	assert.NoError(t, err)
	assert.Equal(t, int64(1514764800000), start)
	assert.Equal(t, int64(1538351999999), end)
}

func Test_PeriodUTC_SameMonth(t *testing.T) {
	start, end, err := GetPeriodsStartAndEndUTC("201809", "201809")
	assert.NoError(t, err)
	assert.Equal(t, int64(1535760000000), start)
	assert.Equal(t, int64(1538351999999), end)
}

func Test_PeriodWrongOrder(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC("201812", "201809")
	assert.Error(t, err)
}

func Test_toTime(t *testing.T) {
	input := "201411"
	result, err := toTime(input)
	expected := time.Date(2014, 11, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
	assert.NoError(t, err)
}

func Test_toTime_InvalidInput(t *testing.T) {
	input := "20141x"
	_, err := toTime(input)
	assert.Error(t, err)
}

func Test_lastDayOfMonth(t *testing.T) {
	input := time.Date(2014, 11, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2014, 11, 30, 23, 59, 59, 999999999, time.UTC)
	result := lastDayOfMonth(input)
	assert.Equal(t, expected, result)
}
