# HAProxy Ingress rewrite

This example demonstrates how to use rewrite options on HAProxy Ingress controller.

## Prerequisites

This document has the following prerequisites:

* Deploy [HAProxy Ingress](/examples/deployment/haproxy) controller, you should
end up with controller, a sample web app and an ingress resource named `app` to
the `foo.bar` domain
* Configure only the default [TLS termination](/examples/tls-termination/haproxy) -
there is no need to create another secret

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Annotations

The following annotations are implemented:

* `ingress.kubernetes.io/ssl-redirect`: Indicates whether a redirect should be
done from HTTP to HTTPS. Possible values are `"true"` to redirect to HTTPS,
or `"false"` meaning requests may be performed as plain HTTP.
* `ingress.kubernetes.io/app-root`: Defines the URL to be redirected when requests
are done to the root context `/`.

### SSL Redirect

Annotate the `app` ingress resource:

```console
$ kubectl annotate ingress/app --overwrite ingress.kubernetes.io/ssl-redirect=false
ingress "app" annotated
```

Try a HTTP request:

```console
$ curl -iL 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Sat, 15 Apr 2017 19:27:30 GMT
Content-Type: text/plain
Transfer-Encoding: chunked

CLIENT VALUES:
client_address=10.2.33.14
command=GET
real path=/
query=nil
...
```

Now turn ssl-redirect true:

```console
$ kubectl annotate ingress/app --overwrite ingress.kubernetes.io/ssl-redirect=true
ingress "app" annotated

$ curl -iL 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 302 Found
Cache-Control: no-cache
Content-length: 0
Location: https://foo.bar/

...
```

The default value of ssl-redirect annotation is true and can be changed globally
using a [ConfigMap](https://github.com/jcmoraisjr/haproxy-ingress#configmap).

### App root context redirect

Annotate the `app` ingress resource with `app-root`, and also `ssl-redirect` to `false` for simplicity:

```console
$ kubectl annotate ingress/app --overwrite ingress.kubernetes.io/app-root=/web
ingress "app" annotated

$ kubectl annotate ingress/app --overwrite ingress.kubernetes.io/ssl-redirect=false
ingress "app" annotated
```

Try a HTTP request:

```console
$ curl -iL 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 302 Found
Cache-Control: no-cache
Content-length: 0
Location: /web

HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Sat, 15 Apr 2017 19:34:49 GMT
Content-Type: text/plain
Transfer-Encoding: chunked

CLIENT VALUES:
client_address=10.2.33.14
command=GET
real path=/web
query=nil
...
```
