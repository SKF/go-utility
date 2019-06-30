package httpmiddleware

import (
	"net/http"

	http_model "github.com/SKF/go-utility/http-model"
	"github.com/SKF/go-utility/log"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField("recover", err).Error("Recovered from a panic")
				http_model.WriteJSONResponse(
					w,
					http.StatusInternalServerError,
					http_model.ErrResponseInternalServerError,
				)
			}

		}()

		next.ServeHTTP(w, r)
	})
}
