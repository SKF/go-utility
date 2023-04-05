package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type SendMessageInputCarrier sqs.SendMessageInput

var _ tracer.TextMapWriter = (*SendMessageInputCarrier)(nil)

func (m *SendMessageInputCarrier) Set(key, value string) {
	if m.MessageAttributes == nil {
		m.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	m.MessageAttributes[key] = types.MessageAttributeValue{
		DataType:    &stringDataType,
		StringValue: &value,
	}
}

type SendMessageBatchInputCarrier sqs.SendMessageBatchInput

var _ tracer.TextMapWriter = (*SendMessageBatchInputCarrier)(nil)

func (m *SendMessageBatchInputCarrier) Set(key, value string) {
	for i := range m.Entries {
		if m.Entries[i].MessageAttributes == nil {
			m.Entries[i].MessageAttributes = make(map[string]types.MessageAttributeValue)
		}

		m.Entries[i].MessageAttributes[key] = types.MessageAttributeValue{
			DataType:    &stringDataType,
			StringValue: &value,
		}
	}
}
