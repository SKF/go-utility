// Package httpmiddleware contains middleware for REST API's built with Gorilla web toolkit (router) and OpenCensus (telemetry).
//
// The package is using on
// "github.com/gorilla/mux"
// "go.opencensus.io/trace"
//
// Examples
//
// An example including creating a router, adding a route and security as well as all middleware.
//   router := mux.NewRouter()
//
//   const pathToCreateCompanyUser = "/companies/{companyID:[a-zA-Z0-9-]+}/users"
//   router.
//       HandleFunc(pathToCreateUser, http_middleware.ContentType(
//           server.createCompanyUserHandler, http_model.MimeJSON,
//       )).
//       Methods(http.MethodPost)
//
//   router.
//       HandleFunc(pathToCreateUser, http_middleware.Options(
//           []string{http.MethodPost},
//           []string{http_model.HeaderContentType},
//       )).
//       Methods(http.MethodOptions)
//
//   http_middleware.
//       HandleSecureEndpoint(pathToCreateCompanyUser).
//       Methods(http.MethodPost).
//       AccessToken().
//       Authorize(ActionIAMCreateUser, http_middleware.NilResourceFunc).
//       Authorize(ActionIAMInviteUser, companyOriginFromPathFunc)
//
//   router.Use(
//       // Middleware is run from top to bottom, order is important
//       http_middleware.TrailingSlashMiddleware,
//       http_middleware.CorsMiddleware,
//       http_middleware.OpenCensusMiddleware,
//       http_middleware.AuthenticateMiddleware("<jwkeyset_url>"),
//       http_middleware.AuthorizeMiddleware(authorizerClient),
//   )
package httpmiddleware
