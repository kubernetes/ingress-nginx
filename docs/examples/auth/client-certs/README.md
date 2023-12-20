# Client Certificate Authentication

It is possible to enable Client-Certificate Authentication by adding additional annotations to your Ingress Resource.

## 1. Prerequisites / Certificates

- Certificate Authority (CA) Certificate ```ca-cert.pem```
- Server Certificate (Signed by CA) and Key ```server-cert.pem``` and ```server-key.pem```
- Client Certificate (Signed by CA), Key and CA Certificate for following client side authentication (See Sub-Section 4 - Test)

:memo: If Intermediate CA-Certificates (Official CA, non-self-signed) used, they all need to be concatenated (CA authority chain) in one CA file.

The following commands let you generate self-signed Certificates and Keys for testing-purpose.

- Generate the CA Key and Certificate:

```bash
openssl req -x509 -sha256 -newkey rsa:4096 -keyout ca-key.der -out ca-cert.der -days 356 -nodes -subj '/CN=My Cert Authority'
```

- Generate the Server Key, and Certificate and Sign with the CA Certificate:

```bash
openssl req -new -newkey rsa:4096 -keyout server-key.der -out server.csr -nodes -subj '/CN=mydomain.com'
openssl x509 -req -sha256 -days 365 -in server.csr -CA ca-cert.der -CAkey ca-key.der -set_serial 01 -out server-cert.der
```

:memo: The CN (Common Name) x.509 attribute for the server Certificate ***must*** match the dns hostname referenced in ingress definition, see example below.

- Generate the Client Key, and Certificate and Sign with the CA Certificate:

```bash
openssl req -new -newkey rsa:4096 -keyout client-key.der -out client.csr -nodes -subj '/CN=My Client'
openssl x509 -req -sha256 -days 365 -in client.csr -CA ca-cert.der -CAkey ca-key.der -set_serial 02 -out client-cert.der
```

## 2. Import Certificates / Keys to Kubernetes Secret-Backend

- Convert all files specified in 1) from .der (binary format) to .pem (base64 encoded):

```bash
openssl x509 -in certificate.der -inform der -out certificate.crt -outform pem
```

:exclamation: Kubernetes Web-Services import relies on .pem Base64-encoded format.

:zap: There is no need to import the CA Private Key, the Private Key is used only to sign new Client Certificates by the CA.

- Import the CA Certificate as Kubernetes sub-type ```generic/ca.crt```

```bash
kubectl create secret generic ca-secret --from-file=ca.crt=./ca-cert.pem
```

- Import the Server Certificate and Key as Kubernetes sub-type ```tls``` for transport layer

```bash
kubectl create secret tls tls-secret --cert ./server-cert.pem --key ./server-key.pem
```

- Optional import CA-cert, Server-cert and Server-Key for TLS and Client-Auth

```bash
kubectl create secret generic tls-and-auth --from-file=tls.crt=./server-crt.pem --from-file=tls.key=./server-key.pem --from-file=ca.crt=./ca-cert.pem
```

- Optional import a CRL (Certificate Revocation List)

```bash
kubectl create secret generic ca-secret --from-file=ca.crt=./ca-cert.pem --from-file=ca.crl=./ca-crl.pem
```

## 3. Annotations / Ingress-Reference

Now we are able to reference the created secrets in the ingress definition.

:memo: The CA Certificate "authentication" will be reference in annotations.

| Annotation                                                                | Description                | Remark             |
|---------------------------------------------------------------------------|----------------------------|--------------------|
| nginx.ingress.kubernetes.io/auth-tls-verify-client: "on"                  | Activate Client-Auth       | If "on", verify client Certificate |
| nginx.ingress.kubernetes.io/auth-tls-secret: "namespace/ca-secret"        | CA "secret" reference      | Secret namespace and service / ingress namespace must match |
| nginx.ingress.kubernetes.io/auth-tls-verify-depth: "1"                    | CA "chain" depth           | How many CA levels should be processed |
| nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream: "true" | Pass Cert / Header         | Pass Certificate to Web-App for e.g. parsing Client E-Mail Address x.509 Property |

:memo: The Server Certificate for transport layer will be referenced in tls .yaml subsection.

```yaml
tls:
  - hosts:
    - mydomain.com
    secretName: tls-secret
```

## 4. Example / Test

The working .yaml Example: [ingress.yaml](ingress.yaml)

- Test by performing a curl / wget against the Ingress Path without the Client Cert and expect a Status Code 400 (Bad Request - No required SSL certificate was sent).
- Test by performing a curl / wget against the Ingress Path with the Client Cert and expect a Status Code 200.

```bash
wget \
--ca-cert=ca-cert.pem \
--certificate=client-cert.pem \
--private-key=client-key.pem \
https://mydomain.com
```

## 5. Remarks

| :exclamation: In future releases, CN verification seems to be "replaced" by SAN (Subject Alternate Name) for verrification, so do not forget to add |
|-----------------------------------------------------------------------------------------------------------------------------------------------------|

```bash
openssl req -addext "subjectAltName = DNS:mydomain.com" ...
```

