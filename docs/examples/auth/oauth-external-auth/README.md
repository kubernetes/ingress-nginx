# External OAUTH Authentication

### Overview

The `auth-url` and `auth-signin` annotations allow you to use an external
authentication provider to protect your Ingress resources.

!!! Important
    This annotation requires `ingress-nginx-controller v0.9.0` or greater.

### Key Detail

This functionality is enabled by deploying multiple Ingress objects for a single host.
One Ingress object has no special annotations and handles authentication.

Other Ingress objects can then be annotated in such a way that require the user to
authenticate against the first Ingress's endpoint, and can redirect `401`s to the
same endpoint.

Sample:

```yaml
...
metadata:
  name: application
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://$host/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://$host/oauth2/start?rd=$escaped_request_uri"
...
```

### Example: OAuth2 Proxy + Kubernetes-Dashboard

This example will show you how to deploy [`oauth2_proxy`](https://github.com/pusher/oauth2_proxy)
into a Kubernetes cluster and use it to protect the Kubernetes Dashboard using GitHub as the OAuth2 provider.

#### Prepare

1. Install the kubernetes dashboard

    ```console
    kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/kubernetes-dashboard/v1.10.1.yaml
    ```

2. Create a [custom GitHub OAuth application](https://github.com/settings/applications/new)

    ![Register OAuth2 Application](images/register-oauth-app.png)

    - Homepage URL is the FQDN in the Ingress rule, like `https://foo.bar.com`
    - Authorization callback URL is the same as the base FQDN plus `/oauth2/callback`, like `https://foo.bar.com/oauth2/callback`

    ![Register OAuth2 Application](images/register-oauth-app-2.png)

3. Configure values in the file [`oauth2-proxy.yaml`](https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/docs/examples/auth/oauth-external-auth/oauth2-proxy.yaml) with the values:

    - OAUTH2_PROXY_CLIENT_ID with the github `<Client ID>`
    - OAUTH2_PROXY_CLIENT_SECRET with the github `<Client Secret>`
    - OAUTH2_PROXY_COOKIE_SECRET with value of `python -c 'import os,base64; print(base64.b64encode(os.urandom(16)).decode("ascii"))'`
    - (optional, but recommended) OAUTH2_PROXY_GITHUB_USERS with GitHub usernames to allow to login
    - `__INGRESS_HOST__` with a valid FQDN (e.g. `foo.bar.com`)
    - `__INGRESS_SECRET__` with a Secret with a valid SSL certificate

4. Deploy the oauth2 proxy and the ingress rules by running:

    ```console
    $ kubectl create -f oauth2-proxy.yaml
    ```

#### Test

Test the integration by accessing the configured URL, e.g. `https://foo.bar.com`

![Register OAuth2 Application](images/github-auth.png)

![GitHub authentication](images/oauth-login.png)

![Kubernetes dashboard](images/dashboard.png)


### Example: Vouch Proxy + Kubernetes-Dashboard

This example will show you how to deploy [`Vouch Proxy`](https://github.com/vouch/vouch-proxy)
into a Kubernetes cluster and use it to protect the Kubernetes Dashboard using GitHub as the OAuth2 provider.

#### Prepare

1. Install the kubernetes dashboard

    ```console
    kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/kubernetes-dashboard/v1.10.1.yaml
    ```

2. Create a [custom GitHub OAuth application](https://github.com/settings/applications/new)

    ![Register OAuth2 Application](images/register-oauth-app.png)

    - Homepage URL is the FQDN in the Ingress rule, like `https://foo.bar.com`
    - Authorization callback URL is the same as the base FQDN plus `/oauth2/auth`, like `https://foo.bar.com/oauth2/auth`

    ![Register OAuth2 Application](images/register-oauth-app-2.png)

3. Configure Vouch Proxy values in the file [`vouch-proxy.yaml`](https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/docs/examples/auth/oauth-external-auth/vouch-proxy.yaml) with the values:

    - VOUCH_COOKIE_DOMAIN with value of `<Ingress Host>`
    - OAUTH_CLIENT_ID with the github `<Client ID>`
    - OAUTH_CLIENT_SECRET with the github `<Client Secret>`
    - (optional, but recommended) VOUCH_WHITELIST with GitHub usernames to allow to login
    - `__INGRESS_HOST__` with a valid FQDN (e.g. `foo.bar.com`)
    - `__INGRESS_SECRET__` with a Secret with a valid SSL certificate

4. Deploy Vouch Proxy and the ingress rules by running:

    ```console
    $ kubectl create -f vouch-proxy.yaml
    ```

#### Test

Test the integration by accessing the configured URL, e.g. `https://foo.bar.com`

![Register OAuth2 Application](images/github-auth.png)

![GitHub authentication](images/oauth-login.png)

![Kubernetes dashboard](images/dashboard.png)
