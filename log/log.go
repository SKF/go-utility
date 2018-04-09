package log

import (
	"github.com/johntdyer/slackrus"
	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields
type Formatter logrus.Formatter
type SlackHook struct {
	AcceptedLevels []logrus.Level
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

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

var origLogger = logrus.New()
var baseLogger = logger{
	entry: logrus.NewEntry(origLogger),
}

func init() {
	SetFormatter(&logrus.JSONFormatter{})
}

func Base() Logger {
	return baseLogger
}

func SetFormatter(formatter Formatter) {
	origLogger.Formatter = formatter
}

func AddSlackHook(hook SlackHook) {
	if hook.HookURL == "" {
		WithField("name", hook.Name).
			Warn("Cannot add slack hook with empty webhook url")
		return
	}

	extra := map[string]interface{}{}
	if hook.Name != "" {
		extra["name"] = hook.Name
	}

	if len(hook.AcceptedLevels) == 0 {
		hook.AcceptedLevels = slackrus.LevelThreshold(logrus.ErrorLevel)
	}

	origLogger.AddHook(&slackrus.SlackrusHook{
		HookURL:        hook.HookURL,
		AcceptedLevels: hook.AcceptedLevels,
		Asynchronous:   hook.Asynchronous,
		Extra:          extra,
	})
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

func Debugln(args ...interface{}) {
	baseLogger.Debugln(args...)
}

func Infoln(args ...interface{}) {
	baseLogger.Infoln(args...)
}

func Println(args ...interface{}) {
	baseLogger.Println(args...)
}

func Warnln(args ...interface{}) {
	baseLogger.Warnln(args...)
}

func Warningln(args ...interface{}) {
	baseLogger.Warnln(args...)
}

func Errorln(args ...interface{}) {
	baseLogger.Errorln(args...)
}

func Fatalln(args ...interface{}) {
	baseLogger.Fatalln(args...)
}

func Panicln(args ...interface{}) {
	baseLogger.Panicln(args...)
}
