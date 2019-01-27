package array

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DistinctString(t *testing.T) {
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

func Test_IntersectString(t *testing.T) {
	var testData = []struct {
		input1   []string
		input2   [][]string
		expected []string
	}{
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{"2", "3", "4"}},
			[]string{"2", "3"},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{"2", "3", "1"}, {"1", "2"}},
			[]string{"1", "2"},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{},
			[]string{"1", "2", "3"},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{}, {}},
			[]string{},
		},
	}

	for _, data := range testData {
		output := IntersectString(data.input1, data.input2...)
		sort.Strings(output)
		assert.EqualValues(t, data.expected, output, fmt.Sprintf("Intersect input1: %+v, input2: %+v", data.input1, data.input2))
	}
}

func Test_DifferenceString(t *testing.T) {
	var testData = []struct {
		input1   []string
		input2   [][]string
		expected []string
	}{
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{"2", "3", "4"}},
			[]string{"1"},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{"2", "3", "1"}, {"1", "2"}},
			[]string{},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{},
			[]string{"1", "2", "3"},
		},
		{
			[]string{"1", "1", "2", "3"},
			[][]string{{}},
			[]string{"1", "2", "3"},
		},
	}

	for _, data := range testData {
		output := DifferenceString(data.input1, data.input2...)
		sort.Strings(output)
		assert.EqualValues(t, data.expected, output, fmt.Sprintf("Difference input1: %+v, input2: %+v", data.input1, data.input2))
	}
}
