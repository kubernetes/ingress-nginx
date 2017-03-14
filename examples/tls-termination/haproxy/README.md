# HAProxy Ingress TLS termination

## Prerequisites

This document has the following prerequisites:

* Deploy [HAProxy Ingress controller](/examples/deployment/haproxy), you should end up with controller, a sample web app and default TLS secret
* Create [*another* secret](/examples/PREREQUISITES.md#tls-certificates) named `foobar-ssl` and subject `'/CN=foo.bar'`

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Using default TLS certificate

Update ingress resource in order to add TLS termination to host `foo.bar`:

```console
$ kubectl replace -f ingress-tls-default.yaml
```

The difference from the starting ingress resource:

```console
 metadata:
   name: app
 spec:
+  tls:
+  - hosts:
+    - foo.bar
   rules:
   - host: foo.bar
     http:
```

Trying default backend:

```console
$ curl -iL 172.17.4.99:30876            
HTTP/1.1 404 Not Found
Date: Tue, 07 Feb 2017 00:06:07 GMT
Content-Length: 21
Content-Type: text/plain; charset=utf-8

default backend - 404
```

Now telling the controller we are `foo.bar`:

```console
$ curl -iL 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 302 Found
Cache-Control: no-cache
Content-length: 0
Location: https://foo.bar/
Connection: close
^C
```

Note the `Location` header - this would redirect us to the correct server.

Checking the default certificate - change below `31692` to the TLS port:

```console
$ openssl s_client -connect 172.17.4.99:31692
...
subject=/CN=localhost
issuer=/CN=localhost
---
```

... and `foo.bar` certificate:

```console
$ openssl s_client -connect 172.17.4.99:31692 -servername foo.bar
...
subject=/CN=localhost
issuer=/CN=localhost
---
```

## Using a new TLS certificate

Now let's reference the new certificate to our domain. Note that secret
`foobar-ssl` should be created as described in the [prerequisites](#prerequisites)

```console
$ kubectl replace -f ingress-tls-foobar.yaml 
```

Here is the difference:

```console
   tls:
   - hosts:
     - foo.bar
+    secretName: foobar-ssl
   rules:
   - host: foo.bar
     http:
```

Now `foo.bar` certificate should be used to terminate TLS:

```console
$ openssl s_client -connect 172.17.4.99:31692
...
subject=/CN=localhost
issuer=/CN=localhost
---

$ openssl s_client -connect 172.17.4.99:31692 -servername foo.bar
...
subject=/CN=foo.bar
issuer=/CN=foo.bar
---
```
