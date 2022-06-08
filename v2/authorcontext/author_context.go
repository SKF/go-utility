package authorcontext

import (
	"context"
)

type authorContextKey struct{}

// FromContext extracts User ID from a context.
func FromContext(ctx context.Context) (string, bool) {
	authorID, ok := ctx.Value(authorContextKey{}).(string)
	return authorID, ok
}

// NewContext adds User ID to a new context.
func NewContext(ctx context.Context, authorID string) context.Context {
	return context.WithValue(ctx, authorContextKey{}, authorID)
}
