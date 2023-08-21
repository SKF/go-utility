package cachedauth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SKF/go-utility/v2/auth"
)

var lock sync.RWMutex

var config *Config
var tokens map[string]auth.Tokens

const latest = "latest"

// Config is the configuration of the package
type Config struct {
	WithDatadogTracing    bool
	WithOpenCensusTracing bool   // default
	ServiceName           string // needed when using lambda and Datadog for tracing
	Stage                 string
}

// Configure will configure the package
// please do not use go-utility/v2/auth together with this package
func Configure(conf Config) {
	lock.Lock()
	defer lock.Unlock()

	conf.WithOpenCensusTracing = !conf.WithDatadogTracing
	config = &conf

	if tokens == nil {
		tokens = map[string]auth.Tokens{}
	}

	auth.Configure(auth.Config{
		WithDatadogTracing:    conf.WithDatadogTracing,
		WithOpenCensusTracing: conf.WithOpenCensusTracing,
		ServiceName:           conf.ServiceName,
		Stage:                 conf.Stage,
	})
}

// GetTokens will return the cached tokens
//
// note: Does not refresh the tokens
func GetTokens() auth.Tokens {
	return GetTokensByUser(latest)
}

// GetTokens will return the cached tokens
//
// note: Does not refresh the tokens
func GetTokensByUser(username string) auth.Tokens {
	lock.RLock()
	defer lock.RUnlock()

	return tokens[username]
}

// SignIn is thread safe and only returns new tokens if the old tokens are about to expire
func SignIn(ctx context.Context, username, password string) error {
	lock.Lock()
	defer lock.Unlock()

	if config == nil {
		return fmt.Errorf("cachedauth is not configured")
	}

	const tokenExpireDurationDiff = 5 * time.Minute

	oldTokens := tokens[username]
	if auth.IsTokenValid(oldTokens.AccessToken, tokenExpireDurationDiff) {
		return nil
	}

	newtokens, err := auth.SignIn(ctx, username, password)
	if err != nil {
		return err
	}

	tokens[username] = newtokens
	tokens[latest] = newtokens

	return nil
}
