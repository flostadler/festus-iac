package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/flostadler/festus/api/pkg/types"
)

type AccountItem struct {
	Pk              string `dynamodbav:"pk"`
	Sk              string `dynamodbav:"sk"`
	AccountName     string `dynamodbav:"accountName"`
	Email           string `dynamodbav:"email"`
	ParentID        string `dynamodbav:"parentID"`
	AwsAccessKey    string `dynamodbav:"awsAccessKey"`
	AwsSecretKey    string `dynamodbav:"awsSecretKey"`
	AwsSessionToken string `dynamodbav:"awsSessionToken"`
	Version int `dynamodbav:"accountVersion"`
	Status int `dynamodbav:"accountStatus"`
}

type AccountDB struct {
	tableName string
	ddb       *dynamodb.DynamoDB
}

func NewAccountDB(ddb *dynamodb.DynamoDB, tableName string) *AccountDB {
	return &AccountDB{ddb: ddb, tableName: tableName}
}

func (db *AccountDB) PutItem(UserID string, orgName string, account *types.Account) (*types.Account, error) {
	// TODO: don't store AWS credentials but rather assume them using ESC
	accountItem := AccountItem{
		Pk:              getAccountPk(UserID),
		Sk:              getAccountSk(orgName, account.AccountName),
		AccountName:     account.AccountName,
		Email:           account.Email,
		ParentID:        account.ParentID,
		AwsAccessKey:    account.AwsAccessKey,
		AwsSecretKey:    account.AwsSecretKey,
		AwsSessionToken: account.AwsSessionToken,
		Status: int(account.Status),
		Version: 0,
	}

	item, err := dynamodbattribute.MarshalMap(accountItem)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	}

	_, err = db.ddb.PutItem(input)
	if err != nil {
		return nil, err
	}

	return db.GetItem(UserID, orgName, account.AccountName, true)
}

func (db *AccountDB) GetItem(userID string, orgName string, accountName string, consistentRead bool) (*types.Account, error) {
	_, acc, err := db.GetItemWithVersion(userID, orgName, accountName, consistentRead)
	return acc, err
}

func (db *AccountDB) GetItemWithVersion(userID string, orgName string, accountName string, consistentRead bool) (int, *types.Account, error) {
	input := &dynamodb.GetItemInput{
		TableName:      aws.String(db.tableName),
		ConsistentRead: aws.Bool(consistentRead),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(getAccountPk(userID)),
			},
			"sk": {
				S: aws.String(getAccountSk(orgName, accountName)),
			},
		},
	}

	result, err := db.ddb.GetItem(input)
	if err != nil {
		return 0, nil, err
	}

	if result.Item == nil {
		return 0, nil, nil
	}

	var acc AccountItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &acc)
	if err != nil {
		return 0, nil, err
	}

	return acc.Version, &types.Account{
		AccountName: acc.AccountName,
		Email: acc.Email,
		ParentID: acc.ParentID,
		AwsAccessKey: acc.AwsAccessKey,
		AwsSecretKey: acc.AwsSecretKey,
		AwsSessionToken: acc.AwsSessionToken,
		Status: types.AccountStatus(acc.Status),
	}, nil
}

func (db *AccountDB) DeleteItem(userID string, orgName string, accountName string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(getAccountPk(userID)),
			},
			"sk": {
				S: aws.String(getAccountSk(orgName, accountName)),
			},
		},
	}

	_, err := db.ddb.DeleteItem(input)
	return err
}

func (db *AccountDB) UpdateStatus(userID string, orgName string, accountName string, expectedVersion int, status types.AccountStatus) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(getAccountPk(userID)),
			},
			"sk": {
				S: aws.String(getAccountSk(orgName, accountName)),
			},
		},
		ConditionExpression: aws.String("accountVersion = :expectedVersion"),
		UpdateExpression: aws.String("SET accountVersion = accountVersion + :increment, accountStatus = :newStatus"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":expectedVersion": {
				N: aws.String(fmt.Sprintf("%d", expectedVersion)),
			},
			":increment": {
				N: aws.String("1"),
			},
			":newStatus": {
				N: aws.String(fmt.Sprintf("%d", int(status))),
			},
		},
	}

	_, err := db.ddb.UpdateItem(input)
	return err
}

func getAccountPk(UserID string) string {
	return fmt.Sprintf("ACC#%s", UserID)
}

func getAccountSk(orgName string, accountName string) string {
	return fmt.Sprintf("ORG#%s#ACC#%s", orgName, accountName)
}
