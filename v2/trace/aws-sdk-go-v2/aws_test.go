package aws_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go/middleware"
)

func AppendValidatorMiddleware(cfg *aws.Config, validator func(middleware.InitializeInput)) {
	cfg.APIOptions = append(cfg.APIOptions, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(
			middleware.InitializeMiddlewareFunc(
				"m",
				func(
					ctx context.Context,
					in middleware.InitializeInput,
					next middleware.InitializeHandler,
				) (middleware.InitializeOutput, middleware.Metadata, error) {
					validator(in)

					return next.HandleInitialize(ctx, in)
				},
			), middleware.After)
	})
}
