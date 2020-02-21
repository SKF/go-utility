package sqs

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/SKF/go-utility/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.opencensus.io/trace"
)

var queueURL string
var sqsClient *sqs.SQS

const errQueueURLNotConfigured = "queue URL not configured"
const errQueueClientNotConfigured = "queue client not configured"

func ConfigureQueue(ctx context.Context, queue string, svc *sqs.SQS) {
	sqsClient = svc
	queueURL = queue
}

func PutObjectOnQueue(ctx context.Context, object interface{}) error {
	ctx, span := trace.StartSpan(ctx, "sqs/PutObjectOnQueue")
	defer span.End()

	if queueURL == "" {
		err := errors.New(errQueueURLNotConfigured)
		log.Warn(errQueueURLNotConfigured)

		return err
	}

	if sqsClient == nil {
		err := errors.New(errQueueClientNotConfigured)
		log.Warn(errQueueClientNotConfigured)

		return err
	}

	jsonStr, marshalErr := requestToJSON(object)
	if marshalErr != nil {
		return marshalErr
	}

	input := sqs.SendMessageInput{
		MessageBody: aws.String(jsonStr),
		QueueUrl:    &queueURL,
	}
	_, err := sqsClient.SendMessageWithContext(ctx, &input)

	if err != nil {
		log.WithError(err).
			Error("Unable to Send message to queue")
		return err
	}

	return nil
}

func requestToJSON(object interface{}) (string, error) {
	bodyByteARR, err := json.Marshal(object)
	if err != nil {
		log.
			WithError(err).
			Error("Unable to marshal object to put on queue to json")

		return "", err
	}

	return string(bodyByteARR), nil
}
