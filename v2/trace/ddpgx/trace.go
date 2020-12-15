package ddpgx

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	beginEndSpace = regexp.MustCompile(`^\s+|\s+$`)
	multipleSpace = regexp.MustCompile(`\s{2,}`)

	flushToLog = strings.EqualFold(os.Getenv("DD_FLUSH_TO_LOG"), "true")
)

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
		span.SetTag(key, value)
	}

	if query, ok := metadata[dd_ext.SQLQuery]; ok {
		q := fmt.Sprintf("%v", query)

		if flushToLog {
			q = multipleSpace.ReplaceAllLiteralString(beginEndSpace.ReplaceAllLiteralString(q, ""), " ")
			span.SetTag(dd_ext.SQLQuery, q)
		}

		span.SetTag(dd_ext.ResourceName, q)
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
