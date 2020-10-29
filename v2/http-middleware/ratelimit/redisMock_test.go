package ratelimit_test

import (
	"github.com/SKF/go-utility/v2/http-middleware/ratelimit"

	"github.com/stretchr/testify/mock"
)

type StoreMock struct {
	mock.Mock
}

type ConnectionMock struct {
	mock.Mock
}

func (m *StoreMock) NewConnection() ratelimit.Connection {
	args := m.Called()
	return args.Get(0).(ratelimit.Connection)
}

func (m *ConnectionMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ConnectionMock) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ConnectionMock) Do(commandName string, input ...interface{}) (interface{}, error) {
	args := m.Called(commandName, input)
	return args.Get(0), args.Error(1)
}

func (m *ConnectionMock) Send(commandName string, input ...interface{}) error {
	args := m.Called(commandName, input)
	return args.Error(0)
}

func (m *ConnectionMock) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ConnectionMock) Receive() (interface{}, error) {
	args := m.Called()
	return args.Get(0), args.Error(1)
}

func (m *ConnectionMock) Incr(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}
