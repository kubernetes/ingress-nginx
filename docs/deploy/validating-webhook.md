# Validating webhook (admission controller)

## Overview

Nginx ingress controller offers the option to validate ingresses before they enter the cluster, ensuring controller will generate a valid configuration.

This controller is called, when [ValidatingAdmissionWebhook][1] is enabled, by the Kubernetes API server each time a new ingress is to enter the cluster, and rejects objects for which the generated nginx configuration fails to be validated.

This feature requires some further configuration of the cluster, hence it is an optional feature, this section explains how to enable it for your cluster.

## Configure the webhook

### Generate the webhook certificate


#### Self signed certificate

Validating webhook must be served using TLS, you need to generate a certificate. Note that kube API server is checking the hostname of the certificate, the common name of your certificate will need to match the service name.

!!! example
    To run the validating webhook with a service named `ingress-validation-webhook` in the namespace `ingress-nginx`, run

    ```bash
    openssl req -x509 -newkey rsa:2048 -keyout certificate.pem -out key.pem -days 365 -nodes -subj "/CN=ingress-validation-webhook.ingress-nginx.svc"
    ```

##### Using Kubernetes CA

Kubernetes also provides primitives to sign a certificate request. Here is an example on how to use it

!!! example
    ```
    #!/bin/bash

    SERVICE_NAME=ingress-nginx
    NAMESPACE=ingress-nginx

    TEMP_DIRECTORY=$(mktemp -d)
    echo "creating certs in directory ${TEMP_DIRECTORY}"

    cat <<EOF >> ${TEMP_DIRECTORY}/csr.conf
    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name
    [req_distinguished_name]
    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    extendedKeyUsage = serverAuth
    subjectAltName = @alt_names
    [alt_names]
    DNS.1 = ${SERVICE_NAME}
    DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
    DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
    EOF

    openssl genrsa -out ${TEMP_DIRECTORY}/server-key.pem 2048
    openssl req -new -key ${TEMP_DIRECTORY}/server-key.pem \
        -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" \
        -out ${TEMP_DIRECTORY}/server.csr \
        -config ${TEMP_DIRECTORY}/csr.conf

    cat <<EOF | kubectl create -f -
    apiVersion: certificates.k8s.io/v1beta1
    kind: CertificateSigningRequest
    metadata:
      name: ${SERVICE_NAME}.${NAMESPACE}.svc
    spec:
      request: $(cat ${TEMP_DIRECTORY}/server.csr | base64 | tr -d '\n')
      usages:
      - digital signature
      - key encipherment
      - server auth
    EOF

    kubectl certificate approve ${SERVICE_NAME}.${NAMESPACE}.svc

    for x in $(seq 10); do
        SERVER_CERT=$(kubectl get csr ${SERVICE_NAME}.${NAMESPACE}.svc -o jsonpath='{.status.certificate}')
        if [[ ${SERVER_CERT} != '' ]]; then
            break
        fi
        sleep 1
    done
    if [[ ${SERVER_CERT} == '' ]]; then
        echo "ERROR: After approving csr ${SERVICE_NAME}.${NAMESPACE}.svc, the signed certificate did not appear on the resource. Giving up after 10 attempts." >&2
        exit 1
    fi
    echo ${SERVER_CERT} | openssl base64 -d -A -out ${TEMP_DIRECTORY}/server-cert.pem

    kubectl create secret generic ingress-nginx.svc \
        --from-file=key.pem=${TEMP_DIRECTORY}/server-key.pem \
        --from-file=cert.pem=${TEMP_DIRECTORY}/server-cert.pem \
        -n ${NAMESPACE}
    ```

#### Using helm

To generate the certificate using helm, you can use the following snippet

!!! example
    ```
    {{- $cn := printf "%s.%s.svc" ( include "nginx-ingress.validatingWebhook.fullname" . ) .Release.Namespace }}
    {{- $ca := genCA (printf "%s-ca" ( include "nginx-ingress.validatingWebhook.fullname" . )) .Values.validatingWebhook.certificateValidity -}}
    {{- $cert := genSignedCert $cn nil nil .Values.validatingWebhook.certificateValidity $ca -}}
    ```

### Ingress controller flags

To enable the feature in the ingress controller, you _need_ to provide 3 flags to the command line.

|flag|description|example usage|
|-|-|-|
|`--validating-webhook`|The address to start an admission controller on|`:8080`|
|`--validating-webhook-certificate`|The certificate the webhook is using for its TLS handling|`/usr/local/certificates/validating-webhook.pem`|
|`--validating-webhook-key`|The key the webhook is using for its TLS handling|`/usr/local/certificates/validating-webhook-key.pem`|

### kube API server flags

Validating webhook feature requires specific setup on the kube API server side. Depending on your kubernetes version, the flag can, or not, be enabled by default.
To check that your kube API server runs with the required flags, please refer to the [kubernetes][1] documentation.

### Additional kubernetes objects

Once both the ingress controller and the kube API server are configured to serve the webhook, add the you can configure the webhook with the following objects:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-validation-webhook
  namespace: ingress-nginx
spec:
  ports:
  - name: admission
    port: 443
    protocol: TCP
    targetPort: 8080
  selector:
    app: nginx-ingress
    component: controller
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: check-ingress
webhooks:
- name: validate.nginx.ingress.kubernetes.io
  rules:
  - apiGroups:
    - extensions
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - ingresses
  failurePolicy: Fail
  clientConfig:
    service:
      namespace: ingress-nginx
      name: ingress-validation-webhook
      path: /extensions/v1beta1/ingress
    caBundle: <pem encoded ca cert that signs the server cert used by the webhook>
```

[1]: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#validatingadmissionwebhook