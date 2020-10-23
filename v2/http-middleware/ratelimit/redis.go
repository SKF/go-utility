package ratelimit

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type redisStore struct {
	url        string
	connection redis.Conn
}

func (s *redisStore) Incr(key string) (int, error) {
	const secondsToExpire = 60

	cnt, err := redis.Int(s.connection.Do("INCR", key))
	if err != nil {
		return -1, err
	}

	_, err = s.connection.Do("EXPIRE", key, secondsToExpire)
	if err != nil {
		return -1, err
	}

	return cnt, nil
}

func (s *redisStore) Connect() error {
	const timeout = 2 * time.Second
	dialConnectTimeout := redis.DialConnectTimeout(timeout)
	readTimeout := redis.DialReadTimeout(timeout)
	writeTimeout := redis.DialWriteTimeout(timeout)

	con, err := redis.Dial("tcp", s.url, dialConnectTimeout, readTimeout, writeTimeout)

	s.connection = con

	return err
}

func (s *redisStore) Disconnect() error {
	if s == nil {
		return fmt.Errorf("redis store is nil")
	}
	if s.connection == nil {
		return fmt.Errorf("redis connection is nil")
	}

	return s.connection.Close()
}

func GetRedisStore(url string) Store {
	return &redisStore{url: url}
}
