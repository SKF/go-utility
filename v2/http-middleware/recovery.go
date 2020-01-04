package httpmiddleware

import (
	"net/http"

	http_model "github.com/SKF/go-utility/v2/http-model"
	http_server "github.com/SKF/go-utility/v2/http-server"
	"github.com/SKF/go-utility/v2/log"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := req.Context()
				log.WithTracing(ctx).WithField("recover", err).Error("Recovered from a panic")
				http_server.WriteJSONResponse(ctx, w, req, http.StatusInternalServerError, http_model.ErrResponseInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}
