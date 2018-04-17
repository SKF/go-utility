package log

import (
	"fmt"
	"go/build"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type logger struct {
	entry *logrus.Entry
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.entry.WithField(key, value)}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.entry.WithFields(logrus.Fields(fields))}
}

func (l logger) WithError(err error) Logger {
	return logger{l.entry.WithError(err)}
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

func (l logger) caller() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<?>"
		line = 1
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	if gopath != "" {
		file = strings.Replace(file, gopath+"/", "", -1)
	}

	return l.entry.WithField("source", fmt.Sprintf("%s:%d", file, line))
}
