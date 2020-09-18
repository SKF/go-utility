package ratelimit

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SKF/go-utility/v2/http-middleware/ratelimit/util"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimitOk(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := StoreMock{}
	storeMock.On("Incr", mock.Anything).Return(0, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(
		Request{Method: http.MethodGet, Path: "/apa"},
		func(req *http.Request) ([]Limit, error) {
			return []Limit{{
				RequestPerMinute: 10,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestRateLimitTooMany(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := StoreMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(
		Request{Method: http.MethodGet, Path: "/apa"},
		func(req *http.Request) ([]Limit, error) {
			return []Limit{{
				RequestPerMinute: 5,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusTooManyRequests, resp.Code)
}

func TestUseCorrectLimit(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := StoreMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	// config GET
	limiter.Configure(
		Request{Method: http.MethodGet, Path: "/apa"},
		func(req *http.Request) ([]Limit, error) {
			return []Limit{{
				RequestPerMinute: 15,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	// config POST
	limiter.Configure(
		Request{Method: http.MethodPost, Path: "/apa"},
		func(req *http.Request) ([]Limit, error) {
			return []Limit{{
				RequestPerMinute: 5,
				Key:              req.URL.Path,
			}}, nil
		},
	)
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
	limiter := CreateLimiter(&StoreMock{})
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestReadBodyInMiddleware(t *testing.T) {
	// ARRANGE
	type testRequest struct {
		SuperKey string
	}

	testBody := `{"SuperKey":"apa"}`

	req, readErr := http.NewRequest(http.MethodPost, "/apa", strings.NewReader(testBody))
	if readErr != nil {
		t.Fatal(readErr)
	}

	r := mux.NewRouter()
	r.HandleFunc("/apa", func(w http.ResponseWriter, r *http.Request) {
		b, readBodyErr := ioutil.ReadAll(r.Body)
		if readBodyErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error")) //nolint: errcheck
			return
		}

		w.Write(b) //nolint: errcheck
	})

	storeMock := StoreMock{}
	storeMock.On("Incr", mock.Anything).Return(10, nil)
	storeMock.On("Connect").Return(nil).Once()
	storeMock.On("Disconnect").Return(nil).Once()

	// ACT
	limiter := CreateLimiter(&storeMock)
	limiter.Configure(
		Request{Method: http.MethodPost, Path: "/apa"},
		func(req *http.Request) ([]Limit, error) {
			a := testRequest{}
			parseErr := util.ParseBody(req, &a)
			if parseErr != nil {
				// limit for invalidJSON
				return []Limit{{
					RequestPerMinute: 10,
					Key:              req.URL.Path,
				}}, parseErr
			}

			// Normal limit
			return []Limit{{
				RequestPerMinute: 15,
				Key:              a.SuperKey,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	require.Equal(t, http.StatusOK, resp.Code)
	res, readErr := ioutil.ReadAll(resp.Body)
	require.NoError(t, readErr)

	require.Equal(t, string(res), testBody)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("apa")) //nolint: errcheck
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
