package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	oc_http "go.opencensus.io/plugin/ochttp"
	dd_http "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/SKF/go-utility/v2/stages"
)

var config *Config

type Config struct {
	WithDatadogTracing    bool   // used when you trace your application with Datadog
	WithOpenCensusTracing bool   // default and used when you trace your application with Open Census
	ServiceName           string // needed when using lambda and Datadog for tracing
	Stage                 string
}

func Configure(conf Config) {
	conf.WithOpenCensusTracing = !conf.WithDatadogTracing
	config = &conf
}

func GetBaseURL() (string, error) {
	if config == nil {
		return "", errors.New("auth is not configured")
	}

	if !allowedStages[config.Stage] {
		return "", errors.Errorf("stage %s is not allowed", config.Stage)
	}

	if config.Stage == stages.StageProd {
		return "https://sso-api.users.enlight.skf.com", nil
	}

	return "https://sso-api." + config.Stage + ".users.enlight.skf.com", nil
}

// SignIn will sign in the user and if needed complete the change password challenge
func SignIn(ctx context.Context, username, password string) (tokens Tokens, err error) {
	var resp SignInResponse

	if resp, err = initiateSignIn(ctx, username, password); err != nil {
		err = errors.Wrap(err, "failed to initiate sign in")
		return
	}

	if resp.Data.Challenge.Type == "" {
		tokens = resp.Data.Tokens
		return
	}

	if resp, err = completeSignIn(ctx, resp.Data.Challenge, username, password); err != nil {
		err = errors.Wrap(err, "failed to complete sign in")
		return
	}

	return resp.Data.Tokens, nil
}

func SignInRefreshToken(ctx context.Context, refreshToken string) (Tokens, error) {
	const endpoint = "/sign-in/initiate"

	jsonBody := `{"refreshToken": "` + refreshToken + `"}`

	resp, err := signIn(ctx, endpoint, jsonBody)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to sign in with refreshtoken: %w", err)
	}

	tokens := Tokens{
		AccessToken:   resp.Data.Tokens.AccessToken,
		IdentityToken: resp.Data.Tokens.IdentityToken,
		RefreshToken:  resp.Data.Tokens.RefreshToken,
	}

	return tokens, nil
}

func initiateSignIn(ctx context.Context, username, password string) (signInResp SignInResponse, err error) {
	const endpoint = "/sign-in/initiate"

	jsonBody := `{"username": "` + username + `", "password": "` + password + `"}`

	return signIn(ctx, endpoint, jsonBody)
}

func completeSignIn(ctx context.Context, challenge Challenge, username, newPassword string) (signInResp SignInResponse, err error) {
	const endpoint = "/sign-in/complete"

	baseJSON := `{"username": "%s", "id": "%s", "type": "%s", "properties": {"newPassword": "%s"}}`
	jsonBody := fmt.Sprintf(baseJSON, username, challenge.ID, challenge.Type, newPassword)

	return signIn(ctx, endpoint, jsonBody)
}

func signIn(ctx context.Context, endpoint, jsonBody string) (signInResp SignInResponse, err error) {
	baseURL, err := GetBaseURL()
	if err != nil {
		err = errors.Wrap(err, "failed to get base URL")
		return
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewBufferString(jsonBody))
	if err != nil {
		err = errors.Wrap(err, "failed to create new HTTP request")
		return
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	var client = new(http.Client)
	if config.WithOpenCensusTracing {
		client.Transport = new(oc_http.Transport)
	}

	if config.WithDatadogTracing {
		client = withDatadogTracing(config.ServiceName, client)
	}

	resp, err := client.Do(req)
	if err != nil {
		err = errors.Wrap(err, "failed to execute HTTP request")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		if err = json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			err = errors.Wrap(err, "failed to decode Error response to JSON")
			return
		}

		err = errors.Errorf("StatusCode: %s, Error Message: %s \n", resp.Status, errorResp.Error.Message)

		return
	}

	if err = json.NewDecoder(resp.Body).Decode(&signInResp); err != nil {
		err = errors.Wrap(err, "failed to decode Sign In response to JSON")
		return
	}

	return signInResp, err
}

type SignInResponse struct {
	Data struct {
		Tokens    Tokens    `json:"tokens"`
		Challenge Challenge `json:"challenge"`
	} `json:"data"`
}

type Tokens struct {
	AccessToken   string `json:"accessToken"`
	IdentityToken string `json:"identityToken"`
	RefreshToken  string `json:"refreshToken"`
}

type Challenge struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

var allowedStages = map[string]bool{
	stages.StageProd:         true,
	stages.StageStaging:      true,
	stages.StageVerification: true,
	stages.StageTest:         true,
	stages.StageSandbox:      true,
}

func withDatadogTracing(serviceName string, client *http.Client) *http.Client {
	resourceNamer := func(req *http.Request) string {
		return fmt.Sprintf("%s %s", req.Method, req.URL.String())
	}

	var opts = []dd_http.RoundTripperOption{
		dd_http.RTWithResourceNamer(resourceNamer),
	}

	if serviceName != "" {
		opts = append(opts, dd_http.RTWithServiceName(serviceName))
	}

	return dd_http.WrapClient(client, opts...)
}
