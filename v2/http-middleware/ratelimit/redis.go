package ratelimit

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Store interface {
	NewConnection() Connection
}

type Connection interface {
	redis.Conn

	Incr(key string) (int, error)
}

type redisStore struct {
	pool *redis.Pool
}

type redisConnection struct {
	redis.Conn
}

func (s *redisStore) NewConnection() Connection {
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

func GetRedisStore(address string) Store {
	var (
		waitingConnections = 10

		dialTimeout  = 1 * time.Second
		idleTimeout  = 4 * time.Minute
		readTimeout  = 1 * time.Second
		writeTimeout = 1 * time.Second
	)

	pool := &redis.Pool{
		MaxIdle:     waitingConnections,
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

	return &redisStore{pool}
}
