package ratelimit

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

type redisPool struct {
	pool *redis.Pool
}

type redisConnection struct {
	redis.Conn
}

func (s *redisPool) Connect() Connection {
	return &redisConnection{s.pool.Get()}
}

func (c *redisConnection) Incr(key string) (int, error) {
	const secondsToExpire = 60

	cnt, err := redis.Int(c.Do("INCR", key))
	if err != nil {
		return -1, err
	}

	_, err = c.Do("EXPIRE", key, secondsToExpire)
	if err != nil {
		return -1, err
	}

	return cnt, nil
}

func GetRedisPool(address string) ConnectionPool {
	var (
		pooledConnections = 10

		dialTimeout  = 1 * time.Second
		idleTimeout  = 4 * time.Minute
		readTimeout  = 1 * time.Second
		writeTimeout = 1 * time.Second
	)

	pool := &redis.Pool{
		MaxIdle:     pooledConnections,
		IdleTimeout: idleTimeout,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialContext(
				ctx,
				"tcp",
				address,
				redis.DialConnectTimeout(dialTimeout),
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout),
			)
		},
	}

	return &redisPool{pool}
}
