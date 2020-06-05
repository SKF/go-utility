package cache

import (
	"fmt"
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

	defer func() {
		statsLogger := c.log.WithField("getsCalled", c.gets).
			WithField("setsCalled", c.sets).
			WithField("expired", c.expired).
			WithField("expiredPercentage", (c.expired*100)/c.gets).
			WithField("misses", c.cache.Metrics.Misses()).
			WithField("missesPercentage", (c.cache.Metrics.Misses()*100)/c.gets).
			WithField("hits", c.cache.Metrics.Hits()).
			WithField("hitsPercentage", (c.cache.Metrics.Hits()*100)/c.gets).
			WithField("ratio", c.cache.Metrics.Ratio()).
			WithField("keysAdded", c.cache.Metrics.KeysAdded()).
			WithField("keysUpdated", c.cache.Metrics.KeysUpdated()).
			WithField("getsEvicted", c.cache.Metrics.KeysEvicted()).
			WithField("getsDropped", c.cache.Metrics.GetsDropped()).
			WithField("getsKept", c.cache.Metrics.GetsKept()).
			WithField("setsDropped", c.cache.Metrics.SetsDropped()).
			WithField("setsRejected", c.cache.Metrics.SetsRejected()).
			WithField("ttl", c.ttl)

		for funcName, metric := range c.perFuncMetrics {
			getsPercentage := (metric.gets * 100) / c.gets
			hitsPercentage := (metric.hits * 100) / metric.gets
			statsLogger = statsLogger.
				WithField(fmt.Sprintf("gets.%s", funcName), metric.gets).
				WithField(fmt.Sprintf("getsPercentage.%s", funcName), getsPercentage).
				WithField(fmt.Sprintf("hitsPercentage.%s", funcName), hitsPercentage)
		}

		statsLogger.Info("Cache stats")
	}()

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
			} else {
				c.expired++
				c.cache.Del(string(key))
			}
		}
	}

	return nil, false
}
