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

func tryTrace(ctx context.Context, startTime time.Time, driver, serviceName, resource string, metadata map[string]interface{}, err error) {
	if _, exists := dd_tracer.SpanFromContext(ctx); !exists {
		return
	}

	operationName := fmt.Sprintf("%s.query", driver)
	span, _ := dd_tracer.StartSpanFromContext(ctx, operationName,
		dd_tracer.ServiceName(serviceName),
		dd_tracer.SpanType(dd_ext.SpanTypeSQL),
		dd_tracer.StartTime(startTime),
	)

	for key, value := range metadata {
		span.SetTag(key, value)
	}

	if query, ok := metadata[dd_ext.SQLQuery]; ok {
		span.SetTag(dd_ext.ResourceName, query)
	} else {
		span.SetTag(dd_ext.ResourceName, resource)
	}

	span.Finish(dd_tracer.WithError(err))
}
