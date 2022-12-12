package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/middleware"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var stringDataType = "String"

func injectMiddleware(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return next.HandleInitialize(ctx, in)
	}

	switch v := in.Parameters.(type) {
	case *sqs.SendMessageBatchInput:
		tracer.Inject(span.Context(), (*SendMessageBatchInputCarrier)(v))
	case *sqs.SendMessageInput:
		tracer.Inject(span.Context(), (*SendMessageInputCarrier)(v))
	case *sns.PublishBatchInput:
		tracer.Inject(span.Context(), (*PublishBatchInputCarrier)(v))
	case *sns.PublishInput:
		tracer.Inject(span.Context(), (*PublishInputCarrier)(v))
	}

	return next.HandleInitialize(ctx, in)
}

func AppendMiddleware(cfg *aws.Config) {
	cfg.APIOptions = append(cfg.APIOptions, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("InitInjectMiddleware", injectMiddleware), middleware.Before)
	})
}

func StartSpan(ctx context.Context, carrier tracer.TextMapReader, operationName string, opts ...tracer.StartSpanOption) (tracer.Span, context.Context) {
	if parent, err := tracer.Extract(carrier); err == nil {
		opts = append(opts, tracer.ChildOf(parent))
	}

	return tracer.StartSpanFromContext(ctx, operationName, opts...)
}
