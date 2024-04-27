import * as apigateway from "@pulumi/aws-apigateway";
import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";
import * as std from "@pulumi/std";
import { authorizerLambda } from "./auth-lambda";

const config = new pulumi.Config();
const authParams = {
    jwksUri: config.require("jwksUri"),
    audience: config.require("audience"),
    issuer: config.require("issuer"),
    wildCardAuth: true,
}

const db = new aws.dynamodb.Table("festus-db", {
    billingMode: "PAY_PER_REQUEST",
    hashKey: "pk",
    rangeKey: "sk",
    name: "festus-db",
    streamEnabled: true,
    streamViewType: "NEW_IMAGE",
    attributes: [{
        name: "pk",
        type: "S",
    }, {
        name: "sk",
        type: "S",
    }],
});

const lambdaRole = new aws.iam.Role("festus-api-handler", {
    name: "festus-api-handler",
    assumeRolePolicy: aws.iam.getPolicyDocument({
        statements: [{
            effect: "Allow",
            principals: [{
                type: "Service",
                identifiers: ["lambda.amazonaws.com"],
            }],
            actions: ["sts:AssumeRole"],
        }],
    }).then(policy => policy.json),
});

const streamHandlerRole = new aws.iam.Role("festus-stream-handler", {
    name: "festus-stream-handler",
    assumeRolePolicy: aws.iam.getPolicyDocument({
        statements: [{
            effect: "Allow",
            principals: [{
                type: "Service",
                identifiers: ["lambda.amazonaws.com"],
            }],
            actions: ["sts:AssumeRole"],
        }],
    }).then(policy => policy.json),
});

new aws.iam.RolePolicyAttachment("festus-stream-handler-basic-execution-role", {
    role: streamHandlerRole,
    policyArn: aws.iam.ManagedPolicy.AWSLambdaBasicExecutionRole,
});

new aws.iam.RolePolicyAttachment("festus-api-handler-basic-execution-role", {
    role: lambdaRole,
    policyArn: aws.iam.ManagedPolicy.AWSLambdaBasicExecutionRole,
});

const ddbAccess = new aws.iam.Policy("festus-ddb-access", {
    policy: pulumi.interpolate`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "dynamodb:BatchGetItem",
                    "dynamodb:BatchWriteItem",
                    "dynamodb:DeleteItem",
                    "dynamodb:GetItem",
                    "dynamodb:PutItem",
                    "dynamodb:Query",
                    "dynamodb:Scan",
                    "dynamodb:UpdateItem"
                ],
                "Resource": "${db.arn}"
            }
        ]
    }`
});

new aws.iam.RolePolicyAttachment("stream handler-ddb-access", {
    role: streamHandlerRole,
    policyArn: ddbAccess.arn,
});

new aws.iam.RolePolicyAttachment("festus-ddb-access", {
    role: lambdaRole,
    policyArn: ddbAccess.arn,
});

const ddbStreamAccess = new aws.iam.Policy("festus-ddb-stream-access", {
    policy: pulumi.interpolate`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "dynamodb:GetRecords",
                    "dynamodb:GetShardIterator",
                    "dynamodb:DescribeStream",
                    "dynamodb:ListStreams"
                ],
                "Resource": "${db.streamArn}"
            }
        ]
    }`
}, );

new aws.iam.RolePolicyAttachment("festus-ddb-stream-access", {
    role: streamHandlerRole,
    policyArn: ddbStreamAccess.arn,
});

const streamProcessor = new aws.lambda.Function("streamProcessor", {
    code: new pulumi.asset.FileArchive("lambdas/out/account-update/function.zip"),
    name: "festus-stream-processor",
    role: streamHandlerRole.arn,
    handler: "dummy",
    timeout: 600,
    memorySize: 1769,
    sourceCodeHash: std.filebase64sha256({
        input: "lambdas/out/account-update/function.zip",
    }).then(invoke => invoke.result),
    runtime: aws.lambda.Runtime.CustomAL2,
    environment: {
        variables: {
            "TABLE_NAME": db.name,
        },
    },
    ephemeralStorage: { size: 2048 }
});

const eventSourceMapping = new aws.lambda.EventSourceMapping("myEventSourceMapping", {
    eventSourceArn: db.streamArn, // the ARN of the DynamoDB Stream
    functionName: streamProcessor.name,
    startingPosition: "TRIM_HORIZON",
    maximumRetryAttempts: 5
});

const apiHandler = new aws.lambda.Function("test_lambda", {
    code: new pulumi.asset.FileArchive("lambdas/out/api/function.zip"),
    name: "festus-api-handler",
    role: lambdaRole.arn,
    handler: "dummy",
    timeout: 30,
    sourceCodeHash: std.filebase64sha256({
        input: "lambdas/out/api/function.zip",
    }).then(invoke => invoke.result),
    runtime: aws.lambda.Runtime.CustomAL2,
    environment: {
        variables: {
            "TABLE_NAME": db.name,
        },
    },
});

const authorizer = {
    authType: "custom",
    authorizerName: "jwt-rsa-custom-authorizer",
    parameterName: "Authorization",
    identityValidationExpression: "^Bearer [-0-9a-zA-Z\._]*$",
    type: "token",
    parameterLocation: "header",
    authorizerResultTtlInSeconds: 300,
    handler: new aws.lambda.CallbackFunction("authorizer", {
        callback: authorizerLambda(authParams),
    }),
}

const api = new apigateway.RestAPI("festus", {
    stageName: "festus",
    routes: [{
        path: "/{proxy+}",
        method: "ANY",
        eventHandler: apiHandler,
        authorizers: [authorizer]
    }],
});

// Export the URL for our API
export const url = api.url;
