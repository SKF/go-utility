package httpmiddleware

import (
	"net/http"
	"strings"
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

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", methodsJoined)
		w.Header().Set("Access-Control-Allow-Headers", headersJoined)
	}
}
