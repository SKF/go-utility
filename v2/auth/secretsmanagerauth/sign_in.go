package secretsmanagerauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/SKF/go-utility/v2/auth"
)

var tokensMutex = new(sync.RWMutex)
var tokens auth.Tokens
var tokenExpireDurationDiff = 5 * time.Minute

var fetchingTokensMutex = new(sync.RWMutex)
var fetchingTokens bool

var config *Config

// Config is the configuration of the package
type Config struct {
	WithDatadogTracing       bool   // used when you trace your application with Datadog
	WithOpenCensusTracing    bool   // default and used when you trace your application with Open Census
	ServiceName              string // needed when using lambda and Datadog for tracing
	AWSConfig                aws.Config
	AWSSecretsManagerAccount string
	AWSSecretsManagerRegion  string
	SecretKey                string
	Stage                    string
}

// Configure will configure the package
func Configure(conf Config) {
	conf.WithOpenCensusTracing = !conf.WithDatadogTracing
	config = &conf

	auth.Configure(auth.Config{
		WithDatadogTracing:    conf.WithDatadogTracing,
		WithOpenCensusTracing: conf.WithOpenCensusTracing,
		ServiceName:           conf.ServiceName,
		Stage:                 conf.Stage,
	})
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

	if auth.IsTokenValid(tokens.AccessToken, tokenExpireDurationDiff) {
		return nil
	}

	tokens, err = signIn(ctx)
	if err != nil {
		tokens = auth.Tokens{}
		return err
	}

	return nil
}

func signIn(ctx context.Context) (tokens auth.Tokens, err error) {
	svc := secretsmanager.NewFromConfig(config.AWSConfig)

	secretKey := "arn:aws:secretsmanager:" + config.AWSSecretsManagerRegion + ":" + config.AWSSecretsManagerAccount + ":secret:" + config.SecretKey

	output, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &secretKey})
	if err != nil {
		err = fmt.Errorf("failed to get secret value: %w", err)
		return
	}

	var secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(output.SecretBinary, &secret); err != nil {
		err = fmt.Errorf("failed to unmarshal secret value: %w", err)
		return
	}

	if tokens, err = auth.SignIn(ctx, secret.Username, secret.Password); err != nil {
		err = fmt.Errorf("failed to sign in; %w", err)
		return
	}

	return tokens, err
}
