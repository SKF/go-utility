package trace_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SKF/go-utility/v2/array"
	"github.com/SKF/go-utility/v2/trace"
)

func Test_AllHeaders(t *testing.T) {
	const expectedLength = 3 + 5

	actual := array.DistinctString(trace.AllHeaders())
	assert.Len(t, actual, expectedLength)
}

func Test_AllB3Headers(t *testing.T) {
	const expectedLength = 3

	actual := array.DistinctString(trace.AllB3Headers())
	assert.Len(t, actual, expectedLength)
}

func Test_AllDatadogHeaders(t *testing.T) {
	const expectedLength = 5

	actual := array.DistinctString(trace.AllDatadogHeaders())
	assert.Len(t, actual, expectedLength)
}
