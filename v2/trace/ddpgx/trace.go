package ddpgx

import (
	"context"
	"fmt"
	"time"

	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func argsToAttributes(args ...interface{}) map[string]interface{} {
	output := map[string]interface{}{}
	for i := range args {
		key := fmt.Sprintf("sql.args.%d", i)
		output[key] = args[i]
	}

	return output
}

func tryTrace(ctx context.Context, startTime time.Time, operationName string, metadata map[string]interface{}, err error) {
	if _, exists := dd_tracer.SpanFromContext(ctx); !exists {
		return
	}

	span, _ := dd_tracer.StartSpanFromContext(ctx, operationName,
		// dd_tracer.ServiceName(tserviceName)
		dd_tracer.SpanType(dd_ext.SpanTypeSQL),
		dd_tracer.StartTime(startTime),
	)

	for key, value := range metadata {
		span.SetTag(key, value)
	}

	// span.SetTag(ext.ResourceName, resource)
	span.Finish(dd_tracer.WithError(err))
}
