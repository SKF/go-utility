package log

import (
	"fmt"
	"go/build"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var MaxStackDepth = 50

type logger struct {
	entry *logrus.Entry

	StackTraceEnabled bool
	SourceEnabled     bool
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.entry.WithField(key, value), l.StackTraceEnabled, l.SourceEnabled}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.entry.WithFields(logrus.Fields(fields)), l.StackTraceEnabled, l.SourceEnabled}
}

func (l logger) WithError(err error) Logger {
	return logger{l.entry.WithError(err), l.StackTraceEnabled, l.SourceEnabled}
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.caller().Debugf(format, args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.caller().Infof(format, args...)
}

func (l logger) Printf(format string, args ...interface{}) {
	l.caller().Printf(format, args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.caller().Warnf(format, args...)
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.caller().Warnf(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.caller().Errorf(format, args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.caller().Fatalf(format, args...)
}

func (l logger) Panicf(format string, args ...interface{}) {
	l.caller().Panicf(format, args...)
}

func (l logger) Debug(args ...interface{}) {
	l.caller().Debug(args...)
}

func (l logger) Info(args ...interface{}) {
	l.caller().Info(args...)
}

func (l logger) Print(args ...interface{}) {
	l.caller().Print(args...)
}

func (l logger) Warn(args ...interface{}) {
	l.caller().Warn(args...)
}

func (l logger) Warning(args ...interface{}) {
	l.caller().Warn(args...)
}

func (l logger) Error(args ...interface{}) {
	l.caller().Error(args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.caller().Fatal(args...)
}

func (l logger) Panic(args ...interface{}) {
	l.caller().Panic(args...)
}

func (l logger) Debugln(args ...interface{}) {
	l.caller().Debugln(args...)
}

func (l logger) Infoln(args ...interface{}) {
	l.caller().Infoln(args...)
}

func (l logger) Println(args ...interface{}) {
	l.caller().Println(args...)
}

func (l logger) Warnln(args ...interface{}) {
	l.caller().Warnln(args...)
}

func (l logger) Warningln(args ...interface{}) {
	l.caller().Warnln(args...)
}

func (l logger) Errorln(args ...interface{}) {
	l.caller().Errorln(args...)
}

func (l logger) Fatalln(args ...interface{}) {
	l.caller().Fatalln(args...)
}

func (l logger) Panicln(args ...interface{}) {
	l.caller().Panicln(args...)
}

func getCaller(skip int) (string, bool) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<?>"
		line = 1
		return fmt.Sprintf("%s:%d", file, line), false
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	if gopath != "" {
		file = strings.Replace(file, gopath+"/", "", -1)
	}

	return fmt.Sprintf("%s:%d", file, line), true
}

func getStackTrace(skip int, maxSteps int) (trace []string) {
	for i := 0; i < maxSteps; i++ {
		source, ok := getCaller(i + skip)
		if !ok {
			return
		}

		trace = append(trace, source)
	}
	return
}

func (l logger) caller() *logrus.Entry {
	entry := l.entry
	if l.SourceEnabled {
		source, _ := getCaller(4)
		entry = entry.WithField("source", source)
	}

	if l.StackTraceEnabled {
		trace := getStackTrace(4, MaxStackDepth)
		entry = entry.WithField("stacktrace", trace)
	}

	return entry
}
