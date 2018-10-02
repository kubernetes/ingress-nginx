# Rewrite

This example demonstrates how to use the Rewrite annotations

## Prerequisites

You will need to make sure your Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](../../user-guide/multiple-ingress.md),
and that you have an ingress controller [running](../../deploy) in your cluster.

## Deployment

Rewriting can be controlled using the following annotations:

|Name|Description|Values|
| --- | --- | --- |
|nginx.ingress.kubernetes.io/rewrite-target|Target URI where the traffic must be redirected|string|
|nginx.ingress.kubernetes.io/add-base-url|indicates if is required to add a base tag in the head of the responses from the upstream servers|bool|
|nginx.ingress.kubernetes.io/base-url-scheme|Override for the scheme passed to the base tag|string|
|nginx.ingress.kubernetes.io/ssl-redirect|Indicates if the location section is accessible SSL only (defaults to True when Ingress contains a Certificate)|bool|
|nginx.ingress.kubernetes.io/force-ssl-redirect|Forces the redirection to HTTPS even if the Ingress is not TLS Enabled|bool|
|nginx.ingress.kubernetes.io/app-root|Defines the Application Root that the Controller must redirect if it's in '/' context|string|
|nginx.ingress.kubernetes.io/use-regex|Indicates if the paths defined on an Ingress use regular expressions|bool|

## Validation

### Rewrite Target

Create an Ingress rule with a rewrite annotation:

```console
$ echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
  name: rewrite
  namespace: default
spec:
  rules:
  - host: rewrite.bar.com
    http:
      paths:
      - backend:
          serviceName: http-svc
          servicePort: 80
        path: /something
" | kubectl create -f -
```

Check the rewrite is working

```
$ curl -v http://172.17.4.99/something -H 'Host: rewrite.bar.com'
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET /something HTTP/1.1
> Host: rewrite.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.0
< Date: Tue, 31 May 2016 16:07:31 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
<
CLIENT VALUES:
client_address=10.2.56.9
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://rewrite.bar.com:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=close
host=rewrite.bar.com
user-agent=curl/7.43.0
x-forwarded-for=10.2.56.1
x-forwarded-host=rewrite.bar.com
x-forwarded-port=80
x-forwarded-proto=http
x-real-ip=10.2.56.1
BODY:
* Connection #0 to host 172.17.4.99 left intact
-no body in request-
```

### App Root

Create an Ingress rule with a app-root annotation:
```
$ echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/app-root: /app1
  name: approot
  namespace: default
spec:
  rules:
  - host: approot.bar.com
    http:
      paths:
      - backend:
          serviceName: http-svc
          servicePort: 80
        path: /
" | kubectl create -f -
```

Check the rewrite is working

```
$ curl -I -k http://approot.bar.com/
HTTP/1.1 302 Moved Temporarily
Server: nginx/1.11.10
Date: Mon, 13 Mar 2017 14:57:15 GMT
Content-Type: text/html
Content-Length: 162
Location: http://stickyingress.example.com/app1
Connection: keep-alive
```
