package httpmiddleware

import (
	"context"
	"net/http"

	rest "github.com/SKF/go-rest-utility/client"
	"github.com/SKF/proto/v2/common"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SKF/go-utility/v2/accesstokensubcontext"
	"github.com/SKF/go-utility/v2/auth"
	"github.com/SKF/go-utility/v2/http-middleware/util"
	http_model "github.com/SKF/go-utility/v2/http-model"
	http_server "github.com/SKF/go-utility/v2/http-server"
	"github.com/SKF/go-utility/v2/jwk"
	"github.com/SKF/go-utility/v2/jwt"
	"github.com/SKF/go-utility/v2/log"
	"github.com/SKF/go-utility/v2/useridcontext"
)

const (
	HeaderAuthorization = "Authorization"
)

type Config struct {
	Stage string

	// Configures the usage of a User ID Cache when using an Access Token
	UseUserIDCache bool
	Client         *rest.Client
}

type ResponseConfig interface {
	InternalErrorResponse() []byte
	UnauthenticateResponse() []byte
	UnauthorizedResponse() []byte
}

var (
	config      Config
	userIDCache map[string]string
	client      *rest.Client
)

func Configure(conf Config) {
	config = conf

	auth.Configure(auth.Config{Stage: conf.Stage})
	jwk.Configure(jwk.Config{Stage: conf.Stage})

	if conf.UseUserIDCache {
		userIDCache = map[string]string{}
	}

	if client = conf.Client; client == nil {
		url, err := auth.GetBaseURL()
		if err != nil {
			panic(err)
		}

		client = rest.NewClient(
			rest.WithBaseURL(url),
			rest.WithOpenCensusTracing(),
		)
	}
}

// AuthenticateMiddleware retrieves the security configuration for the matched route
// and handles Access Token validation and stores the token claims in the request context.
// Deprecated: Use AuthenticateMiddlewareV3() instead
func AuthenticateMiddleware(keySetURL string) mux.MiddlewareFunc {
	jwk.KeySetURL = keySetURL
	return AuthenticateMiddlewareV3()
}

