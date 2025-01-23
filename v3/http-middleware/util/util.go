package util

import (
	"context"

	"go.opencensus.io/trace"
)

func StartSpanNoRoot(ctx context.Context, name string, o ...trace.StartOption) (context.Context, *trace.Span) {
	if trace.FromContext(ctx) == nil {
		return ctx, nil
	}

	return trace.StartSpan(ctx, name, o...)
}
