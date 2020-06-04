package log

import (
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
)

type sampleSyncer struct {
	zapcore.WriteSyncer
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
	now := time.Now()
	if now.Sub(s.lastTick) > s.tick {
		s.count = 0
	}
	s.count++
	s.lastTick = now

	if s.count > s.first && s.count%s.thereafter != 0 {
		return ioutil.Discard.Write(b)
	}

	return s.WriteSyncer.Write(b)
}

func (s *sampleSyncer) Sync() error {
	return s.WriteSyncer.Sync()
}

func NewSampleLogger(tick time.Duration, first, thereafter int) Logger {
	syncer := newSampleSyncer(tick, first, thereafter)
	return logger{newLogger(syncer).With(fields...).Sugar()}
}
