import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

type EscAssumableIamRoleArgs = Omit<aws.iam.RoleArgs, "assumeRolePolicy"> & {
    environmentName: string,
}

export class EscAssumableIamRole extends pulumi.ComponentResource {
    public iamRole: aws.iam.Role;

    /**
     * Creates an IAM trust policy that grants the specified environment in the current pulumi org to vend credentials for the IAM role
     */
    private getEscTrustPolicy(environmentName: string): aws.iam.PolicyDocument {
        return {
            Version: "2012-10-17",
            Statement: [{
                Action: "sts:AssumeRole",
                Principal: {
                    Federated: `arn:aws:iam::${aws.getCallerIdentity()}:oidc-provider/api.pulumi.com/oidc`
                },
                Effect: "Allow",
                Condition: {
                    "StringEquals": {
                        "api.pulumi.com/oidc:aud": pulumi.getOrganization(),
                        "api.pulumi.com/oidc:sub": `pulumi:environments:org:${pulumi.getOrganization()}:env:${environmentName}`
                    }
                }
            }]
        }
    };

    constructor(name: string, args: EscAssumableIamRoleArgs, opts?: pulumi.ComponentResourceOptions) {
        super("pkg:index:EscAssumableIamRole", name, args, opts);

        this.iamRole = new aws.iam.Role(name, {
            ...args,
            assumeRolePolicy: this.getEscTrustPolicy(args.environmentName)
        });

        this.registerOutputs({
            name: this.iamRole.name,
            arn: this.iamRole.arn
        });
    }
}
