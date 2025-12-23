# Ingress Path Matching

## Regular Expression Support

!!! important
    Regular expressions is not supported in the `spec.rules.host` field. The wildcard character '\*' must appear by itself as the first DNS label and matches only a single label. You cannot have a wildcard label by itself (e.g. Host == "\*").

!!! note
    Please see the [FAQ](../faq.md#validation-of-path) for Validation Of __`path`__

The ingress controller supports **case insensitive** regular expressions in the `spec.rules.http.paths.path` field.
This can be enabled by setting the `nginx.ingress.kubernetes.io/use-regex` annotation to `true` (the default is false).

See the [description](./nginx-configuration/annotations.md#use-regex) of the `use-regex` annotation for more details.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  annotations:
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  rules:
  - host: test.com
    http:
      paths:
      - path: /foo/.*
        pathType: ImplementationSpecific
        backend:
          service:
            name: test
            port:
              number: 80
```

The preceding ingress definition would translate to the following location block within the NGINX configuration for the `test.com` server:

```txt
location ~* "^/foo/.*" {
  ...
}
```

## Path Priority

In NGINX, regular expressions follow a **first match** policy. In order to enable more accurate path matching, ingress-nginx first orders the paths by descending length before writing them to the NGINX template as location blocks.

**Please read the [warning](#warning) before using regular expressions in your ingress definitions.**

### Example

Let the following two ingress definitions be created:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress-1
spec:
  ingressClassName: nginx
  rules:
  - host: test.com
    http:
      paths:
      - path: /foo/bar
        pathType: Prefix
        backend:
          service:
            name: service1
            port:
              number: 80
      - path: /foo/bar/
        pathType: Prefix
        backend:
          service:
            name: service2
            port:
              number: 80
```

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress-2
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$1
spec:
  ingressClassName: nginx
  rules:
  - host: test.com
    http:
      paths:
      - path: /foo/bar/(.+)
        pathType: ImplementationSpecific
        backend:
          service:
            name: service3
            port: 
              number: 80
```

The ingress controller would define the following location blocks, in order of descending length, within the NGINX template for the `test.com` server:

```txt
location ~* ^/foo/bar/.+ {
  ...
}

location ~* "^/foo/bar/" {
  ...
}

location ~* "^/foo/bar" {
  ...
}
```

The following request URI's would match the corresponding location blocks:

- `test.com/foo/bar/1` matches `~* ^/foo/bar/.+` and will go to service 3.
- `test.com/foo/bar/` matches `~* ^/foo/bar/` and will go to service 2.
- `test.com/foo/bar` matches `~* ^/foo/bar` and will go to service 1.

**IMPORTANT NOTES**:

- If the `use-regex` OR `rewrite-target` annotation is used on any Ingress for a given host, then the case insensitive regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) will be enforced on ALL paths for a given host regardless of what Ingress they are defined on.

## Warning

The following example describes a case that may inflict unwanted path matching behavior.

This case is expected and a result of NGINX's a first match policy for paths that use the regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location). For more information about how a path is chosen, please read the following article: ["Understanding Nginx Server and Location Block Selection Algorithms"](https://www.digitalocean.com/community/tutorials/understanding-nginx-server-and-location-block-selection-algorithms).

### Example

Let the following ingress be defined:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress-3
  annotations:
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  rules:
  - host: test.com
    http:
      paths:
      - path: /foo/bar/bar
        pathType: Prefix
        backend:
          service:
            name: test
            port: 
              number: 80
      - path: /foo/bar/[A-Z0-9]{3}
        pathType: ImplementationSpecific
        backend:
          service:
            name: test
            port: 
              number: 80
```

The ingress controller would define the following location blocks (in this order) within the NGINX template for the `test.com` server:

```txt
location ~* "^/foo/bar/[A-Z0-9]{3}" {
  ...
}

location ~* "^/foo/bar/bar" {
  ...
}
```

A request to `test.com/foo/bar/bar` would match the `^/foo/bar/[A-Z0-9]{3}` location block instead of the longest EXACT matching path.
