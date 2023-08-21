package awscloudwatchlogevents

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	datadog "github.com/SKF/go-utility/v2/datadog/client"
	"github.com/SKF/go-utility/v2/datadog/tags"
)

const (
	lambdaEventType     = "lambda"
	ecsEventType        = "ecs"
	cloudwatchEventType = "cloudwatch"
	defaultNoOfWorkers  = 1
)

type Processor struct {
	tags        tags.Tags
	service     string
	client      datadog.Client
	noOfWorkers int
	errs        []error
}

func NewProcessor(service string, client datadog.Client) *Processor {
	return &Processor{
		service:     service,
		client:      client,
		noOfWorkers: defaultNoOfWorkers,
	}
}

func (p *Processor) Withtags(tags tags.Tags) *Processor {
	p.tags = tags
	return p
}

func (p *Processor) WithNoOfWorkers(noOfWorkers int) *Processor {
	p.noOfWorkers = noOfWorkers
	return p
}

func (p *Processor) Errors() []error {
	return p.errs
}

// Process will parse and push AWS Cloudwatch Log Events as Datadog logs to Datadog
func (p *Processor) Process(ctx context.Context, request events.CloudwatchLogsEvent) {
	logsData, err := request.AWSLogs.Parse()
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("failed to parse raw AWS logs: %w", err))
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

	workers := make([]*worker, p.noOfWorkers)
	for idx := range workers {
		w := newWorker(idx, p.service, p.client).
			withTags(p.tags).
			withEventType(eventType).
			withSource(source)

		go w.start(done, work)
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
