import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

type EscAssumableIamRoleArgs = Omit<aws.iam.RoleArgs, "assumeRolePolicy"> & {
    environmentName: pulumi.Input<string>,
    projectName: pulumi.Input<string>,
    accountId?: pulumi.Input<string>
}

export class EscAssumableIamRole extends pulumi.ComponentResource {
    public iamRole: aws.iam.Role;

    /**
     * Creates an IAM trust policy that grants the specified environment in the current pulumi org to vend credentials for the IAM role
     */
    private getEscTrustPolicy(environmentName: pulumi.Input<string>, accountId: pulumi.Input<string>, projectName: pulumi.Input<string>): aws.iam.PolicyDocument {
        return {
            Version: "2012-10-17",
            Statement: [
                {
                    Action: "sts:AssumeRoleWithWebIdentity",
                    Principal: {
                        Federated: pulumi.interpolate`arn:aws:iam::${accountId}:oidc-provider/api.pulumi.com/oidc`
                    },
                    Effect: "Allow",
                    Condition: {
                        "StringEquals": {
                            "api.pulumi.com/oidc:aud": pulumi.getOrganization(),
                            "api.pulumi.com/oidc:sub": pulumi.interpolate`pulumi:environments:org:${pulumi.getOrganization()}:env:${environmentName}`
                        }
                    }
                },
                {
                    Action: "sts:AssumeRoleWithWebIdentity",
                    Principal: {
                        Federated: pulumi.interpolate`arn:aws:iam::${accountId}:oidc-provider/api.pulumi.com/oidc`
                    },
                    Effect: "Allow",
                    Condition: {
                        "StringEquals": {
                            "api.pulumi.com/oidc:aud": pulumi.getOrganization(),
                          },
                          "StringLike": {
                            "api.pulumi.com/oidc:sub": pulumi.interpolate`pulumi:deploy:org:${pulumi.getOrganization()}:project:${projectName}:*`
                          }
                    }
                }
            ]
        }
    };

    constructor(name: string, args: EscAssumableIamRoleArgs, opts?: pulumi.ComponentResourceOptions) {
        super("pkg:index:EscAssumableIamRole", name, args, opts);

        const accountId: pulumi.Input<string> = args.accountId ? args.accountId : aws.getCallerIdentity({}).then(id => id.accountId)
        this.iamRole = new aws.iam.Role(name, {
            ...args,
            assumeRolePolicy: this.getEscTrustPolicy(args.environmentName, accountId, args.projectName)
        }, opts);

        this.registerOutputs({
            name: this.iamRole.name,
            arn: this.iamRole.arn
        });
    }
}
