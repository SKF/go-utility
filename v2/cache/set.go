package cache

import (
	"time"

	"github.com/SKF/go-utility/v2/log"
)

func (c *Cache) Set(key ObjectKey, value interface{}) bool {
	if c.ttl > 0 {
		c.sets++

		data := item{
			expiration: time.Now().Add(c.ttl),
			data:       value,
		}

		ok := c.cache.Set(string(key), data, 1)
		if !ok {
			log.Warnf("Cache insert dropped: %s", key)
		}

		return ok
	}

	return false
}
