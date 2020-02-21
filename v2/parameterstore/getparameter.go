package paramaterstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"go.opencensus.io/trace"
)

const errSSMClientNotConfigured = "ssm client not configured"

var ssmClient *ssm.SSM

func ConfigureParameterStore(ctx context.Context, client *ssm.SSM) {
	ssmClient = client
}

func GetParameter(ctx context.Context, parameterKey string) (string, error) {
	ctx, span := trace.StartSpan(ctx, "parameterstore/GetParameter")
	defer span.End()

	if ssmClient == nil {
		return "", errors.New(errSSMClientNotConfigured)
	}

	output, err := ssmClient.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(parameterKey),
		WithDecryption: aws.Bool(false),
	})

	if err != nil {
		if awsError, ok := err.(awserr.Error); ok {
			if awsError.Code() == ssm.ErrCodeParameterNotFound ||
				awsError.Code() == ssm.ErrCodeParameterVersionNotFound {
				return "", nil
			}
		}

		return "", nil
	}

	if output.Parameter == nil {
		return "", fmt.Errorf("param '%v' was nil", parameterKey)
	}

	value := *output.Parameter.Value

	return value, nil
}
