package ssoClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SKF/go-utility/v2/cmd/getTokens/model"
)

type Client struct {
	client http.Client
}

type SSOClient interface {
	SignInInitiate(config model.Config) (model.Tokens, error)
}

func (sso Client) SignInInitiate(config model.Config) (model.Tokens, error) {
	type initRequest struct {
		Username     string `json:"username"`
		RefreshToken string `json:"refreshToken"`
	}
	type initResponse struct {
		Data struct {
			Tokens model.Tokens
		}
	}
	body := initRequest{
		Username:     config.Username,
		RefreshToken: config.RefreshToken,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return model.Tokens{}, fmt.Errorf("failed to marshal initiate request: %w", err)
	}

	url := fmt.Sprintf("%s/sign-in/initiate", config.SSOURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return model.Tokens{}, fmt.Errorf("failed to post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return model.Tokens{}, fmt.Errorf("failed to sign in got %d status code", resp.StatusCode)
	}

	response := initResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response.Data.Tokens, err
}
