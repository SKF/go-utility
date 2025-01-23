package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/SKF/go-utility/v2/log"
)

const (
	defaultNumCounters = 1000000 // default number of keys to track frequency of (1M).
	defaultBufferItems = 64      // default number of keys per Get buffer.

	megaByte = 1 << 10
)

type Cache struct {
	cache          *ristretto.Cache
	log            log.Logger
	gets           uint64
	sets           uint64
	perFuncMetrics map[string]*perFuncMetric
	ttl            time.Duration
}

type perFuncMetric struct {
	gets uint64
	hits uint64
}

func New(ttl time.Duration, cacheSizeMaxMB int64) (*Cache, error) {
	if ttl <= 0 {
		log.Infof("Caching disabled, TTL: %d", ttl)
	}

	maxCache := cacheSizeMaxMB * megaByte

	memcache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: defaultNumCounters,
		MaxCost:     maxCache, // maximum cost of cache.
		BufferItems: defaultBufferItems,
		Metrics:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating cache: %w", err)
	}

	obj := Cache{
		cache:          memcache,
		ttl:            ttl,
		log:            log.Base(),
		perFuncMetrics: make(map[string]*perFuncMetric),
	}

	return &obj, nil
}

func (c *Cache) SetLogger(logger log.Logger) {
	c.log = logger
}

func (c *Cache) Clear() {
	c.cache.Clear()
}

func (c *Cache) Sets() uint64 {
	return c.sets
}

func (c *Cache) Gets() uint64 {
	return c.gets
}

func (c *Cache) Misses() uint64 {
	return c.cache.Metrics.Misses()
}

func (c *Cache) Hits() uint64 {
	return c.cache.Metrics.Hits()
}

func (c *Cache) SetTTL(value time.Duration) {
	c.ttl = value
}
