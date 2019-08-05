package awscloudwatchlogevents_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aws_cloudwatch_log_events "github.com/SKF/go-utility/datadog/aws-cloudwatch-log-events"
	lambda_report_messages "github.com/SKF/go-utility/datadog/aws-cloudwatch-log-events/lambda-report-messages"
	"github.com/SKF/go-utility/datadog/tags"
)

func Test_Process_Valid_Request(t *testing.T) {
	service := "keckebarn"

	tags := tags.Tags{}
	tags.AddTagsAsString("Apa,bepa,cEPA")
	tags.AddTag("apa", "Bepa")
	tags.AddTag("Apa", "bEPA")

	client := &mockDatadogClient{}

	rawLog, err := encode(validLambdaLogEvent)
	require.Nil(t, err)
	request := events.CloudwatchLogsEvent{AWSLogs: rawLog}

	p := &aws_cloudwatch_log_events.Processor{}
	p.WithClient(client).
		Withtags(tags).
		WithService(service).
		Process(context.Background(), request)

	assert.Len(t, p.Errors(), 0)
	assert.EqualValues(t, expectedDatadogLogs, client.logEntries)
}

const source = "Monkey"

var validLambdaLogEvent = events.CloudwatchLogsData{
	LogGroup: "/aws/lambda/" + source,
	LogEvents: []events.CloudwatchLogsLogEvent{
		{
			Timestamp: time.Now().UnixNano(),
			Message:   "msg",
		},
		{
			Timestamp: time.Now().UnixNano(),
			Message:   `{"msg": "msg", "dd.trace_id": 123, "dd.span_id": 456}`,
		},
		{
			Timestamp: time.Now().UnixNano(),
			Message:   "START RequestId: Apa123",
		},
		{
			Timestamp: time.Now().UnixNano(),
			Message:   "MONITORING apa",
		},
	},
}
var expectedDatadogLogs = []interface{}{
	map[string]interface{}{
		"application":      "backend",
		"ddsource":         "Monkey",
		"ddsourcecategory": "lambda",
		"ddtags":           "Apa,bepa,cEPA,apa:Bepa,Apa:bEPA",
		"lambda":           lambda_report_messages.LambdaBaseMsg{Type: "message"},
		"message":          "msg",
		"service":          "keckebarn",
		"timestamp":        validLambdaLogEvent.LogEvents[0].Timestamp,
	},
	map[string]interface{}{
		"application":      "backend",
		"ddsource":         "Monkey",
		"ddsourcecategory": "lambda",
		"ddtags":           "Apa,bepa,cEPA,apa:Bepa,Apa:bEPA",
		"lambda":           lambda_report_messages.LambdaBaseMsg{Type: "message"},
		"message":          "msg",
		"service":          "keckebarn",
		"timestamp":        validLambdaLogEvent.LogEvents[1].Timestamp,
		"dd.trace_id":      uint64(123),
		"dd.span_id":       uint64(456),
	},
	map[string]interface{}{
		"application":      "backend",
		"ddsource":         "Monkey",
		"ddsourcecategory": "lambda",
		"ddtags":           "Apa,bepa,cEPA,apa:Bepa,Apa:bEPA",
		"lambda":           &lambda_report_messages.LambdaBaseMsg{Type: "start", RequestID: "Apa123"},
		"message":          "START RequestId: Apa123",
		"service":          "keckebarn",
		"timestamp":        validLambdaLogEvent.LogEvents[2].Timestamp,
	},
	"MONITORING apa",
}

type mockDatadogClient struct {
	logEntries []interface{}
}

func (c *mockDatadogClient) PostLogEntry(request interface{}) (err error) {
	c.logEntries = append(c.logEntries, request)
	return
}

func encode(d events.CloudwatchLogsData) (c events.CloudwatchLogsRawData, err error) {
	data, err := json.Marshal(d)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err = zw.Write(data); err != nil {
		return
	}

	if err = zw.Close(); err != nil {
		return
	}

	c.Data = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}
