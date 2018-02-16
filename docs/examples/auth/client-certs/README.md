# Client Certificate Authentication

It is possible to enable Client Certificate Authentication using additional annotations in the Ingress.

## Setup instructions
1. Create a file named `ca.crt` containing the trusted certificate authority chain (all ca certificates in PEM format) to verify client certificates. 
 
2. Create a secret from this file:
`kubectl create secret generic auth-tls-chain --from-file=ca.crt --namespace=default`

3. Add the annotations as provided in the [ingress.yaml](ingress.yaml) example to your ingress object.