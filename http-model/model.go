package httpmodel

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

var ErrResponseInternalServerError = []byte(`{"error": {"message": "internal server error"}}`)
var ErrResponseUnauthorized = []byte(`{"error": {"message": "unauthorized"}}`)
var ErrResponseNotFound = []byte(`{"error": {"message": "not found"}}`)
var ErrResponseUnsupportedMediaType = []byte(`{"error": {"message": "unsupported media type"}}`)

var ErrMessageInternalServerError = "internal server error"
var ErrMessageUnauthorized = "unauthorized"
var ErrMessageNotFound = "not found"
var ErrMessageUnsupportedMediaType = "unsupported media type"
