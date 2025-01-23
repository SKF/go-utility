package awstrace

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	oc_trace "go.opencensus.io/trace"
	dd_mocktracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-utility/v3/array"
	"github.com/SKF/go-utility/v3/trace"
)

func Test_Inject_Basic(t *testing.T) {
	ctx := context.Background()

	input1 := injectSNSPublish(ctx, nil)
	assert.Nil(t, input1)

	input1 = injectSNSPublish(ctx, &sns.PublishInput{})
	assert.Len(t, input1.MessageAttributes, 0)

	input2 := injectSQSSendMessage(ctx, nil)
	assert.Nil(t, input2)

	input2 = injectSQSSendMessage(ctx, &sqs.SendMessageInput{})
	assert.Len(t, input2.MessageAttributes, 0)

	input3 := injectSQSSendMessageBatch(ctx, nil)
	assert.Nil(t, input3)

	input3 = injectSQSSendMessageBatch(ctx, &sqs.SendMessageBatchInput{})
	assert.Len(t, input3.Entries, 0)
}

func Test_InjectDatadog_HappyCase(t *testing.T) {
	var message = "test message"

	tracer, ctx := startDatadogSpan()
	defer tracer.Stop()

	input := injectSNSPublish(ctx, &sns.PublishInput{
		Message: aws.String(message),
	})
	assert.Equal(t, message, *input.Message)
	assert.Len(t, input.MessageAttributes, 2)

	attributesKeys := []string{}
	for key := range input.MessageAttributes {
		attributesKeys = append(attributesKeys, key)
	}

	assert.True(t, array.ContainsString(attributesKeys, trace.DatadogTraceIDHeader))
	assert.True(t, array.ContainsString(attributesKeys, trace.DatadogParentIDHeader))
}

func Test_InjectOC_HappyCase(t *testing.T) {
	var message = "test message"

	ctx := startOCSpan()
	input := injectSQSSendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(message),
	})
	assert.Equal(t, message, *input.MessageBody)
	assert.Len(t, input.MessageAttributes, 2)

	attributesKeys := []string{}
	for key := range input.MessageAttributes {
		attributesKeys = append(attributesKeys, key)
	}

	assert.True(t, array.ContainsString(attributesKeys, trace.DatadogTraceIDHeader))
	assert.True(t, array.ContainsString(attributesKeys, trace.DatadogParentIDHeader))
}

func startOCSpan() context.Context {
	ctx := context.Background()
	ctx, _ = oc_trace.StartSpan(ctx, "test")

	return ctx
}

func Test_getTraceAttributesFromContextB3_HappyCase(t *testing.T) {
	ctx := startOCSpan()
	attributes := getTraceAttributesFromContext(ctx)
	attributesKeys := []string{}

	for key := range attributes {
		attributesKeys = append(attributesKeys, key)
	}

	require.True(t, array.ContainsString(attributesKeys, trace.DatadogTraceIDHeader))
	assert.NotEmpty(t, attributes[trace.DatadogTraceIDHeader])

	require.True(t, array.ContainsString(attributesKeys, trace.DatadogParentIDHeader))
	assert.NotEmpty(t, attributes[trace.DatadogParentIDHeader])
}

func startDatadogSpan() (dd_mocktracer.Tracer, context.Context) {
	mt := dd_mocktracer.Start()
	ctx := context.Background()
	_, ctx = dd_tracer.StartSpanFromContext(ctx, "test")

	return mt, ctx
}

func Test_getTraceAttributesFromContextDatadog_HappyCase(t *testing.T) {
	tracer, ctx := startDatadogSpan()
	defer tracer.Stop()

	attributes := getTraceAttributesFromContext(ctx)
	attributesKeys := []string{}

	for key := range attributes {
		attributesKeys = append(attributesKeys, key)
	}

	require.True(t, array.ContainsString(attributesKeys, trace.DatadogTraceIDHeader))
	assert.NotEmpty(t, attributes[trace.DatadogTraceIDHeader])
	assert.NotEqual(t, "0", attributes[trace.DatadogTraceIDHeader])

	require.True(t, array.ContainsString(attributesKeys, trace.DatadogParentIDHeader))
	assert.NotEmpty(t, attributes[trace.DatadogParentIDHeader])
	assert.NotEqual(t, "0", attributes[trace.DatadogParentIDHeader])
}
