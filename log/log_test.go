package log_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SKF/go-utility/log"
)

func TestLog(t *testing.T) {
	log.SetDefaultService("service")

	log.WithField("application", "backend").WithError(fmt.Errorf("A test error")).Info("A info msg")
	log.WithField("application", "backend").Debug("This is a debug message")
	log.WithField("token", "1234").Warning("A warning msg")
	log.Error("A test error, should have stacktrace")

	assert.Panics(t, panicLog)
}

func panicLog() {
	log.WithField("token", "1234").Panic("A panic msg")
}
