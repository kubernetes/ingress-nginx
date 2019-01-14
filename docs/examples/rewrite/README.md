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
|nginx.ingress.kubernetes.io/ssl-redirect|Indicates if the location section is accessible SSL only (defaults to True when Ingress contains a Certificate)|bool|
|nginx.ingress.kubernetes.io/force-ssl-redirect|Forces the redirection to HTTPS even if the Ingress is not TLS Enabled|bool|
|nginx.ingress.kubernetes.io/app-root|Defines the Application Root that the Controller must redirect if it's in '/' context|string|
|nginx.ingress.kubernetes.io/use-regex|Indicates if the paths defined on an Ingress use regular expressions|bool|

## Examples

### Rewrite Target

!!! attention
    Starting in Version 0.22.0, ingress definitions using the annotation `nginx.ingress.kubernetes.io/rewrite-target` are not backwards compatible with previous versions. In Version 0.22.0 and beyond, any substrings within the request URI that need to be passed to the rewritten path must explicitly be defined in a [capture group](https://www.regular-expressions.info/refcapture.html).
    
!!! note
    [Captured groups](https://www.regular-expressions.info/refcapture.html) are saved in numbered placeholders, chronologically, in the form `$1`, `$2` ... `$n`. These placeholders can be used as parameters in the `rewrite-target` annotation. 

Create an Ingress rule with a rewrite annotation:

```console
$ echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$1
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
        path: /something/?(.*)
" | kubectl create -f -
```

In this ingress definition, any characters captured by `(.*)` will be assigned to the placeholder `$1`, which is then used as a parameter in the `rewrite-target` annotation. 

For example, the ingress definition above will result in the following rewrites:
- `rewrite.bar.com/something` rewrites to `rewrite.bar.com/`
- `rewrite.bar.com/something/` rewrites to `rewrite.bar.com/`
- `rewrite.bar.com/something/new` rewrites to `rewrite.bar.com/new`

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
