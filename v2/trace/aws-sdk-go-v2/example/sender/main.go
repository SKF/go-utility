package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
)

func sendMessage(ctx context.Context, cfg aws.Config, msg string) error {
	// Create a span to propagate
	tracer.StartSpanFromContext(ctx, "operation")

	client := sqs.NewFromConfig(cfg)

	// Trace information is injected in the input by the middleware
	_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: &msg,
	})

	return err
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	trace.AppendMiddleware(&cfg)
}
