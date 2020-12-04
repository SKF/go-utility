package ochttp_test

import (
	"net/http"
	"reflect"
	"testing"

	oc_trace "go.opencensus.io/trace"

	"github.com/SKF/go-utility/v2/trace"
	oc_http "github.com/SKF/go-utility/v2/trace/ochttp"
)

func TestHTTPFormat_B3_FromRequest(t *testing.T) {
	tests := []struct {
		name    string
		makeReq func() *http.Request
		wantSc  oc_trace.SpanContext
		wantOk  bool
	}{
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				req.Header.Set(trace.B3SampledHeader, "1")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "short trace ID + short span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "000102")
				req.Header.Set(trace.B3SpanIDHeader, "000102")
				req.Header.Set(trace.B3SampledHeader, "1")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
				SpanID:       oc_trace.SpanID{0, 0, 0, 0, 0, 0, 1, 2},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=0",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "0020000000000001")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				req.Header.Set(trace.B3SampledHeader, "0")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 1},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "invalid trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "invalid >128-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "0020000000000001002000000000000111")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID; invalid span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID; invalid >64 bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "002000000000000111")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=true",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				req.Header.Set(trace.B3SampledHeader, "true")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "128-bit trace ID + 64-bit span ID; sampled=false",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.B3TraceIDHeader, "463ac35c9f6413ad48485a3953bb6124")
				req.Header.Set(trace.B3SpanIDHeader, "0020000000000001")
				req.Header.Set(trace.B3SampledHeader, "false")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
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
	tests := []struct {
		name    string
		makeReq func() *http.Request
		wantSc  oc_trace.SpanContext
		wantOk  bool
	}{
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				req.Header.Set(trace.DatadogSamplingPriorityHeader, "1")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "short trace ID + short span ID; sampled=1",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "258")
				req.Header.Set(trace.DatadogParentIDHeader, "258")
				req.Header.Set(trace.DatadogSamplingPriorityHeader, "1")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
				SpanID:       oc_trace.SpanID{0, 0, 0, 0, 0, 0, 1, 2},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=0",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "9007199254740993")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				req.Header.Set(trace.DatadogSamplingPriorityHeader, "0")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 1},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
			},
			wantOk: true,
		},
		{
			name: "invalid trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "invalid >64-bit trace ID + 64-bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "18446744073709552000")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID; invalid span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID; invalid >64 bit span ID; no sampling header",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "18446744073709552000")
				return req
			},
			wantSc: oc_trace.SpanContext{},
			wantOk: false,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=2",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				req.Header.Set(trace.DatadogSamplingPriorityHeader, "2")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name: "64-bit trace ID + 64-bit span ID; sampled=false",
			makeReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com", nil) //nolint: errcheck
				req.Header.Set(trace.DatadogTraceIDHeader, "5208512171318403364")
				req.Header.Set(trace.DatadogParentIDHeader, "9007199254740993")
				req.Header.Set(trace.DatadogSamplingPriorityHeader, "false")
				return req
			},
			wantSc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
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
	tests := []struct {
		name        string
		sc          oc_trace.SpanContext
		wantHeaders map[string]string
	}{
		{
			name: "valid traceID, header ID, sampled=1",
			sc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(1),
			},
			wantHeaders: map[string]string{
				trace.B3TraceIDHeader:               "463ac35c9f6413ad48485a3953bb6124",
				trace.B3SpanIDHeader:                "0020000000000001",
				trace.B3SampledHeader:               "1",
				trace.DatadogTraceIDHeader:          "5208512171318403364",
				trace.DatadogParentIDHeader:         "9007199254740993",
				trace.DatadogSamplingPriorityHeader: "1",
			},
		},
		{
			name: "valid traceID, header ID, sampled=0",
			sc: oc_trace.SpanContext{
				TraceID:      oc_trace.TraceID{70, 58, 195, 92, 159, 100, 19, 173, 72, 72, 90, 57, 83, 187, 97, 36},
				SpanID:       oc_trace.SpanID{0, 32, 0, 0, 0, 0, 0, 1},
				TraceOptions: oc_trace.TraceOptions(0),
			},
			wantHeaders: map[string]string{
				trace.B3TraceIDHeader:               "463ac35c9f6413ad48485a3953bb6124",
				trace.B3SpanIDHeader:                "0020000000000001",
				trace.B3SampledHeader:               "0",
				trace.DatadogTraceIDHeader:          "5208512171318403364",
				trace.DatadogParentIDHeader:         "9007199254740993",
				trace.DatadogSamplingPriorityHeader: "0",
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
