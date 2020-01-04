package httpmiddleware

import (
	"net/http"
	"strings"
)

// CorsMiddleware adds Access-Control-Allow-Origin header to responses.
func CorsMiddleware(next http.Handler) http.Handler {
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
