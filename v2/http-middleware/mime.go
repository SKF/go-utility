package httpmiddleware

import (
	"net/http"
	"strings"

	http_model "github.com/SKF/go-utility/v2/http-model"
	http_server "github.com/SKF/go-utility/v2/http-server"
	"github.com/SKF/go-utility/v2/log"
)

// ContentType wraps a HandlerFunc and checks the incoming
// content-type with a list of allowed content types.
func ContentType(next http.HandlerFunc, contentTypes ...string) http.HandlerFunc {
	var validContentTypes = make(map[string]bool)
	for _, contentType := range contentTypes {
		validContentTypes[strings.ToLower(contentType)] = true
	}

	return func(w http.ResponseWriter, req *http.Request) {
		reqContentType := req.Header.Get(http_model.HeaderContentType)
		reqContentType = strings.ToLower(reqContentType)

		parts := strings.Split(reqContentType, ";")

		if !validContentTypes[parts[0]] {
			ctx := req.Context()
			log.WithTracing(ctx).WithField("contentType", reqContentType).Warn("Unsupported Content-Type")
			http_server.WriteJSONResponse(ctx, w, nil, http.StatusUnsupportedMediaType, http_model.ErrResponseUnsupportedMediaType)
			return
		}

		if len(parts) == 1 || strings.TrimSpace(parts[1]) == http_model.MimeParameterUTF8 {
			next(w, req)
			return
		}

		ctx := req.Context()
		log.WithTracing(ctx).WithField("contentType", reqContentType).Warn("Unsupported Content-Type Parameter")
		http_server.WriteJSONResponse(ctx, w, nil, http.StatusUnsupportedMediaType, http_model.ErrResponseUnsupportedMediaType)
	}
}
