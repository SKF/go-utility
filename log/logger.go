package log

import (
	"go.uber.org/zap"
)

type logger struct {
	entry *zap.SugaredLogger

	CallTraceEnabled bool
	SourceEnabled    bool
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.entry.With(zap.Any(key, value)), l.CallTraceEnabled, l.SourceEnabled}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.entry.With(fields), l.CallTraceEnabled, l.SourceEnabled}
}

func (l logger) WithError(err error) Logger {
	return logger{l.entry.With(zap.Error(err)), l.CallTraceEnabled, l.SourceEnabled}
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l logger) Printf(format string, args ...interface{}) {
	l.entry.Infof(format, args...) // Printf not present in zap
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l logger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

func (l logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l logger) Print(args ...interface{}) {
	l.entry.Info(args...) // Print not present in zap
}

func (l logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l logger) Warning(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l logger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}
