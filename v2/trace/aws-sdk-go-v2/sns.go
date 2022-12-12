package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type PublishInputCarrier sns.PublishInput

var _ tracer.TextMapWriter = (*PublishInputCarrier)(nil)

func (p *PublishInputCarrier) Set(key, value string) {
	if p.MessageAttributes == nil {
		p.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	p.MessageAttributes[key] = types.MessageAttributeValue{
		DataType:    &stringDataType,
		StringValue: &value,
	}
}

type PublishBatchInputCarrier sns.PublishBatchInput

var _ tracer.TextMapWriter = (*PublishBatchInputCarrier)(nil)

func (p *PublishBatchInputCarrier) Set(key, value string) {
	for i := range p.PublishBatchRequestEntries {
		if p.PublishBatchRequestEntries[i].MessageAttributes == nil {
			p.PublishBatchRequestEntries[i].MessageAttributes = make(map[string]types.MessageAttributeValue)
		}

		p.PublishBatchRequestEntries[i].MessageAttributes[key] = types.MessageAttributeValue{
			DataType:    &stringDataType,
			StringValue: &value,
		}
	}
}