// AuthenticateMiddlewareV3 retrieves the security configuration for the matched route
// and handles Access Token validation and stores the token claims in the request context.
func AuthenticateMiddlewareV3() mux.MiddlewareFunc {
	if err := jwk.RefreshKeySets(); err != nil {
		log.WithError(err).
			Error("Couldn't refresh JWKeySets")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := util.StartSpanNoRoot(req.Context(), "AuthenticateMiddlewareV3/Handler")
			defer span.End()

			secConfig := lookupSecurityConfig(req)
			if secConfig.accessTokenHeader != "" {
				if err := handleAccessOrIDToken(ctx, req, secConfig.accessTokenHeader); err != nil {
					responseBody := GetUnauthenticedErrorResponseBody(http_model.ErrResponseUnauthorized, secConfig)
					writeAndLogResponse(ctx, w, req, http.StatusUnauthorized, responseBody)

					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func handleAccessOrIDToken(ctx context.Context, req *http.Request, header string) error {
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
		if config.UseUserIDCache {
			var found bool
			if userID, found = userIDCache[claims.Subject]; found {
				break
			}
		}

		if userID, err = getUserIDByToken(ctx, base64Token); err != nil {
			return errors.Wrap(err, "couldn't get User by token")
		}

		if config.UseUserIDCache {
			userIDCache[claims.Subject] = userID
		}
	default:
		return errors.Errorf("invalid token use %s", claims.TokenUse)
	}

	ctx = accesstokensubcontext.NewContext(ctx, claims.Subject)
	ctx = useridcontext.NewContext(ctx, userID)
	*req = *req.WithContext(ctx)

	return nil
}

func getUserIDByToken(ctx context.Context, accessToken string) (_ string, err error) {
	req := rest.Get("/users/me").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", accessToken)

	resp, err := client.Do(ctx, req)
	if err != nil {
		err = errors.Wrap(err, "failed to execute HTTP request")
		return
	}
	defer resp.Body.Close()

	var myUserResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err = resp.Unmarshal(&myUserResp); err != nil {
		err = errors.Wrap(err, "failed to decode My User response to JSON")
		return
	}

	return myUserResp.Data.ID, err
}

type Authorizer interface {
	IsAuthorizedWithContext(ctx context.Context, userID, action string, resource *common.Origin) (bool, error)
}

// AuthorizeMiddleware retrieves the security configuration for the matched
// route and handles the configured authorizations. If any of the configured
// ResourceFuncs returns a HTTPError or an error wrapping a HTTPError, the error
// code and message from that error is written. Other errors from the
// ResourceFuncs results in a http.StatusInternalServerError response being
// written. If the request fails the authorization check,
// http.StatusUnauthorized is returned to the client.
func AuthorizeMiddleware(authorizer Authorizer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := util.StartSpanNoRoot(req.Context(), "AuthorizeMiddleware/Handler")
			defer span.End()

			// If current route doesn't need to be authenicated
			secConfig := lookupSecurityConfig(req)
			if len(secConfig.authorizations) == 0 {
				next.ServeHTTP(w, req)
				return
			}

			userID, ok := useridcontext.FromContext(req.Context())
			if !ok {
				responseBody := GetInternalServerErrorResponseBody(http_model.ErrResponseInternalServerError, secConfig)
				writeAndLogResponse(ctx, w, req, http.StatusInternalServerError, responseBody)

				return
			}

			isAuthorized, err := checkAuthorization(ctx, req, authorizer, userID, secConfig.authorizations)
			var httpErr *http_model.HTTPError
			if errors.As(err, &httpErr) {
				if secConfig.responses != nil {
					writeAndLogResponse(ctx, w, req, http.StatusInternalServerError, secConfig.responses.InternalErrorResponse())
				} else {
					writeAndLogResponse(ctx, w, req, httpErr.StatusCode, httpErr.Message())
				}

				return
			}

			if err != nil {
				responseBody := GetInternalServerErrorResponseBody(http_model.ErrResponseInternalServerError, secConfig)
				writeAndLogResponse(ctx, w, req, http.StatusInternalServerError, responseBody)

				return
			}

			if !isAuthorized {
				responseBody := GetUnauthorizedErrorResponseBody(http_model.ErrResponseUnauthorized, secConfig)
				writeAndLogResponse(ctx, w, req, http.StatusUnauthorized, responseBody)

				return
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func writeAndLogResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, body []byte) {
	if status == http.StatusInternalServerError {
		log.
			WithTracing(ctx).
			WithUserID(ctx).
			WithField("method", r.Method).
			WithField("url", r.URL.String()).
			Error("AuthenticateMiddlewareV3 returned an 500 error")
	}

	http_server.WriteJSONResponse(ctx, w, r, status, body)
}

func checkAuthorization(ctx context.Context, req *http.Request, authorizer Authorizer, userID string, configuredAuthorizations []authorizationConfig) (bool, error) {
	logFields := log.
		WithTracing(ctx).
		WithUserID(ctx).
		WithField("method", req.Method).
		WithField("url", req.URL.String())

	for _, authorizeConfig := range configuredAuthorizations {
		resource, err := authorizeConfig.resourceFunc(req)
		if err != nil {
			return false, err
		}

		ok, err := authorizer.IsAuthorizedWithContext(
			ctx,
			userID,
			authorizeConfig.action,
			resource,
		)
		if err != nil {
			if grpcStatus := status.Code(err); grpcStatus == codes.Canceled {
				return false, nil
			}

			return false, err
		}

		if !ok {
			logFields.
				WithField("userId", userID).
				WithField("action", authorizeConfig.action).
				WithField("resource", resource).
				Debug("User is not Authorized")

			return false, nil
		}
	}

	return true, nil
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
	responses         ResponseConfig
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

func HandleSecureEndpointCustomErrorResponse(endpoint string, responses ResponseConfig) *SecurityConfig {
	s := &SecurityConfig{
		endpoint:  endpoint,
		responses: responses,
	}
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
// If the ResourceFunc fails because of invalid input data or a missing resource,
// return a HttpError, or an error wrapping a HTTPError.
// The following example ResourceFunc expects an input struct with a non-empty field
//
//     func fieldFromBodyFunc(r *http.Request) (*common.Origin, error) {
//         var inputData struct {
//             field string `json:"field,omitempty"`
//         }
//         body, err := ioutil.ReadAll(r.Body)
//         if err != nil {
//             return nil, err
//         }
//         r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
//         if err := json.Unmarshal(body, &inputData); err != nil {
//             return nil, &http_model.HTTPError{
//                 Msg:        "Failed to unmarshal body",
//                 StatusCode: http.StatusBadRequest,
//             }
//         }
//         if inputData.field == "" || uuid.UUID(inputData.field) == uuid.EmptyUUID {
//             return nil, &http_model.HTTPError{
//                 Msg:        "Required field 'field' is empty",
//                 StatusCode: http.StatusBadRequest,
//             }
//         }
//         return &common.Origin{Id: inputData.field, Type: "example"}, nil
//     }
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

func GetInternalServerErrorResponseBody(defaultResponse []byte, secConfig SecurityConfig) []byte {
	responsebody := defaultResponse

	if secConfig.responses != nil && len(secConfig.responses.InternalErrorResponse()) > 0 {
		responsebody = secConfig.responses.InternalErrorResponse()
	}

	return responsebody
}

func GetUnauthenticedErrorResponseBody(defaultResponse []byte, secConfig SecurityConfig) []byte {
	responsebody := defaultResponse

	if secConfig.responses != nil && len(secConfig.responses.UnauthenticateResponse()) > 0 {
		responsebody = secConfig.responses.UnauthenticateResponse()
	}

	return responsebody
}

func GetUnauthorizedErrorResponseBody(defaultResponse []byte, secConfig SecurityConfig) []byte {
	responsebody := defaultResponse

	if secConfig.responses != nil && len(secConfig.responses.UnauthorizedResponse()) > 0 {
		responsebody = secConfig.responses.UnauthorizedResponse()
	}

	return responsebody
}
