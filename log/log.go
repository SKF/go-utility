package log

import (
	"os"
	"time"

	"github.com/bluele/zapslack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Fields = zapcore.Field

type SlackHook struct {
	AcceptedLevels []zapcore.Level
	HookURL        string
	Name           string
	Asynchronous   bool
}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Sync()
}

var origLogger *zap.SugaredLogger
var baseLogger logger

func init() {
	encoderConf := zap.NewProductionEncoderConfig()

	// Set RFC3339 timestamp encoding format
	encoderConf.TimeKey = "timestamp"
	encoderConf.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}

	l := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConf),
			zapcore.Lock(os.Stdout),
			zap.NewAtomicLevel(),
		))
	origLogger = l.Sugar()

	baseLogger = logger{
		entry: origLogger,
	}

}

func Base() Logger {
	return baseLogger
}

func AddSlackHook(hook SlackHook) {
	if hook.HookURL == "" {
		WithField("name", hook.Name).
			Warn("Cannot add slack hook with empty webhook url")
		return
	}

	if len(hook.AcceptedLevels) == 0 {
		hook.AcceptedLevels = []zapcore.Level{zap.ErrorLevel}
	}

	zl := zapslack.SlackHook{
		HookURL:        hook.HookURL,
		AcceptedLevels: hook.AcceptedLevels,
		Async:          hook.Asynchronous,
		FieldHeader:    hook.Name,
	}
	l := origLogger.Desugar()
	l.WithOptions(
		zap.Hooks(
			zl.GetHook()))
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

func Debugf(format string, args ...interface{}) {
	baseLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	baseLogger.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	baseLogger.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	baseLogger.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	baseLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	baseLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	baseLogger.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	baseLogger.Panicf(format, args...)
}

func Debug(args ...interface{}) {
	baseLogger.Debug(args...)
}

func Info(args ...interface{}) {
	baseLogger.Info(args...)
}

func Print(args ...interface{}) {
	baseLogger.Print(args...)
}

func Warn(args ...interface{}) {
	baseLogger.Warn(args...)
}

func Warning(args ...interface{}) {
	baseLogger.Warn(args...)
}

func Error(args ...interface{}) {
	baseLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	baseLogger.Fatal(args...)
}

func Panic(args ...interface{}) {
	baseLogger.Panic(args...)
}
