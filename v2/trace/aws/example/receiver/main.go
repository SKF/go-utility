// Example lamba receiving trace headers from SQS record

package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/SKF/go-utility/v2/env"
	"github.com/SKF/go-utility/v2/log"
	aws_skf_trace "github.com/SKF/go-utility/v2/trace/aws"
)

type handler struct {
	sqs         *sqs.SQS
	queueURL    string
	serviceName string
}

func (h *handler) handler(ctx context.Context, event events.SQSEvent) {
	// Traverse all records
	for _, record := range event.Records {
		// Handle one record at the time
		if err := h.handleRecord(ctx, record); err != nil {
			log.WithTracing(ctx).
				WithError(err).
				Error("Error handling record")
		}
	}
}

func (h *handler) handleRecord(ctx context.Context, record events.SQSMessage) error {
	// Start a new span, using the incomming trace as parent (if any)
	span, ctx := aws_skf_trace.StartDatadogSpanFromMessage(ctx, h.serviceName, record)
	defer span.Finish()

	// Logging that the record has been handled,
	// but logs it to the new span which may have its trace parent from the record
	log.WithTracing(ctx).Infof("Record has been handled")

	// Remove record in SQS
	_, err := h.sqs.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(h.queueURL),
		ReceiptHandle: aws.String(record.ReceiptHandle),
	})

	return err
}

func main() {
	sess := session.Must(session.NewSession())

	h := handler{
		sqs:         sqs.New(sess),
		queueURL:    env.MustGetAsString("QUEUE_URL"),
		serviceName: env.MustGetAsString("SERVICE_NAME"),
	}

	lambda.Start(h.handler)
}
