package httpmodel

import (
	"net/http"

	"github.com/SKF/go-utility/log"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

var ErrResponseInternalServerError = []byte(`{"error": {"message": "internal server error"}}`)
var ErrResponseUnauthorized = []byte(`{"error": {"message": "unauthorized"}}`)

var ErrMessageInternalServerError = "internal server error"
var ErrMessageUnauthorized = "unauthorized"

func WriteJSONResponse(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		log.WithError(err).Error("Failed to write response")
	}
}
