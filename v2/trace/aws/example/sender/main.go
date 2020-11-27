// Example using AWS Wrapper to send trace headers

package main

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	oc_trace "go.opencensus.io/trace"

	"github.com/SKF/go-utility/v2/log"
	aws_skf_trace "github.com/SKF/go-utility/v2/trace/aws"
)

func sendMessage(ctx context.Context, sess *session.Session, msg string) error {
	// Context need to containing Datadog or B3 trace for this to work
	span, ctx := oc_trace.StartSpan(ctx, "sendMessage")
	defer span.End()

	// Sending message and injecting tracing headers
	client := sqs.New(sess)
	_, err := client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(msg),
	})
	return err
}

func main() {
	sess := session.Must(session.NewSession())

	// Pass trace headers by wrapping the session with WrapSession
	sess = aws_skf_trace.WrapSession(sess)

	ctx := context.Background()
	if err := sendMessage(ctx, sess, "Hello SKF!"); err != nil {
		log.WithTracing(ctx).
			WithError(err).
			Error("Error sending message")
	}
}
