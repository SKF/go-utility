package httpmiddleware

import (
	"fmt"
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

		for key, value := range mux.Vars(req) {
			span.AddAttributes(trace.StringAttribute(fmt.Sprintf("vars.%s", key), value))
		}

		for key, values := range req.URL.Query() {
			key = fmt.Sprintf("query.%s", key)
			switch len(values) {
			case 0:
				continue
			case 1: // nolint: gomnd
				span.AddAttributes(trace.StringAttribute(key, values[0]))
			default:
				for i := range values {
					span.AddAttributes(trace.StringAttribute(fmt.Sprintf("%s.%d", key, i), values[i]))
				}
			}
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
