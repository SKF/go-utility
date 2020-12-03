package ochttp

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"strconv"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
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

// SpanContextFromRequest extracts an OC span context from incoming requests.
// Will first try to extract from Datadog headers and then from B3 headers.
func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	if sc, ok = f.spanContextFromDatadogHeaders(req); ok {
		return
	}

	if sc, ok = f.spanContextFromB3Headers(req); ok {
		return
	}

	return
}

func (f *HTTPFormat) spanContextFromDatadogHeaders(req *http.Request) (sc trace.SpanContext, ok bool) {
	if err := parseUint64ToByteSlice(req.Header.Get(datadogTraceIDHeader), sc.TraceID[8:16]); err != nil {
		return trace.SpanContext{}, false
	}

	if err := parseUint64ToByteSlice(req.Header.Get(datadogSpanIDHeader), sc.SpanID[0:8]); err != nil {
		return trace.SpanContext{}, false
	}

	sampled, _ := strconv.Atoi(req.Header.Get(datadogSampledHeader)) //nolint: errcheck
	if sampled >= ext.PriorityAutoKeep {
		sampled = 1
	}

	sc.TraceOptions = trace.TraceOptions(sampled)

	return sc, true
}

func parseUint64ToByteSlice(str string, v []byte) error {
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint64(v, id)

	return nil
}

func (f *HTTPFormat) spanContextFromB3Headers(req *http.Request) (sc trace.SpanContext, ok bool) {
	if sc.TraceID, ok = parseTraceID(req.Header.Get(b3TraceIDHeader)); !ok {
		return trace.SpanContext{}, false
	}

	if sc.SpanID, ok = parseSpanID(req.Header.Get(b3SpanIDHeader)); !ok {
		return trace.SpanContext{}, false
	}

	sc.TraceOptions, _ = parseSampled(req.Header.Get(b3SampledHeader))

	return sc, true
}

const eightBytes = 8
const sixteenBytes = 16

func parseTraceID(tid string) (trace.TraceID, bool) {
	if tid == "" {
		return trace.TraceID{}, false
	}

	b, err := hex.DecodeString(tid)
	if err != nil || len(b) > sixteenBytes {
		return trace.TraceID{}, false
	}

	var traceID trace.TraceID

	start := sixteenBytes - len(b)
	copy(traceID[start:], b)

	return traceID, true
}

func parseSpanID(sid string) (spanID trace.SpanID, ok bool) {
	if sid == "" {
		return trace.SpanID{}, false
	}

	b, err := hex.DecodeString(sid)
	if err != nil || len(b) > eightBytes {
		return trace.SpanID{}, false
	}

	start := eightBytes - len(b)
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

	req.Header.Set(datadogTraceIDHeader, strconv.FormatUint(binary.BigEndian.Uint64(sc.TraceID[8:16]), 10))
	req.Header.Set(datadogSpanIDHeader, strconv.FormatUint(binary.BigEndian.Uint64(sc.SpanID[0:8]), 10))

	var sampled string
	if sc.IsSampled() {
		sampled = "1"
	} else {
		sampled = "0"
	}

	req.Header.Set(b3SampledHeader, sampled)
	req.Header.Set(datadogSampledHeader, sampled)
}
