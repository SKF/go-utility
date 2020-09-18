package ratelimit

import "github.com/stretchr/testify/mock"

type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) Incr(key string) (int, error) {
	args := m.Called(key)
	return args.Int(0), args.Error(1)
}

func (m *StoreMock) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *StoreMock) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}
