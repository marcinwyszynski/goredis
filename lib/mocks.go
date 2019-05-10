package lib

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/mock"
)

type mockDynamo struct {
	dynamodbiface.DynamoDBAPI
	mock.Mock
}

func (m *mockDynamo) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamo) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

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
