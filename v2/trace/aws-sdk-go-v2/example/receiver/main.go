package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/SKF/go-utility/v2/log"
	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
)

type handler struct {
	sqs *sqs.Client
}

func (h *handler) handle(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := h.handleRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) traceRecord(ctx context.Context, record events.SQSMessage) error {
	span, ctx := trace.SQSMessageCarrier(record).StartSpan(ctx, "operation")
	defer span.Finish()

	return h.handleRecord(ctx, record)
}

func (h *handler) handleRecord(ctx context.Context, record events.SQSMessage) error {
	log.WithTracing(ctx).Info("Doing stuff")

	return nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	trace.AppendMiddleware(&cfg)

	h := &handler{
		sqs: sqs.NewFromConfig(cfg),
	}

	lambda.Start(h.handle)
}
