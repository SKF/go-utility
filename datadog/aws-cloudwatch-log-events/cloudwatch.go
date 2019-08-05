package awscloudwatchlogevents

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"

	datadog "github.com/SKF/go-utility/datadog/client"
	"github.com/SKF/go-utility/datadog/tags"
)

const (
	lambdaEventType     = "lambda"
	ecsEventType        = "ecs"
	cloudwatchEventType = "cloudwatch"
	noOfWorkers         = 1
)

type Processor struct {
	tags    tags.Tags
	service string
	client  datadog.Client
	errs    []error
}

func (p *Processor) WithClient(client datadog.Client) *Processor {
	p.client = client
	return p
}

func (p *Processor) Withtags(tags tags.Tags) *Processor {
	p.tags = tags
	return p
}

func (p *Processor) WithService(service string) *Processor {
	p.service = service
	return p
}

func (p *Processor) Errors() []error {
	return p.errs
}

// Process will parse and push AWS Cloudwatch Log Events as Datadog logs to Datadog
func (p *Processor) Process(ctx context.Context, request events.CloudwatchLogsEvent) {
	logsData, err := request.AWSLogs.Parse()
	if err != nil {
		p.errs = append(p.errs, errors.Wrap(err, "failed to parse raw AWS logs"))
		return
	}

	eventType := parseEventType(logsData.LogGroup)
	source := parseSource(eventType, logsData)

	work := make(chan events.CloudwatchLogsLogEvent)
	go func() {
		for _, event := range logsData.LogEvents {
			select {
			case <-ctx.Done():
			case work <- event:
			}
		}
		close(work)
	}()

	done := make(chan int)
	workers := make([]*worker, noOfWorkers)
	for idx := range workers {
		w := &worker{id: idx}
		go w.withDatadogClient(p.client).
			withTags(p.tags).
			withService(p.service).
			withEventType(eventType).
			withSource(source).
			start(done, work)
		workers[idx] = w
	}

	for range workers {
		idx := <-done
		w := workers[idx]
		p.errs = append(p.errs, w.errors()...)
	}
}

func parseEventType(logGroup string) string {
	if strings.Contains(logGroup, "/aws/lambda/") {
		return lambdaEventType
	}
	if strings.Contains(logGroup, "/aws/ecs/") {
		return ecsEventType
	}
	return cloudwatchEventType
}

func parseSource(eventType string, logsData events.CloudwatchLogsData) string {
	switch eventType {
	case lambdaEventType:
		return getLambdaName(logsData.LogGroup)
	case ecsEventType:
		return getECSFargateServiceName(logsData.LogGroup)
	default:
		return cloudwatchEventType
	}
}

func getLambdaName(logGroup string) string {
	return strings.Replace(logGroup, "/aws/lambda/", "", 1)
}

func getECSFargateServiceName(logGroup string) string {
	return strings.Replace(logGroup, "/aws/ecs/fargate/", "", 1)
}
