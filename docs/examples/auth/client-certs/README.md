# Client Certificate Authentication

It is possible to enable Client-Certificate Authentication by adding additional annotations to your Ingress Resource.
Before getting started you must have the following Certificates Setup:

1. CA certificate and Key(Intermediate Certs need to be in CA)
2. Server Certificate(Signed by CA) and Key (CN should be equal the hostname you will use)
3. Client Certificate(Signed by CA) and Key

For more details on the generation process, checkout the Prerequisite [docs](../../PREREQUISITES.md#client-certificate-authentication).

You can have as many certificates as you want. If they're in the binary DER format, you can convert them as the following:

```bash
openssl x509 -in certificate.der -inform der -out certificate.crt -outform pem
```

Then, you can concatenate them all in only one file, named 'ca.crt' as the following:

```bash
cat certificate1.crt certificate2.crt certificate3.crt >> ca.crt
```

**Note:** Make sure that the Key Size is greater than 1024 and Hashing Algorithm(Digest) is something better than md5
for each certificate generated. Otherwise you will receive an error.

## Creating Certificate Secrets

There are many different ways of configuring your secrets to enable Client-Certificate
Authentication to work properly.

1. You can create a secret containing just the CA certificate and another
    Secret containing the Server Certificate which is Signed by the CA.

    ```bash
    kubectl create secret generic ca-secret --from-file=ca.crt=ca.crt
    kubectl create secret generic tls-secret --from-file=tls.crt=server.crt --from-file=tls.key=server.key
    ```

2. You can create a secret containing CA certificate along with the Server
    Certificate, that can be used for both TLS and Client Auth.

    ```bash
    kubectl create secret generic ca-secret --from-file=tls.crt=server.crt --from-file=tls.key=server.key --from-file=ca.crt=ca.crt
    ```

3. If you want to also enable Certificate Revocation List verification you can 
   create the secret also containing the CRL file in PEM format:
   ```bash
   kubectl create secret generic ca-secret --from-file=ca.crt=ca.crt --from-file=ca.crl=ca.crl
   ```

Note: The CA Certificate must contain the trusted certificate authority chain to verify client certificates.

## Setup Instructions

1. Add the annotations as provided in the [ingress.yaml](ingress.yaml) example to your own ingress resources as required.
2. Test by performing a curl against the Ingress Path without the Client Cert and expect a Status Code 400.
3. Test by performing a curl against the Ingress Path with the Client Cert and expect a Status Code 200.
