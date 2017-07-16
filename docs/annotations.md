# Ingress Annotations

This file defines a list of annotations which are supported by various Ingress controllers (both those based on the common ingress code, and alternative implementations).  The intention is to ensure the maximum amount of compatibility between different implementations.

All annotations are assumed to be prefixed with `ingress.kubernetes.io/` except where otherwise specified. There is no attempt to record implementation-specific annotations using other prefixes.  (Traefik in particular defines several of its own annotations which are not described here, and does not seem to support any of the standard annotations.)

Key:

* `nginx`: the `kubernetes/ingress` nginx controller
* `gce`: the `kubernetes/ingress` GCE controller
* `traefik`: Traefik's built-in Ingress controller 
* `haproxy`: Joao Morais' [HAProxy Ingress controller](https://github.com/jcmoraisjr/haproxy-ingress)
* `trafficserver`: Torchbox's [Apache Traffic Server controller plugin](https://github.com/torchbox/k8s-ts-ingress)

## TLS-related

| Name | Meaning
| --- | ---
| `ssl-passthrough` | Pass TLS connections directly to backend; do not offload.  Default `false`.  (nginx, haproxy)
| `ssl-redirect` | Redirect non-TLS requests to TLS when TLS is enabled.  Default `true`.  (nginx, haproxy, trafficserver)
| `force-ssl-redirect` | Redirect non-TLS requests to TLS even when TLS is not configured.  Default `false`.  (nginx, trafficserver).
| `secure-backends` | Use TLS to communicate with origin (pods).  Default `false`. (nginx, haproxy, trafficserver)
| `kubernetes.io/ingress.allow-http` | Whether to accept non-TLS HTTP connections.  (gce)
| `hsts-max-age` | Set an HSTS header with this lifetime. (trafficserver)
| `hsts-include-subdomains` | Add includeSubdomains to the HSTS header. (trafficserver)

## Authentication related

| Name | Meaning
| --- | ---
| `auth-type` | Authentication type: `basic`, `digest`, ... (nginx, haproxy, trafficserver)
| `auth-secret` | Secret name for authentication. (nginx, haproxy, trafficserver)
| `auth-realm` | Authentication realm. (nginx, haproxy, trafficserver)
| `auth-tls-secret` | Name of secret for TLS client certification validation. (nginx, haproxy)
| `auth-tls-verify-depth` | Maximum chain length of TLS client certificate. (nginx)
| `auth-satisfy` | Behaviour when more than one of `auth-type`, `auth-tls-secret` or `whitelist-source-range` are configured: `all` (default) or `any`. (trafficserver) | `trafficserver`
| `whitelist-source-range` | Comma-separate list of IP addresses to restrict access to. (nginx, haproxy, trafficserver)

## URL related

| Name | Meaning
| --- | ---
| `app-root` | Redirect requests without a path (i.e., for `/`) to this location. (nginx, haproxy, trafficserver)
| `rewrite-target` | Replace matched Ingress `path` with this value. (nginx, trafficserver)
| `add-base-url` | Add `<base>` tag to HTML. (nginx)
| `preserve-host` | Whether to pass the client request host (`true`) or the origin hostname (`false`) in the HTTP Host field.  (trafficserver)

## Miscellaneous

| Name | Meaning
| --- | ---
| `configuration-snippet` | Arbitrary text to put in the generated configuration file. (nginx) 
| `enable-cors` | Enable CORS headers in response. (nginx) 
| `limit-connections` | Limit concurrent connections per IP address[1]. (nginx) 
| `limit-rps` | Limit requests per second per IP address[1]. (nginx) 
| `affinity` | Specify a method to stick clients to origins across requests.  Found in `nginx`, where the only supported value is `cookie`. (nginx) 
| `session-cookie-name` | When `affinity` is set to `cookie`, the name of the cookie to use. (nginx) 
| `session-cookie-hash` | When `affinity` is set to `cookie`, the hash algorithm used: `md5`, `sha`, `index`. (nginx) 
| `proxy-body-size` | Maximum request body size. (nginx, haproxy)
| `follow-redirects` | Follow HTTP redirects in the response and deliver the redirect target to the client.  (trafficserver)

[1] The documentation for the `nginx` controller says that only one of `limit-connections` or `limit-rps` may be specified; it's not clear why this is.

## Caching

| Name | Meaning
| --- | ---
| `cache-enable` | Cache responses according to Expires or Cache-Control headers (trafficserver)
| `cache-generation` | An arbitrary numeric value included in the cache key; changing this effectively clears the cache for this ingress.  (trafficserver)
| `cache-ignore-query-params` | Space-separate list of globs matching URL parameters to ignore when doing cache lookups.  (trafficserver)
| `cache-whitelist-query-params` | Ignore any URL parameters not in this whitespace-separate list of globs.  (trafficserver)
| `cache-sort-query-params` | Lexically sort the query parameters by name before cache lookup. (trafficserver)
| `cache-ignore-cookies` | Requests containing a `Cookie:` header will not use the cache unless all the cookie names match this whitespace-separate list of globs.  (trafficserver)
