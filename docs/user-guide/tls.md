# TLS/HTTPS

## TLS Secrets

Anytime we reference a TLS secret, we mean a PEM-encoded X.509, RSA (2048) secret.

!!! warning
    Ensure that the certificate order is leaf->intermediate->root, otherwise the controller will not be able to import the certificate, and you'll see this error in the logs ```W1012 09:15:45.920000       6 backend_ssl.go:46] Error obtaining X.509 certificate: unexpected error creating SSL Cert: certificate and private key does not have a matching public key: tls: private key does not match public key```

You can generate a self-signed certificate and private key with:

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${HOST}/O=${HOST}" -addext "subjectAltName = DNS:${HOST}"
```

Then create the secret in the cluster via:

```bash
kubectl create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}
```

The resulting secret will be of type `kubernetes.io/tls`.

## Host names

Ensure that the relevant [ingress rules specify a matching host name](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls).

## Default SSL Certificate

NGINX provides the option to configure a server as a catch-all with
[server_name](https://nginx.org/en/docs/http/server_names.html)
for requests that do not match any of the configured server names.
This configuration works out-of-the-box for HTTP traffic.
For HTTPS, a certificate is naturally required.

For this reason the Ingress controller provides the flag `--default-ssl-certificate`.
The secret referred to by this flag contains the default certificate to be used when
accessing the catch-all server.
If this flag is not provided NGINX will use a self-signed certificate.

For instance, if you have a TLS secret `foo-tls` in the `default` namespace,
add `--default-ssl-certificate=default/foo-tls` in the `nginx-controller` deployment.

The default certificate will also be used for ingress `tls:` sections that do not
have a `secretName` option.

To force redirects for Ingresses that do not specify a TLS-block at all, take a look at `force-ssl-redirect` in [ConfigMap][ConfigMap].

## SSL Passthrough

The [`--enable-ssl-passthrough`](cli-arguments.md) flag enables the SSL Passthrough feature, which is disabled by
default. This is required to enable passthrough backends in Ingress objects.

!!! warning
    This feature is implemented by intercepting **all traffic** on the configured HTTPS port (default: 443) and handing
    it over to a local TCP proxy. This bypasses NGINX completely and introduces a non-negligible performance penalty.

SSL Passthrough leverages [SNI][SNI] and reads the virtual domain from the TLS negotiation, which requires compatible
clients. After a connection has been accepted by the TLS listener, it is handled by the controller itself and piped back
and forth between the backend and the client.

If there is no hostname matching the requested host name, the request is handed over to NGINX on the configured
passthrough proxy port (default: 442), which proxies the request to the default backend.

!!! note
    Unlike HTTP backends, traffic to Passthrough backends is sent to the *clusterIP* of the backing Service instead of
    individual Endpoints.

## HTTP Strict Transport Security

HTTP Strict Transport Security (HSTS) is an opt-in security enhancement specified
through the use of a special response header. Once a supported browser receives
this header that browser will prevent any communications from being sent over
HTTP to the specified domain and will instead send all communications over HTTPS.

HSTS is enabled by default.

To disable this behavior use `hsts: "false"` in the configuration [ConfigMap][ConfigMap].

## Server-side HTTPS enforcement through redirect

By default the controller redirects HTTP clients to the HTTPS port
443 using a 308 Permanent Redirect response if TLS is enabled for that Ingress.

This can be disabled globally using `ssl-redirect: "false"` in the NGINX [config map][ConfigMap],
or per-Ingress with the `nginx.ingress.kubernetes.io/ssl-redirect: "false"`
annotation in the particular resource.

!!! tip
    When using SSL offloading outside of cluster (e.g. AWS ELB) it may be useful to enforce a
    redirect to HTTPS even when there is no TLS certificate available.
    This can be achieved by using the `nginx.ingress.kubernetes.io/force-ssl-redirect: "true"`
    annotation in the particular resource.

## Automated Certificate Management with cert-manager

[cert-manager] automatically requests missing or expired certificates from a range of 
[supported issuers][cert-manager-issuer-config] (including [Let's Encrypt]) by monitoring 
ingress resources.

To set up cert-manager you should take a look at this [full example][full-cert-manager-example].

To enable it for an ingress resource you have to deploy cert-manager, configure a certificate 
issuer update the manifest:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-demo
  annotations:
    cert-manager.io/issuer: "letsencrypt-staging" # Replace this with a production issuer once you've tested it
    [..]
spec:
  tls:
    - hosts:
        - ingress-demo.example.com
      secretName: ingress-demo-tls
    [...]
```

## Default TLS Version and Ciphers

To provide the most secure baseline configuration possible,

ingress-nginx defaults to using TLS 1.2 and 1.3 only, with a [secure set of TLS ciphers][ssl-ciphers].

### Legacy TLS

The default configuration, though secure, does not support some older browsers and operating systems.

For instance, TLS 1.1+ is only enabled by default from Android 5.0 on. At the time of writing,
May 2018, [approximately 15% of Android devices](https://developer.android.com/about/dashboards/#Platform)
are not compatible with ingress-nginx's default configuration.

To change this default behavior, use a [ConfigMap][ConfigMap].

A sample ConfigMap fragment to allow these older clients to connect could look something like the following
(generated using the Mozilla SSL Configuration Generator)[mozilla-ssl-config-old]:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: nginx-config
data:
  ssl-ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA256:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:DES-CBC3-SHA"
  ssl-protocols: "TLSv1 TLSv1.1 TLSv1.2 TLSv1.3"
```



[Let's Encrypt]:https://letsencrypt.org
[ConfigMap]: ./nginx-configuration/configmap.md
[ssl-ciphers]: ./nginx-configuration/configmap.md#ssl-ciphers
[SNI]: https://en.wikipedia.org/wiki/Server_Name_Indication
[mozilla-ssl-config-old]: https://ssl-config.mozilla.org/#server=nginx&config=old
[cert-manager]: https://github.com/jetstack/cert-manager/
[full-cert-manager-example]:https://cert-manager.io/docs/tutorials/acme/nginx-ingress/
[cert-manager-issuer-config]:https://cert-manager.io/docs/configuration/
