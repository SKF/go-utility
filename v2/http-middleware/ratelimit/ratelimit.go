package ratelimit

import (
	"fmt"
	"github.com/SKF/go-utility/v2/http-middleware/util"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Store interface {
	Incr(string) (int, error)
}

type EndpointConfig struct {
	path   Request
	config []Config
}

type Config struct {
	requestPerMinute int
	getKeyFunc       func(req *http.Request) (string, error)
}

type Request struct {
	method string
	path   string
}

type Limiter struct {
	store   Store
	configs map[Request][]Config
}

func SetStore(s Store) Limiter {
	return Limiter{
		store:   s,
		configs: map[Request][]Config{},
	}
}

func (s *Limiter) Configure(config EndpointConfig) {
	s.configs[config.path] = config.config
}

type redisStore struct {
	url string
}

func (s *redisStore) Incr(key string) (int, error) {
	conn, err := redis.Dial("tcp", s.url)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	cnt, err := redis.Int(conn.Do("INCR", key))
	if err != nil {
		return -1, err
	}

	_, err = conn.Do("EXPIRE", key, 59)
	if err != nil {
		return -1, err
	}

	return cnt, nil
}

func GetRedisStore(url string) Store {
	return &redisStore{url: "localhost:6379"}
}

func (s *Limiter) Middleware() mux.MiddlewareFunc {
	if s.store == nil {
		panic("store is not configured")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, span := util.StartSpanNoRoot(req.Context(), "RateLimitMiddleware/Handler")
			defer span.End()

			now := time.Now()

			cfgs, ok := s.configs[Request{method: req.Method, path: req.URL.Path}]
			if ok {
				for _, config := range cfgs {
					key, err := config.getKeyFunc(req)
					key = fmt.Sprintf("%s:%d", key, now.Minute())
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("Internal Server error"))
						return
					}

					resp, err := s.store.Incr(key)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("Internal Server error"))
						return
					}

					if resp > config.requestPerMinute {
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
				}
			}

			next.ServeHTTP(w, req)
		})
	}
}
