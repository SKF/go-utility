package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SKF/go-utility/v2/http-middleware/util"
	"github.com/SKF/go-utility/v2/log"

	"github.com/gorilla/mux"
)

type Store interface {
	Incr(string) (int, error)
	Connect() error
	Disconnect() error
}

type Limit struct {
	RequestPerMinute int
	Key              string
}

type Request struct {
	Method string
	Path   string
}

type Limiter struct {
	store   Store
	configs map[Request]limitGenerator
}

type limitGenerator func(*http.Request) ([]Limit, error)

func CreateLimiter(s Store) Limiter {
	return Limiter{
		store:   s,
		configs: map[Request]limitGenerator{},
	}
}

//TOOD: fix
// The GetKeyFunc should return a key that will be stored
// request using that key.
//
// If you give multiple configs for 1 endpoint. The most restrictive one will apply
func (s *Limiter) Configure(path Request, gen limitGenerator) {
	s.configs[path] = gen
}

// Rate limiting middleware, you can configure 1 or many limits for each endpoint using a limitGenerator
// The algorithm is inspired from: https://redislabs.com/redis-best-practices/basic-rate-limiting/
//
// The key will be stored in clear text in the cache. If the key contains personal data please consider hashing the key
func (s *Limiter) Middleware() mux.MiddlewareFunc {
	if s.store == nil {
		panic("store is not configured")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := util.StartSpanNoRoot(req.Context(), "RateLimitMiddleware/Handler")

			now := time.Now()

			configGenerator, ok := s.configs[Request{Method: req.Method, Path: req.URL.Path}]
			if ok {
				cfgs, err := configGenerator(req)
				if err != nil {
					log.WithTracing(ctx).WithError(err).Error("Failed to generate limits")
					span.End()
					next.ServeHTTP(w, req)

					return
				}

				tooManyRequest, err := s.checkAccessCounts(cfgs, now)
				if err != nil {
					log.WithTracing(ctx).WithError(err).Errorf("failed to check limit")
					span.End()
					next.ServeHTTP(w, req)

					return
				}

				if tooManyRequest {
					w.WriteHeader(http.StatusTooManyRequests) //nolint:errcheck
					w.Write([]byte("Too many requests"))      //nolint:errcheck
					span.End()

					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func (s *Limiter) checkAccessCounts(cfgs []Limit, now time.Time) (tooManyRequests bool, err error) {
	if err := s.store.Connect(); err != nil {
		return false, err
	}
	defer s.store.Disconnect() //nolint: errcheck

	for _, config := range cfgs {
		key := fmt.Sprintf("%s:%d", config.Key, now.Minute())

		resp, err := s.store.Incr(key)
		if err != nil {
			return false, err
		}

		if resp > config.RequestPerMinute {
			return true, nil
		}
	}

	return false, nil
}
