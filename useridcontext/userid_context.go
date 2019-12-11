package useridcontext

import (
	"context"
)

type userIDContextKey struct{}

// FromContext extracts User ID from a context.
func FromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(string)
	return userID, ok
}

// NewContext adds User ID to a new context.
func NewContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}
