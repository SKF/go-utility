package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
		return "", fmt.Errorf("auth is not configured")
	}

	if !allowedStages[config.Stage] {
		return "", fmt.Errorf("stage %s is not allowed", config.Stage)
	}

	if config.Stage == stages.StageProd {
		return "https://sso-api.users.enlight.skf.com", nil
	}

	return "https://sso-api." + config.Stage + ".users.enlight.skf.com", nil
}

// SignIn will sign in the user and if needed complete the change password challenge
func SignIn(ctx context.Context, username, password string) (Tokens, error) {
	resp, err := initiateSignIn(ctx, username, password)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to initiate sign in: %w", err)
	}

	if resp.Data.Challenge.Type == "" {
		return resp.Data.Tokens, nil
	}

	resp, err = completeSignIn(ctx, resp.Data.Challenge, username, password)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to complete sign in: %w", err)
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

func completeSignIn(ctx context.Context, challenge Challenge, username, newPassword string) (SignInResponse, error) {
	const endpoint = "/sign-in/complete"

	baseJSON := `{"username": "%s", "id": "%s", "type": "%s", "properties": {"newPassword": "%s"}}`
	jsonBody := fmt.Sprintf(baseJSON, username, challenge.ID, challenge.Type, newPassword)

	return signIn(ctx, endpoint, jsonBody)
}

func signIn(ctx context.Context, endpoint, jsonBody string) (SignInResponse, error) {
	baseURL, err := GetBaseURL()
	if err != nil {
		return SignInResponse{}, fmt.Errorf("failed to get base URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewBufferString(jsonBody))
	if err != nil {
		return SignInResponse{}, fmt.Errorf("failed to create new HTTP request: %w", err)
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
		return SignInResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		if err = json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return SignInResponse{}, fmt.Errorf("failed to decode Error response to JSON: %w", err)
		}

		return SignInResponse{}, fmt.Errorf("status code: %s, error message: %s", resp.Status, errorResp.Error.Message)
	}

	var signInResp SignInResponse

	if err = json.NewDecoder(resp.Body).Decode(&signInResp); err != nil {
		return SignInResponse{}, fmt.Errorf("failed to decode Sign In response to JSON: %w", err)
	}

	return signInResp, nil
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
