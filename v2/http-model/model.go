package httpmodel

import "encoding/json"

const (
	HeaderAuthorization           = "Authorization"
	HeaderContentType             = "Content-Type"
	HeaderClientID                = "X-Client-ID"
	HeaderCacheControl            = "Cache-Control"
	HeaderDataDogTraceID          = "x-datadog-trace-id"
	HeaderDataDogParentID         = "x-datadog-parent-id"
	HeaderDataDogSamplingPriority = "x-datadog-sampling-priority"
	HeaderDataDogOrigin           = "x-datadog-origin"
	HeaderB3TraceID               = "x-b3-traceid"
	HeaderB3SpanID                = "x-b3-spanid"
	HeaderB3SampledHeader         = "x-b3-sampled"

	CacheControlNoCache = "no-cache"
	MimeJSON            = "application/json"
	MimeParameterUTF8   = "charset=utf-8"
)

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type HTTPError struct {
	Msg        string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return e.Msg
}

func (e *HTTPError) Message() []byte {
	errStruct := ErrorResponse{
		Error: struct {
			Message string `json:"message"`
		}{Message: e.Msg}}
	data, _ := json.Marshal(errStruct) // nolint:errcheck

	return data
}

var ErrResponseUnsupportedMediaType = []byte(`{"error": {"message": "unsupported media type"}}`)
var ErrResponseInternalServerError = []byte(`{"error": {"message": "internal server error"}}`)
var ErrResponseBadRequest = []byte(`{"error": {"message": "bad request"}}`)
var ErrResponseTooManyRequests = []byte(`{"error": {"message": "Too many requests"}}`)
var ErrResponseUnauthorized = []byte(`{"error": {"message": "unauthorized"}}`)
var ErrResponseNotFound = []byte(`{"error": {"message": "not found"}}`)
var ErrResponseMethodNotAllowed = []byte(`{"error": {"message": "method not allowed"}}`)

var ErrMessageInternalServerError = "internal server error"
var ErrMessageUnauthorized = "unauthorized"
var ErrMessageNotFound = "not found"
