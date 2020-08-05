package log

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

type sampleSyncer struct {
	zapcore.WriteSyncer
	sync.Mutex
	count             int
	tick              time.Duration
	lastTick          time.Time
	first, thereafter int
}

func newSampleSyncer(tick time.Duration, first, thereafter int) zapcore.WriteSyncer {
	return &sampleSyncer{
		WriteSyncer: zapcore.Lock(os.Stdout),
		count:       0,
		tick:        tick,
		first:       first,
		thereafter:  thereafter,
	}
}

func (s *sampleSyncer) Write(b []byte) (int, error) {
	s.Lock()
	now := time.Now()
	if now.Sub(s.lastTick) > s.tick {
		s.count = 0
	}
	s.count++
	s.lastTick = now

	if s.count > s.first && s.count%s.thereafter != 0 {
		s.Unlock()
		return ioutil.Discard.Write(b)
	}
	s.Unlock()

	return s.WriteSyncer.Write(b)
}

func (s *sampleSyncer) Sync() error {
	return s.WriteSyncer.Sync()
}

// NewSampleLogger will create a new logger
// which will sample logs and dropped if requested to often
// Logs which is written within `tick` be applied by first and thereafter
// `first` is amount of logs that will always will be outputted
// `thereafter` is used for outputten every n after within the `tick`
func NewSampleLogger(tick time.Duration, first, thereafter int) Logger {
	syncer := newSampleSyncer(tick, first, thereafter)
	return logger{newLogger(syncer).With(fields...).Sugar()}
}
