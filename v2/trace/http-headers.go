package trace

const (
	// B3 headers, used by Open Census.
	B3TraceIDHeader = "X-B3-TraceId"
	B3SpanIDHeader  = "X-B3-SpanId"
	B3SampledHeader = "X-B3-Sampled"

	// Datadog headers.
	DatadogOriginHeader           = "x-datadog-origin"
	DatadogParentIDHeader         = "x-datadog-parent-id"
	DatadogSampledHeader          = "x-datadog-sampled"
	DatadogSamplingPriorityHeader = "x-datadog-sampling-priority"
	DatadogTraceIDHeader          = "x-datadog-trace-id"
)

func AllB3Headers() []string {
	return []string{B3TraceIDHeader, B3SpanIDHeader, B3SampledHeader}
}

func AllDatadogHeaders() []string {
	return []string{DatadogOriginHeader, DatadogParentIDHeader, DatadogSampledHeader, DatadogSamplingPriorityHeader, DatadogTraceIDHeader}
}
