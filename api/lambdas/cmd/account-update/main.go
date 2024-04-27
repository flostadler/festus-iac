package main

import (
	"context"
	"fmt"
	"strings"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/flostadler/festus/api/pkg/db"
	"github.com/flostadler/festus/api/pkg/types"
	"github.com/flostadler/festus/api/pkg/iac"
)

var accountsDb *db.AccountDB
var orgDb *db.OrganizationDB

func init() {
	sess := session.Must(session.NewSession())
    ddb := dynamodb.New(sess, &aws.Config{
		LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
	})
	tableName := os.Getenv("TABLE_NAME")
	accountsDb = db.NewAccountDB(ddb, tableName)
	orgDb = db.NewOrganizationDB(ddb, tableName)
}

func Handler(ctx context.Context, e events.DynamoDBEvent) (error) {
	for _, record := range e.Records {
		if record.EventName != "INSERT" && record.EventName != "MODIFY" {
			continue
		}

		fmt.Printf("Processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)
		handleRecord := false
		for name, value := range record.Change.Keys {
			if name == "pk" {
				if value.DataType() != events.DataTypeString {
					fmt.Printf("Received invalid record that does not have a string as pk")
					break
				}

				if strings.HasPrefix(value.String(), "ACC#") {
					handleRecord = true
					break
				}
			}
		}

		if handleRecord {
			var acc db.AccountItem
			newImage := AttributeValueMapFrom(record.Change.NewImage)
			err := dynamodbattribute.UnmarshalMap(*newImage, &acc)
			if err != nil {
				return err
			}

			// the PK has the form of "ACC#:userId" => index 1 is the username
			userId := strings.Split(*(*newImage)["pk"].S, "#")[1]
			// the SK has the form of "ORG#:orgName#ACC#:accountName" => index 1 is the name of the org
			orgName := strings.Split(*(*newImage)["sk"].S, "#")[1]
			
			if acc.Status != int(types.Pending) {
				fmt.Printf("Ignoring account '%s' in org '%s'. Only handling new accounts", acc.AccountName, orgName)
				continue
			}

			version, err := strconv.Atoi(*(*newImage)["accountVersion"].N)
			if err != nil {
				println("item version is not of number type")
				return err
			}

			fmt.Printf("Creating account '%s' in org '%s' (current version %d)\n", acc.AccountName, orgName, version)

			err = accountsDb.UpdateStatus(userId, orgName, acc.AccountName, version, types.CreatingAccount)
			if err != nil {
				fmt.Printf("failed to update item type: %s", err.Error())
				return err
			}

			org, err := orgDb.GetItem(userId, orgName, false)
			if err != nil {
				fmt.Printf("failed to retrieve org: %s", err.Error())
				return err
			}

			if org == nil {
				fmt.Printf("org does not exist")
				return nil
			}

			account, err := accountsDb.GetItem(userId, orgName, acc.AccountName, true)
			if err != nil {
				fmt.Printf("failed to retrieve account: %s", err.Error())
				return err
			}
			if account == nil {
				fmt.Printf("account does not exist anymore")
				return nil
			}

			message, err := iac.CreateAccount(ctx, account, org)
			if err != nil {
				fmt.Printf("failed to apply stack: %s", err.Error())
				return err
			}

			println(message)
		}
	}
	return nil
}

func AttributeValueMapFrom(m map[string]events.DynamoDBAttributeValue) *map[string]*dynamodb.AttributeValue {
	result := map[string]*dynamodb.AttributeValue{}
	for k, v := range m {
		result[k] = AttributeValueFrom(v)
	}
	return &result
}

// AttributeValueFrom converts from events.DynamoDBAttributeValue to dynamodb.AttributeValue
func AttributeValueFrom(from events.DynamoDBAttributeValue) *dynamodb.AttributeValue {
	attr := dynamodb.AttributeValue{}
	switch from.DataType() {
	case events.DataTypeBinary:
		return attr.SetB(from.Binary())
	case events.DataTypeBinarySet:
		return attr.SetBS(from.BinarySet())
	case events.DataTypeBoolean:
		return attr.SetBOOL(from.Boolean())
	case events.DataTypeList:
		var vs []*dynamodb.AttributeValue
		for _, v := range from.List() {
			lv := AttributeValueFrom(v)
			vs = append(vs, lv)
		}
		return attr.SetL(vs)
	case events.DataTypeMap:
		mv := map[string]*dynamodb.AttributeValue{}
		for k, v := range from.Map() {
			mv[k] = AttributeValueFrom(v)
		}
		return attr.SetM(mv)
	case events.DataTypeNull:
		return attr.SetNULL(from.IsNull())
	case events.DataTypeNumber:
		return attr.SetN(from.Number())
	case events.DataTypeNumberSet:
		var ns []*string
		for _, v := range from.NumberSet() {
			ns = append(ns, &v)
		}
		return attr.SetNS(ns)
	case events.DataTypeString:
		return attr.SetS(from.String())
	case events.DataTypeStringSet:
		var ss []*string
		for _, v := range from.StringSet() {
			ss = append(ss, &v)
		}
		return attr.SetSS(ss)
	default:
		panic(fmt.Errorf("unknown ddb type: %v", from.DataType()))
	}
}

func main() {
	lambda.Start(Handler)
}
