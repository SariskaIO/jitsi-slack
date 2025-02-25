package jitsi

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	KeyTeamID      = "team-id"      // primary key; slack team id
	KeyAccessToken = "access-token" // oauth access token
)

// TokenData is the access token data stored from oauth.
type TokenData struct {
	TeamID      string `json:"team-id"`
	AccessToken string `json:"access-token"`
}

// TokenStore stores and retrieves access tokens from aws dynamodb.
type TokenStore struct {
	TableName string
	DB        *dynamodb.Client
}

// GetToken retrieves the access token stored with the provided team id.
func (t *TokenStore) GetTokenForTeam(teamID string) (*TokenData, error) {
	keyCond := expression.Key(KeyTeamID).Equal(expression.Value(teamID))
	keyCond1 := expression.Key("sariska").Equal(expression.Value("sariska"))
	builder := expression.NewBuilder().WithKeyCondition(keyCond).WithKeyCondition(keyCond1)
	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}
	queryInput := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 aws.String(t.TableName),
	}
	result, err := t.DB.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, err
	}

	if len(result.Items) < 1 {
		return nil, errors.New(errMissingAuthToken)
	}

	var token string

	err = attributevalue.Unmarshal(result.Items[0]["AccessToken"], &token)

	if err != nil {
		return nil, err
	}

	return &TokenData{
		TeamID:      teamID,
		AccessToken: token,
	}, nil
}

// Store will store access token data.
func (t *TokenStore) Store(data *TokenData) error {

	_, err := t.DB.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(t.TableName),
		Item: map[string]types.AttributeValue{
			"sariska":     &types.AttributeValueMemberS{Value: "sariska"},
			"TeamID":      &types.AttributeValueMemberS{Value: data.TeamID},
			"AccessToken": &types.AttributeValueMemberS{Value: data.AccessToken},
		},
	})
	return err

}

// Remove will remove access token data for the user.
func (t *TokenStore) Remove(teamID string) error {
	av, err := attributevalue.MarshalMap(map[string]string{
		"TeamID":  teamID,
		"sariska": "sariska",
	})
	dii := &dynamodb.DeleteItemInput{
		TableName: aws.String(t.TableName),
		Key:       av,
	}
	_, err = t.DB.DeleteItem(context.TODO(), dii)
	return err
}
