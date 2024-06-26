import * as awslambda from "aws-lambda";
import * as jwt from "jsonwebtoken";
import * as jwksClient from "jwks-rsa";
import * as util from "util";

type AuthorizerLambda = (event: awslambda.APIGatewayAuthorizerEvent) => Promise<awslambda.APIGatewayAuthorizerResult>
type AuthParameters = {
    jwksUri: string,
    audience: string,
    issuer: string,
    wildCardAuth?: boolean,
  }

export function authorizerLambda(params: AuthParameters): AuthorizerLambda {
    return async (event: awslambda.APIGatewayAuthorizerEvent) => {
        try {
            return await authenticate(event, params);
        }
        catch (err) {
            console.log(err);
            // Tells API Gateway to return a 401 Unauthorized response
            throw new Error("Unauthorized");
        }
    }
}

/**
 * Below is all code that gets added to the Authorizer Lambda. The code was copied and
 * converted to TypeScript from [Auth0's GitHub
 * Example](https://github.com/auth0-samples/jwt-rsa-aws-custom-authorizer)
 */

// Extract and return the Bearer Token from the Lambda event parameters
function getToken(event: awslambda.APIGatewayAuthorizerEvent): string {
    if (!event.type || event.type !== "TOKEN") {
        throw new Error('Expected "event.type" parameter to have value "TOKEN"');
    }

    const tokenString = event.authorizationToken;
    if (!tokenString) {
        throw new Error('Expected "event.authorizationToken" parameter to be set');
    }

    const match = tokenString.match(/^Bearer (.*)$/);
    if (!match) {
        throw new Error(`Invalid Authorization token - ${tokenString} does not match "Bearer .*"`);
    }
    return match[1];
}

// Check the Token is valid with Auth0
async function authenticate(event: awslambda.APIGatewayAuthorizerEvent, params: AuthParameters): Promise<awslambda.APIGatewayAuthorizerResult> {
    console.log(event);
    const token = getToken(event);

    const decoded = jwt.decode(token, { complete: true });
    if (!decoded || typeof decoded === "string" || !decoded.header || !decoded.header.kid) {
        throw new Error("invalid token");
    }

    const client = jwksClient({
        cache: true,
        rateLimit: true,
        jwksRequestsPerMinute: 10, // Default value
        jwksUri: params.jwksUri,
    });

    const key = await util.promisify(client.getSigningKey)(decoded.header.kid);
    const signingKey = key.getPublicKey();
    if (!signingKey) {
        throw new Error("could not get signing key");
    }

    const verifiedJWT = await jwt.verify(token, signingKey, { audience: params.audience, issuer: params.issuer });
    if (!verifiedJWT || typeof verifiedJWT === "string" || !isVerifiedJWT(verifiedJWT)) {
        throw new Error("could not verify JWT");
    }

    const methodArn = getMethodArn(event, params);
    console.log(`Method ARN: ${methodArn}`);
    return {
        principalId: verifiedJWT.sub,
        policyDocument: {
            Version: "2012-10-17",
            Statement: [{
                Action: "execute-api:Invoke",
                Effect: "Allow",
                Resource: methodArn,
            }],
        },
    };
}

function getMethodArn(event: awslambda.APIGatewayAuthorizerEvent, params: AuthParameters): string {
    if (!event.methodArn) {
        throw new Error('Expected "event.methodArn" parameter to be set');
    }

    if (params.wildCardAuth && params.wildCardAuth === true) {
        // split methodarn by /
        const arnPartials = event.methodArn.split("/");
        return arnPartials.slice(0, 2).join("/") + "/*";
    }

    return event.methodArn;
}

interface VerifiedJWT {
    sub: string;
}

function isVerifiedJWT(toBeDetermined: VerifiedJWT | Object): toBeDetermined is VerifiedJWT {
    return (<VerifiedJWT>toBeDetermined).sub !== undefined;
}
