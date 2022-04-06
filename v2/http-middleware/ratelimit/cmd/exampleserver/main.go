package main

import (
	"net/http"

	"github.com/SKF/go-utility/v2/log"

	"github.com/SKF/go-utility/v2/http-middleware/ratelimit"

	"github.com/gorilla/mux"
)

func main() {
	const requestsPerMinute = 2

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	limiter := ratelimit.Limiter{}
	limiter.SetConnectionPool(ratelimit.GetRedisPool("localhost:6379"))
	limiter.Configure(
		ratelimit.Request{Method: http.MethodGet, PathTemplate: "/"},
		func(request *http.Request) ([]ratelimit.Limit, error) {
			return []ratelimit.Limit{{
				RequestPerMinute: requestsPerMinute,
				Key:              "/",
			}}, nil
		})
	r.Use(limiter.Middleware())

	err := http.ListenAndServe(":8080", r)
	log.Errorf(err.Error())
}

func HomeHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("foo\n")) //nolint:errcheck
}
