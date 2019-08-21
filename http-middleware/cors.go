package httpmiddleware

import (
	"net/http"
	"strings"

	http_model "github.com/SKF/go-utility/http-model"
	http_server "github.com/SKF/go-utility/http-server"
	"github.com/SKF/go-utility/log"
)

// CorsMiddleware adds CORS headers to requests.
// Shouldn't be use, instead use the combination of helper
// functions below.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT, POST, GET, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CorsMiddleware adds CORS Origin header to responses.
func CorsMiddlewareV2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	})
}

// Options takes a list of methods and headers and returns an Options HandlerFunc
func Options(methods, headers []string) http.HandlerFunc {
	methodsJoined := strings.Join(methods, ", ")
	headersJoined := strings.Join(headers, ", ")

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", methodsJoined)
		w.Header().Set("Access-Control-Allow-Headers", headersJoined)
	}
}

// ContentType wraps a HandlerFunc and checks the incoming
// content-type with a list of allowed content types.
func ContentType(next http.HandlerFunc, contentTypes ...string) http.HandlerFunc {
	var validContentTypes = make(map[string]bool)
	for _, contentType := range contentTypes {
		validContentTypes[contentType] = true
	}

	return func(w http.ResponseWriter, req *http.Request) {
		reqContentType := req.Header.Get("Content-Type")
		if validContentTypes[reqContentType] {
			next(w, req)
			return
		}

		ctx := req.Context()
		log.WithTracing(ctx).WithField("contentType", reqContentType).Warn("Unsupported Content-Type")
		http_server.WriteJSONResponse(ctx, w, http.StatusUnsupportedMediaType, http_model.ErrResponseUnsupportedMediaType)
	}
}
