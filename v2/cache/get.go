package cache

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

	data, found := c.cache.Get(string(key))

	c.gets++
	c.perFuncMetrics[key.FuncName()].gets++

	if found {
		c.perFuncMetrics[key.FuncName()].hits++
	}

	return data, found
}
