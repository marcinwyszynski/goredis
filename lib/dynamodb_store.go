package lib

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

const (
	apiErrorMessage = "DynamoDB API error"
	keyField        = "key"
	valueField      = "value"
)

var (
	// ErrNoValue is returned when there's no value field in the DynamoDB
	// record retrieved by key.
	ErrNoValue = errors.New("value field not found in DynamoDB record")

	// ErrNilValue is returned when there's a value field in the DynamoDB
	// record retrieved by key, but it
	ErrNilValue = errors.New("value field nil in DynamoDB record")
)

// DynamoDBStore is an implementation of the Store interface, backed by
// DynamoDB.
type DynamoDBStore struct {
	API       dynamodbiface.DynamoDBAPI
	TableName string
}

// Get is a DynamoDB implementation of the Store's Get method.
func (d *DynamoDBStore) Get(key string) (value string, found bool, err error) {
	// Let's make sure that requests never take more than a second. Anything
	// longer will return an API error.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := d.API.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		Key:       dynamoDBKey(key),
		TableName: aws.String(d.TableName),
	})
	if err != nil {
		err = errors.Wrap(err, apiErrorMessage)
		return
	}

	if len(out.Item) == 0 {
		return
	}

	valueField, exists := out.Item[valueField]
	if !exists {
		err = ErrNoValue
		return
	}

	if valueField.S == nil {
		err = ErrNilValue
		return
	}

	value, found = *valueField.S, true
	return
}

// Set is a DynamoDB implementation of the Store's Set method.
func (d *DynamoDBStore) Set(key string, value string) error {
	// Let's make sure that requests never take more than a second. Anything
	// longer will return an API error.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	item := dynamoDBKey(key)
	item[valueField] = &dynamodb.AttributeValue{S: aws.String(value)}

	_, err := d.API.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(d.TableName),
	})

	return errors.Wrap(err, apiErrorMessage)
}

func dynamoDBKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{keyField: {S: aws.String(key)}}
}
