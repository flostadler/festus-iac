import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as awsx from "@pulumi/awsx";
import * as thumbprints from './thumbprints'
import * as esc from './esc'

type OrgElement

const orgStructure = {
    "Sandbox": {},
    "Festus": ["Dev", "Prod"]
}

const oidcUrl = 'https://api.pulumi.com/oidc'

export = async () => {
    const pulumiOrg = pulumi.getOrganization()
    const thumbprint = await thumbprints.downloadThumbprint(oidcUrl)

    const organization = new aws.organizations.Organization("flostadler");

    const role = new esc.EscAssumableIamRole("OrgAdmin", {
        environmentName: "AwsOrgManagement"
    })
    new aws.iam.RolePolicyAttachment("OrgAdmin", {
        role: role.iamRole,
        policyArn: aws.iam.ManagedPolicy.AWSOrganizationsFullAccess,
    });



    const devOrgUnit = new aws.organizations.OrganizationalUnit("orgUnit", {
        parentId: organization.roots[0].id,
        name: "Development",
    });

    devOrgUnit.id

    return {

    }
}

function createOrganization(parentId: pulumi.Output<String>, name: string, childs: )
