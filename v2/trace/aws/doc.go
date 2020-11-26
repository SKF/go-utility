// Package awstrace contains aws session wrapper for tracing through SNS and SQS
// supporting Datadog native and B3 tracing contexts.
//
// Examples
//
// An example of tracing between two services through SQS
//
// Using WrapSession, service 1
//
//		func sendMessage(msg string) error {
//			// Pass trace headers by wrapping the session with WrapSession
//			sess := session.Must(session.NewSession())
//			sess = awstrace.WrapSession(sess)
//
//			// Context need to containing Datadog or B3 trace for this to work
//			span, ctx := oc_trace.StartSpan(ctx, "example")
//			defer span.End()
//
//			// Sending message and injecting tracing headers
//			client := sqs.New(sess)
//			_, err := client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
//				MessageBody: aws.String(msg),
//			})
//			return err
//		}
//
// Using StartDatadogSpanFromMessage, service 2
//
//		func handleRecord(ctx context.Context, record *sqs.SQSMessage) {
//			// Creating a new datadog span and extracing the trace headers
//			span, ctx = awstrace.StartDatadogSpanFromMessage(ctx, serviceName, record)
//			defer span.Finish()
//
//			...
//		}
//
package awstrace
