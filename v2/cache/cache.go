package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"

	"github.com/SKF/go-utility/v2/log"
)

type Cache struct {
	cache          *ristretto.Cache
	log            log.Logger
	gets           uint64
	sets           uint64
	perFuncMetrics map[string]*perFuncMetric
	expired        uint64
	ttl            time.Duration
}

type perFuncMetric struct {
	gets uint64
	hits uint64
}

type item struct {
	expiration time.Time
	data       interface{}
}

func New(ttl time.Duration, cacheSizeMaxMB int64) (*Cache, error) {
	if ttl <= 0 {
		log.Infof("Caching disabled, TTL: %d", ttl)
	}

	// nolint: gomnd
	maxCache := cacheSizeMaxMB * 1024 * 1024 // byte to mb
	log.Infof("Cache size in bytes: %d", maxCache)

	memcache, err := ristretto.NewCache(&ristretto.Config{
		// nolint: gomnd
		NumCounters: 1000000,  // number of keys to track frequency of (1M).
		MaxCost:     maxCache, // maximum cost of cache.
		BufferItems: 64,       //  number of keys per Get buffer.
		Metrics:     true,
	})

	if err != nil {
		err = errors.Wrap(err, "Error creating cache")
		log.WithError(err).Error("Error creating cache")

		return nil, err
	}

	log.WithField("ttl", ttl).Info("Creating in memory cache")

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
