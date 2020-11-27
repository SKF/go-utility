package awstrace

var (
	b3TraceHeader = "x-b3-traceid"
	b3SpanHeader  = "x-b3-spanid"

	datadogTraceHeader  = "x-datadog-trace-id"
	datadogParentHeader = "x-datadog-parent-id"

	allHeaders = []string{b3TraceHeader, b3SpanHeader, datadogTraceHeader, datadogParentHeader}
)
