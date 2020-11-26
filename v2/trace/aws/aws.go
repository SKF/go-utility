package aws_trace

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func WrapSession(sess *session.Session) *session.Session {
	sess.Handlers.Build.PushFront(matchingHandler("sns/Publish", snsPublishHandler))
	sess.Handlers.Build.PushFront(matchingHandler("sqs/SendMessage", sqsSendMessageHandler))
	sess.Handlers.Build.PushFront(matchingHandler("sqs/SendMessageBatch", sqsSendMessageBatchHandler))
	return sess
}

func matchingHandler(operationName string, handler func(*request.Request)) func(*request.Request) {
	return func(r *request.Request) {
		reqOperationName := fmt.Sprintf("%s/%s", r.ClientInfo.ServiceName, r.Operation.Name)
		if reqOperationName == operationName {
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
