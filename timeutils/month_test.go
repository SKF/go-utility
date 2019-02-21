package timeutils

import (
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func Test_Month_HappyPath(t *testing.T) {
	start, end, err := getMonthStartAndEnd("201805", *time.UTC)
	assert.Nil(t, err)
	assert.Equal(t, int64(1525132800000), start)
	assert.Equal(t, int64(1527811199999), end)
}

func Test_Month_Default(t *testing.T) {
	_, _, err := getMonthStartAndEnd("", *time.UTC)
	assert.Nil(t, err)
	// values for 201805:
	//	assert.Equal(t, int64(1525132800000), start)
	//	assert.Equal(t, int64(1527811199999), end)
}

func Test_NotNumeric(t *testing.T) {
	_, _, err := getMonthStartAndEnd("abcdef", *time.UTC)
	assert.NotNil(t, err)
}

func Test_OutOfRange(t *testing.T) {
	_, _, err := getMonthStartAndEnd("987613", *time.UTC)
	assert.NotNil(t, err)
	_, _, err = getMonthStartAndEnd("299913", *time.UTC)
	assert.NotNil(t, err)
	_, _, err = getMonthStartAndEnd("201813", *time.UTC)
	assert.NotNil(t, err)
	_, _, err = getMonthStartAndEnd("201800", *time.UTC)
	assert.NotNil(t, err)
}

func Test_WrongLength(t *testing.T) {
	_, _, err := getMonthStartAndEnd(" ", *time.UTC)
	assert.NotNil(t, err)

	_, _, err = getMonthStartAndEnd("2018", *time.UTC)
	assert.NotNil(t, err)

	_, _, err = getMonthStartAndEnd("20181", *time.UTC)
	assert.NotNil(t, err)

	_, _, err = getMonthStartAndEnd("2018 12", *time.UTC)
	assert.NotNil(t, err)
}

func Test_Periods_HappyPath(t *testing.T) {
	start, end, err := GetPeriodsStartAndEndUTC("201801", "201809")
	assert.Nil(t, err)
	assert.Equal(t, int64(1514764800000), start)
	assert.Equal(t, int64(1538351999999), end)

	start, end, err = GetPeriodsStartAndEndUTC("201809", "201809") // same month
	assert.Nil(t, err)
	assert.Equal(t, int64(1535760000000), start)
	assert.Equal(t, int64(1538351999999), end)
}

func Test_PeriodUTC_HappyPath(t *testing.T) {
	start, end, err := getPeriodsStartAndEnd("201801", "201809", *time.UTC)
	assert.Nil(t, err)
	assert.Equal(t, int64(1514764800000), start)
	assert.Equal(t, int64(1538351999999), end)
}

func Test_PeriodWrongOrder(t *testing.T) {
	_, _, err := GetPeriodsStartAndEndUTC("201812", "201809")
	assert.NotNil(t, err)
}

func Test_validFormat(t *testing.T) {
	_, _, err := validFormat("123400")
	assert.NotNil(t, err)

	_, _, err = validFormat("20171")
	assert.NotNil(t, err)

	yyyy, mm, err := validFormat("201712")
	assert.Nil(t, err)
	assert.Equal(t, "2017", yyyy)
	assert.Equal(t, "12", mm)
}
