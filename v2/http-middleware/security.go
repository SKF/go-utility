package httpmiddleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"github.com/SKF/go-utility/auth"
	http_model "github.com/SKF/go-utility/http-model"
	http_server "github.com/SKF/go-utility/http-server"
	"github.com/SKF/go-utility/jwk"
	"github.com/SKF/go-utility/jwt"
	"github.com/SKF/go-utility/log"
	"github.com/SKF/go-utility/useridcontext"
	"github.com/SKF/proto/common"
)

const (
	HeaderAuthorization = "Authorization"
)

type Config = auth.Config

func Configure(conf Config) {
	auth.Configure(conf)
}

// AuthenticateMiddleware retrieves the security configuration for the matched route
// and handles Access Token validation and stores the token claims in the request context.
func AuthenticateMiddleware(keySetURL string) mux.MiddlewareFunc {
	jwk.KeySetURL = keySetURL
	if err := jwk.RefreshKeySets(); err != nil {
		log.
			WithError(err).
			Error("Couldn't refresh JWKeySets")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := trace.StartSpan(req.Context(), "Authenticator")
			defer span.End()
			*req = *req.WithContext(ctx)

			logFields := log.
				WithTracing(ctx).
				WithField("method", req.Method).
				WithField("url", req.URL.String())

			secConfig := lookupSecurityConfig(req)
			if secConfig.accessTokenHeader != "" {
				if err := handleAccessOrIDToken(req, secConfig.accessTokenHeader); err != nil {
					logFields.WithError(err).Warn("User is not authorized")
					http_server.WriteJSONResponse(ctx, w, req, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func handleAccessOrIDToken(req *http.Request, header string) error {
	ctx := req.Context()

	base64Token := req.Header.Get(header)
	if base64Token == "" {
		return errors.Errorf("auth header [%s] was empty", header)
	}

	token, err := jwt.Parse(base64Token)
	if err != nil {
		return errors.Wrap(err, "authorization token not valid")
	}

	var userID string

	claims := token.GetClaims()
	switch claims.TokenUse {
	case jwt.TokenUseID:
		userID = claims.EnlightUserID
	case jwt.TokenUseAccess:
		if userID, err = getUserIDByToken(ctx, base64Token); err != nil {
			return errors.Wrap(err, "couldn't get User by token")
		}
	default:
		return errors.Errorf("invalid token use %s", claims.TokenUse)
	}

	ctx = useridcontext.NewContext(ctx, userID)
	*req = *req.WithContext(ctx)

	return nil
}

func getUserIDByToken(ctx context.Context, accessToken string) (_ string, err error) {
	const endpoint = "/users/me"

	baseURL, err := auth.GetBaseURL()
	if err != nil {
		err = errors.Wrap(err, "failed to get base URL")
		return
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		err = errors.Wrap(err, "failed to create new HTTP request")
		return
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", accessToken)

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

		err = errors.Errorf("StatusCode: %s, Error Message: %s \n", resp.Status, errorResp.Error.Message)

		return
	}

	var myUserResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&myUserResp); err != nil {
		err = errors.Wrap(err, "failed to decode My User response to JSON")
		return
	}

	return myUserResp.Data.ID, err
}

type Authorizer interface {
	IsAuthorizedWithContext(ctx context.Context, userID, action string, resource *common.Origin) (bool, error)
}

// AuthorizeMiddleware retrieves the security configuration for the matched route
// and handles the configured authorizations.
func AuthorizeMiddleware(authorizer Authorizer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := trace.StartSpan(req.Context(), "Authorizer")
			defer span.End()
			req = req.WithContext(ctx)

			logFields := log.
				WithTracing(ctx).
				WithField("method", req.Method).
				WithField("url", req.URL.String())

			secConfig := lookupSecurityConfig(req)
			if len(secConfig.authorizations) == 0 {
				span.End()
				next.ServeHTTP(w, req)
				return
			}

			userID, ok := useridcontext.FromContext(req.Context())
			if !ok {
				logFields.Error("Couldn't extract User ID from context.")
				http_server.WriteJSONResponse(ctx, w, req, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
				return
			}

			logFields = logFields.WithUserID(ctx)

			for _, authorizeConfig := range secConfig.authorizations {
				resource, err := authorizeConfig.resourceFunc(req)
				if err != nil {
					logFields.WithError(err).Error("ResourceFunc failed.")
					http_server.WriteJSONResponse(ctx, w, req, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
					return
				}

				ok, err := authorizer.IsAuthorizedWithContext(
					ctx,
					userID,
					authorizeConfig.action,
					resource,
				)
				if !ok || err != nil {
					logFields.
						WithField("userId", userID).
						WithField("action", authorizeConfig.action).
						WithField("resource", resource).
						Warn("User is not Authorized")
					http_server.WriteJSONResponse(ctx, w, req, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

var securityConfigurations []*SecurityConfig

func lookupSecurityConfig(req *http.Request) (_ SecurityConfig) {
	route := mux.CurrentRoute(req)
	if route == nil {
		return
	}

	if pathTemplate, err := route.GetPathTemplate(); err == nil {
		for _, config := range securityConfigurations {
			if config.endpoint == pathTemplate {
				for _, method := range config.methods {
					if method == req.Method {
						return *config
					}
				}
			}
		}
	}

	return
}

// SecurityConfig represents how to authenticate and authorize a given endpoint and method.
type SecurityConfig struct {
	endpoint          string
	methods           []string
	accessTokenHeader string
	authorizations    []authorizationConfig
}

type authorizationConfig struct {
	action       string
	resourceFunc ResourceFunc
}

// HandleSecureEndpoint creates a new SecurityConfig for the specified endpoint.
func HandleSecureEndpoint(endpoint string) *SecurityConfig {
	s := &SecurityConfig{endpoint: endpoint}
	securityConfigurations = append(securityConfigurations, s)
	return s
}

// Methods adds methods to the SecurityConfig.
func (s *SecurityConfig) Methods(methods ...string) *SecurityConfig {
	s.methods = methods
	return s
}

// AccessToken adds Access Token as a mean for Authentication to the SecurityConfig.
// The header defaults to "Authorization".
func (s *SecurityConfig) AccessToken(headers ...string) *SecurityConfig {
	s.accessTokenHeader = HeaderAuthorization
	if len(headers) > 0 {
		s.accessTokenHeader = headers[0]
	}
	return s
}

// ResourceFunc takes a *http.Request and returns the resource to use for authorization.
type ResourceFunc func(*http.Request) (*common.Origin, error)

// NilResourceFunc represents the Zero Value ResourceFunc.
var NilResourceFunc = func(req *http.Request) (*common.Origin, error) {
	return nil, nil
}

// Authorize adds an Authorization Configuration to the SecurityConfig.
func (s *SecurityConfig) Authorize(action string, resourceFunc ResourceFunc) *SecurityConfig {
	s.authorizations = append(
		s.authorizations,
		authorizationConfig{action, resourceFunc},
	)
	return s
}
