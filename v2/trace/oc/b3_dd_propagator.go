// Package oc contains a propagation.HTTPFormat implementation for B3 and Datadog propagation.
// The HTTPFormat is based on ochttp and ddtrace.
// - ochttp: https://github.com/census-instrumentation/opencensus-go/tree/master/plugin/ochttp/propagation/b3
// - ddtrace: https://github.com/DataDog/dd-trace-go/blob/v1/ddtrace/tracer/textmap.go
// See https://github.com/openzipkin/b3-propagation for more details on B3 propagation.
package oc

import (
	"encoding/hex"
	"net/http"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const (
	// B3 headers that OpenCensus understands.
	b3TraceIDHeader = "X-B3-TraceId"
	b3SpanIDHeader  = "X-B3-SpanId"
	b3SampledHeader = "X-B3-Sampled"

	// Datadog headers that OpenCensus understands.
	datadogTraceIDHeader = "x-datadog-trace-id"
	datadogSpanIDHeader  = "x-datadog-parent-id"
	datadogSampledHeader = "x-datadog-sampling-priority"
)

// HTTPFormat implements propagation.HTTPFormat to propagate
// traces in HTTP headers in B3 and Datadog propagation format.
// HTTPFormat skips the X-B3-ParentId and X-B3-Flags headers
// because there are additional fields not represented in the
// OpenCensus span context. Spans created from the incoming
// header will be the direct children of the client-side span.
// Similarly, receiver of the outgoing spans should use client-side
// span created by OpenCensus as the parent.
type HTTPFormat struct{}

var _ propagation.HTTPFormat = (*HTTPFormat)(nil)

type headers struct {
	traceIDHeader string
	spanIDHeader  string
	sampledHeader string
}

// SpanContextFromRequest extracts an OC span context from incoming requests.
// Will first try to extract from Datadog headers and then from B3 headers.
func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	ddHeaders := headers{traceIDHeader: datadogTraceIDHeader, spanIDHeader: datadogSpanIDHeader, sampledHeader: datadogSampledHeader}
	if sc, ok = f.spanContextFromRequest(req, ddHeaders); ok {
		return
	}

	b3Headers := headers{traceIDHeader: b3TraceIDHeader, spanIDHeader: b3SpanIDHeader, sampledHeader: b3SampledHeader}
	if sc, ok = f.spanContextFromRequest(req, b3Headers); ok {
		return
	}

	return
}

func (f *HTTPFormat) spanContextFromRequest(req *http.Request, headers headers) (sc trace.SpanContext, ok bool) {
	tid, ok := parseTraceID(req.Header.Get(headers.traceIDHeader))
	if !ok {
		return trace.SpanContext{}, false
	}

	sid, ok := parseSpanID(req.Header.Get(headers.spanIDHeader))
	if !ok {
		return trace.SpanContext{}, false
	}

	sampled, _ := parseSampled(req.Header.Get(headers.sampledHeader))

	return trace.SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: sampled,
	}, true
}

func parseTraceID(tid string) (trace.TraceID, bool) {
	if tid == "" {
		return trace.TraceID{}, false
	}

	b, err := hex.DecodeString(tid)
	if err != nil {
		return trace.TraceID{}, false
	}

	var traceID trace.TraceID
	if len(b) <= 8 {
		// The lower 64-bits.
		start := 8 + (8 - len(b))
		copy(traceID[start:], b)
	} else {
		start := 16 - len(b)
		copy(traceID[start:], b)
	}

	return traceID, true
}

func parseSpanID(sid string) (spanID trace.SpanID, ok bool) {
	if sid == "" {
		return trace.SpanID{}, false
	}

	b, err := hex.DecodeString(sid)
	if err != nil {
		return trace.SpanID{}, false
	}

	start := 8 - len(b)
	copy(spanID[start:], b)

	return spanID, true
}

func parseSampled(sampled string) (trace.TraceOptions, bool) {
	switch sampled {
	case "true", "1":
		return trace.TraceOptions(1), true
	default:
		return trace.TraceOptions(0), false
	}
}

// SpanContextToRequest modifies the given request to include B3 and Datadog headers.
func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	req.Header.Set(b3TraceIDHeader, hex.EncodeToString(sc.TraceID[:]))
	req.Header.Set(b3SpanIDHeader, hex.EncodeToString(sc.SpanID[:]))

	var sampled string
	if sc.IsSampled() {
		sampled = "1"
	} else {
		sampled = "0"
	}

	req.Header.Set(b3SampledHeader, sampled)

	req.Header.Set(datadogTraceIDHeader, hex.EncodeToString(sc.TraceID[8:]))
	req.Header.Set(datadogSpanIDHeader, hex.EncodeToString(sc.SpanID[:]))
	req.Header.Set(datadogSampledHeader, sampled)
}
