package secretsmanagerauth

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"

	"github.com/SKF/go-utility/v2/auth"
)

var tokensMutex = new(sync.RWMutex)
var tokens auth.Tokens

var fetchingTokensMutex = new(sync.RWMutex)
var fetchingTokens bool

var config *Config

// Config is the configuration of the package
type Config struct {
	AWSSession               *session.Session
	AWSSecretsManagerAccount string
	AWSSecretsManagerRegion  string
	SecretKey                string
	Stage                    string
}

// Configure will configure the package
func Configure(conf Config) {
	config = &conf

	auth.Configure(auth.Config{Stage: conf.Stage})
}

// GetTokens will return the cached tokens
func GetTokens() auth.Tokens {
	tokensMutex.RLock()
	defer tokensMutex.RUnlock()

	return tokens
}

// SignIn will fetch credentials from the Secret Manager and Sign In using those credentials
func SignIn(ctx context.Context) (err error) {
	if config == nil {
		return errors.New("secretsmanagerauth is not configured")
	}

	// handle multiple concurrent calls to secretsmanagerlogin.SignIn
	fetchingTokensMutex.RLock()
	if fetchingTokens {
		fetchingTokensMutex.RUnlock()
		return nil
	}
	fetchingTokensMutex.RUnlock()

	// will make calls to secretsmanagerlogin.GetTokens to wait for secretsmanagerlogin.SignIn to finish
	tokensMutex.Lock()
	defer tokensMutex.Unlock()

	fetchingTokensMutex.Lock()
	fetchingTokens = true
	fetchingTokensMutex.Unlock()

	defer func() {
		fetchingTokensMutex.Lock()
		fetchingTokens = false
		fetchingTokensMutex.Unlock()
	}()

	tokens, err = signIn(ctx)

	return
}

func signIn(ctx context.Context) (tokens auth.Tokens, err error) {
	svc := secretsmanager.New(config.AWSSession)

	secretKey := "arn:aws:secretsmanager:" + config.AWSSecretsManagerRegion + ":" + config.AWSSecretsManagerAccount + ":secret:" + config.SecretKey

	output, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: &secretKey})
	if err != nil {
		err = errors.Wrap(err, "failed to get secret value")
		return
	}

	var secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(output.SecretBinary, &secret); err != nil {
		err = errors.Wrap(err, "failed to unmarshal secret value")
		return
	}

	if tokens, err = auth.SignIn(ctx, secret.Username, secret.Password); err != nil {
		err = errors.Wrap(err, "failed to sign in")
		return
	}

	return tokens, err
}
