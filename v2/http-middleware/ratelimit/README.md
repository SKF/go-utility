# Rate limit Middleware

## Example usage

Create a webserver that limits the number of requests to / to 2 requests per minute
Complete example can be found [here](cmd/exampleserver/main.go)
``` go
   func main() {
   	r := mux.NewRouter()
   	r.HandleFunc("/", HomeHandler)
   
   	limiter := ratelimit.CreateLimiter(ratelimit.GetRedisStore("localhost:6379"))
   	limiter.Configure(
   		ratelimit.Request{Method: http.MethodGet, Path: "/"},
   		func(request *http.Request) ([]ratelimit.Limit, error) {
   			return []ratelimit.Limit{{
   				RequestPerMinute: 2,
   				Key:              "/",
   			}}, nil
   
   		})
   	r.Use(limiter.Middleware())
   
   	err := http.ListenAndServe(":8080", r)
   	log.Errorf(err.Error())
   }
```