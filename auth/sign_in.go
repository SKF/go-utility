package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
)

const stageProd = "prod"

var config *Config

type Config struct {
	Stage string
}

func Configure(conf Config) {
	config = &conf
}

func GetBaseURL() (string, error) {
	if config == nil {
		return "", errors.New("auth is not configured")
	}

	if config.Stage == stageProd {
		return "https://sso-api.users.enlight.skf.com", nil
	}

	const ssoBaseURL = "https://sso-api.%s.users.enlight.skf.com"
	url := fmt.Sprintf(ssoBaseURL, config.Stage)

	return url, nil
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

	client := &http.Client{Transport: &ochttp.Transport{}}

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

		err = errors.Errorf("StatusCode: %s, Error Message: %s \n", resp.Status, errorResp.Error.Message) //nolint: revive

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
