package awstrace

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// WrapSession will wrap the AWS session and
// injecting tracing headers into the outgoing requests for these functions:
// - sns.PublishWithContext(...)
// - sqs.SendMessageWithContext(...)
// - sqs.SendMessageBatchWithContext(...)
// the wrapper support B3 and Datadog
func WrapSession(sess *session.Session) *session.Session {
	sess.Handlers.Build.PushFront(matchingHandler("sns", "Publish", snsPublishHandler))
	sess.Handlers.Build.PushFront(matchingHandler("sqs", "SendMessage", sqsSendMessageHandler))
	sess.Handlers.Build.PushFront(matchingHandler("sqs", "SendMessageBatch", sqsSendMessageBatchHandler))

	return sess
}

func matchingHandler(service, operation string, handler func(*request.Request)) func(*request.Request) {
	return func(r *request.Request) {
		if service == r.ClientInfo.ServiceName && operation == r.Operation.Name {
			handler(r)
		}
	}
}

func snsPublishHandler(r *request.Request) {
	if input, ok := r.Params.(*sns.PublishInput); ok {
		r.Params = injectSNSPublish(r.Context(), input)
	}
}

func sqsSendMessageBatchHandler(r *request.Request) {
	if input, ok := r.Params.(*sqs.SendMessageInput); ok {
		r.Params = injectSQSSendMessage(r.Context(), input)
	}
}

func sqsSendMessageHandler(r *request.Request) {
	if input, ok := r.Params.(*sqs.SendMessageBatchInput); ok {
		r.Params = injectSQSSendMessageBatch(r.Context(), input)
	}
}
