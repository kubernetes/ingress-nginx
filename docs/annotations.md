# Ingress Annotations

This file defines a list of annotations which are supported by various Ingress controllers (both those based on the common ingress code, and alternative implementations).
The intention is to ensure the maximum amount of compatibility between different implementations.

All annotations are assumed to be prefixed with `nginx.ingress.kubernetes.io/` except where otherwise specified.
There is no attempt to record implementation-specific annotations using other prefixes.
(Traefik in particular defines several of its own annotations which are not described here, and does not seem to support any of the standard annotations.)

Key:

* `nginx`: the `kubernetes/ingress` nginx controller
* `gce`: the `kubernetes/ingress` GCE controller
* `traefik`: Traefik's built-in Ingress controller
* `voyager`: [Voyager by AppsCode](https://github.com/appscode/voyager) - Secure HAProxy based Ingress Controller for Kubernetes
* `haproxy`: Joao Morais' [HAProxy Ingress controller](https://github.com/jcmoraisjr/haproxy-ingress)
* `trafficserver`: Torchbox's [Apache Traffic Server controller plugin](https://github.com/torchbox/k8s-ts-ingress)

## TLS-related

| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `ssl-passthrough` | Pass TLS connections directly to backend; do not offload. | `false` | nginx, voyager, haproxy
| `ssl-redirect` | Redirect non-TLS requests to TLS when TLS is enabled. | `true` | nginx, voyager, haproxy, trafficserver
| `force-ssl-redirect` | Redirect non-TLS requests to TLS even when TLS is not configured. | `false` | nginx, voyager, trafficserver
| `secure-backends` | Use TLS to communicate with origin (pods). | `false` | nginx, voyager, haproxy, trafficserver
| `kubernetes.io/ingress.allow-http` | Whether to accept non-TLS HTTP connections. | `true` | gce
| `pre-shared-cert` | Name of the TLS certificate in GCP to use when provisioning the HTTPS load balancer. | empty string | gce
| `hsts-max-age` | Set an HSTS header with this lifetime. | | voyager, trafficserver
| `hsts-include-subdomains` | Add includeSubdomains to the HSTS header. | | voyager, trafficserver

## Authentication related

| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `auth-type` | Authentication type: `basic`, `digest`, ... | | nginx, voyager, haproxy, trafficserver
| `auth-secret` | Secret name for authentication. | | nginx, voyager, haproxy, trafficserver
| `auth-realm` | Authentication realm. | | nginx, voyager, haproxy, trafficserver
| `auth-tls-secret` | Name of secret for TLS client certification validation. | | nginx, voyager, haproxy
| `auth-tls-verify-depth` | Maximum chain length of TLS client certificate. | | nginx
| `auth-tls-error-page` | The page that user should be redirected in case of Auth error | | nginx, voyager
| `auth-satisfy` | Behaviour when more than one of `auth-type`, `auth-tls-secret` or `whitelist-source-range` are configured: `all` or `any`. | `all` | trafficserver | `trafficserver`
| `whitelist-source-range` | Comma-separate list of IP addresses to enable access to. | | nginx, voyager, haproxy, trafficserver

## URL related

| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `app-root` | Redirect requests without a path (i.e., for `/`) to this location. | | nginx, haproxy, trafficserver
| `rewrite-target` | Replace matched Ingress `path` with this value. | | nginx, trafficserver
| `add-base-url` | Add `<base>` tag to HTML. | | nginx
| `base-url-scheme` | Specify the scheme of the `<base>` tags. | | nginx
| `preserve-host` | Whether to pass the client request host (`true`) or the origin hostname (`false`) in the HTTP Host field. | | trafficserver
| `x-forwarded-prefix` | Add the non-standard `X-Forwarded-Prefix` header to the request with the value of the matched location. | | nginx

## CORS Related
| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `enable-cors` | Enable CORS headers in response. | false | nginx, voyager
| `cors-allow-origin` | Specifies the Origin allowed in CORS (Access-Control-Allow-Origin) | * | nginx
| `cors-allow-headers` | Specifies the Headers allowed in CORS (Access-Control-Allow-Headers) | DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization | nginx
| `cors-allow-methods` | Specifies the Methods allowed in CORS (Access-Control-Allow-Methods) | GET, PUT, POST, DELETE, PATCH, OPTIONS | nginx
| `cors-allow-credentials` | Specifies the Access-Control-Allow-Credentials | true | nginx
| `cors-max-age` | Specifies the Access-Control-Max-Age | 1728000 | nginx

## Miscellaneous

| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `configuration-snippet` | Arbitrary text to put in the generated configuration file. | | nginx
| `limit-connections` | Limit concurrent connections per IP address[1]. | | nginx, voyager
| `limit-rps` | Limit requests per second per IP address[1]. | | nginx, voyager
| `limit-rpm` | Limit requests per minute per IP address. | | nginx, voyager
| `affinity` | Specify a method to stick clients to origins across requests.  Found in `nginx`, where the only supported value is `cookie`. | | nginx, voyager
| `session-cookie-name` | When `affinity` is set to `cookie`, the name of the cookie to use. | | nginx, voyager
| `session-cookie-hash` | When `affinity` is set to `cookie`, the hash algorithm used: `md5`, `sha`, `index`. | | nginx
| `proxy-body-size` | Maximum request body size. | | nginx, voyager, haproxy
| `proxy-pass-params` | Parameters for proxy-pass directives. | |
| `follow-redirects` | Follow HTTP redirects in the response and deliver the redirect target to the client. | | trafficserver
| `kubernetes.io/ingress.global-static-ip-name` | Name of the static global IP address in GCP to use when provisioning the HTTPS load balancer. | empty string | gce

[1] The documentation for the `nginx` controller says that only one of `limit-connections` or `limit-rps` may be specified; it's not clear why this is.

## Caching

| Name | Meaning | Default | Controller
| --- | --- | --- | --- |
| `cache-enable` | Cache responses according to Expires or Cache-Control headers. | | trafficserver
| `cache-generation` | An arbitrary numeric value included in the cache key; changing this effectively clears the cache for this ingress. | | trafficserver
| `cache-ignore-query-params` | Space-separate list of globs matching URL parameters to ignore when doing cache lookups. | | trafficserver
| `cache-whitelist-query-params` | Ignore any URL parameters not in this whitespace-separate list of globs. | | trafficserver
| `cache-sort-query-params` | Lexically sort the query parameters by name before cache lookup. | | trafficserver
| `cache-ignore-cookies` | Requests containing a `Cookie:` header will not use the cache unless all the cookie names match this whitespace-separate list of globs. | | trafficserver
