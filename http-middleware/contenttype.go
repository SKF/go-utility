package httpmiddleware

import (
	"net/http"

	http_model "github.com/SKF/go-utility/http-model"
	http_server "github.com/SKF/go-utility/http-server"
)

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if (req.Method == "PUT" || req.Method == "POST" || req.Method == "PATHCH") &&
			req.Header.Get("Content-Type") != "application/json" {
			ctx := req.Context()
			http_server.WriteJSONResponse(ctx, w, http.StatusUnsupportedMediaType, http_model.ErrResponseUnsupportedMediaType)
		}
		next.ServeHTTP(w, req)
	})
}
