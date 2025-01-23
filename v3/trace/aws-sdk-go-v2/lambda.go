package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type SQSMessageCarrier events.SQSMessage

func (s SQSMessageCarrier) ForeachKey(handler func(key, value string) error) error {
	for k, v := range s.MessageAttributes {
		if v.DataType == stringDataType && v.StringValue != nil {
			if err := handler(k, *v.StringValue); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s SQSMessageCarrier) StartSpan(ctx context.Context, operationName string, opts ...tracer.StartSpanOption) (tracer.Span, context.Context) {
	spanOpts := []tracer.StartSpanOption{
		tracer.SpanType("serverless"),
	}

	if lambdaCtx, ok := lambdacontext.FromContext(ctx); ok {
		functionARN := strings.ToLower(lambdaCtx.InvokedFunctionArn)

		spanOpts = append(spanOpts,
			tracer.ResourceName(lambdacontext.FunctionName),
			tracer.Tag("cold_start", ctx.Value("cold_start")),
			tracer.Tag("function_arn", functionARN),
			tracer.Tag("request_id", lambdaCtx.AwsRequestID),
		)
	}

	opts = append(opts, spanOpts...)

	return StartSpan(ctx, s, operationName, opts...)
}
