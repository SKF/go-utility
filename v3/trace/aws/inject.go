package awstrace

import (
	"context"
	"encoding/binary"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	sns_types "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqs_types "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	oc_trace "go.opencensus.io/trace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-utility/v2/trace"
)

func injectSNSPublish(ctx context.Context, input *sns.PublishInput) *sns.PublishInput {
	if input == nil {
		return nil
	}

	if input.MessageAttributes == nil {
		input.MessageAttributes = map[string]sns_types.MessageAttributeValue{}
	}

	for key, value := range getTraceAttributesFromContext(ctx) {
		input.MessageAttributes[key] = sns_types.MessageAttributeValue{
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

func extendMessageAttributes(attributes map[string]sqs_types.MessageAttributeValue, data map[string]string) map[string]sqs_types.MessageAttributeValue {
	if attributes == nil {
		attributes = map[string]sqs_types.MessageAttributeValue{}
	}

	for key, value := range data {
		attributes[key] = sqs_types.MessageAttributeValue{
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

		attributes[trace.DatadogTraceIDHeader] = strconv.FormatUint(binary.BigEndian.Uint64(spanCtx.TraceID[8:]), 10) //nolint: gomnd
		attributes[trace.DatadogParentIDHeader] = strconv.FormatUint(binary.BigEndian.Uint64(spanCtx.SpanID[:]), 10)  //nolint: gomnd
	}

	if span, exists := dd_tracer.SpanFromContext(ctx); exists {
		traceID := strconv.FormatUint(span.Context().TraceID(), 10) //nolint: gomnd
		spanID := strconv.FormatUint(span.Context().SpanID(), 10)   //nolint: gomnd

		attributes[trace.DatadogTraceIDHeader] = traceID
		attributes[trace.DatadogParentIDHeader] = spanID
	}

	return attributes
}
