# HAProxy Multi TLS certificate termination

This example uses 2 different certificates to terminate SSL for 2 hostnames.

## Prerequisites

This document has the following prerequisites:

* Deploy [HAProxy Ingress controller](/examples/deployment/haproxy), you should end up with controller, a sample web app and default TLS secret
* Create [*two* secrets](/examples/PREREQUISITES.md#tls-certificates) named `foobar-ssl` with subject `'/CN=foo.bar'` and `barfoo-ssl` with subject `'/CN=bar.foo'`

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Using a new TLS certificate

Update ingress resource in order to add TLS termination to two hosts:

```console
$ kubectl replace -f ingress-multi-tls.yaml
```

Trying without host:

```console
$ curl -iL 10.129.51.55:30221           
HTTP/1.1 404 Not Found
Date: Tue, 28 Mar 2017 07:32:34 GMT
Content-Length: 21
Content-Type: text/plain; charset=utf-8

default backend - 404
```

Telling the controller we are `foo.bar` or `bar.foo`:

```console
$ curl -iL 10.129.51.55:36462 -H 'Host: foo.bar'
HTTP/1.1 302 Found
Cache-Control: no-cache
Content-length: 0
Location: https://foo.bar/
Connection: close
$ curl -iL 10.129.51.55:36462 -H 'Host: bar.foo'
HTTP/1.1 302 Found
Cache-Control: no-cache
Content-length: 0
Location: https://bar.foo/
Connection: close
^C
```

Note the `Location` header - this would redirect us to the correct server.

Checking the certificate - change below `31578` to the TLS port:

```console
$ openssl s_client -connect 10.129.51.55:31578 -servername foo.bar
...
subject=/CN=foo.bar
issuer=/CN=foo.bar
---
```

... and `bar.foo` certificate:

```console
$ openssl s_client -connect 10.129.51.55:31578 -servername bar.foo
...
subject=/CN=bar.foo
issuer=/CN=bar.foo
---
```
