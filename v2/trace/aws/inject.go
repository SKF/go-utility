package awstrace

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	oc_trace "go.opencensus.io/trace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func injectSNSPublish(ctx context.Context, input *sns.PublishInput) *sns.PublishInput {
	if input == nil {
		return nil
	}

	if input.MessageAttributes == nil {
		input.MessageAttributes = map[string]*sns.MessageAttributeValue{}
	}

	for key, value := range getTraceAttributesFromContext(ctx) {
		input.MessageAttributes[key] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}

	return input
}

func injectSQSSendMessage(ctx context.Context, input *sqs.SendMessageInput) *sqs.SendMessageInput {
	if input == nil {
		return nil
	}

	input.MessageAttributes = extendMessageAttributes(input.MessageAttributes, getTraceAttributesFromContext(ctx))

	return input
}

func injectSQSSendMessageBatch(ctx context.Context, input *sqs.SendMessageBatchInput) *sqs.SendMessageBatchInput {
	if input == nil || input.Entries == nil {
		return input
	}

	for _, el := range input.Entries {
		el.MessageAttributes = extendMessageAttributes(el.MessageAttributes, getTraceAttributesFromContext(ctx))
	}

	return input
}

func extendMessageAttributes(attributes map[string]*sqs.MessageAttributeValue, data map[string]string) map[string]*sqs.MessageAttributeValue {
	if attributes == nil {
		attributes = map[string]*sqs.MessageAttributeValue{}
	}

	for key, value := range data {
		attributes[key] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}

	return attributes
}

func getTraceAttributesFromContext(ctx context.Context) map[string]string {
	attributes := map[string]string{}

	if span := oc_trace.FromContext(ctx); span != nil {
		spanCtx := span.SpanContext()

		attributes[b3TraceHeader] = hex.EncodeToString(spanCtx.TraceID[:])
		attributes[b3SpanHeader] = hex.EncodeToString(spanCtx.SpanID[:])
		attributes[datadogTraceHeader] = strconv.FormatUint(binary.BigEndian.Uint64(spanCtx.TraceID[8:]), 10)
		attributes[datadogParentHeader] = strconv.FormatUint(binary.BigEndian.Uint64(spanCtx.SpanID[:]), 10)
	}

	if span, exists := dd_tracer.SpanFromContext(ctx); exists {
		traceID := strconv.FormatUint(span.Context().TraceID(), 10)
		spanID := strconv.FormatUint(span.Context().SpanID(), 10)

		attributes[datadogTraceHeader] = traceID
		attributes[datadogParentHeader] = spanID
	}

	return attributes
}
