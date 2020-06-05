package secretsmanagerauth

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/env"
)

func setupTest() {
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			Profile: "users_playground",
			Config: aws.Config{
				Region: aws.String(endpoints.EuWest1RegionID),
			},
			SharedConfigState: session.SharedConfigEnable,
		},
	))

	Configure(Config{
		AWSSession:               sess,
		AWSSecretsManagerAccount: env.GetAsString("ACCOUNT_ID", "633888256817"),
		AWSSecretsManagerRegion:  env.GetAsString("ACCOUNT_REGION", endpoints.EuWest1RegionID),
		Stage:                    env.GetAsString("STAGE", "sandbox"),
		SecretKey:                "user-credentials/measurement_service",
	})
}

func Test_SignInValid_HappyCase(t *testing.T) {
	setupTest()

	ctx := context.Background()
	err := SignIn(ctx)
	require.NoError(t, err)

	tokens1 := GetTokens()
	assert.NotEmpty(t, tokens1.AccessToken)

	err = SignIn(ctx)
	require.NoError(t, err)

	tokens2 := GetTokens()
	assert.Equal(t, tokens1.AccessToken, tokens2.AccessToken)
}

func Test_SignInExpired_HappyCase(t *testing.T) {
	setupTest()

	ctx := context.Background()

	err := SignIn(ctx)
	require.NoError(t, err)

	tokens1 := GetTokens()
	assert.NotEmpty(t, tokens1.AccessToken)

	// Override global tokens storage
	expiredToken := "eyJraWQiOiJ3TktkUUtMQURMdmVoRzR5V2h0RjRHZSsyUW9Rdm1DXC9vRzdFWkU0cVI2ND0iLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxZmM2MDM2Ni1hODc4LTQ5NzItYjBlNS04NTIxZGUyZGRkMGIiLCJldmVudF9pZCI6IjgzOTI0MTY4LTEzYTctNDM1Ny1iZGMyLTRjNmU1NjhjY2VhNCIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoiYXdzLmNvZ25pdG8uc2lnbmluLnVzZXIuYWRtaW4iLCJhdXRoX3RpbWUiOjE1OTA2NjA3NDcsImlzcyI6Imh0dHBzOlwvXC9jb2duaXRvLWlkcC5ldS13ZXN0LTEuYW1hem9uYXdzLmNvbVwvZXUtd2VzdC0xX0RnN3BURGpXYyIsImV4cCI6MTU5MDY2NDM0NywiaWF0IjoxNTkwNjYwNzQ3LCJqdGkiOiJiNjg3ODZkMy0xMWVhLTQ0NjItYWZkYy02NDE5OGE5OGZiZWQiLCJjbGllbnRfaWQiOiIxOThhaXEwdXBwajBtbzZjMjh1bHZ1aWYxYiIsInVzZXJuYW1lIjoibWVhc3VyZW1lbnRfc2VydmljZStzYW5kYm94QHNzby5zaGFyZWQtc2VydmljZXMuc2tmLmNvbSJ9.HEpFR7lMimsRe0YCdxZkeuABhpDQNEDvPUN3sHYwXz2-RaBrbNHyQK2IFnKP74QeoRKUSFwrHzwdWVif4NGj5val5ACm-DPxcGRGEJyueHMj5VMw9JBsU_ZG6g7G0DfzNO33AKurUzm-T-zy58jHKBsZnIloVl-6pMHZ_Pm7M4q6n_TmLOuIbQ-d6XMRN4wNFKLf9CRhtmZ18-dMvO9oysMW8Z8EmYg5_Yuu2yLVQOmKgBe2b8ctO-6YNgl07lfdBAaoR6SrBZ37GEl_zUFlWc05mqDdsXabeWLLcBBFts5yli6bP162DpuNnYQ69PWcKHNQc0HFYysI9Yq8sfhGNQ"
	tokens = auth.Tokens{
		AccessToken: expiredToken,
	}

	// Override global variable
	tokenExpireDurationDiff = -10 * time.Minute

	err = SignIn(ctx)
	require.NoError(t, err)

	tokens2 := GetTokens()
	assert.NotEqual(t, expiredToken, tokens2.AccessToken)
}
