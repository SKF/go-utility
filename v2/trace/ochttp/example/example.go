package example

import (
	"net/http"

	"github.com/gorilla/mux"
	oc_http "go.opencensus.io/plugin/ochttp"

	"github.com/SKF/go-utility/v2/log"
	skf_oc_http "github.com/SKF/go-utility/v2/trace/ochttp"
)

type Server struct {
	mux        *mux.Router
	httpServer *http.Server
}

func (s *Server) ListenAndServe(port string) {
	handler := new(oc_http.Handler)
	handler.Handler = s.mux
	handler.Propagation = new(skf_oc_http.HTTPFormat)

	s.httpServer = &http.Server{ // nolint: gosec
		Addr:    ":" + port,
		Handler: handler,
	}

	log.WithField("port", port).Info("Will start to listen and serve")

	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.WithError(err).Error("HTTP server ListenAndServe")
	}
}
