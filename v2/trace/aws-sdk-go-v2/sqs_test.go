package aws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	trace "github.com/SKF/go-utility/v2/trace/aws-sdk-go-v2"
	"github.com/SKF/go-utility/v2/uuid"
)

func Test_Injection_SendMessageInput(t *testing.T) {
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
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           s.URL,
				SigningRegion: "eu-west-1",
			}, nil
		})),
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
		input, ok := in.Parameters.(*sqs.SendMessageInput)
		require.True(t, ok, "unexpected type %T", in.Parameters)

		require.NotEmpty(t, input.MessageAttributes)

		assert.Contains(t, input.MessageAttributes, tracer.DefaultTraceIDHeader)
		assert.Contains(t, input.MessageAttributes, tracer.DefaultParentIDHeader)
		assert.Contains(t, input.MessageAttributes, tracer.DefaultPriorityHeader)
	})

	var (
		queueURL = ""
		message  = ""
	)

	input := &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &message,
	}

	_, err = sqs.NewFromConfig(cfg).SendMessage(ctx, input)
	require.NoError(t, err)

	span.Finish()
	tracer.Stop()
}

func Test_Injection_SendMessageBatch(t *testing.T) {
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
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           s.URL,
				SigningRegion: "eu-west-1",
			}, nil
		})),
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
		input, ok := in.Parameters.(*sqs.SendMessageBatchInput)
		require.True(t, ok, "unexpected type %T", in.Parameters)
		require.NotEmpty(t, input.Entries)
		require.Len(t, input.Entries, 1)

		entry := input.Entries[0]

		require.NotEmpty(t, entry.MessageAttributes)

		assert.Contains(t, entry.MessageAttributes, tracer.DefaultTraceIDHeader)
		assert.Contains(t, entry.MessageAttributes, tracer.DefaultParentIDHeader)
		assert.Contains(t, entry.MessageAttributes, tracer.DefaultPriorityHeader)
	})

	var (
		queueURL = ""
		message  = ""
		id       = uuid.New().String()
	)

	input := &sqs.SendMessageBatchInput{
		QueueUrl: &queueURL,
		Entries: []types.SendMessageBatchRequestEntry{
			{
				Id:          &id,
				MessageBody: &message,
			},
		},
	}

	_, err = sqs.NewFromConfig(cfg).SendMessageBatch(ctx, input)
	require.NoError(t, err)

	span.Finish()
	tracer.Stop()
}
