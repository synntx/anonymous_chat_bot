package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type User struct {
	ChatId       int64 `dynamodbav:"ChatId"`
	IsConnecting int   `dynamodbav:"IsConnecting"`
	IsConnected  bool  `dynamodbav:"IsConnected"`
	Partner      int64 `dynamodbav:"Partner,omitempty"`
}

type DynamoDBStore struct {
	Client    *dynamodb.Client
	TableName string
}

func New(ctx context.Context, tableName string) (*DynamoDBStore, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if os.Getenv("AWS_SAM_LOCAL") == "true" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://host.docker.internal:8000",
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBStore{Client: client, TableName: tableName}, nil
}

func (s *DynamoDBStore) GetUser(ctx context.Context, chatId int64) (*User, error) {
	key, err := attributevalue.Marshal(chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key:       map[string]types.AttributeValue{"ChatId": key},
	}

	result, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}
	if result.Item == nil {
		return nil, errors.New("user not found")
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %w", err)
	}
	return &user, nil
}

func (s *DynamoDBStore) UpdateUser(ctx context.Context, user *User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user into DynamoDB item: %w", err)
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item:      item,
	}
	_, err = s.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item to DynamoDB: %w", err)
	}
	return nil
}

func (s *DynamoDBStore) FindAndConnectPartner(ctx context.Context, me *User) (*User, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.TableName),
		IndexName:              aws.String("IsConnectingIndex"),
		KeyConditionExpression: aws.String("IsConnecting = :connecting"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":connecting": &types.AttributeValueMemberN{Value: "1"},
		},
		Limit: aws.Int32(20),
	}

	result, err := s.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query for partners: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var partner *User
	for _, item := range result.Items {
		var p User
		if err := attributevalue.UnmarshalMap(item, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal partner item: %w", err)
		}
		if p.ChatId != me.ChatId {
			partner = &p
			break
		}
	}

	if partner == nil {
		return nil, nil
	}

	me.IsConnected = true
	me.IsConnecting = 0
	me.Partner = partner.ChatId

	partner.IsConnected = true
	partner.IsConnecting = 0
	partner.Partner = me.ChatId

	mePut, err := s.createPut(me)
	if err != nil {
		return nil, err
	}
	partnerPut, err := s.createPut(partner)
	if err != nil {
		return nil, err
	}

	txInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{Put: mePut},
			{Put: partnerPut},
		},
	}

	_, err = s.Client.TransactWriteItems(ctx, txInput)
	if err != nil {
		return nil, fmt.Errorf("failed to execute connect transaction: %w", err)
	}

	return partner, nil
}

func (s *DynamoDBStore) createPut(user *User) (*types.Put, error) {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user for transaction: %w", err)
	}
	return &types.Put{
		TableName: aws.String(s.TableName),
		Item:      item,
	}, nil
}
