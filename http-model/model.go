package httpmodel

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

var ErrResponseInternalServerError = []byte(`{"error": {"message": "internal server error"}}`)
var ErrResponseUnauthorized = []byte(`{"error": {"message": "unauthorized"}}`)

var ErrMessageInternalServerError = "internal server error"
var ErrMessageUnauthorized = "unauthorized"
