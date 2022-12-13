package aws

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type SQSMessageCarrier events.SQSMessage

func (s SQSMessageCarrier) ForeachKey(handler func(key, value string) error) error {
	for k, v := range s.MessageAttributes {
		if v.DataType == stringDataType && v.StringValue != nil {
			if err := handler(k, *v.StringValue); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s SQSMessageCarrier) StartSpan(ctx context.Context, operationName string, opts ...tracer.StartSpanOption) (tracer.Span, context.Context) {
	return StartSpan(ctx, s, operationName, opts...)
}
