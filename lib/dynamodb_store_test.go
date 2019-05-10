package lib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type dynamoDBStoreTestSuite struct {
	suite.Suite

	api *mockDynamo
	sut *DynamoDBStore
}

func (d *dynamoDBStoreTestSuite) SetupTest() {
	d.api = new(mockDynamo)
	d.sut = &DynamoDBStore{API: d.api, TableName: "table"}
}

func (d *dynamoDBStoreTestSuite) TestGet_OK() {
	const key = "key"
	const value = "value"

	d.api.On(
		"GetItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*dynamodb.GetItemInput)
			if !ok {
				return false
			}

			d.Len(input.Key, 1)
			d.Equal(key, *input.Key["key"].S)
			d.Equal("table", *input.TableName)

			return true
		}),
		[]request.Option(nil),
	).Return(&dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{"value": {S: aws.String(value)}},
	}, nil)

	ret, found, err := d.sut.Get(key)

	d.Equal(value, ret)
	d.True(found)
	d.NoError(err)
}

func (d *dynamoDBStoreTestSuite) TestGet_APIError() {
	const key = "key"

	d.api.On(
		"GetItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("*dynamodb.GetItemInput"),
		[]request.Option(nil),
	).Return((*dynamodb.GetItemOutput)(nil), errors.New("bacon"))

	ret, found, err := d.sut.Get(key)

	d.Empty(ret)
	d.False(found)
	d.EqualError(err, "DynamoDB API error: bacon")
}

func (d *dynamoDBStoreTestSuite) TestGet_NotFound() {
	const key = "key"

	d.api.On(
		"GetItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("*dynamodb.GetItemInput"),
		[]request.Option(nil),
	).Return(&dynamodb.GetItemOutput{Item: nil}, nil)

	ret, found, err := d.sut.Get(key)

	d.Empty(ret)
	d.False(found)
	d.NoError(err)
}

func (d *dynamoDBStoreTestSuite) TestGet_NoValueField() {
	const key = "key"

	d.api.On(
		"GetItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("*dynamodb.GetItemInput"),
		[]request.Option(nil),
	).Return(&dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{"bacon": {S: aws.String("tasty")}},
	}, nil)

	ret, found, err := d.sut.Get(key)

	d.Empty(ret)
	d.False(found)
	d.EqualError(err, "value field not found in DynamoDB record")
}

func (d *dynamoDBStoreTestSuite) TestGet_NilValueField() {
	const key = "key"

	d.api.On(
		"GetItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("*dynamodb.GetItemInput"),
		[]request.Option(nil),
	).Return(&dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{"value": {}},
	}, nil)

	ret, found, err := d.sut.Get(key)

	d.Empty(ret)
	d.False(found)
	d.EqualError(err, "value field nil in DynamoDB record")
}

func (d *dynamoDBStoreTestSuite) TestSet_OK() {
	const key = "key"
	const value = "value"

	d.api.On(
		"PutItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.MatchedBy(func(in interface{}) bool {
			input, ok := in.(*dynamodb.PutItemInput)
			if !ok {
				return false
			}

			d.Len(input.Item, 2)
			d.Equal(key, *input.Item["key"].S)
			d.Equal(value, *input.Item["value"].S)

			d.Equal("table", *input.TableName)

			return true
		}),
		[]request.Option(nil),
	).Return((*dynamodb.PutItemOutput)(nil), nil)

	d.NoError(d.sut.Set(key, value))
}

func (d *dynamoDBStoreTestSuite) TestSet_APIError() {
	const key = "key"
	const value = "value"

	d.api.On(
		"PutItemWithContext",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("*dynamodb.PutItemInput"),
		[]request.Option(nil),
	).Return((*dynamodb.PutItemOutput)(nil), errors.New("bacon"))

	d.EqualError(d.sut.Set(key, value), "DynamoDB API error: bacon")
}

func TestDynamoDBStore(t *testing.T) {
	suite.Run(t, new(dynamoDBStoreTestSuite))
}
