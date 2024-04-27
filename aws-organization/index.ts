import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as thumbprints from './thumbprints'
import * as esc from './esc'

type OidcConfig = { audience: string, thumbprint: string, url: string }

type Account = string
type OrgUnit = { name: string, children: OrgElement[] }
type OrgElement = Account | OrgUnit

const orgStructure: OrgElement[] = [
    { name: "sandbox", children: [] },
    { name: "festus", children: ["dev", "prod"]}
]

const oidcUrl = 'https://api.pulumi.com/oidc'
const orgManagementRole = "OrganizationalAccountAccessRole"
const emailContact = "flrnstdlr@gmail.com"

const config = new pulumi.Config();

export = async () => {
    const pulumiOrg = pulumi.getOrganization()
    const thumbprint = await thumbprints.downloadThumbprint(oidcUrl)
    const oidcConfig: OidcConfig = { audience: pulumiOrg, thumbprint, url: oidcUrl }

    const organization = new aws.organizations.Organization("flostadler");
    const rootOidcProvider = createOidcProvider("RootOidcProvider", oidcConfig)

    const role = new esc.EscAssumableIamRole("OrgAdmin", {
        name: "OrgAdmin",
        environmentName: "AwsOrgManagement",
        projectName: "AwsOrgManagement",
    })
    new aws.iam.RolePolicyAttachment("OrgAdmin", {
        role: role.iamRole,
        policyArn: aws.iam.ManagedPolicy.AWSOrganizationsFullAccess,
    });

    // TODO: Create ESC environment for each account
    orgStructure.forEach(element => {
        createOrgElement(organization.roots[0].id, element, oidcConfig)
    })

    return {

    }
}

function createOrgElement(parentId: pulumi.Output<string>, element: OrgElement, oidcConfig: OidcConfig, parentUnit?: OrgUnit) {
    if (typeof element === 'string') {
        createAccount(parentId, element, oidcConfig, parentUnit)
    } else {
        createOrganizationalUnit(parentId, element, oidcConfig)
    }
}

function createOrganizationalUnit(parentId: pulumi.Output<string>, orgUnit: OrgUnit, oidcConfig: OidcConfig) {
    const ou = new aws.organizations.OrganizationalUnit(orgUnit.name, {
        parentId: parentId,
        name: orgUnit.name,
    });

    orgUnit.children.forEach(child => {
        createOrgElement(ou.id, child, oidcConfig, orgUnit)
    })
}

function createAccount(parentId: pulumi.Output<string>, name: string, oidcConfig: OidcConfig, parentUnit?: OrgUnit) {
    const fqn = parentUnit ? `${parentUnit.name}-${name}` : name
    const account =new aws.organizations.Account(
        fqn,
        {
            name: fqn,
            parentId: parentId,
            email: createEmailAlias(emailContact, fqn),
            roleName: orgManagementRole,
            closeOnDeletion: true,
        },
        { protect: true }
    );

    const accountProvider = new aws.Provider(`${fqn}-provider`, {
        allowedAccountIds: [account.id],
        region: pulumi.output(aws.getRegion()).apply(region => region.name as aws.Region),
        assumeRole: {
            roleArn: pulumi.interpolate`arn:aws:iam::${account.id}:role/${orgManagementRole}`,
        },
    });

    createOidcProvider(`${fqn}-oidc`, oidcConfig, { provider: accountProvider })

    // create an Admin and Developer Role in every account
    const developerRole = new esc.EscAssumableIamRole(`${fqn}-developer`, {
        name: "Developer",
        environmentName: `${fqn}-developer`,
        accountId: account.id,
        projectName: parentUnit ? parentUnit.name : fqn
    }, { provider: accountProvider })
    new aws.iam.RolePolicyAttachment(`${fqn}-developer`, {
        role: developerRole.iamRole,
        policyArn: aws.iam.ManagedPolicy.PowerUserAccess,
    }, { provider: accountProvider });

    const adminRole = new esc.EscAssumableIamRole(`${fqn}-admin`, {
        name: "Admin",
        environmentName: `${fqn}-admin`,
        accountId: account.id,
        projectName: parentUnit ? parentUnit.name : fqn
    }, { provider: accountProvider })
    new aws.iam.RolePolicyAttachment(`${fqn}-admin`, {
        role: adminRole.iamRole,
        policyArn: aws.iam.ManagedPolicy.AdministratorAccess,
    }, { provider: accountProvider });
}

function createOidcProvider(id: string, oidcConfig: OidcConfig, opts?: pulumi.CustomResourceOptions) {
    return new aws.iam.OpenIdConnectProvider(id, {
        clientIdLists: [oidcConfig.audience],
        thumbprintLists: [oidcConfig.thumbprint],
        url: oidcConfig.url
    }, {...opts, ignoreChanges: ["thumbprintLists"]})
}

function createEmailAlias(email: string, alias: string): string {
    const [localPart, domain] = email.split("@");
    return `${localPart}+${alias}@${domain}`;
}
