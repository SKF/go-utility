package awstrace

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	dd_trace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-utility/v2/array"
	"github.com/SKF/go-utility/v2/trace"
)

type messageAttribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type snsEntityBody struct {
	MessageAttributes map[string]messageAttribute `json:"MessageAttributes"`
}

var propagator = dd_tracer.NewPropagator(nil)

// StartDatadogSpanFromMessage will start a Datadog span based on message attributes found in sqs message.
// It will extract trace and span id either depending on the extraction style.
// Configure the extraction style by using the environment variable: DD_PROPAGATION_STYLE_EXTRACT=Datadog,B3
// Default extraction style is set to Datadog
// For this to work for SNS subscriptions, you need to enable raw message delivery.
func StartDatadogSpanFromMessage(ctx context.Context, serviceName string, msg events.SQSMessage) (dd_tracer.Span, context.Context) {
	spanContext, err := getRecordSpanContext(ctx, msg)
	if err != nil {
		return startSpan(ctx, serviceName, nil)
	}

	return startSpan(ctx, serviceName, spanContext)
}

func getRecordSpanContext(ctx context.Context, msg events.SQSMessage) (dd_trace.SpanContext, error) {
	traceHeaders := getTraceHeadersFromAttributes(ctx, msg)
	if len(traceHeaders) == 0 {
		return nil, errors.New("no trace headers")
	}

	recordSpanContext, err := propagator.Extract(dd_tracer.TextMapCarrier(traceHeaders))
	if err != nil {
		return nil, err
	}

	return recordSpanContext, nil
}

func getTraceHeadersFromAttributes(_ context.Context, msg events.SQSMessage) map[string]string {
	traceAttributes := map[string]string{}
	allTraceHeaders := trace.AllHeaders()

	// Get SQS message attributes
	for key, attr := range msg.MessageAttributes {
		dataType := strings.ToLower(attr.DataType)
		if array.ContainsString(allTraceHeaders, key) && dataType == "string" {
			traceAttributes[key] = *attr.StringValue
		}
	}

	// Get SNS message attributes
	var snsEvent snsEntityBody
	if err := json.Unmarshal([]byte(msg.Body), &snsEvent); err != nil {
		return traceAttributes
	}

	if snsEvent.MessageAttributes != nil {
		for key, attr := range snsEvent.MessageAttributes {
			if array.ContainsString(allTraceHeaders, key) && strings.ToLower(attr.Type) == "string" {
				traceAttributes[key] = attr.Value
			}
		}
	}

	return traceAttributes
}

func startSpan(ctx context.Context, serviceName string, parentSpanContext dd_trace.SpanContext) (dd_tracer.Span, context.Context) {
	operationName := "record.handler"
	spanOpts := []dd_trace.StartSpanOption{
		dd_tracer.SpanType("serverless"),
		dd_tracer.ServiceName(serviceName),
	}

	// Populate span with lambda information
	lambdaCtx, ok := lambdacontext.FromContext(ctx)
	if ok {
		functionArn := lambdaCtx.InvokedFunctionArn
		functionArn = strings.ToLower(functionArn)
		functionArn, functionVersion := separateVersionFromFunctionArn(functionArn)

		spanOpts = append(spanOpts,
			dd_tracer.ResourceName(lambdacontext.FunctionName),
			dd_tracer.Tag("cold_start", ctx.Value("cold_start")),
			dd_tracer.Tag("function_arn", functionArn),
			dd_tracer.Tag("function_version", functionVersion),
			dd_tracer.Tag("request_id", lambdaCtx.AwsRequestID),
			dd_tracer.Tag("resource_names", lambdacontext.FunctionName),
		)
	}

	if parentSpanContext == nil {
		return dd_tracer.StartSpanFromContext(ctx, operationName, spanOpts...)
	}

	spanOpts = append(spanOpts, dd_tracer.ChildOf(parentSpanContext))

	recordSpan := dd_tracer.StartSpan(operationName, spanOpts...)
	recordCtx := dd_tracer.ContextWithSpan(ctx, recordSpan)

	return recordSpan, recordCtx
}

func separateVersionFromFunctionArn(functionArn string) (arnWithoutVersion string, functionVersion string) {
	// Example arn: arn:aws:lambda:us-east-2:123456789012:function:my-function:1
	arnSegments := strings.Split(functionArn, ":")
	functionVersion = "$LATEST"
	arnWithoutVersion = strings.Join(arnSegments[0:7], ":")

	const lastPartOfArn = 7
	if len(arnSegments) > lastPartOfArn {
		functionVersion = arnSegments[lastPartOfArn]
	}

	return arnWithoutVersion, functionVersion
}
