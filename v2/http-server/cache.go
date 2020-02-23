package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/SKF/go-utility/v2/accesstokensubcontext"
)

type GeneralCacheKey struct {
	Method   string
	RawPath  string
	RawQuery string
}

type UserSpecificCacheKey struct {
	GeneralCacheKey
	AccessTokenSubject string
}

func NewGeneralCacheKey(req *http.Request) GeneralCacheKey {
	return GeneralCacheKey{
		Method:   req.Method,
		RawPath:  req.URL.RawPath,
		RawQuery: req.URL.RawQuery,
	}
}

func NewUserSpecificCacheKey(ctx context.Context, req *http.Request) (_ UserSpecificCacheKey, err error) {
	subject, found := accesstokensubcontext.FromContext(ctx)
	if !found {
		err = errors.New("failed to extract Access Token Subject from context")
		return
	}

	key := UserSpecificCacheKey{
		GeneralCacheKey:    NewGeneralCacheKey(req),
		AccessTokenSubject: subject,
	}

	return key, nil
}
