package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
	"github.com/SKF/go-utility/v2/uuid"
)

type snsresolver struct {
	s *httptest.Server
}

func (r *snsresolver) ResolveEndpoint(ctx context.Context, params sns.EndpointParameters) (smithyendpoints.Endpoint, error) {
	params.Endpoint = &r.s.URL
	return sns.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

func Test_Injection_PublishInput(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	host, err := url.Parse(s.URL)
	require.NoError(t, err)

	t.Setenv("DD_TRACE_STARTUP_LOGS", "false")
	t.Setenv("DD_AGENT_HOST", host.Hostname())
	t.Setenv("DD_TRACE_AGENT_PORT", host.Port())
	t.Setenv("DD_PROPAGATION_STYLE_INJECT", "Datadog")

	tracer.Start()
	defer tracer.Stop()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(_ context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "0",
					SecretAccessKey: "0",
				}, nil
			},
		)),
	)
	require.NoError(t, err)

	span, ctx := tracer.StartSpanFromContext(context.TODO(), "testcase")
	defer span.Finish()

	trace.AppendMiddleware(&cfg)
	AppendValidatorMiddleware(&cfg, func(in middleware.InitializeInput) {
		input, ok := in.Parameters.(*sns.PublishInput)
		require.True(t, ok, "unexpected type %T", in.Parameters)

		require.NotEmpty(t, input.MessageAttributes)

		assert.Contains(t, input.MessageAttributes, tracer.DefaultTraceIDHeader)
		assert.Contains(t, input.MessageAttributes, tracer.DefaultParentIDHeader)
		assert.Contains(t, input.MessageAttributes, tracer.DefaultPriorityHeader)
	})

	message := "profound"

	input := &sns.PublishInput{
		Message: &message,
	}

	_, err = sns.NewFromConfig(cfg, sns.WithEndpointResolverV2(&snsresolver{s})).Publish(ctx, input)
	require.NoError(t, err)

	span.Finish()
	tracer.Stop()
}

func Test_Injection_PublishBatchInput(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	host, err := url.Parse(s.URL)
	require.NoError(t, err)

	t.Setenv("DD_TRACE_STARTUP_LOGS", "false")
	t.Setenv("DD_AGENT_HOST", host.Hostname())
	t.Setenv("DD_TRACE_AGENT_PORT", host.Port())
	t.Setenv("DD_PROPAGATION_STYLE_INJECT", "DataDog")

	tracer.Start()
	defer tracer.Stop()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(_ context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "0",
					SecretAccessKey: "0",
				}, nil
			},
		)),
	)
	require.NoError(t, err)

	span, ctx := tracer.StartSpanFromContext(context.TODO(), "testcase")
	defer span.Finish()

	trace.AppendMiddleware(&cfg)
	AppendValidatorMiddleware(&cfg, func(in middleware.InitializeInput) {
		input, ok := in.Parameters.(*sns.PublishBatchInput)
		require.True(t, ok, "unexpected type %T", in.Parameters)
		require.NotEmpty(t, input.PublishBatchRequestEntries)
		require.Len(t, input.PublishBatchRequestEntries, 1)

		entry := input.PublishBatchRequestEntries[0]

		require.NotEmpty(t, entry)

		assert.Contains(t, entry.MessageAttributes, tracer.DefaultTraceIDHeader)
		assert.Contains(t, entry.MessageAttributes, tracer.DefaultParentIDHeader)
		assert.Contains(t, entry.MessageAttributes, tracer.DefaultPriorityHeader)
	})

	var (
		topicARN = "topic-arn"
		message  = "profound"
		id       = uuid.New().String()
	)

	input := &sns.PublishBatchInput{
		TopicArn: &topicARN,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      &id,
				Message: &message,
			},
		},
	}

	_, err = sns.NewFromConfig(cfg, sns.WithEndpointResolverV2(&snsresolver{s})).PublishBatch(ctx, input)
	require.NoError(t, err)

	err = tracer.Inject(span.Context(), (*trace.PublishBatchInputCarrier)(input))
	require.NoError(t, err)

	span.Finish()
	tracer.Stop()
}
