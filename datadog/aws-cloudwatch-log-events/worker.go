package awscloudwatchlogevents

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"

	lambda_report_messages "github.com/SKF/go-utility/datadog/aws-cloudwatch-log-events/lambda-report-messages"
	datadog "github.com/SKF/go-utility/datadog/client"
	"github.com/SKF/go-utility/datadog/tags"
)

type worker struct {
	id        int
	service   string
	client    datadog.Client
	errs      []error
	tags      tags.Tags
	eventType string
	source    string
}

func newWorker(id int, service string, client datadog.Client) *worker {
	return &worker{
		id:      id,
		service: service,
		client:  client,
	}
}

func (w *worker) withTags(tags tags.Tags) *worker {
	w.tags = tags
	return w
}

func (w *worker) withEventType(eventType string) *worker {
	w.eventType = eventType
	return w
}

func (w *worker) withSource(source string) *worker {
	w.source = source
	return w
}

func (w *worker) errors() []error {
	return w.errs
}

func (w *worker) start(done chan int, work chan events.CloudwatchLogsLogEvent) {
	for event := range work {
		logEntry, err := w.mapToDatadogLog(event)
		if err != nil {
			err = errors.Wrapf(err, "failed to map AWS log event [%s] to a Datadog log", event.ID)
			w.errs = append(w.errs, err)
			continue
		}

		if logEntry == nil {
			continue
		}

		if err = w.client.PostLogEntry(logEntry); err != nil {
			err = errors.Wrapf(err, "failed to send AWS log event [%s] to Datadog", event.ID)
			w.errs = append(w.errs, err)
			continue
		}
	}

	done <- w.id
}

// mapToDatadogLog converts the Cloudwatch Log Event into a dictionary for Datadog
func (w *worker) mapToDatadogLog(event events.CloudwatchLogsLogEvent) (_ interface{}, err error) {
	msg := strings.TrimSpace(event.Message)
	if msg == "" {
		return
	}

	if strings.HasPrefix(msg, "MONITORING") {
		// MONITORING messages are metrics, and should be passed through as-is
		return msg, nil
	}

	jsonDict := map[string]interface{}{}

	if strings.HasPrefix(msg, "{") && strings.HasSuffix(msg, "}") {
		// probably json
		if err = handleJSON(msg, jsonDict); err != nil {
			return
		}
	} else {
		// probably not json, just set the message
		jsonDict["message"] = msg
	}

	// check if we have the application property set, if it's not there, then we assume it's backend
	if _, ok := jsonDict["application"]; !ok {
		jsonDict["application"] = "backend"
	}

	jsonDict["ddsourcecategory"] = w.eventType
	jsonDict["service"] = w.service
	jsonDict["timestamp"] = event.Timestamp
	jsonDict["ddsource"] = w.source

	if w.eventType == lambdaEventType {
		if lambdaMsg := lambda_report_messages.Parse(event.Message); lambdaMsg != nil {
			jsonDict[lambdaEventType] = lambdaMsg
		} else {
			jsonDict[lambdaEventType] = lambda_report_messages.LambdaBaseMsg{Type: "message"}
		}
	}

	jsonDict["ddtags"] = w.tags.String()

	return jsonDict, nil
}

func handleJSON(msg string, jsonDict map[string]interface{}) (err error) {
	// trim any UTF-8 encoding prefix from the incoming message data
	var messageBytes = bytes.TrimPrefix([]byte(msg), []byte("\xef\xbb\xbf"))
	if err = json.Unmarshal(messageBytes, &jsonDict); err != nil {
		return
	}

	// Datadog tracing information
	// https://docs.datadoghq.com/tracing/advanced/connect_logs_and_traces/#manual-trace-id-injection
	tracing := struct {
		TraceID *uint64 `json:"dd.trace_id,omitempty"`
		SpanID  *uint64 `json:"dd.span_id,omitempty"`
	}{}
	if err = json.Unmarshal(messageBytes, &tracing); err != nil {
		return
	}
	if tracing.TraceID != nil {
		jsonDict["dd.trace_id"] = *tracing.TraceID
		jsonDict["dd.span_id"] = *tracing.SpanID
	}

	// change any "msg" field to "message" for Datadog
	// https://docs.datadoghq.com/logs/processing/#message-attribute
	if msgValue, ok := jsonDict["msg"]; ok {
		jsonDict["message"] = msgValue
		delete(jsonDict, "msg")
	}

	return
}
