// Package ochttp contains:
//
// - a propagation.HTTPFormat implementation for B3 and Datadog propagation
//
// The HTTPFormat is based on ochttp and ddtrace.
// - ochttp: https://github.com/census-instrumentation/opencensus-go/tree/master/plugin/ochttp/propagation/b3
// - ddtrace: https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/tracer/textmap.go
// See https://github.com/openzipkin/b3-propagation for more details on B3 propagation.
//
// Examples
//
// An example of using the propagator
//
//		func ListenAndServe(port string, handler http.Handler) error {
//			ocHandler := new(oc_http.Handler)
//			ocHandler.Handler = handler
//			ocHandler.Propagation = new(HTTPFormat)
//
//			server = http.Server{
//				Addr:    ":" + port,
//				Handler: ocHandler,
//			}
//
//			return server.ListenAndServe()
//		}
package ochttp
