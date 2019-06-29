package httpmiddleware

import (
	"context"
	"net/http"

	http_model "github.com/SKF/go-utility/http-model"
	"github.com/SKF/go-utility/jwk"
	"github.com/SKF/go-utility/jwt"
	"github.com/SKF/go-utility/log"
	"github.com/SKF/proto/common"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type claimsContextKey struct{}

const (
	HeaderAuthorization = "Authorization"
)

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
				WithField("method", req.Method).
				WithField("url", req.URL.String())

			secConfig := lookupSecurityConfig(req)
			if secConfig.accessTokenHeader != "" {
				if err := handleAccessToken(req, secConfig.accessTokenHeader); err != nil {
					logFields.WithError(err).Warn("User is not authorized")
					http_model.WriteJSONResponse(w, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func handleAccessToken(req *http.Request, header string) error {
	base64Token := req.Header.Get(header)
	if base64Token == "" {
		return errors.Errorf("auth header [%s] was empty", header)
	}

	token, err := jwt.Parse(base64Token)
	if err != nil {
		return errors.Wrap(err, "authorization token not valid")
	}

	*req = *req.WithContext(
		context.WithValue(req.Context(), claimsContextKey{}, token.GetClaims()),
	)
	return nil
}

// ExtractClaimsFromContext extracts JWT claims from a context.
func ExtractClaimsFromContext(ctx context.Context) (_ jwt.Claims, err error) {
	v := ctx.Value(claimsContextKey{})
	if v == nil {
		err = errors.New("unable to parse Claims from context")
		return
	}

	claims := v.(jwt.Claims)
	return claims, nil
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
			*req = *req.WithContext(ctx)

			logFields := log.
				WithField("method", req.Method).
				WithField("url", req.URL.String())

			secConfig := lookupSecurityConfig(req)
			if len(secConfig.authorizations) == 0 {
				span.End()
				next.ServeHTTP(w, req)
				return
			}

			claims, err := ExtractClaimsFromContext(req.Context())
			if err != nil {
				logFields.Error("Couldn't extract claims from context.")
				http_model.WriteJSONResponse(w, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
				return
			}

			for _, authorizeConfig := range secConfig.authorizations {
				resource, err := authorizeConfig.resourceFunc(req)
				if err != nil {
					logFields.WithError(err).Error("ResourceFunc failed.")
					http_model.WriteJSONResponse(w, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
					return
				}

				ok, err := authorizer.IsAuthorizedWithContext(
					ctx,
					claims.UserID,
					authorizeConfig.action,
					resource,
				)
				if !ok || err != nil {
					logFields.
						WithField("userId", claims.UserID).
						WithField("action", authorizeConfig.action).
						WithField("resource", resource).
						Warn("User is not Authorized")
					http_model.WriteJSONResponse(w, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
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
