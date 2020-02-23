package accesstokensubcontext

import (
	"context"
)

type accessTokensSubContextKey struct{}

// FromContext extracts the Access Token Subject from a context.
func FromContext(ctx context.Context) (string, bool) {
	accessTokenSubject, ok := ctx.Value(accessTokensSubContextKey{}).(string)
	return accessTokenSubject, ok
}

// NewContext adds an Access Token Subject to a new context.
func NewContext(ctx context.Context, accessTokenSubject string) context.Context {
	return context.WithValue(ctx, accessTokensSubContextKey{}, accessTokenSubject)
}
