package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
)

func Test_Lambda_StartFromSQS(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	host, err := url.Parse(s.URL)
	require.NoError(t, err)

	t.Setenv("DD_TRACE_STARTUP_LOGS", "false")
	t.Setenv("DD_AGENT_HOST", host.Hostname())
	t.Setenv("DD_TRACE_AGENT_PORT", host.Port())
	t.Setenv("DD_PROPAGATION_STYLE_EXTRACT", "DataDog")

	tracer.Start()

	var (
		traceID          = strconv.FormatUint(1, 10)
		parentID         = strconv.FormatUint(2, 10)
		samplingPriority = "1"
		tags             = "_dd.p.dm=-1"
	)

	event := events.SQSMessage{
		MessageAttributes: map[string]events.SQSMessageAttribute{
			tracer.DefaultTraceIDHeader: {
				DataType:    "String",
				StringValue: &traceID,
			},
			tracer.DefaultParentIDHeader: {
				DataType:    "String",
				StringValue: &parentID,
			},
			tracer.DefaultPriorityHeader: {
				DataType:    "String",
				StringValue: &samplingPriority,
			},
			"x-datadog-tags": {
				DataType:    "String",
				StringValue: &tags,
			},
		},
	}

	span, ctx := (trace.SQSMessageCarrier)(event).StartSpan(context.TODO(), "operation")

	_, ok := tracer.SpanFromContext(ctx)
	assert.True(t, ok)

	spanContext := span.Context()

	assert.Equal(t, spanContext.TraceID(), uint64(1))
}
