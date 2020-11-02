package ratelimit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	http_model "github.com/SKF/go-utility/v2/http-model"
	http_server "github.com/SKF/go-utility/v2/http-server"

	"github.com/SKF/go-utility/v2/log"

	"github.com/gorilla/mux"
	"go.opencensus.io/trace"
)

type ConnectionPool interface {
	Connect() Connection
}

type Connection interface {
	io.Closer

	Incr(key string) (int, error)
}

type Limit struct {
	RequestPerMinute int
	Key              string
}

type Request struct {
	Method       string
	PathTemplate string
}

type Limiter struct {
	connectionPool ConnectionPool
	configs        map[Request]limitGenerator
}

type limitGenerator func(*http.Request) ([]Limit, error)

func (l *Limiter) SetConnectionPool(p ConnectionPool) *Limiter {
	l.connectionPool = p
	return l
}

// The LimitGenerator function should return the rate limit with corresponding key
// that should be used for the given path template
// path.pathTemplate should be the pathtemplate used to route the request
//
// If you give multiple configs for 1 endpoint. The most restrictive one will apply
func (l *Limiter) Configure(path Request, gen limitGenerator) {
	if l.configs == nil {
		l.configs = map[Request]limitGenerator{}
	}

	l.configs[path] = gen
}

// Rate limiting middleware, you can configure 1 or many limits for each path template using a limitGenerator
// The algorithm is inspired from: https://redislabs.com/redis-best-practices/basic-rate-limiting/
//
// The key will be stored in clear text in the cache. If the key contains personal data please consider hashing the key
func (l *Limiter) Middleware() mux.MiddlewareFunc {
	if l.connectionPool == nil {
		panic("connectionPool is not configured")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span := trace.StartSpan(req.Context(), "RateLimitMiddleware/Handler")

			now := time.Now()

			pathTemplate, err := mux.CurrentRoute(req).GetPathTemplate()
			if err != nil {
				log.WithTracing(ctx).WithError(err).Errorf("failed to parse mux path template from request: %s", req.URL.Path)
				span.End()
				next.ServeHTTP(w, req)
				return
			}

			configGenerator, ok := l.configs[Request{Method: req.Method, PathTemplate: pathTemplate}]
			if ok {
				cfgs, err := configGenerator(req)
				if err != nil {
					log.WithTracing(ctx).WithError(err).Error("Failed to generate limits")
					span.End()
					next.ServeHTTP(w, req)

					return
				}

				tooManyRequest, err := l.checkAccessCounts(ctx, cfgs, now)
				if err != nil {
					log.WithTracing(ctx).WithError(err).Errorf("failed to check limit")
					span.End()
					next.ServeHTTP(w, req)

					return
				}

				if tooManyRequest {
					http_server.WriteJSONResponse(ctx, w, req, http.StatusTooManyRequests, http_model.ErrResponseTooManyRequests)
					span.End()

					return
				}
			}

			span.End()
			next.ServeHTTP(w, req)
		})
	}
}

func (l *Limiter) checkAccessCounts(ctx context.Context, cfgs []Limit, now time.Time) (tooManyRequests bool, err error) {
	_, span := trace.StartSpan(ctx, "RateLimitMiddleware/checkAccessCounts")
	defer span.End()

	db := l.connectionPool.Connect()
	defer db.Close()

	for _, config := range cfgs {
		key := fmt.Sprintf("%s:%d", config.Key, now.Minute())

		resp, err := db.Incr(key)
		if err != nil {
			return false, fmt.Errorf("incr failed: %w", err)
		}

		if resp > config.RequestPerMinute {
			return true, nil
		}
	}

	return false, nil
}
