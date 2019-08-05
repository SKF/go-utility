package lambdareportmessages_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lambda_report_messages "github.com/SKF/go-utility/datadog/aws-cloudwatch-log-events/lambda-report-messages"
)

func Test_Parse_Report_Msg(t *testing.T) {
	msg := "REPORT RequestId: 83de0fc2-cd31-11e8-8c01-5da1f4bca768	Duration: 363.36 ms	Billed Duration: 400 ms Memory Size: 1024 MB	Max Memory Used: 30 MB"
	parsedMsg := lambda_report_messages.Parse(msg)
	require.NotNil(t, parsedMsg)

	reportMsg, ok := parsedMsg.(*lambda_report_messages.LambdaReportMsg)
	require.True(t, ok)

	assert.Equal(t, "report", reportMsg.Type)
	assert.Equal(t, "83de0fc2-cd31-11e8-8c01-5da1f4bca768", reportMsg.RequestID)
	assert.Equal(t, "363.36", reportMsg.Duration)
	assert.Equal(t, "ms", reportMsg.DurationUnit)
	assert.Equal(t, "400", reportMsg.BilledDuration)
	assert.Equal(t, "ms", reportMsg.BilledDurationUnit)
	assert.Equal(t, "1024", reportMsg.MemorySize)
	assert.Equal(t, "MB", reportMsg.MemorySizeUnit)
	assert.Equal(t, "30", reportMsg.MaxMemoryUsed)
	assert.Equal(t, "MB", reportMsg.MaxMemoryUsedUnit)
}

func Test_Parse_Start_Msg(t *testing.T) {
	msg := "START RequestId: 32e3e48c-d05a-11e8-8599-d93918f49152"
	parsedMsg := lambda_report_messages.Parse(msg)
	require.NotNil(t, parsedMsg)

	baseMsg, ok := parsedMsg.(*lambda_report_messages.LambdaBaseMsg)
	require.True(t, ok)

	assert.Equal(t, "start", baseMsg.Type)
	assert.Equal(t, "32e3e48c-d05a-11e8-8599-d93918f49152", baseMsg.RequestID)

	msg = "START RequestId: 44e1de11-d059-11e8-9634-87afec859d44 Version: $LATEST"
	parsedMsg = lambda_report_messages.Parse(msg)
	require.NotNil(t, parsedMsg)

	baseMsg, ok = parsedMsg.(*lambda_report_messages.LambdaBaseMsg)
	require.True(t, ok)

	assert.Equal(t, "start", baseMsg.Type)
	assert.Equal(t, "44e1de11-d059-11e8-9634-87afec859d44", baseMsg.RequestID)
}

func Test_Parse_End_Msg(t *testing.T) {
	msg := "END RequestId: 32e3e48c-d05a-11e8-8599-d93918f49152"
	parsedMsg := lambda_report_messages.Parse(msg)
	require.NotNil(t, parsedMsg)

	baseMsg, ok := parsedMsg.(*lambda_report_messages.LambdaBaseMsg)
	require.True(t, ok)

	assert.Equal(t, "end", baseMsg.Type)
	assert.Equal(t, "32e3e48c-d05a-11e8-8599-d93918f49152", baseMsg.RequestID)
}

func Test_Parse_Bad_Start_Msg(t *testing.T) {
	msg := "NOTSTART RequestId: 44e1de11-d059-11e8-9634-87afec859d44 Version: $LATEST"
	parsedMsg := lambda_report_messages.Parse(msg)
	assert.Nil(t, parsedMsg)
}

func Test_Parse_Bad_End_Msg(t *testing.T) {
	msg := "NOTEND RequestId: 44e1de11-d059-11e8-9634-87afec859d44"
	parsedMsg := lambda_report_messages.Parse(msg)
	assert.Nil(t, parsedMsg)
}

func Test_Parse_Empty_Msg(t *testing.T) {
	msg := ""
	parsedMsg := lambda_report_messages.Parse(msg)
	assert.Nil(t, parsedMsg)
}
