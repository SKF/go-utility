package ratelimit

import "github.com/gomodule/redigo/redis"

type redisStore struct {
	url        string
	connection redis.Conn
}

func (s *redisStore) Incr(key string) (int, error) {
	cnt, err := redis.Int(s.connection.Do("INCR", key))
	if err != nil {
		return -1, err
	}

	_, err = s.connection.Do("EXPIRE", key, 59)
	if err != nil {
		return -1, err
	}

	return cnt, nil
}

func (s *redisStore) Connect() error {
	con, err := redis.Dial("tcp", s.url)
	s.connection = con

	return err
}

func (s *redisStore) Disconnect() error {
	return s.connection.Close()
}

func GetRedisStore(url string) Store {
	return &redisStore{url: url}
}
