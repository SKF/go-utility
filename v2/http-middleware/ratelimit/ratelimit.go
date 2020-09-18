package ratelimit

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SKF/go-utility/v2/http-middleware/util"

	"github.com/gorilla/mux"
)

type Store interface {
	Incr(string) (int, error)
	Connect() error
	Disconnect() error
}

type Limit struct {
	RequestPerMinute int
	key              string
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

// The GetKeyFunc should return a key that will be stored
// as sha256(key) + <current minute in the cache to limit the number of
// request using that key.
//
// If you give multiple configs for 1 endpoint. The most restrictive one will apply
func (s *Limiter) Configure(path Request, gen limitGenerator) {
	s.configs[path] = gen
}

// Rate limiting middleware, you can configure 1 or many limits for each endpoint using a limitGenerator
// The algorithm is inspired from: https://redislabs.com/redis-best-practices/basic-rate-limiting/
func (s *Limiter) Middleware() mux.MiddlewareFunc {
	if s.store == nil {
		panic("store is not configured")
	}

	hasher := sha256.New()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, span := util.StartSpanNoRoot(req.Context(), "RateLimitMiddleware/Handler")
			defer span.End()

			now := time.Now()

			// TODO: handle errors maybe allow access if we get error here?
			configGenerator, ok := s.configs[Request{Method: req.Method, Path: req.URL.Path}]
			if ok {
				cfgs, err := configGenerator(req)
				if err != nil {
					log.Fatalf("Failed to generate limits: %v", err)
					w.WriteHeader(http.StatusInternalServerError) //nolint:errcheck
					w.Write([]byte("Internal Server error"))      // nolint:errcheck
					return
				}
				if err := s.store.Connect(); err != nil {
					w.WriteHeader(http.StatusInternalServerError) // nolint:errcheck
					w.Write([]byte("Internal Server error"))      // nolint:errcheck
					return
				}

				defer s.store.Disconnect() //nolint: errcheck
				for _, config := range cfgs {
					key := fmt.Sprintf("%x:%d", hasher.Sum([]byte(config.key)), now.Minute())

					resp, err := s.store.Incr(key)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError) // nolint:errcheck
						w.Write([]byte("Internal Server error"))      // nolint:errcheck
						return
					}

					if resp > config.RequestPerMinute {
						w.WriteHeader(http.StatusTooManyRequests) // nolint:errcheck
						w.Write([]byte("Too many requests"))      //nolint:errcheck
						return
					}
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}
