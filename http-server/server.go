package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	datadog "github.com/Datadog/opencensus-go-exporter-datadog"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	http_model "github.com/SKF/go-utility/http-model"
	"github.com/SKF/go-utility/log"
)

func StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
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

func UnmarshalRequest(r *http.Request, v interface{}) (err error) {
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		err = errors.Wrap(err, "failed to unmarshal request body")
	}
	return
}

func MarshalAndWriteJSONResponse(w http.ResponseWriter, code int, v interface{}) {
	response, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).
			WithField("type", fmt.Sprintf("%T", v)).
			Error("Failed to marshal response body")
		response = http_model.ErrResponseInternalServerError
	}
	WriteJSONResponse(w, code, response)
}

func WriteJSONResponse(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		log.WithError(err).Error("Failed to write response")
	}
}
