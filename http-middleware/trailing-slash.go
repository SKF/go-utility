package httpmiddleware

import (
	"net/http"
	"strings"
)

// TrailingSlashMiddleware removes trailing slash from URL's
func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
