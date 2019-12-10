package secretsmanagerlogin

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"

	"github.com/SKF/go-utility/auth"
)

var tokensMutex = new(sync.RWMutex)
var tokens auth.Tokens

var fetchingTokensMutex = new(sync.RWMutex)
var fetchingTokens bool

var config *Config

// Config is the configuration of the package
type Config struct {
	AWSSession *session.Session
	SecretKey  string
	Stage      string
}

// Configure will configure the package
func Configure(conf Config) {
	config = &conf
	auth.Configure(auth.Config{Stage: conf.Stage}) //nolint: wsl
}

// GetTokens will return the cached tokens
func GetTokens(ctx context.Context) auth.Tokens {
	tokensMutex.RLock()
	defer tokensMutex.RUnlock()

	return tokens
}

// SignIn will fetch credentials from the Secret Manager and Sign In using those credentials
func SignIn(ctx context.Context) error {
	if config == nil {
		return errors.New("secretmanagerlogin is not configured")
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

	return signIn(ctx)
}

func signIn(ctx context.Context) error {
	svc := secretsmanager.New(config.AWSSession)

	output, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: &config.SecretKey})
	if err != nil {
		return errors.Wrap(err, "failed to get secret value")
	}

	var secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(output.SecretBinary, &secret); err != nil {
		return errors.Wrap(err, "failed to unmarshal secret value")
	}

	if tokens, err = auth.SignIn(ctx, secret.Username, secret.Password); err != nil {
		return errors.Wrap(err, "failed to sign in")
	}

	return nil
}
