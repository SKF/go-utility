package awstrace

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/middleware"
)

type withTrace struct{}

var _ middleware.InitializeMiddleware = (*withTrace)(nil)

func (mw withTrace) ID() string {
	return "withTrace"
}

func (mw withTrace) HandleInitialize(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (out middleware.InitializeOutput, metadata middleware.Metadata, err error) {
	switch v := in.Parameters.(type) {
	case *sns.PublishInput:
		injectSNSPublish(ctx, v)
	case *sqs.SendMessageInput:
		injectSQSSendMessage(ctx, v)
	case *sqs.SendMessageBatchInput:
		injectSQSSendMessageBatch(ctx, v)
	}

	return next.HandleInitialize(ctx, in)
}

// Adds middleware to injecting tracing headers into the outgoing requests for these functions:
// - sns.Publish(...)
// - sqs.SendMessage(...)
// - sqs.SendMessageBatch(...)
// the wrapper support B3 and Datadog
func AddTraceMiddleware(cfg aws.Config) aws.Config {
	cfg.APIOptions = append(cfg.APIOptions, func(stack *middleware.Stack) error {
		// Attach the custom middleware to the beginning of the Initialize step
		return stack.Initialize.Add(&withTrace{}, middleware.Before)
	})

	return cfg
}
