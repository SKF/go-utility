package ddpgx

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var newlineIndent = regexp.MustCompile(`\n\s+`)

type internalTracer struct {
	serviceName string
	driver      string
}

func newTracer(serviceName, driver string) internalTracer {
	return internalTracer{
		serviceName: serviceName,
		driver:      driver,
	}
}

func (t internalTracer) ServiceName() string {
	return t.serviceName
}

func (t internalTracer) TryTrace(ctx context.Context, startTime time.Time, resource string, metadata map[string]interface{}, err error) {
	if _, exists := dd_tracer.SpanFromContext(ctx); !exists {
		return
	}

	operationName := fmt.Sprintf("%s.query", t.driver)
	span, _ := dd_tracer.StartSpanFromContext(ctx, operationName,
		dd_tracer.ServiceName(t.serviceName),
		dd_tracer.SpanType(dd_ext.SpanTypeSQL),
		dd_tracer.StartTime(startTime),
	)

	span.SetTag("sql.method", resource)

	for key, value := range metadata {
		if key == dd_ext.SQLQuery {
			q := fmt.Sprintf("%v", value)
			q = strings.TrimSpace(q)
			q = newlineIndent.ReplaceAllString(q, " ")

			metadata[key] = q
		}

		span.SetTag(key, value)
	}

	if query, ok := metadata[dd_ext.SQLQuery]; ok {
		span.SetTag(dd_ext.ResourceName, query)
	} else {
		span.SetTag(dd_ext.ResourceName, resource)
	}

	span.Finish(dd_tracer.WithError(err))
}

func argsToAttributes(args ...interface{}) map[string]interface{} {
	output := map[string]interface{}{}

	for i := range args {
		key := fmt.Sprintf("sql.args.%d", i)
		output[key] = args[i]
	}

	return output
}
