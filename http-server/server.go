package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	datadog "github.com/DataDog/opencensus-go-exporter-datadog"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	http_model "github.com/SKF/go-utility/http-model"
	"github.com/SKF/go-utility/log"
)

func StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		WriteJSONResponse(req.Context(), w, http.StatusOK, []byte(`{"status": "ok"}`))
	})

	log.Infof("Starting health server on port %s", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
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

func UnmarshalRequest(body io.ReadCloser, v interface{}) (err error) {
	defer body.Close()
	if err = json.NewDecoder(body).Decode(v); err != nil {
		err = errors.Wrap(err, "failed to unmarshal request body")
	}
	return
}

func MarshalAndWriteJSONResponse(ctx context.Context, w http.ResponseWriter, code int, v interface{}) {
	_, span := trace.StartSpan(ctx, "MarshalResponse")
	response, err := json.Marshal(v)
	span.End()
	if err != nil {
		log.WithError(err).
			WithTracing(ctx).
			WithField("type", fmt.Sprintf("%T", v)).
			Error("Failed to marshal response body")
		response = http_model.ErrResponseInternalServerError
	}
	WriteJSONResponse(ctx, w, code, response)
}

func WriteJSONResponse(ctx context.Context, w http.ResponseWriter, code int, body []byte) {
	ctx, span := trace.StartSpan(ctx, "WriteJSONResponse")
	defer span.End()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(body); err != nil {
		log.WithError(err).
			WithTracing(ctx).
			Error("Failed to write response")
	}
}

func StatusNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSONResponse(r.Context(), w, http.StatusNotFound, http_model.ErrResponseNotFound)
	})
}
