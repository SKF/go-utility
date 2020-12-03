package ochttp_test

import (
	"net/http"
	"reflect"
	"testing"

	"go.opencensus.io/trace"

	oc_http "github.com/SKF/go-utility/v2/trace/ochttp"
)

func TestHTTPFormat_B3_FromRequest(t *testing.T) {
	const (
		// B3 headers that OpenCensus understands.
		traceIDHeader = "X-B3-TraceId"
		spanIDHeader  = "X-B3-SpanId"
		sampledHeader = "X-B3-Sampled"
	)

	tests := []struct {
		name    string
		makeReq func() *http.Request
		wantSc  trace.SpanContext
		wantOk  bool
	}{
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "0020000000000001")
				req.Header.Set(sampledHeader, "1")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "short trace ID + short span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "000102")
				req.Header.Set(spanIDHeader, "000102")
				req.Header.Set(sampledHeader, "1")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
				SpanID:       trace.SpanID{0, 0, 0, 0, 0, 0, 1, 2},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=0",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "0020000000000001")
				req.Header.Set(spanIDHeader, "0020000000000001")
				req.Header.Set(sampledHeader, "0")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 1},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "0020000000000001")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "invalid trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "")
				req.Header.Set(spanIDHeader, "0020000000000001")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "invalid >128-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "0020000000000001002000000000000111")
				req.Header.Set(spanIDHeader, "0020000000000001")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID; invalid span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID; invalid >64 bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "002000000000000111")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=true",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "0020000000000001")
				req.Header.Set(sampledHeader, "true")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=false",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(spanIDHeader, "0020000000000001")
				req.Header.Set(sampledHeader, "false")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &oc_http.HTTPFormat{}
			sc, ok := f.SpanContextFromRequest(tt.makeReq())
			if ok != tt.wantOk {
				t.Errorf("HTTPFormat.SpanContextFromRequest() got ok = %v, want %v", ok, tt.wantOk)
			}
			if !reflect.DeepEqual(sc, tt.wantSc) {
				t.Errorf("HTTPFormat.SpanContextFromRequest() got span context = %v, want %v", sc, tt.wantSc)
			}
		})
	}
}

func TestHTTPFormat_DD_FromRequest(t *testing.T) {
	const (
		// Datadog headers that OpenCensus understands.
		traceIDHeader = "x-datadog-trace-id"
		spanIDHeader  = "x-datadog-parent-id"
		sampledHeader = "x-datadog-sampling-priority"
	)

	tests := []struct {
		name    string
		makeReq func() *http.Request
		wantSc  trace.SpanContext
		wantOk  bool
	}{
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "9007199254740993")
				req.Header.Set(sampledHeader, "1")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "short trace ID + short span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "258")
				req.Header.Set(spanIDHeader, "258")
				req.Header.Set(sampledHeader, "1")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
				SpanID:       trace.SpanID{0, 0, 0, 0, 0, 0, 1, 2},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=0",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "9007199254740993")
				req.Header.Set(spanIDHeader, "9007199254740993")
				req.Header.Set(sampledHeader, "0")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 1},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "9007199254740993")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "invalid trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "")
				req.Header.Set(spanIDHeader, "9007199254740993")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "invalid >64-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "18446744073709552000")
				req.Header.Set(spanIDHeader, "9007199254740993")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID; invalid span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID; invalid >64 bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "18446744073709552000")
				return req
			},
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=2",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "9007199254740993")
				req.Header.Set(sampledHeader, "2")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=false",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(traceIDHeader, "5208512171318403364")
				req.Header.Set(spanIDHeader, "9007199254740993")
				req.Header.Set(sampledHeader, "false")
				return req
			},
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &oc_http.HTTPFormat{}
			sc, ok := f.SpanContextFromRequest(tt.makeReq())
			if ok != tt.wantOk {
				t.Errorf("HTTPFormat.SpanContextFromRequest() got ok = %v, want %v", ok, tt.wantOk)
			}
			if !reflect.DeepEqual(sc, tt.wantSc) {
				t.Errorf("HTTPFormat.SpanContextFromRequest() got span context = %v, want %v", sc, tt.wantSc)
			}
		})
	}
}

func TestHTTPFormat_ToRequest(t *testing.T) {
	const (
		// B3 headers that OpenCensus understands.
		b3TraceIDHeader = "X-B3-TraceId"
		b3SpanIDHeader  = "X-B3-SpanId"
		b3SampledHeader = "X-B3-Sampled"

		// Datadog headers that OpenCensus understands.
		ddTraceIDHeader = "x-datadog-trace-id"
		ddSpanIDHeader  = "x-datadog-parent-id"
		ddSampledHeader = "x-datadog-sampling-priority"
	)

	tests := []struct {
		name        string
		sc          trace.SpanContext
		wantHeaders map[string]string
	}{
		{
			name: "valid traceID, header ID, sampled=1",
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(1),
			},
			wantHeaders: map[string]string{
				b3TraceIDHeader: "463ac35c9f6413ad48485a3953bb6124",
				b3SpanIDHeader:  "0020000000000001",
				b3SampledHeader: "1",
				ddTraceIDHeader: "5208512171318403364",
				ddSpanIDHeader:  "9007199254740993",
				ddSampledHeader: "1",
			},
		},
		{
			name: "valid traceID, header ID, sampled=0",
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: trace.TraceOptions(0),
			},
			wantHeaders: map[string]string{
				b3TraceIDHeader: "463ac35c9f6413ad48485a3953bb6124",
				b3SpanIDHeader:  "0020000000000001",
				b3SampledHeader: "0",
				ddTraceIDHeader: "5208512171318403364",
				ddSpanIDHeader:  "9007199254740993",
				ddSampledHeader: "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &oc_http.HTTPFormat{}
			req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
			f.SpanContextToRequest(tt.sc, req)

			for k, v := range tt.wantHeaders {
				if got, want := req.Header.Get(k), v; got != want {
					t.Errorf("req.Header.Get(%q) = %q; want %q", k, got, want)
				}
			}
		})
	}
}
