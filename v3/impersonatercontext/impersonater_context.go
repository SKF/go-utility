package impersonatercontext

import (
	"context"
)

type impersonaterContextKey struct{}

// FromContext extracts User ID from a context.
func FromContext(ctx context.Context) (string, bool) {
	authorID, ok := ctx.Value(impersonaterContextKey{}).(string)
	return authorID, ok
}

// NewContext adds Impersonater User ID to a new context.
func NewContext(ctx context.Context, impersonaterID string) context.Context {
	return context.WithValue(ctx, impersonaterContextKey{}, impersonaterID)
}
