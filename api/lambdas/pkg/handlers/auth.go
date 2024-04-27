package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

const UserIDKey = "UserID"
const RequestIDKey = "RequestID"

func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestIDKey)
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ctx, ok := core.GetAPIGatewayContextFromContext(c.Request.Context()); ok {
			c.Set(RequestIDKey, ctx.RequestID)
			if ctx.Authorizer != nil && ctx.Authorizer["principalId"] != nil && ctx.Authorizer["principalId"].(string) != "" {
				c.Set(UserIDKey, ctx.Authorizer["principalId"])
				c.Next()
			} else {
				c.AbortWithError(500, core.NewLoggedError("No principalId found in authorizer"))
			}
		} else {
			c.AbortWithError(500, core.NewLoggedError("No API Gateway context found"))
		}
	}
}
