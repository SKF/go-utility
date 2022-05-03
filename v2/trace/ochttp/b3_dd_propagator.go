package ochttp

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"strconv"

	oc_trace "go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"

	"github.com/SKF/go-utility/v2/trace"
)

// HTTPFormat implements propagation.HTTPFormat to propagate
// traces in HTTP headers in B3 and Datadog propagation format.
// HTTPFormat skips the X-B3-ParentId and X-B3-Flags headers
// because there are additional fields not represented in the
// OpenCensus span context. Spans created from the incoming
// header will be the direct children of the client-side span.
// Similarly, receiver of the outgoing spans should use client-side
// span created by OpenCensus as the parent.
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
type HTTPFormat struct{}

var _ propagation.HTTPFormat = (*HTTPFormat)(nil)

// SpanContextFromRequest extracts an OC span context from incoming requests.
// Will first try to extract from Datadog headers and then from B3 headers.
func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc oc_trace.SpanContext, ok bool) {
	if sc, ok = f.spanContextFromDatadogHeaders(req); ok {
		return
	}

	if sc, ok = f.spanContextFromB3Headers(req); ok {
		return
	}

	return
}

func (f *HTTPFormat) spanContextFromDatadogHeaders(req *http.Request) (sc oc_trace.SpanContext, ok bool) {
	if err := parseUint64ToByteSlice(req.Header.Get(trace.DatadogTraceIDHeader), sc.TraceID[8:16]); err != nil {
		return oc_trace.SpanContext{}, false
	}

	if err := parseUint64ToByteSlice(req.Header.Get(trace.DatadogParentIDHeader), sc.SpanID[0:8]); err != nil {
		return oc_trace.SpanContext{}, false
	}

	sampled, _ := strconv.Atoi(req.Header.Get(trace.DatadogSamplingPriorityHeader)) //nolint: errcheck
	if sampled >= ext.PriorityAutoKeep {
		sampled = 1
	}

	sc.TraceOptions = oc_trace.TraceOptions(sampled)

	return sc, true
}

func parseUint64ToByteSlice(str string, v []byte) error {
	id, err := strconv.ParseUint(str, 10, 64) //nolint: gomnd
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint64(v, id)

	return nil
}

func (f *HTTPFormat) spanContextFromB3Headers(req *http.Request) (sc oc_trace.SpanContext, ok bool) {
	if sc.TraceID, ok = parseTraceID(req.Header.Get(trace.B3TraceIDHeader)); !ok {
		return oc_trace.SpanContext{}, false
	}

	if sc.SpanID, ok = parseSpanID(req.Header.Get(trace.B3SpanIDHeader)); !ok {
		return oc_trace.SpanContext{}, false
	}

	sc.TraceOptions, _ = parseSampled(req.Header.Get(trace.B3SampledHeader))

	return sc, true
}

const eightBytes = 8
const sixteenBytes = 16

func parseTraceID(tid string) (oc_trace.TraceID, bool) {
	if tid == "" {
		return oc_trace.TraceID{}, false
	}

	b, err := hex.DecodeString(tid)
	if err != nil || len(b) > sixteenBytes {
		return oc_trace.TraceID{}, false
	}

	var traceID oc_trace.TraceID

	start := sixteenBytes - len(b)
	copy(traceID[start:], b)

	return traceID, true
}

func parseSpanID(sid string) (spanID oc_trace.SpanID, ok bool) {
	if sid == "" {
		return oc_trace.SpanID{}, false
	}

	b, err := hex.DecodeString(sid)
	if err != nil || len(b) > eightBytes {
		return oc_trace.SpanID{}, false
	}

	start := eightBytes - len(b)
	copy(spanID[start:], b)

	return spanID, true
}

func parseSampled(sampled string) (oc_trace.TraceOptions, bool) {
	switch sampled {
	case "true", "1":
		return oc_trace.TraceOptions(1), true
	default:
		return oc_trace.TraceOptions(0), false
	}
}

// SpanContextToRequest modifies the given request to include B3 and Datadog headers.
func (f *HTTPFormat) SpanContextToRequest(sc oc_trace.SpanContext, req *http.Request) {
	req.Header.Set(trace.B3TraceIDHeader, hex.EncodeToString(sc.TraceID[:]))
	req.Header.Set(trace.B3SpanIDHeader, hex.EncodeToString(sc.SpanID[:]))

	req.Header.Set(trace.DatadogTraceIDHeader, strconv.FormatUint(binary.BigEndian.Uint64(sc.TraceID[8:16]), 10)) //nolint: gomnd
	req.Header.Set(trace.DatadogParentIDHeader, strconv.FormatUint(binary.BigEndian.Uint64(sc.SpanID[0:8]), 10))  //nolint: gomnd

	var sampled string
	if sc.IsSampled() {
		sampled = "1"
	} else {
		sampled = "0"
	}

	req.Header.Set(trace.B3SampledHeader, sampled)
	req.Header.Set(trace.DatadogSamplingPriorityHeader, sampled)
}
