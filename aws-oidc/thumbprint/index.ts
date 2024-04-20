import { X509Certificate } from 'node:crypto';
import * as tls from 'tls';
import * as url from 'url';

/**
 * Downloads the CA thumbprint from the issuer URL
 */
export async function downloadThumbprint(issuerUrl: string) {
    return new Promise<string>((ok, ko) => {
        const purl = url.parse(issuerUrl);
        const port = purl.port ? parseInt(purl.port, 10) : 443;

        if (!purl.host) {
            return ko(new Error(`unable to determine host from issuer url ${issuerUrl}`));
        }

        console.log(`Fetching x509 certificate chain from issuer ${issuerUrl}`);

        const socket = tls.connect(port, purl.host, { rejectUnauthorized: false, servername: purl.host });
        socket.once('error', ko);

        socket.once('secureConnect', () => {
            let cert = socket.getPeerX509Certificate();
            if (!cert) {
                throw new Error(`Unable to retrieve X509 certificate from host ${purl.host}`);
            }
            while (cert.issuerCertificate) {
                printCertificate(cert);
                cert = cert.issuerCertificate;
            }
            const validTo = new Date(cert.validTo);
            const certificateValidity = getCertificateValidity(validTo);

            if (certificateValidity < 0) {
                return ko(new Error(`The certificate has already expired on: ${validTo.toUTCString()}`));
            }

            if (certificateValidity < 180) {
                console.warn(`The root certificate obtained would expire in ${certificateValidity} days!`);
            }

            socket.end();

            const thumbprint = extractThumbprint(cert);
            console.log(`Certificate Authority thumbprint for ${issuerUrl} is ${thumbprint}`);

            ok(thumbprint);
        });
    });
}

function extractThumbprint(cert: X509Certificate) {
    return cert.fingerprint.split(':').join('');
}

function printCertificate(cert: X509Certificate) {
    console.log('-------------BEGIN CERT----------------');
    console.log(`Thumbprint: ${extractThumbprint(cert)}`);
    console.log(`Valid To: ${cert.validTo}`);
    if (cert.issuerCertificate) {
        console.log(`Issuer Thumbprint: ${extractThumbprint(cert.issuerCertificate)}`);
    }
    console.log(`Issuer: ${cert.issuer}`);
    console.log(`Subject: ${cert.subject}`);
    console.log('-------------END CERT------------------');
}

function getCertificateValidity(certDate: Date): number {
    const millisecondsInDay = 24 * 60 * 60 * 1000;
    const currentDate = new Date();

    const validity = Math.round((certDate.getTime() - currentDate.getTime()) / millisecondsInDay);

    return validity;
}