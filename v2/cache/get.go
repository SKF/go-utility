package cache

import (
	"time"
)

func (c *Cache) Exist(key ObjectKey) bool {
	_, ok := c.Get(key)
	return ok
}

func (c *Cache) Get(key ObjectKey) (obj interface{}, ok bool) {
	if c.ttl <= 0 {
		return nil, false
	}

	if _, ok := c.perFuncMetrics[key.FuncName()]; !ok {
		c.perFuncMetrics[key.FuncName()] = &perFuncMetric{}
	}

	c.gets++
	c.perFuncMetrics[key.FuncName()].gets++

	data, found := c.cache.Get(string(key))
	if found && data != nil {
		if dataItem, ok := data.(item); ok {
			if time.Now().Before(dataItem.expiration) {
				c.perFuncMetrics[key.FuncName()].hits++
				return dataItem.data, true
			}

			c.expired++
			c.cache.Del(string(key))
		}
	}

	return nil, false
}
