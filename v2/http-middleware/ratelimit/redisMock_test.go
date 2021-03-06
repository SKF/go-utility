package ratelimit_test

import (
	"github.com/SKF/go-utility/v2/http-middleware/ratelimit"

	"github.com/stretchr/testify/mock"
)

type ConnectionPoolMock struct {
	mock.Mock
}

type ConnectionMock struct {
	mock.Mock
}

func (m *ConnectionPoolMock) Connect() ratelimit.Connection {
	args := m.Called()
	return args.Get(0).(ratelimit.Connection)
}

func (m *ConnectionMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ConnectionMock) Incr(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}
