package lib

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type mockReadWriteCloser struct {
	io.ReadWriter
	mock.Mock
}

func (m *mockReadWriteCloser) Close() error {
	return m.Called().Error(0)
}

type mockStore struct {
	mock.Mock
}

func (m *mockStore) Get(key string) (value string, found bool, err error) {
	args := m.Called(key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *mockStore) Set(key string, value string) error {
	return m.Called(key, value).Error(0)
}
