package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/flostadler/festus/api/pkg/db"
	"github.com/flostadler/festus/api/pkg/handlers"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin cold start")

	sess := session.Must(session.NewSession())
    ddb := dynamodb.New(sess)

	orgDB := db.NewOrganizationDB(ddb, os.Getenv("TABLE_NAME"))
	accountDB := db.NewAccountDB(ddb, os.Getenv("TABLE_NAME"))
	orgHandler := handlers.NewOrganizationHandler(orgDB)
	accountsHandler := handlers.NewAccountsHandler(orgDB, accountDB)

	r := gin.Default()
	r.Use(handlers.Auth())

	root := r.Group("/", handlers.Auth())

	orgs := root.Group("/organizations")
	{
		orgs.POST("", orgHandler.CreateOrganization)
		orgs.GET("/:organizationName", orgHandler.GetOrganization)
		orgs.DELETE("/:organizationName", orgHandler.DeleteOrganization)
		accounts := orgs.Group("/:organizationName/accounts")
		{
			accounts.POST("", accountsHandler.CreateAccount)
			accounts.GET("/:accountName", accountsHandler.GetAccount)
			accounts.DELETE("/:accountName", accountsHandler.DeleteAccount)
		}
	}

	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"userID":  handlers.GetUserID(c),
			"requestID": handlers.GetRequestID(c),
		})
	})

	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Handling request for path: %s method: %s request ID: %s", req.Path, req.HTTPMethod, req.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
