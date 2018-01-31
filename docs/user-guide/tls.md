# TLS

- [Default SSL Certificate](#default-ssl-certificate)
- [SSL Passthrough](#ssl-passthrough)
- [HTTPS enforcement](#server-side-https-enforcement)
- [HSTS](#http-strict-transport-security)
- [Server-side HTTPS enforcement through redirect](#server-side-https-enforcement-through-redirect) 
- [Kube-Lego](#automated-certificate-management-with-kube-lego)
- [Default TLS Version and Ciphers](#default-tls-version-and-ciphers)
- [Legacy TLS](#legacy-tls)

## Default SSL Certificate

NGINX provides the option to configure a server as a catch-all with [server_name](http://nginx.org/en/docs/http/server_names.html) for requests that do not match any of the configured server names. This configuration works without issues for HTTP traffic.
In case of HTTPS, NGINX requires a certificate.
For this reason the Ingress controller provides the flag `--default-ssl-certificate`. The secret behind this flag contains the default certificate to be used in the mentioned scenario. If this flag is not provided NGINX will use a self signed certificate.

Running without the flag `--default-ssl-certificate`:

```console
$ curl -v https://10.2.78.7:443 -k
* Rebuilt URL to: https://10.2.78.7:443/
*   Trying 10.2.78.4...
* Connected to 10.2.78.7 (10.2.78.7) port 443 (#0)
* ALPN, offering http/1.1
* Cipher selection: ALL:!EXPORT:!EXPORT40:!EXPORT56:!aNULL:!LOW:!RC4:@STRENGTH
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.2 (OUT), TLS header, Certificate Status (22):
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use http/1.1
* Server certificate:
*    subject: CN=foo.bar.com
*    start date: Apr 13 00:50:56 2016 GMT
*    expire date: Apr 13 00:50:56 2017 GMT
*    issuer: CN=foo.bar.com
*    SSL certificate verify result: self signed certificate (18), continuing anyway.
> GET / HTTP/1.1
> Host: 10.2.78.7
> User-Agent: curl/7.47.1
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.11.1
< Date: Thu, 21 Jul 2016 15:38:46 GMT
< Content-Type: text/html
< Transfer-Encoding: chunked
< Connection: keep-alive
< Strict-Transport-Security: max-age=15724800; includeSubDomains; preload
<
<span>The page you're looking for could not be found.</span>

* Connection #0 to host 10.2.78.7 left intact
```

Specifying `--default-ssl-certificate=default/foo-tls`:

```console
core@localhost ~ $ curl -v https://10.2.78.7:443 -k
* Rebuilt URL to: https://10.2.78.7:443/
*   Trying 10.2.78.7...
* Connected to 10.2.78.7 (10.2.78.7) port 443 (#0)
* ALPN, offering http/1.1
* Cipher selection: ALL:!EXPORT:!EXPORT40:!EXPORT56:!aNULL:!LOW:!RC4:@STRENGTH
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.2 (OUT), TLS header, Certificate Status (22):
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Client hello (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-AES128-GCM-SHA256
* ALPN, server accepted to use http/1.1
* Server certificate:
*    subject: CN=foo.bar.com
*    start date: Apr 13 00:50:56 2016 GMT
*    expire date: Apr 13 00:50:56 2017 GMT
*    issuer: CN=foo.bar.com
*    SSL certificate verify result: self signed certificate (18), continuing anyway.
> GET / HTTP/1.1
> Host: 10.2.78.7
> User-Agent: curl/7.47.1
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.11.1
< Date: Mon, 18 Jul 2016 21:02:59 GMT
< Content-Type: text/html
< Transfer-Encoding: chunked
< Connection: keep-alive
< Strict-Transport-Security: max-age=15724800; includeSubDomains; preload
<
<span>The page you're looking for could not be found.</span>

* Connection #0 to host 10.2.78.7 left intact
```

## SSL Passthrough

The flag `--enable-ssl-passthrough` enables SSL passthrough feature.
By default this feature is disabled

## HTTP Strict Transport Security

HTTP Strict Transport Security (HSTS) is an opt-in security enhancement specified through the use of a special response header. Once a supported browser receives this header that browser will prevent any communications from being sent over HTTP to the specified domain and will instead send all communications over HTTPS.

By default the controller redirects (301) to HTTPS if there is a TLS Ingress rule.

To disable this behavior use `hsts: "false"` in the configuration ConfigMap.

## Server-side HTTPS enforcement through redirect

By default the controller redirects (301) to `HTTPS` if TLS is enabled for that ingress. If you want to disable that behavior globally, you can use `ssl-redirect: "false"` in the NGINX config map.

To configure this feature for specific ingress resources, you can use the `nginx.ingress.kubernetes.io/ssl-redirect: "false"` annotation in the particular resource.

When using SSL offloading outside of cluster (e.g. AWS ELB) it may be useful to enforce a redirect to `HTTPS` even when there is not TLS cert available. This can be achieved by using the `nginx.ingress.kubernetes.io/force-ssl-redirect: "true"` annotation in the particular resource.

## Automated Certificate Management with Kube-Lego

[Kube-Lego] automatically requests missing or expired certificates from [Let's Encrypt] by monitoring ingress resources and their referenced secrets. To enable this for an ingress resource you have to add an annotation:

```console
kubectl annotate ing ingress-demo kubernetes.io/tls-acme="true"
```

To setup Kube-Lego you can take a look at this [full example]. The first
version to fully support Kube-Lego is nginx Ingress controller 0.8.

[full example]:https://github.com/jetstack/kube-lego/tree/master/examples
[Kube-Lego]:https://github.com/jetstack/kube-lego
[Let's Encrypt]:https://letsencrypt.org

## Default TLS Version and Ciphers

To provide the most secure baseline configuration possible, nginx-ingress defaults to using TLS 1.2 and a [secure set of TLS ciphers](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/configmap.md#ssl-ciphers)

## Legacy TLS 
The default configuration, though secure, does not support some older browsers and operating systems. For instance, 20% of Android phones in use today are not compatible with nginx-ingress's default configuration. To change this default behavior, use a [ConfigMap](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/configmap.md#ssl-ciphers).

A sample ConfigMap to allow these older clients connect could look something like the following:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: nginx-config
data:
  ssl-ciphers: "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA"
  ssl-protocols: "TLSv1 TLSv1.1 TLSv1.2"
```
