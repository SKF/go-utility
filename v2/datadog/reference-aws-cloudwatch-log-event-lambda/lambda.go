package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	aws_cloudwatch_log_events "github.com/SKF/go-utility/v2/datadog/aws-cloudwatch-log-events"
	datadog "github.com/SKF/go-utility/v2/datadog/client"
	"github.com/SKF/go-utility/v2/datadog/tags"
	"github.com/SKF/go-utility/v2/env"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/stages"
)

var (
	stage           = strings.ToLower(env.GetAsString("STAGE", stages.StageSandbox))
	entryTags       = strings.ToLower(env.GetAsString("TAGS", ""))
	service         = env.GetAsString("SERVICE", "")
	awsRegion       = env.GetAsString("AWS_REGION", "")
	awsAccountID    = env.GetAsString("AWS_ACCOUNT_ID", "")
	vstsReleaseName = env.GetAsString("VSTS_RELEASE_NAME", "")
	vstsReleaseDef  = env.GetAsString("VSTS_RELEASE_DEF", "")
	vstsBuildNumber = env.GetAsString("VSTS_BUILD_NUMBER", "")
)

var client datadog.Client

func init() {
	ddHost := env.GetAsString("DD_HOST", "lambda-intake.logs.datadoghq.com")
	ddPort := env.GetAsString("DD_PORT", "10516")
	ddAPIKey := env.MustGetAsString("DD_API_KEY")
	useSSL := env.GetAsBool("DD_USE_SSL", true)

	client = datadog.NewTCPClient(ddHost, ddPort, ddAPIKey, useSSL)
}

// Handler parses the incoming Cloudwatch Log Events and pushes them to Datadog
func Handler(ctx context.Context, request events.CloudwatchLogsEvent) (err error) {
	tags := tags.Tags{}
	tags.AddTagsAsString("Enlight software")
	tags.AddTagsAsString(entryTags)
	tags.AddTag("env", stage)
	tags.AddTag("aws_account_id", awsAccountID)
	tags.AddTag("aws_region", awsRegion)
	tags.AddTag("vsts_release_name", vstsReleaseName)
	tags.AddTag("vsts_release_def", vstsReleaseDef)
	tags.AddTag("vsts_build_number", vstsBuildNumber)

	p := aws_cloudwatch_log_events.NewProcessor(service, client).Withtags(tags)
	p.Process(ctx, request)

	for _, err := range p.Errors() {
		log.WithError(err).Error("failed to send log events to Datadog")
	}

	return
}
