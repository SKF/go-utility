package log_test

import (
	"fmt"
	"testing"

	"github.com/SKF/go-utility/log"
)

func TestLog(t *testing.T) {
	log.WithField("application", "backend").WithError(fmt.Errorf("A test error")).Info("A info msg")
	log.WithField("token", "1234").Warning("A warning msg")
	log.Error("A test error, should have stacktrace")
}
