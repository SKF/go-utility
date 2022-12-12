package aws_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
)

func Test_Injection_PublishInput(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	host, err := url.Parse(s.URL)
	require.NoError(t, err)

	t.Setenv("DD_AGENT_HOST", host.Hostname())
	t.Setenv("DD_TRACE_AGENT_PORT", host.Port())
	t.Setenv("DD_PROPAGATION_STYLE_INJECT", "DataDog,B3")

	tracer.Start()

	input := &sns.PublishInput{}

	span := tracer.StartSpan("testcase")

	err = tracer.Inject(span.Context(), (*trace.PublishInputCarrier)(input))
	require.NoError(t, err)

	messageAttributes := input.MessageAttributes
	require.NotZero(t, messageAttributes)

	assert.Contains(t, messageAttributes, tracer.DefaultTraceIDHeader)
	assert.Contains(t, messageAttributes, tracer.DefaultParentIDHeader)
	assert.Contains(t, messageAttributes, tracer.DefaultPriorityHeader)
	assert.Contains(t, messageAttributes, b3TraceIDHeader)
	assert.Contains(t, messageAttributes, b3SpanIDHeader)
	assert.Contains(t, messageAttributes, b3SampledHeader)
}

func Test_Injection_PublishBatchInput(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	host, err := url.Parse(s.URL)
	require.NoError(t, err)

	t.Setenv("DD_AGENT_HOST", host.Hostname())
	t.Setenv("DD_TRACE_AGENT_PORT", host.Port())
	t.Setenv("DD_PROPAGATION_STYLE_INJECT", "DataDog,B3")

	tracer.Start()

	input := &sns.PublishBatchInput{
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{},
		},
	}

	span := tracer.StartSpan("testcase")

	err = tracer.Inject(span.Context(), (*trace.PublishBatchInputCarrier)(input))
	require.NoError(t, err)

	require.Len(t, input.PublishBatchRequestEntries, 1)

	messageAttributes := input.PublishBatchRequestEntries[0].MessageAttributes
	require.NotZero(t, messageAttributes)

	assert.Contains(t, messageAttributes, tracer.DefaultTraceIDHeader)
	assert.Contains(t, messageAttributes, tracer.DefaultParentIDHeader)
	assert.Contains(t, messageAttributes, tracer.DefaultPriorityHeader)
	assert.Contains(t, messageAttributes, b3TraceIDHeader)
	assert.Contains(t, messageAttributes, b3SpanIDHeader)
	assert.Contains(t, messageAttributes, b3SampledHeader)
}
