package httpmiddleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opencensus.io/trace"
)

// OpenCensusMiddleware adds request method and path template as span name.
func OpenCensusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		span := trace.FromContext(req.Context())
		if span == nil {
			next.ServeHTTP(w, req)
			return
		}

		route := mux.CurrentRoute(req)
		if route == nil {
			next.ServeHTTP(w, req)
			return
		}

		if name, err := route.GetPathTemplate(); err == nil {
			span.SetName(req.Method + " " + name)
		}

		next.ServeHTTP(w, req)
	})
}
