package ratelimit_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	http_model "github.com/SKF/go-utility/v2/http-model"

	"github.com/SKF/go-utility/v2/http-middleware/ratelimit"

	"github.com/SKF/go-utility/v2/http-middleware/ratelimit/util"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimitOk(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := &StoreMock{}
	connMock := &ConnectionMock{}
	storeMock.On("NewConnection").Return(connMock).Once()

	connMock.On("Incr", mock.Anything).Return(0, nil)
	connMock.On("Close").Return(nil).Once()

	// ACT
	limiter := ratelimit.Limiter{}
	limiter.SetStore(storeMock)
	limiter.Configure(
		ratelimit.Request{Method: http.MethodGet, PathTemplate: "/apa"},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			return []ratelimit.Limit{{
				RequestPerMinute: 10,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	storeMock.AssertExpectations(t)
	connMock.AssertExpectations(t)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestRateLimitTooMany(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := &StoreMock{}
	connMock := &ConnectionMock{}
	storeMock.On("NewConnection").Return(connMock).Once()

	connMock.On("Incr", mock.Anything).Return(10, nil)
	connMock.On("Close").Return(nil).Once()

	// ACT
	limiter := &ratelimit.Limiter{}
	limiter.SetStore(storeMock)
	limiter.Configure(
		ratelimit.Request{Method: http.MethodGet, PathTemplate: "/apa"},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			return []ratelimit.Limit{{
				RequestPerMinute: 5,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	storeMock.AssertExpectations(t)
	connMock.AssertExpectations(t)

	require.Equal(t, http.StatusTooManyRequests, resp.Code)
	require.Equal(t, http_model.ErrResponseTooManyRequests, resp.Body.Bytes())
}

func TestUseCorrectLimit(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	storeMock := &StoreMock{}
	connMock := &ConnectionMock{}
	storeMock.On("NewConnection").Return(connMock).Once()

	connMock.On("Incr", mock.Anything).Return(10, nil)
	connMock.On("Close").Return(nil).Once()

	// ACT
	limiter := &ratelimit.Limiter{}
	limiter.SetStore(storeMock)
	// config GET
	limiter.Configure(
		ratelimit.Request{Method: http.MethodGet, PathTemplate: "/apa"},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			return []ratelimit.Limit{{
				RequestPerMinute: 15,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	// config POST
	limiter.Configure(
		ratelimit.Request{Method: http.MethodPost, PathTemplate: "/apa"},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			return []ratelimit.Limit{{
				RequestPerMinute: 5,
				Key:              req.URL.Path,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	storeMock.AssertExpectations(t)
	connMock.AssertExpectations(t)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestUnconfiguredIsOk(t *testing.T) {
	// ARRANGE
	req, r := getRouterAndRequest(t)

	// ACT
	limiter := &ratelimit.Limiter{}
	limiter.SetStore(&StoreMock{})
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

	storeMock := &StoreMock{}
	connMock := &ConnectionMock{}
	storeMock.On("NewConnection").Return(connMock).Once()

	connMock.On("Incr", mock.Anything).Return(10, nil)
	connMock.On("Close").Return(nil).Once()

	// ACT
	limiter := &ratelimit.Limiter{}
	limiter.SetStore(storeMock)
	limiter.Configure(
		ratelimit.Request{Method: http.MethodPost, PathTemplate: "/apa"},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			a := testRequest{}
			parseErr := util.ParseBody(req, &a)
			if parseErr != nil {
				// limit for invalidJSON
				return []ratelimit.Limit{{
					RequestPerMinute: 10,
					Key:              req.URL.Path,
				}}, parseErr
			}

			// Normal limit
			return []ratelimit.Limit{{
				RequestPerMinute: 15,
				Key:              a.SuperKey,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	// ASSERT
	storeMock.AssertExpectations(t)
	connMock.AssertExpectations(t)

	require.Equal(t, http.StatusOK, resp.Code)
	res, readErr := ioutil.ReadAll(resp.Body)
	require.NoError(t, readErr)

	require.Equal(t, string(res), testBody)
}

func TestUseDynamicRoute(t *testing.T) {
	// ARRANGE
	req, _ := http.NewRequest(http.MethodGet, "/apa/1", nil)  //nolint:errcheck
	req2, _ := http.NewRequest(http.MethodGet, "/apa/2", nil) //nolint:errcheck
	req3, _ := http.NewRequest(http.MethodGet, "/bepa", nil)  //nolint:errcheck

	const pathTemplate = "/apa/{id:[0-9]}"

	r := mux.NewRouter()
	r.HandleFunc(pathTemplate, func(writer http.ResponseWriter, request *http.Request) {
		args := mux.Vars(request)
		writer.Write([]byte(args["id"])) //nolint:errcheck
	})
	r.HandleFunc("/bepa", handler)

	storeMock := &StoreMock{}
	connMock := &ConnectionMock{}
	storeMock.On("NewConnection").Return(connMock)

	connMock.On("Incr", mock.Anything).Return(10, nil)
	connMock.On("Close").Return(nil)

	// ACT
	limiter := &ratelimit.Limiter{}
	limiter.SetStore(storeMock)
	limiter.Configure(
		ratelimit.Request{Method: http.MethodGet, PathTemplate: pathTemplate},
		func(req *http.Request) ([]ratelimit.Limit, error) {
			template, err := mux.CurrentRoute(req).GetPathTemplate()
			if err != nil {
				return nil, err
			}

			return []ratelimit.Limit{{
				RequestPerMinute: 5,
				Key:              template,
			}}, nil
		},
	)
	r.Use(limiter.Middleware())

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)

	resp3 := httptest.NewRecorder()
	r.ServeHTTP(resp3, req3)

	// ASSERT
	storeMock.AssertExpectations(t)
	connMock.AssertExpectations(t)

	require.Equal(t, http.StatusTooManyRequests, resp.Code)
	require.Equal(t, http.StatusTooManyRequests, resp2.Code)
	require.Equal(t, http.StatusOK, resp3.Code)
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
