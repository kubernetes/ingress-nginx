# Default SSL Certificate

NGINX provides the option to configure a server as a catch-all with [server name _](http://nginx.org/en/docs/http/server_names.html) for requests that do not match any of the configured server names. This configuration works without issues for HTTP traffic.
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
