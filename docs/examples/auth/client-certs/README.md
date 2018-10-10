# Client Certificate Authentication
It is possible to enable Client Certificate Authentication using additional annotations in Ingress resources, created by you.

## Setup Instructions
1. Create a file named `ca.crt` containing the trusted certificate authority chain to verify client certificates. All of the certificates must be in PEM format.  
   *NB:* The file containing the trusted certificates must be named `ca.crt` exactly - this is expected to be found in the secret.

2. Create a secret from this file:  
`kubectl create secret generic auth-tls-chain --from-file=ca.crt --namespace=default`

3. Add the annotations as provided in the [ingress.yaml](ingress.yaml) example to your own ingress resources as required.
