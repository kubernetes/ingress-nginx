# Ingress Path Matching

## Regular Expression Support

The ingress controller supports **case insensitive** regular expressions in the `spec.rules.http.paths.path` field.


See the [description](./nginx-configuration/annotations.md#use-regex) of the `use-regex` annotation for more details. 

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress
  annotations:
    nginx.ingress.kubernetes.io/use-regex: true
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /foo/.*
        backend:
          serviceName: test
          servicePort: 80
```

The preceding ingress definition would translate to the following location block within the NGINX configuration for the `test.com` server:

```
location ~* ^/foo/.* {
  ...
}
```

## Path Priority

In NGINX, regular expressions follow a **first match** policy. In order to enable more acurate path matching, ingress-nginx first orders the paths by descending length before writing them to the NGINX template as location blocks. 

__Please read the [warning](#warning) before using regular expressions in your ingress definitions.__

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



The ingress controller would define the following location blocks, in order of descending length, within the NGINX template for the `test.com` server: 

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
- `test.com/foo/bar/1` matches `~* ^/foo/bar/.+\/?(?<baseuri>.*)`
- `test.com/foo/bar/` matches `~* ^/foo/bar/`
- `test.com/foo/bar` matches `~* ^/foo/bar`

__IMPORTANT NOTES__: 
- paths created under the `rewrite-ingress` are sorted before `\/?(?<baseuri>.*)` is appended. For example if the path defined within `test-ingress-2` was `/foo/.+` then the location block for `^/foo/.+\/?(?<baseuri>.*)` would be the LAST block listed.
- If the `use-regex` OR `rewrite-target` annotation is used on any Ingress for a given host, then the case insensitive regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) will be enforced on ALL paths for a given host regardless of what Ingress they are defined on.  


## Warning
The following example describes a case that may inflict unwanted path matching behaviour. 

This case is expected and a result of NGINX's a first match policy for paths that use the regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location). For more information about how a path is chosen, please read the following article: ["Understanding Nginx Server and Location Block Selection Algorithms"](https://www.digitalocean.com/community/tutorials/understanding-nginx-server-and-location-block-selection-algorithms). 

### Example

Let the following ingress be defined:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress-1
  annotations:
    nginx.ingress.kubernetes.io/use-regex: true
spec:
  host: test.com
  rules:
  - http:
      paths:
      - path: /foo/bar/bar
        backend:
          serviceName: test
          servicePort: 80
      - path: /foo/bar/[A-Z0-9]{3}
        backend:
          serviceName: test
          servicePort: 80
```

The ingress controller would define the following location blocks (in this order) within the NGINX template for the `test.com` server: 

```
location ~* ^/foo/bar/[A-Z0-9]{3} {
  ...
}

location ~* ^/foo/bar/bar {
  ...
}
```

A request to `test.com/foo/bar/bar` would match the `^/foo/[A-Z0-9]{3}` location block instead of the longest EXACT matching path.

