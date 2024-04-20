import * as apigateway from "@pulumi/aws-apigateway";
import * as awsx from "@pulumi/awsx/classic";
import * as pulumi from "@pulumi/pulumi";
import { authorizerLambda } from "./auth-lambda";

const config = new pulumi.Config();
const authParams = {
    jwksUri: config.require("jwksUri"),
    audience: config.require("audience"),
    issuer: config.require("issuer")
}

// Create our API and reference the Lambda authorizer
const api = new awsx.apigateway.API("myapi", {
    routes: [{
        path: "/hello",
        method: "GET",
        eventHandler: async () => {
            return {
                statusCode: 200,
                body: "<h1>Hello world!</h1>",
            };
        },
        authorizers: awsx.apigateway.getTokenLambdaAuthorizer({
            authorizerName: "jwt-rsa-custom-authorizer",
            header: "Authorization",
            handler: authorizerLambda(authParams),
            identityValidationExpression: "^Bearer [-0-9a-zA-Z\._]*$",
            authorizerResultTtlInSeconds: 3600,
        }),
    }],
});

// Export the URL for our API
export const url = api.url;
