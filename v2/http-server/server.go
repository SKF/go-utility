package httpserver

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	http_model "github.com/SKF/go-utility/v2/http-model"
	"github.com/SKF/go-utility/v2/log"
)

func StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		WriteJSONResponse(req.Context(), w, req, http.StatusOK, []byte(`{"status": "ok"}`))
	})

	log.Infof("Starting health server on port %s", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil { // nolint: gosec
		log.WithError(err).Error("ListenAndServe")
	}
}

func SetupDatadogInstrumentation(service, awsRegion, awsAccountID, stage string) *datadog.Exporter {
	ddTracer, err := datadog.NewExporter(datadog.Options{
		Service: service,
		GlobalTags: map[string]interface{}{
			"aws_region":     awsRegion,
			"aws_account_id": awsAccountID,
			"env":            stage,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create the Datadog exporter: %v", err)
	}

	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register server views for HTTP metrics: %v", err)
	}

	view.RegisterExporter(ddTracer)
	trace.RegisterExporter(ddTracer)

	// Allow Datadog to calculate APM metrics and do the sampling.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	return ddTracer
}

func UnmarshalRequest(body io.ReadCloser, v interface{}) error {
	defer body.Close()

	if err := json.NewDecoder(body).Decode(v); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	return nil
}

// MarshalAndWriteJSONResponse will JSON marshal the incoming data and return the serialized version
func MarshalAndWriteJSONResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, v interface{}) []byte {
	response, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).
			WithTracing(ctx).
			WithField("type", fmt.Sprintf("%T", v)).
			Error("Failed to marshal response body")

		code = http.StatusInternalServerError
		response = http_model.ErrResponseInternalServerError
	}

	WriteJSONResponse(ctx, w, r, code, response)

	return response
}

// Don't gzip body if smaller than one packet, as it will be transmitted as a full packet anyway.
const gzipMinBodySize = 1400

func WriteJSONResponse(ctx context.Context, w http.ResponseWriter, r *http.Request, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json")

	var err error

	if r != nil && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && len(body) > gzipMinBodySize {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(code)
		gz := gzip.NewWriter(w)

		defer gz.Close()
		_, err = gz.Write(body)
	} else {
		w.WriteHeader(code)
		_, err = w.Write(body)
	}

	if err != nil {
		log.WithError(err).
			WithTracing(ctx).
			Error("Failed to write response")
	}
}

func StatusNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusNotFound
		resp := http_model.ErrResponseNotFound
		WriteJSONResponse(r.Context(), w, r, code, resp)
	})
}

func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusMethodNotAllowed
		resp := http_model.ErrResponseMethodNotAllowed
		WriteJSONResponse(r.Context(), w, r, code, resp)
	})
}
