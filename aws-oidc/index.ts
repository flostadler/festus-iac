import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import { downloadThumbprint } from "./thumbprint";

export = async () => {
    const audience = pulumi.getOrganization()
    const oidcUrl = 'https://api.pulumi.com/oidc'
    const thumbprint = await downloadThumbprint(oidcUrl)

    new aws.iam.OpenIdConnectProvider("PulumiOidcProvider", {
        clientIdLists: [audience],
        thumbprintLists: [thumbprint],
        url: oidcUrl
    })
}
