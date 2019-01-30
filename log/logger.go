package log

import (
	"go.uber.org/zap"
)

type logger struct {
	logger *zap.SugaredLogger
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.logger.With(zap.Any(key, value))}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.logger.With(fields)}
}

func (l logger) WithError(err error) Logger {
	return logger{l.logger.With(zap.Error(err))}
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

func (l logger) Sync() {
	l.logger.Sync()
}
