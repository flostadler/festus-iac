# Pulumi experience project

## Feedback

- Import terraform modules as pulumi component libraries for awsx
    - For example https://registry.terraform.io/modules/terraform-aws-modules/iam/aws/latest
    - Or the AWS EKS module
        - I see we have https://github.com/pulumi/pulumi-eks , are we gonna be able to provide enough support for those “special components”. The EKS one alone seems like it’s quite a lot of work to maintain and keep up to date (e.g. it’s missing Karpenter support). If we could somehow reuse the official AWS modules like https://github.com/terraform-aws-modules/terraform-aws-eks it might take work off of our plate
        - pulumi-docs
- pulumi new. CLI doesn’t properly do line wrapping if description is longer than the current window
- Aws-go-lambda examples use outdated/deprecated AWS resources (e.g. lambda runtime go1.x)
- https://github.com/pulumi/examples/blob/master/aws-ts-apigateway-auth0/index.ts uses awsx instead of awsx/classic. Seems to be a 0.x migration leftover
    - Also some references seem to be wrong
- Cannot create esc environment with pulumi? Seems to be missing from pulumi cloud provider
- ESC login into the AWS console? I was able to get this working with a script, but this is something we could add to the aws-login provider
  - Why not have that in the web ui as well? Federated login into envs would mean a standardized way for engineers to get access to all tools.
- Doc: https://www.pulumi.com/docs/esc/reference/
    - Wrong reference ${context.organization.login} should be ${context.pulumi.organization.login}
- AWS IAM assume role policy for ESC oidc out of the box in cross walk?
- https://www.pulumi.com/docs/clouds/aws/guides/iam/#iam-roles
    - Wrong/outdated reference
        * aws.iam.IAMReadOnlyAccess vs aws.iam.ManagedPolicy.AWSOrganizationsFullAccess
