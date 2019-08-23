package httpmiddleware

import (
	"context"
	"net/http"

	http_model "github.com/SKF/go-utility/http-model"
	http_server "github.com/SKF/go-utility/http-server"
	"github.com/SKF/go-utility/jwk"
	"github.com/SKF/go-utility/jwt"
	"github.com/SKF/go-utility/log"
	"github.com/SKF/proto/common"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

const (
	HeaderAuthorization = "Authorization"
)

type Users interface {
	GetUserIDByEmail(ctx context.Context, email string) (string, error)
}

// AuthenticateMiddleware retrieves the security configuration for the matched route
// and handles Access Token validation and stores the token claims in the request context.
func AuthenticateMiddleware(users Users, keySetURL string) mux.MiddlewareFunc {
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
				if err := handleAccessOrIDToken(users, req, secConfig.accessTokenHeader); err != nil {
					logFields.WithError(err).Warn("User is not authorized")
					http_server.WriteJSONResponse(ctx, w, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

type UserIDContextKey struct{}

const maxNumberOfCognitoUsers = 2 * (10 ^ 7)

var userIDs = make(map[string]string, maxNumberOfCognitoUsers)

func handleAccessOrIDToken(users Users, req *http.Request, header string) error {
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
		var exists bool
		email := claims.Username
		userID, exists = userIDs[email]
		if users != nil && !exists {
			if userID, err = users.GetUserIDByEmail(ctx, email); err != nil {
				return errors.Wrap(err, "couldn't get User ID by email")
			}
			userIDs[email] = userID
		}
	default:
		return errors.Errorf("invalid token use %s", claims.TokenUse)
	}

	ctx = context.WithValue(ctx, UserIDContextKey{}, userID)
	*req = *req.WithContext(ctx)

	return nil
}

// ExtractUserIDFromContext extracts User ID from a context.
func ExtractUserIDFromContext(ctx context.Context) (_ string, err error) {
	v := ctx.Value(UserIDContextKey{})
	if v == nil {
		err = errors.New("unable to parse User ID from context")
		return
	}
	return v.(string), nil
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
				WithTracing(ctx).
				WithField("method", req.Method).
				WithField("url", req.URL.String())

			secConfig := lookupSecurityConfig(req)
			if len(secConfig.authorizations) == 0 {
				span.End()
				next.ServeHTTP(w, req)
				return
			}

			userID, err := ExtractUserIDFromContext(req.Context())
			if err != nil {
				logFields.Error("Couldn't extract User ID from context.")
				http_server.WriteJSONResponse(ctx, w, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
				return
			}

			for _, authorizeConfig := range secConfig.authorizations {
				resource, err := authorizeConfig.resourceFunc(req)
				if err != nil {
					logFields.WithError(err).Error("ResourceFunc failed.")
					http_server.WriteJSONResponse(ctx, w, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
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
					http_server.WriteJSONResponse(ctx, w, http.StatusUnauthorized, http_model.ErrResponseUnauthorized)
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
