package ratelimit

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitOk(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(0, nil)

	// ACT
	limiter := SetStore(&storeMock)
	limiter.Configure(EndpointConfig{
		path:              Request{method: http.MethodGet, path: "/apa"},
		config: []Config{{
			requestPerMinute: 10,
			getKeyFunc: func(req *http.Request) (string, error) {
				return req.URL.Path, nil

			}},
		},
	})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func getRouterAndRequest(t *testing.T) (*http.Request, *mux.Router) {
	req, err := http.NewRequest("GET", "/apa", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", handler)
	return req, r
}

func TestRateLimitTooMany(t *testing.T) {
	// ARRANGE
	req, err := http.NewRequest("GET", "/apa", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", handler)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)

	// ACT
	limiter := SetStore(&storeMock)
	limiter.Configure(EndpointConfig{
		path:              Request{method: http.MethodGet, path: "/apa"},
		config: []Config{{
			requestPerMinute: 5,
			getKeyFunc: func(req *http.Request) (string, error) {
				return req.URL.Path, nil

			}},
		},
	})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusTooManyRequests, resp.Code)
}

func TestUseCorrectLimit(t *testing.T) {
	// ARRANGE
	req, err := http.NewRequest("GET", "/apa", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", handler)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)

	// ACT
	limiter := SetStore(&storeMock)
	// config GET
	limiter.Configure(EndpointConfig{
		path:              Request{method: http.MethodGet, path: "/apa"},
		config: []Config{{
			requestPerMinute: 15,
			getKeyFunc: func(req *http.Request) (string, error) {
				return req.URL.Path, nil

			}},
		},
	})
	// config POST
	limiter.Configure(EndpointConfig{
		path:              Request{method: http.MethodPost, path: "/apa"},
		config: []Config{{
			requestPerMinute: 5,
			getKeyFunc: func(req *http.Request) (string, error) {
				return req.URL.Path, nil

			}},
		},
	})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestUnconfiguredIsOk(t *testing.T) {
	// ARRANGE
	req, err := http.NewRequest("GET", "/apa", nil)
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", handler)

	limiter := SetStore(&storeMock{})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("apa"))
}
