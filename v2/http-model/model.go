package httpmodel

const (
	HeaderAuthorization = "Authorization"
	HeaderClientID      = "X-Client-ID"
	HeaderContentType = "Content-Type"

	HeaderCacheControl  = "Cache-Control"
	CacheControlNoCache = "no-cache"

	MimeJSON          = "application/json"
	MimeParameterUTF8 = "charset=utf-8"

)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

var ErrResponseUnsupportedMediaType = []byte(`{"error": {"message": "unsupported media type"}}`)
var ErrResponseInternalServerError = []byte(`{"error": {"message": "internal server error"}}`)
var ErrResponseUnauthorized = []byte(`{"error": {"message": "unauthorized"}}`)
var ErrResponseNotFound = []byte(`{"error": {"message": "not found"}}`)
var ErrResponseMethodNotAllowed = []byte(`{"error": {"message": "method not allowed"}}`)

var ErrMessageInternalServerError = "internal server error"
var ErrMessageUnauthorized = "unauthorized"
var ErrMessageNotFound = "not found"
