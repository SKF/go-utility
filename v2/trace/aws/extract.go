package aws_trace

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SKF/go-utility/v2/array"
	"github.com/SKF/go-utility/v2/log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	dd_trace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type MessageAttribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type SNSEntityBody struct {
	MessageAttributes map[string]MessageAttribute `json:"MessageAttributes"`
}

var propagator = dd_tracer.NewPropagator(nil)

func StartDatadogSpanFromMessage(ctx context.Context, serviceName string, msg events.SQSMessage) (dd_tracer.Span, context.Context) {
	traceHeaders := getTraceHeadersFromAttributes(ctx, msg)
	if len(traceHeaders) == 0 {

		fmt.Println("HEJ 1")
		return startSpan(ctx, serviceName, nil)
	}

	recordSpanContext, err := propagator.Extract(dd_tracer.TextMapCarrier(traceHeaders))
	if err != nil {
		log.WithTracing(ctx).
			WithError(err).
			Debug("couldnt create span from headers, using incomming span as parent")

		fmt.Println("HEJ 2")
		return startSpan(ctx, serviceName, nil)
	}

	fmt.Println("HEJ 3")

	return startSpan(ctx, serviceName, recordSpanContext)
}

func getTraceHeadersFromAttributes(ctx context.Context, msg events.SQSMessage) map[string]string {
	traceAttributes := map[string]string{}

	// Get SQS message attributes
	for key, attr := range msg.MessageAttributes {
		dataType := strings.ToLower(attr.DataType)
		if array.ContainsString(allHeaders, key) && dataType == "string" {
			traceAttributes[key] = *attr.StringValue
		}
	}

	// Get SNS message attributes
	var snsEvent SNSEntityBody
	if err := json.Unmarshal([]byte(msg.Body), &snsEvent); err != nil {
		log.WithTracing(ctx).WithError(err).Debug("failed to unmarshal sns body")
		return traceAttributes
	}

	if snsEvent.MessageAttributes != nil {
		for key, attr := range snsEvent.MessageAttributes {
			if array.ContainsString(allHeaders, key) && strings.ToLower(attr.Type) == "string" {
				traceAttributes[key] = attr.Value
			}
		}
	}

	return traceAttributes
}

func startSpan(ctx context.Context, serviceName string, parentSpanContext dd_trace.SpanContext) (dd_tracer.Span, context.Context) {
	operationName := "sns.handler"
	spanOpts := []dd_trace.StartSpanOption{
		dd_tracer.SpanType("serverless"),
		dd_tracer.ServiceName(serviceName),
	}

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
	return recordSpan, dd_tracer.ContextWithSpan(ctx, recordSpan)
}

func separateVersionFromFunctionArn(functionArn string) (arnWithoutVersion string, functionVersion string) {
	arnSegments := strings.Split(functionArn, ":")
	functionVersion = "$LATEST"
	arnWithoutVersion = strings.Join(arnSegments[0:7], ":")
	if len(arnSegments) > 7 {
		functionVersion = arnSegments[7]
	}
	return arnWithoutVersion, functionVersion
}
