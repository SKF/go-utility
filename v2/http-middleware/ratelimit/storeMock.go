package ratelimit

import "github.com/stretchr/testify/mock"

type storeMock struct {
	mock.Mock
}

func (m *storeMock) Incr(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}

func (m *storeMock) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *storeMock) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}
