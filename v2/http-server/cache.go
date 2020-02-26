package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/SKF/go-utility/v2/accesstokensubcontext"
)

func NewGeneralCacheKey(req *http.Request) string {
	return req.Method + " :: " + req.URL.Path + " :: " + req.URL.RawQuery
}

func NewUserSpecificCacheKey(ctx context.Context, req *http.Request) (_ string, err error) {
	subject, found := accesstokensubcontext.FromContext(ctx)
	if !found {
		err = errors.New("failed to extract Access Token Subject from context")
		return
	}

	return subject + " :: " + NewGeneralCacheKey(req), nil
}
