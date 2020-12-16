package trace_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SKF/go-utility/v2/array"
	"github.com/SKF/go-utility/v2/trace"
)

const (
	noOfB3Headers      = 3
	noOfDatadogHeaders = 5
)

func Test_AllHeaders(t *testing.T) {
	const expectedLength = noOfB3Headers + noOfDatadogHeaders

	actual := array.DistinctString(trace.AllHeaders())
	assert.Len(t, actual, expectedLength)
}

func Test_AllB3Headers(t *testing.T) {
	actual := array.DistinctString(trace.AllB3Headers())
	assert.Len(t, actual, noOfB3Headers)
}

func Test_AllDatadogHeaders(t *testing.T) {
	actual := array.DistinctString(trace.AllDatadogHeaders())
	assert.Len(t, actual, noOfDatadogHeaders)
}
