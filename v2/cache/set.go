package cache

func (c *Cache) Set(key ObjectKey, value interface{}) bool {
	if c.ttl <= 0 {
		return false
	}

	c.sets++

	return c.cache.SetWithTTL(string(key), value, 1, c.ttl)
}
