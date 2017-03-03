## External Authentication

### Overview

The `auth-url` and `auth-signin` annotations allow you to use an external
authentication provider to protect your Ingress resources.

(Note, this annotation requires `nginx-ingress-controller v0.9.0` or greater.)

### Key Detail

This functionality is enabled by deploying multiple Ingress objects for a single host.
One Ingress object has no special annotations and handles authentication.

Other Ingress objects can then be annotated in such a way that require the user to
authenticate against the first Ingress's endpoint, and can redirect `401`s to the
same endpoint.

Sample:

```
...
metadata:
  name: application
  annotations:
    "ingress.kubernetes.io/auth-url": "https://$host/oauth2/auth"
    "ingress.kubernetes.io/signin-url": "https://$host/oauth2/sign_in"
...
```

### Example: OAuth2 Proxy + Kubernetes-Dashboard

This example will show you how to deploy [`oauth2_proxy`](https://github.com/bitly/oauth2_proxy)
into a Kubernetes cluster and use it to protect the Kubernetes Dashboard.

#### Prepare:

1. `export DOMAIN="somedomain.io"`
2. Install `nginx-ingress`. If you haven't already, consider using `helm`: `$ helm install stable/nginx-ingress`
3. Make sure you have a TLS cert added as a Secret named `ingress-tls` that corresponds to your `$DOMAIN`.

### Deploy: `oauth2_proxy`

This is the Deployment object that runs `oauth2_proxy`.

1. Configure `oauth2proxy.deployment.yaml` to use the desired provider.

2. Create a secret with the appropriate name and values, matching what is specified in
   `oauth2proxy.deployment.yaml`.

   For example, as-is with Azure AD, you can use this command:
   ```
   kubectl create-secret generic oauth2proxy \
     --from-literal=COOKIE_SECRET="$(uuidgen)" \
     --from-literal=TENANT_ID="${TENANT_ID}" \
     --from-literal=CLIENT_ID="${CLIENT_ID}" \
     --from-literal=CLIENT_SECRET="${CLIENT_SECRET}"
   ```

3. Deploy it all: `./deploy.sh`

See the script for further details.
