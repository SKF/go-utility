package log

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/SKF/go-utility/env"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field
type Fields = []Field
type Level = zapcore.Level

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
	WithTracing(ctx context.Context) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	CheckWrite(lvl Level, msg string, fields ...Field)
	Sync() error
}

var baseLogger logger

func init() {
	encoder := getEncoder()

	origLogger := zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevelAt(getLogLevel()),
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	baseLogger = logger{origLogger.Sugar()}
}

func getLogLevel() zapcore.Level {
	levelStr := env.GetAsString("LOG_LEVEL", "info")

	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func getEncoder() zapcore.Encoder {
	encoderConf := zap.NewProductionEncoderConfig()

	encoderConf.MessageKey = "message"
	encoderConf.TimeKey = "timestamp"
	encoderConf.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		// Set RFC3339 timestamp encoding format
		enc.AppendString(t.Format(time.RFC3339))
	}
	encoderConf.CallerKey = "source"

	useConsoleEncoder := strings.EqualFold(os.Getenv("CONSOLE_LOGGER"), "true")
	encoder := zapcore.NewJSONEncoder(encoderConf)
	if useConsoleEncoder {
		encoder = zapcore.NewConsoleEncoder(encoderConf)
	}
	return encoder
}

func Base() Logger {
	return baseLogger
}

func WithField(key string, value interface{}) Logger {
	return baseLogger.WithField(key, value)
}

func WithFields(fields Fields) Logger {
	return baseLogger.WithFields(fields)
}

func WithError(err error) Logger {
	return baseLogger.WithError(err)
}

// WithTracing will take an OpenCensus trace and add log fields for Datadog.
func WithTracing(ctx context.Context) Logger {
	return baseLogger.WithTracing(ctx)
}

// We must directly call the bundled logger here (whenever a func instead of
// method is used), reason is for the "caller skip" calculation to be correct
// in all instances.
// When we are called as a function `log.Info("msg")` vs method
// `log.WithField("key", "val").Info("msg")` we would otherwise end up with
// 3 vs 2 stack entries.

func Debugf(format string, args ...interface{}) {
	baseLogger.logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	baseLogger.logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	baseLogger.logger.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	baseLogger.logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	baseLogger.logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	baseLogger.logger.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	baseLogger.logger.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	baseLogger.logger.Debug(args...)
}

func Info(args ...interface{}) {
	baseLogger.logger.Info(args...)
}

func Warn(args ...interface{}) {
	baseLogger.logger.Warn(args...)
}

func Warning(args ...interface{}) {
	baseLogger.logger.Warn(args...)
}

func Error(args ...interface{}) {
	baseLogger.logger.Error(args...)
}

func Fatal(args ...interface{}) {
	baseLogger.logger.Fatal(args...)
}

func Panic(args ...interface{}) {
	baseLogger.logger.Panic(args...)
}

func CheckWrite(lvl Level, msg string, fields ...Field) {
	baseLogger.CheckWrite(lvl, msg, fields...)
}
