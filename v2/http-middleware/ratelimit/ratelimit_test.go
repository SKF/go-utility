package ratelimit

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimitOk(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(0, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(EndpointConfig{
		Path: Request{Method: http.MethodGet, Path: "/apa"},
		Configs: []Config{{
			RequestPerMinute: 10,
			GetKeyFunc: func(req *http.Request) (string, error) {
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

func TestRateLimitTooMany(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(EndpointConfig{
		Path: Request{Method: http.MethodGet, Path: "/apa"},
		Configs: []Config{{
			RequestPerMinute: 5,
			GetKeyFunc: func(req *http.Request) (string, error) {
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
	req, r := getRouterAndRequest(t)

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	// config GET
	limiter.Configure(EndpointConfig{
		Path: Request{Method: http.MethodGet, Path: "/apa"},
		Configs: []Config{{
			RequestPerMinute: 15,
			GetKeyFunc: func(req *http.Request) (string, error) {
				return req.URL.Path, nil

			}},
		},
	})
	// config POST
	limiter.Configure(EndpointConfig{
		Path: Request{Method: http.MethodPost, Path: "/apa"},
		Configs: []Config{{
			RequestPerMinute: 5,
			GetKeyFunc: func(req *http.Request) (string, error) {
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
	req, r := getRouterAndRequest(t)

	// ACT
	limiter := CreateLimiter(&storeMock{})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestReadBodyInMiddleware(t *testing.T) {
	// ARRANGE
	type testRequest struct {
		Apa string
	}
	testBody := `{"Apa":"bepa"}`
	req, err := http.NewRequest(http.MethodPost, "/apa", strings.NewReader(testBody))
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error"))
			return
		}

		w.Write(b)
	})

	storeMock := storeMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(EndpointConfig{
		Path: Request{Method: http.MethodPost, Path: "/apa"},
		Configs: []Config{{
			RequestPerMinute: 15,
			GetKeyFunc: func(req *http.Request) (string, error) {
				bodybytes, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return "", err
				}

				a := testRequest{}
				err = json.Unmarshal(bodybytes, &a)
				if err != nil {
					return "", err
				}

				req.Body = ioutil.NopCloser(bytes.NewReader(bodybytes))

				return a.Apa, nil
			}},
		},
	})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
	res, err := ioutil.ReadAll(resp.Body)
	require.Equal(t, string(res), testBody)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("apa"))
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
