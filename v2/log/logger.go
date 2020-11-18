package log

import (
	"context"
	"encoding/binary"

	oc_trace "go.opencensus.io/trace"
	"go.uber.org/zap"
	dd_tracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/SKF/go-utility/v2/useridcontext"
)

type logger struct {
	logger *zap.SugaredLogger
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.logger.With(zap.Any(key, value))}
}

func (l logger) CheckWrite(lvl Level, msg string, fields ...Field) {
	if ce := l.logger.Desugar().Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.logger.Desugar().With(fields...).Sugar()}
}

func (l logger) WithError(err error) Logger {
	return logger{l.logger.With(zap.Error(err))}
}

// WithTracing will take either an OpenCensus or Datadog trace and add log fields for Datadog.
// The OpenCensus conversion is based:
// https://github.com/DataDog/opencensus-go-exporter-datadog/blob/master/span.go
// https://docs.datadoghq.com/tracing/advanced/connect_logs_and_traces/?tab=go
func (l logger) WithTracing(ctx context.Context) Logger {
	if span := oc_trace.FromContext(ctx); span != nil {
		traceID := span.SpanContext().TraceID
		spanID := span.SpanContext().SpanID

		return l.
			WithField("dd.trace_id", binary.BigEndian.Uint64(traceID[8:])).
			WithField("dd.span_id", binary.BigEndian.Uint64(spanID[:]))
	}

	if span, exists := dd_tracer.SpanFromContext(ctx); exists {
		return l.
			WithField("dd.trace_id", span.Context().TraceID()).
			WithField("dd.span_id", span.Context().SpanID())
	}

	return l
}

func (l logger) WithUserID(ctx context.Context) Logger {
	if userID, ok := useridcontext.FromContext(ctx); ok {
		return l.WithField("userId", userID)
	}

	return l
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l logger) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

func (l logger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l logger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l logger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l logger) Warning(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l logger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l logger) Panic(args ...interface{}) {
	l.logger.Panic(args...)
}

func (l logger) Sync() error {
	return l.logger.Sync()
}
