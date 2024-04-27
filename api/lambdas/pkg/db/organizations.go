package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/flostadler/festus/api/pkg/types"
)

type OrganizationItem struct {
	Pk 					   	   string `dynamodbav:"pk"`
	OrgName                    string `dynamodbav:"sk"`
    PulumiAccessToken          string `dynamodbav:"pulumiAccessToken"`
    OrgManagementEnvironment   string `dynamodbav:"orgManagementEnvironment"`
}

type OrganizationDB struct {
	tableName string
	ddb *dynamodb.DynamoDB
}

func NewOrganizationDB(ddb *dynamodb.DynamoDB, tableName string) *OrganizationDB {
	return &OrganizationDB{ddb: ddb, tableName: tableName}
}

func (db *OrganizationDB) PutItem(UserID string, org *types.Organization) (*types.Organization, error) {
	// TODO: pulumi access token should be encrypted
	orgItem := OrganizationItem{
		Pk: getOrgPk(UserID),
		OrgName: org.OrgName,
		PulumiAccessToken: org.PulumiAccessToken,
		OrgManagementEnvironment: org.OrgManagementEnvironment,
	}

	item, err := dynamodbattribute.MarshalMap(orgItem)
    if err != nil {
        return nil, err
    }

    input := &dynamodb.PutItemInput{
        TableName: aws.String(db.tableName),
        Item: item,
    }

    _, err = db.ddb.PutItem(input)
	if err != nil {
        return nil, err
    }

	return db.GetItem(UserID, org.OrgName, true)
}

func (db *OrganizationDB) GetItem(userID string, orgName string, consistentRead bool) (*types.Organization, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		ConsistentRead: aws.Bool(consistentRead),
		Key: map[string]*dynamodb.AttributeValue{
            "pk": {
                S: aws.String(getOrgPk(userID)),
            },
            "sk": {
                S: aws.String(orgName),
            },
        },
	}

    result, err := db.ddb.GetItem(input)
    if err != nil {
        return nil, err
    }

    if result.Item == nil {
        return nil, nil
    }

    var org OrganizationItem
    err = dynamodbattribute.UnmarshalMap(result.Item, &org)
    if err != nil {
        return nil, err
    }

    return &types.Organization{
		OrgName: org.OrgName,
		PulumiAccessToken: org.PulumiAccessToken,
		OrgManagementEnvironment: org.OrgManagementEnvironment,
	}, nil
}

// Delete item
func (db *OrganizationDB) DeleteItem(userID string, orgName string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(getOrgPk(userID)),
			},
			"sk": {
				S: aws.String(orgName),
			},
		},
	}

	_, err := db.ddb.DeleteItem(input)
	return err
}

func getOrgPk(UserID string) string {
	return fmt.Sprintf("ORG#%s", UserID)
}
