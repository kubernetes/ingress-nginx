# HAProxy Ingress Client Certificate Authentication

This example demonstrates how to configure client certificate
authentication on HAProxy Ingress controller.

## Prerequisites

This document has the following prerequisites:

* Deploy [HAProxy Ingress](/examples/deployment/haproxy) controller, you should
end up with controller, a sample web app and an ingress resource named `app` to
the `foo.bar` domain
* Configure [TLS termination](/examples/tls-termination/haproxy)
* Create a [CA, certificate and private key](/examples/PREREQUISITES.md#ca-authentication),
following these steps you should have a secret named `caingress`, a certificate file
`client.crt` and it's private key `client.key`
* Use these same steps and create another CA and generate another certificate and private
key `fake.crt` and `fake.key` just for testing

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

Secret, certificates and keys can be created using these shortcuts:

CA and it's secret:

```console
$ openssl req -x509 -newkey rsa:2048 -nodes -subj '/CN=example-ca' -keyout ca.key -out ca.crt
$ kubectl create secret generic caingress --from-file=ca.crt
```

Valid certificate and private key:

```console
$ openssl req -new -newkey rsa:2048 -nodes -subj '/CN=client' -keyout client.key | \
  openssl x509 -req -CA ca.crt -CAkey ca.key -set_serial 1 -out client.crt
```

Another CA, certificate and private key that should be refused by ingress:

```console
$ openssl req -x509 -newkey rsa:2048 -nodes -subj '/CN=example-ca' -keyout ca-fake.key -out ca-fake.crt
$ openssl req -new -newkey rsa:2048 -nodes -subj '/CN=client' -keyout fake.key | \
  openssl x509 -req -CA ca-fake.crt -CAkey ca-fake.key -set_serial 1 -out fake.crt
```

## Using Client Certificate Authentication

HAProxy Ingress read one or a bundle of certificate authorities from a secret.
Only client certificates signed by one of these certificate authorities should be
allowed to make requests.

Annotate the ingress resource to use our valid certificate authority. The ingress resource and the
secret `caingress` were created on the prerequisites.

```console
$ kubectl annotate ingress/app ingress.kubernetes.io/auth-tls-secret=default/caingress
```

Make some SSL requests against domain `foo.bar`. Change `31692:172.17.4.99` below to the IP and
port of HAProxy Ingress controller.

Note: `curl`'s `--cert` and `-k` options on macOS (since 10.9 Mavericks) doesn't work as
expected, see troubleshooting below if using macOS.

Connect without a certificate:

```console
$ curl -ik https://foo.bar:31692 --resolve 'foo.bar:31692:172.17.4.99'
curl: (35) error:14094410:SSL routines:ssl3_read_bytes:sslv3 alert handshake failure
```

Connect using the correct certificate and private key:

```console
$ curl -ik https://foo.bar:31692 --resolve 'foo.bar:31692:172.17.4.99' --cert client.crt --key client.key
HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Fri, 26 Mar 2017 13:41:26 GMT
Content-Type: text/plain
Transfer-Encoding: chunked
Strict-Transport-Security: max-age=15768000

CLIENT VALUES:
...
```

Now connect using a private key and certificate signed by another CA:

```console
$ curl -ik https://foo.bar:31692 --resolve 'foo.bar:31692:172.17.4.99' --cert fake.crt --key fake.key
curl: (35) error:1409441B:SSL routines:ssl3_read_bytes:tlsv1 alert decrypt error
```

## Troubleshooting

`curl` on macOS since 10.9 Mavericks has some issues regarding certificate on command line
parameters:

* [sni](https://en.wikipedia.org/wiki/Server_Name_Indication) TLS extension isn't used
if `-k` (unsecure connection) is provided. The sni extension is used by HAProxy to identify
the host of the request. Without sni, the default backend will be used. The TLS
certificate should be added to Keychain instead and `-k` should be avoided.

* `--cert` option is broken.

These issues and it's workarounds are described on
[this message](https://curl.haxx.se/mail/archive-2013-10/0036.html) from curl mailing list.
In short, in order to test client auth use a Linux VM or the options below on macOS.

### Using wget

Add `foo.bar` to `/etc/hosts` and change `31692` below to the
port of HAProxy Ingress controller:

```console
$ wget https://foo.bar:31692 -S -nv -O- --no-check-certificate                                        
OpenSSL: error:14094410:SSL routines:SSL3_READ_BYTES:sslv3 alert handshake failure
Unable to establish SSL connection.
```

Now with certificate and private key:

```console
$ wget https://foo.bar:31692 -S -nv -O- --no-check-certificate --certificate client.crt --private-key client.key
WARNING: cannot verify foo.bar's certificate, issued by ‘/CN=foo.bar’:
  Self-signed certificate encountered.
  HTTP/1.1 200 OK
  Server: nginx/1.9.11
  Date: Sun, 26 Mar 2017 13:57:53 GMT
  Content-Type: text/plain
  Transfer-Encoding: chunked
  Strict-Transport-Security: max-age=15768000
CLIENT VALUES:
```

### Using openssl

Change `31692` below to the port of HAProxy Ingress controller:

```console
$ openssl s_client -connect 172.17.4.99:31692 -servername foo.bar                             
CONNECTED(00000003)
depth=0 /CN=foo.bar
verify error:num=18:self signed certificate
verify return:1
depth=0 /CN=foo.bar
verify return:1
91929:error:14094410:SSL routines:SSL3_READ_BYTES:sslv3 alert handshake failure:/BuildRoot/Library/Caches/com.apple.xbs/Sources/OpenSSL098/OpenSSL098-59.60.1/src/ssl/s3_pkt.c:1145:SSL alert number 40
91929:error:140790E5:SSL routines:SSL23_WRITE:ssl handshake failure:/BuildRoot/Library/Caches/com.apple.xbs/Sources/OpenSSL098/OpenSSL098-59.60.1/src/ssl/s23_lib.c:185:
```

Now with certificate and private key - copy these two lines to the clipboard
(HAProxy will timeout after 5 seconds waiting a http request):

```
GET / HTTP/1.0
Host: foo.bar
```

Type the command below, paste the http request (two lines above) and send a blank line
pressing enter twice:

```console
$ openssl s_client -connect 172.17.4.99:31692 -servername foo.bar -cert client.crt -key client.key
...
---
GET / HTTP/1.0
Host: foo.bar

HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Sun, 26 Mar 2017 14:06:30 GMT
Content-Type: text/plain
Content-Length: 268
Connection: close
Strict-Transport-Security: max-age=15768000

CLIENT VALUES:
...
```
