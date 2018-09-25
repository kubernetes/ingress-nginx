# Ingress Path Matching

## Regular Expression Support

The ingress controller supports **case insensitive** regular expressions in the `spec.rules.http.paths.path` feild. 

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /testpath/.*
        backend:
          serviceName: test
          servicePort: 80
```

The preceding ingress definition would translate to the following location block within the NGINX configuration for the `test.com` server:

```
location ~* ^/testpath/.* {
  ...
}
```

## Path Priority

In NGINX, regular expressions follow a **first match** policy. In order to enable more acurate path matching, ingress-nginx first orders the paths by descending length before writing them to the NGINX template as location blocks. Therefore, longest path matching will be used in most cases. 

### Example

Let the following two ingress definitions be created:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress-1
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /foo/bar
        backend:
          serviceName: test
          servicePort: 80
      - path: /foo/bar/
        backend:
          serviceName: test
          servicePort: 80
```

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress-2
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /foo/bar/.+
        backend:
          serviceName: test
          servicePort: 80
```



The ingress controller would define the following location blocks (in this order) within the NGINX template for the `test.com` server: 

```
location ~* ^/foo/bar/.+\/?(?<baseuri>.*) {
  ...
}

location ~* ^/foo/bar/ {
  ...
}

location ~* ^/foo/bar {
  ...
}
```
The following request URI's would match the corresponding location blocks:
- `test.com/blog/topics/announcements` matches `~* ^/blog/topics/.+\/?(?<baseuri>.*)`
- `test.com/blog/topics/` matches `^~ /blog/topics/`
- `test.com/blog/topics` matches `^~ /blog/topics`

_NOTE: paths created under the `rewrite-ingress` are sorted before `\/?(?<baseuri>.*)` is appended. For example if the path defined within `test-ingress-2` was `/foo/.+` then the location block for `^/foo/.+\/?(?<baseuri>.*)` would be the LAST block listed._


## Path Priority Edge case
The exception to this the longest match policy will occur when long regular expressions are used.

### Example

Let the following ingress be defined:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress-1
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /foo/bar/
        backend:
          serviceName: test
          servicePort: 80
      - path: /foo/bar/[A-Z0-9]{32}
        backend:
          serviceName: test
          servicePort: 80
```

The ingress controller would define the following location blocks (in this order) within the NGINX template for the `test.com` server: 

```
location ~* ^/foo/bar/[A-Z0-9]{32} {
  ...
}

location ~* ^/foo/bar/ {
  ...
}
```

A request to `test.com/foo/bar` would match the `^/foo/[A-Z0-9]{32}` location block. 

