package array

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Intersect(t *testing.T) {
	var testData = []struct {
		input    [][]string
		expected []string
	}{
		{
			[][]string{{"1", "1", "2", "3"}, {"2", "3", "4"}},
			[]string{"2", "3"},
		},
		{
			[][]string{{"1", "1", "2", "3"}, {"2", "3", "1"}, {"1", "2"}},
			[]string{"1", "2"},
		},
		{
			[][]string{{"1", "1", "2", "3"}},
			[]string{"1", "2", "3"},
		},
		{
			[][]string{{"1", "1", "2", "3"}, {}},
			[]string{},
		},
	}

	for _, data := range testData {
		output := IntersectString(data.input...)
		sort.Strings(output)
		assert.EqualValues(t, data.expected, output, fmt.Sprintf("Intersect input: %+v", data.input))
	}
}

func Test_Distinct(t *testing.T) {
	var testData = []struct {
		input    []string
		expected []string
	}{
		{
			[]string{"1", "1", "2", "3"},
			[]string{"1", "2", "3"},
		},
		{
			[]string{"a", "b", "A", "B", "A", "b"},
			[]string{"A", "B", "a", "b"},
		},
	}

	for _, data := range testData {
		output := DistinctString(data.input)
		sort.Strings(output)
		assert.EqualValues(t, data.expected, output, fmt.Sprintf("Distinct input: %+v", data.input))
	}
}

func Test_Difference(t *testing.T) {
	var testData = []struct {
		input    [][]string
		expected []string
	}{
		{
			[][]string{{"1", "1", "2", "3"}, {"2", "3", "4"}},
			[]string{"1", "4"},
		},
		{
			[][]string{{"1", "1", "2", "3"}, {"2", "3", "1"}, {"1", "2"}},
			[]string{},
		},
		{
			[][]string{{"1", "1", "2", "3"}},
			[]string{"1", "2", "3"},
		},
		{
			[][]string{{"1", "1", "2", "3"}, {}},
			[]string{"1", "2", "3"},
		},
	}

	for _, data := range testData {
		output := DifferenceString(data.input...)
		sort.Strings(output)
		assert.EqualValues(t, data.expected, output, fmt.Sprintf("Difference input: %+v", data.input))
	}
}
