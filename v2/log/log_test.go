package log_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SKF/go-utility/v2/log"
)

func TestLog(t *testing.T) {
	log.SetDefaultService("service")

	// nolint: gomnd
	sampler := log.NewSampleLogger(100*time.Millisecond, 10, 50)

	for i := 0; i < 100; i++ {
		// nolint: gomnd
		sampler.Info(fmt.Sprintf("Called %d", i+1))
		// nolint: gomnd
		time.Sleep(10 * time.Millisecond)
	}

	log.WithField("application", "backend").Info("A info msg")
	log.WithField("application", "backend").Debug("This is a debug message")
	log.WithField("token", "1234").Warning("A warning msg")
	log.WithError(errors.New("A test error")).Error("A test error, should have stacktrace")

	assert.Panics(t, panicLog)
}

func panicLog() {
	log.WithField("token", "1234").Panic("A panic msg")
}
