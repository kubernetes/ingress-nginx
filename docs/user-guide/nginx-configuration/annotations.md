# Annotations

You can add these Kubernetes annotations to specific Ingress objects to customize their behavior.

!!! tip
    Annotation keys and values can only be strings.
    Other types, such as boolean or numeric values must be quoted,
    i.e. `"true"`, `"false"`, `"100"`.

!!! note
    The annotation prefix can be changed using the
    [`--annotations-prefix` command line argument](../cli-arguments.md),
    but the default is `nginx.ingress.kubernetes.io`, as described in the
    table below.

|Name                       | type |
|---------------------------|------|
|[nginx.ingress.kubernetes.io/app-root](#rewrite)|string|
|[nginx.ingress.kubernetes.io/affinity](#session-affinity)|cookie|
|[nginx.ingress.kubernetes.io/affinity-mode](#session-affinity)|"balanced" or "persistent"|
|[nginx.ingress.kubernetes.io/affinity-canary-behavior](#session-affinity)|"sticky" or "legacy"|
|[nginx.ingress.kubernetes.io/auth-realm](#authentication)|string|
|[nginx.ingress.kubernetes.io/auth-secret](#authentication)|string|
|[nginx.ingress.kubernetes.io/auth-secret-type](#authentication)|string|
|[nginx.ingress.kubernetes.io/auth-type](#authentication)|"basic" or "digest"|
|[nginx.ingress.kubernetes.io/auth-tls-secret](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-verify-depth](#client-certificate-authentication)|number|
|[nginx.ingress.kubernetes.io/auth-tls-verify-client](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-error-page](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream](#client-certificate-authentication)|"true" or "false"|
|[nginx.ingress.kubernetes.io/auth-tls-match-cn](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-url](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-cache-key](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-cache-duration](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-keepalive](#external-authentication)|number|
|[nginx.ingress.kubernetes.io/auth-keepalive-share-vars](#external-authentication)|"true" or "false"|
|[nginx.ingress.kubernetes.io/auth-keepalive-requests](#external-authentication)|number|
|[nginx.ingress.kubernetes.io/auth-keepalive-timeout](#external-authentication)|number|
|[nginx.ingress.kubernetes.io/auth-proxy-set-headers](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-snippet](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/enable-global-auth](#external-authentication)|"true" or "false"|
|[nginx.ingress.kubernetes.io/backend-protocol](#backend-protocol)|string|
|[nginx.ingress.kubernetes.io/canary](#canary)|"true" or "false"|
|[nginx.ingress.kubernetes.io/canary-by-header](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-by-header-value](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-by-header-pattern](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-by-cookie](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-weight](#canary)|number|
|[nginx.ingress.kubernetes.io/canary-weight-total](#canary)|number|
|[nginx.ingress.kubernetes.io/client-body-buffer-size](#client-body-buffer-size)|string|
|[nginx.ingress.kubernetes.io/configuration-snippet](#configuration-snippet)|string|
|[nginx.ingress.kubernetes.io/custom-http-errors](#custom-http-errors)|[]int|
|[nginx.ingress.kubernetes.io/custom-headers](#custom-headers)|string|
|[nginx.ingress.kubernetes.io/default-backend](#default-backend)|string|
|[nginx.ingress.kubernetes.io/enable-cors](#enable-cors)|"true" or "false"|
|[nginx.ingress.kubernetes.io/cors-allow-origin](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-methods](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-headers](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-expose-headers](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-credentials](#enable-cors)|"true" or "false"|
|[nginx.ingress.kubernetes.io/cors-max-age](#enable-cors)|number|
|[nginx.ingress.kubernetes.io/force-ssl-redirect](#server-side-https-enforcement-through-redirect)|"true" or "false"|
|[nginx.ingress.kubernetes.io/from-to-www-redirect](#redirect-fromto-www)|"true" or "false"|
|[nginx.ingress.kubernetes.io/http2-push-preload](#http2-push-preload)|"true" or "false"|
|[nginx.ingress.kubernetes.io/limit-connections](#rate-limiting)|number|
|[nginx.ingress.kubernetes.io/limit-rps](#rate-limiting)|number|
|[nginx.ingress.kubernetes.io/permanent-redirect](#permanent-redirect)|string|
|[nginx.ingress.kubernetes.io/permanent-redirect-code](#permanent-redirect-code)|number|
|[nginx.ingress.kubernetes.io/temporal-redirect](#temporal-redirect)|string|
|[nginx.ingress.kubernetes.io/temporal-redirect-code](#temporal-redirect-code)|number|
|[nginx.ingress.kubernetes.io/preserve-trailing-slash](#server-side-https-enforcement-through-redirect)|"true" or "false"|
|[nginx.ingress.kubernetes.io/proxy-body-size](#custom-max-body-size)|string|
|[nginx.ingress.kubernetes.io/proxy-cookie-domain](#proxy-cookie-domain)|string|
|[nginx.ingress.kubernetes.io/proxy-cookie-path](#proxy-cookie-path)|string|
|[nginx.ingress.kubernetes.io/proxy-connect-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-send-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-read-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-next-upstream](#custom-timeouts)|string|
|[nginx.ingress.kubernetes.io/proxy-next-upstream-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-next-upstream-tries](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-request-buffering](#custom-timeouts)|string|
|[nginx.ingress.kubernetes.io/proxy-redirect-from](#proxy-redirect)|string|
|[nginx.ingress.kubernetes.io/proxy-redirect-to](#proxy-redirect)|string|
|[nginx.ingress.kubernetes.io/proxy-http-version](#proxy-http-version)|"1.0" or "1.1"|
|[nginx.ingress.kubernetes.io/proxy-ssl-secret](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/proxy-ssl-ciphers](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/proxy-ssl-name](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/proxy-ssl-protocols](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/proxy-ssl-verify](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/proxy-ssl-verify-depth](#backend-certificate-authentication)|number|
|[nginx.ingress.kubernetes.io/proxy-ssl-server-name](#backend-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/enable-rewrite-log](#enable-rewrite-log)|"true" or "false"|
|[nginx.ingress.kubernetes.io/rewrite-target](#rewrite)|URI|
|[nginx.ingress.kubernetes.io/satisfy](#satisfy)|string|
|[nginx.ingress.kubernetes.io/server-alias](#server-alias)|string|
|[nginx.ingress.kubernetes.io/server-snippet](#server-snippet)|string|
|[nginx.ingress.kubernetes.io/service-upstream](#service-upstream)|"true" or "false"|
|[nginx.ingress.kubernetes.io/session-cookie-change-on-failure](#cookie-affinity)|"true" or "false"|
|[nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none](#cookie-affinity)|"true" or "false"|
|[nginx.ingress.kubernetes.io/session-cookie-domain](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-expires](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-max-age](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-name](#cookie-affinity)|string|default "INGRESSCOOKIE"|
|[nginx.ingress.kubernetes.io/session-cookie-path](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-samesite](#cookie-affinity)|string|"None", "Lax" or "Strict"|
|[nginx.ingress.kubernetes.io/session-cookie-secure](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/ssl-redirect](#server-side-https-enforcement-through-redirect)|"true" or "false"|
|[nginx.ingress.kubernetes.io/ssl-passthrough](#ssl-passthrough)|"true" or "false"|
|[nginx.ingress.kubernetes.io/stream-snippet](#stream-snippet)|string|
|[nginx.ingress.kubernetes.io/upstream-hash-by](#custom-nginx-upstream-hashing)|string|
|[nginx.ingress.kubernetes.io/x-forwarded-prefix](#x-forwarded-prefix-header)|string|
|[nginx.ingress.kubernetes.io/load-balance](#custom-nginx-load-balancing)|string|
|[nginx.ingress.kubernetes.io/upstream-vhost](#custom-nginx-upstream-vhost)|string|
|[nginx.ingress.kubernetes.io/denylist-source-range](#denylist-source-range)|CIDR|
|[nginx.ingress.kubernetes.io/whitelist-source-range](#whitelist-source-range)|CIDR|
|[nginx.ingress.kubernetes.io/proxy-buffering](#proxy-buffering)|string|
|[nginx.ingress.kubernetes.io/proxy-buffers-number](#proxy-buffers-number)|number|
|[nginx.ingress.kubernetes.io/proxy-buffer-size](#proxy-buffer-size)|string|
|[nginx.ingress.kubernetes.io/proxy-busy-buffers-size](#proxy-busy-buffers-size)|string|
|[nginx.ingress.kubernetes.io/proxy-max-temp-file-size](#proxy-max-temp-file-size)|string|
|[nginx.ingress.kubernetes.io/ssl-ciphers](#ssl-ciphers)|string|
|[nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers](#ssl-ciphers)|"true" or "false"|
|[nginx.ingress.kubernetes.io/connection-proxy-header](#connection-proxy-header)|string|
|[nginx.ingress.kubernetes.io/enable-access-log](#enable-access-log)|"true" or "false"|
|[nginx.ingress.kubernetes.io/enable-opentelemetry](#enable-opentelemetry)|"true" or "false"|
|[nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-span](#opentelemetry-trust-incoming-spans)|"true" or "false"|
|[nginx.ingress.kubernetes.io/use-regex](#use-regex)|bool|
|[nginx.ingress.kubernetes.io/enable-modsecurity](#modsecurity)|bool|
|[nginx.ingress.kubernetes.io/enable-owasp-core-rules](#modsecurity)|bool|
|[nginx.ingress.kubernetes.io/modsecurity-transaction-id](#modsecurity)|string|
|[nginx.ingress.kubernetes.io/modsecurity-snippet](#modsecurity)|string|
|[nginx.ingress.kubernetes.io/mirror-request-body](#mirror)|string|
|[nginx.ingress.kubernetes.io/mirror-target](#mirror)|string|
|[nginx.ingress.kubernetes.io/mirror-host](#mirror)|string|

### Canary

In some cases, you may want to "canary" a new set of changes by sending a small number of requests to a different service than the production service. The canary annotation enables the Ingress spec to act as an alternative service for requests to route to depending on the rules applied. The following annotations to configure canary can be enabled after `nginx.ingress.kubernetes.io/canary: "true"` is set:

* `nginx.ingress.kubernetes.io/canary-by-header`: The header to use for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the request header is set to `always`, it will be routed to the canary. When the header is set to `never`, it will never be routed to the canary. For any other value, the header will be ignored and the request compared against the other canary rules by precedence.

* `nginx.ingress.kubernetes.io/canary-by-header-value`: The header value to match for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the request header is set to this value, it will be routed to the canary. For any other header value, the header will be ignored and the request compared against the other canary rules by precedence. This annotation has to be used together with `nginx.ingress.kubernetes.io/canary-by-header`. The annotation is an extension of the `nginx.ingress.kubernetes.io/canary-by-header` to allow customizing the header value instead of using hardcoded values. It doesn't have any effect if the `nginx.ingress.kubernetes.io/canary-by-header` annotation is not defined.

* `nginx.ingress.kubernetes.io/canary-by-header-pattern`: This works the same way as `canary-by-header-value` except it does PCRE Regex matching. Note that when `canary-by-header-value` is set this annotation will be ignored. When the given Regex causes error during request processing, the request will be considered as not matching.

* `nginx.ingress.kubernetes.io/canary-by-cookie`: The cookie to use for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the cookie value is set to `always`, it will be routed to the canary. When the cookie is set to `never`, it will never be routed to the canary. For any other value, the cookie will be ignored and the request compared against the other canary rules by precedence.

* `nginx.ingress.kubernetes.io/canary-weight`: The integer based (0 - <weight-total>) percent of random requests that should be routed to the service specified in the canary Ingress. A weight of 0 implies that no requests will be sent to the service in the Canary ingress by this canary rule. A weight of `<weight-total>` means implies all requests will be sent to the alternative service specified in the Ingress. `<weight-total>` defaults to 100, and can be increased via `nginx.ingress.kubernetes.io/canary-weight-total`.

* `nginx.ingress.kubernetes.io/canary-weight-total`: The total weight of traffic. If unspecified, it defaults to 100.

Canary rules are evaluated in order of precedence. Precedence is as follows:
`canary-by-header -> canary-by-cookie -> canary-weight`

**Note** that when you mark an ingress as canary, then all the other non-canary annotations will be ignored (inherited from the corresponding main ingress) except `nginx.ingress.kubernetes.io/load-balance`, `nginx.ingress.kubernetes.io/upstream-hash-by`, and [annotations related to session affinity](#session-affinity). If you want to restore the original behavior of canaries when session affinity was ignored, set `nginx.ingress.kubernetes.io/affinity-canary-behavior` annotation with value `legacy` on the canary ingress definition.

**Known Limitations**

Currently a maximum of one canary ingress can be applied per Ingress rule.

### Rewrite

In some scenarios the exposed URL in the backend service differs from the specified path in the Ingress rule. Without a rewrite any request will return 404.
Set the annotation `nginx.ingress.kubernetes.io/rewrite-target` to the path expected by the service.

If the Application Root is exposed in a different path and needs to be redirected, set the annotation `nginx.ingress.kubernetes.io/app-root` to redirect requests for `/`.

!!! example
    Please check the [rewrite](../../examples/rewrite/README.md) example.

### Session Affinity

The annotation `nginx.ingress.kubernetes.io/affinity` enables and sets the affinity type in all Upstreams of an Ingress. This way, a request will always be directed to the same upstream server.
The only affinity type available for NGINX is `cookie`.

The annotation `nginx.ingress.kubernetes.io/affinity-mode` defines the stickiness of a session. Setting this to `balanced` (default) will redistribute some sessions if a deployment gets scaled up, therefore rebalancing the load on the servers. Setting this to `persistent` will not rebalance sessions to new servers, therefore providing maximum stickiness.

The annotation `nginx.ingress.kubernetes.io/affinity-canary-behavior` defines the behavior of canaries when session affinity is enabled. Setting this to `sticky` (default) will ensure that users that were served by canaries, will continue to be served by canaries. Setting this to `legacy` will restore original canary behavior, when session affinity was ignored.

!!! attention
    If more than one Ingress is defined for a host and at least one Ingress uses `nginx.ingress.kubernetes.io/affinity: cookie`, then only paths on the Ingress using `nginx.ingress.kubernetes.io/affinity` will use session cookie affinity. All paths defined on other Ingresses for the host will be load balanced through the random selection of a backend server.

!!! example
    Please check the [affinity](../../examples/affinity/cookie/README.md) example.

#### Cookie affinity

If you use the ``cookie`` affinity type you can also specify the name of the cookie that will be used to route the requests with the annotation `nginx.ingress.kubernetes.io/session-cookie-name`. The default is to create a cookie named 'INGRESSCOOKIE'.

The NGINX annotation `nginx.ingress.kubernetes.io/session-cookie-path` defines the path that will be set on the cookie. This is optional unless the annotation `nginx.ingress.kubernetes.io/use-regex` is set to true; Session cookie paths do not support regex.

Use `nginx.ingress.kubernetes.io/session-cookie-domain` to set the `Domain` attribute of the sticky cookie.

Use `nginx.ingress.kubernetes.io/session-cookie-samesite` to apply a `SameSite` attribute to the sticky cookie. Browser accepted values are `None`, `Lax`, and `Strict`. Some browsers reject cookies with `SameSite=None`, including those created before the `SameSite=None` specification (e.g. Chrome 5X). Other browsers mistakenly treat `SameSite=None` cookies as `SameSite=Strict` (e.g. Safari running on OSX 14). To omit `SameSite=None` from browsers with these incompatibilities, add the annotation `nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none: "true"`.

Use `nginx.ingress.kubernetes.io/session-cookie-expires` to control the cookie expires, its value is a number of seconds until the cookie expires.

Use `nginx.ingress.kubernetes.io/session-cookie-path` to control the cookie path when use-regex is set to true.

Use `nginx.ingress.kubernetes.io/session-cookie-change-on-failure` to control the cookie change after request failure.

### Authentication

It is possible to add authentication by adding additional annotations in the Ingress rule. The source of the authentication is a secret that contains usernames and passwords.

The annotations are:
```
nginx.ingress.kubernetes.io/auth-type: [basic|digest]
```

Indicates the [HTTP Authentication Type: Basic or Digest Access Authentication](https://tools.ietf.org/html/rfc2617).

```
nginx.ingress.kubernetes.io/auth-secret: secretName
```

The name of the Secret that contains the usernames and passwords which are granted access to the `path`s defined in the Ingress rules.
This annotation also accepts the alternative form "namespace/secretName", in which case the Secret lookup is performed in the referenced namespace instead of the Ingress namespace.

```
nginx.ingress.kubernetes.io/auth-secret-type: [auth-file|auth-map]
```

The `auth-secret` can have two forms:

- `auth-file` - default, an htpasswd file in the key `auth` within the secret
- `auth-map` - the keys of the secret are the usernames, and the values are the hashed passwords

```
nginx.ingress.kubernetes.io/auth-realm: "realm string"
```

!!! example
    Please check the [auth](../../examples/auth/basic/README.md) example.

### Custom NGINX upstream hashing

NGINX supports load balancing by client-server mapping based on [consistent hashing](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#hash) for a given key. The key can contain text, variables or any combination thereof. This feature allows for request stickiness other than client IP or cookies. The [ketama](https://www.last.fm/user/RJ/journal/2007/04/10/rz_libketama_-_a_consistent_hashing_algo_for_memcache_clients) consistent hashing method will be used which ensures only a few keys would be remapped to different servers on upstream group changes.

There is a special mode of upstream hashing called subset. In this mode, upstream servers are grouped into subsets, and stickiness works by mapping keys to a subset instead of individual upstream servers. Specific server is chosen uniformly at random from the selected sticky subset. It provides a balance between stickiness and load distribution.

To enable consistent hashing for a backend:

`nginx.ingress.kubernetes.io/upstream-hash-by`: the nginx variable, text value or any combination thereof to use for consistent hashing. For example: `nginx.ingress.kubernetes.io/upstream-hash-by: "$request_uri"` or `nginx.ingress.kubernetes.io/upstream-hash-by: "$request_uri$host"` or `nginx.ingress.kubernetes.io/upstream-hash-by: "${request_uri}-text-value"` to consistently hash upstream requests by the current request URI.

"subset" hashing can be enabled setting `nginx.ingress.kubernetes.io/upstream-hash-by-subset`: "true". This maps requests to subset of nodes instead of a single one. `nginx.ingress.kubernetes.io/upstream-hash-by-subset-size` determines the size of each subset (default 3).

Please check the [chashsubset](../../examples/chashsubset/deployment.yaml) example.

### Custom NGINX load balancing

This is similar to [`load-balance` in ConfigMap](./configmap.md#load-balance), but configures load balancing algorithm per ingress.
>Note that `nginx.ingress.kubernetes.io/upstream-hash-by` takes preference over this. If this and `nginx.ingress.kubernetes.io/upstream-hash-by` are not set then we fallback to using globally configured load balancing algorithm.

### Custom NGINX upstream vhost

This configuration setting allows you to control the value for host in the following statement: `proxy_set_header Host $host`, which forms part of the location block.  This is useful if you need to call the upstream server by something other than `$host`.

### Client Certificate Authentication

It is possible to enable Client Certificate Authentication using additional annotations in Ingress Rule.

Client Certificate Authentication is applied per host and it is not possible to specify rules that differ for individual paths.

To enable, add the annotation `nginx.ingress.kubernetes.io/auth-tls-secret: namespace/secretName`. This secret must have a file named `ca.crt` containing the full Certificate Authority chain `ca.crt` that is enabled to authenticate against this Ingress.

You can further customize client certificate authentication and behavior with these annotations:

* `nginx.ingress.kubernetes.io/auth-tls-verify-depth`: The validation depth between the provided client certificate and the Certification Authority chain. (default: 1)
* `nginx.ingress.kubernetes.io/auth-tls-verify-client`: Enables verification of client certificates. Possible values are:
    * `on`: Request a client certificate that must be signed by a certificate that is included in the secret key `ca.crt` of the secret specified by `nginx.ingress.kubernetes.io/auth-tls-secret: namespace/secretName`. Failed certificate verification will result in a status code 400 (Bad Request) (default)
    * `off`: Don't request client certificates and don't do client certificate verification.
    * `optional`: Do optional client certificate validation against the CAs from `auth-tls-secret`. The request fails with status code 400 (Bad Request) when a certificate is provided that is not signed by the CA. When no or an otherwise invalid certificate is provided, the request does not fail, but instead the verification result is sent to the upstream service.
    * `optional_no_ca`: Do optional client certificate validation, but do not fail the request when the client certificate is not signed by the CAs from `auth-tls-secret`. Certificate verification result is sent to the upstream service.
* `nginx.ingress.kubernetes.io/auth-tls-error-page`: The URL/Page that user should be redirected in case of a Certificate Authentication Error
* `nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream`: Indicates if the received certificates should be passed or not to the upstream server in the header `ssl-client-cert`. Possible values are "true" or "false" (default).
* `nginx.ingress.kubernetes.io/auth-tls-match-cn`: Adds a sanity check for the CN of the client certificate that is sent over using a string / regex starting with "CN=", example: `"CN=myvalidclient"`. If the certificate CN sent during mTLS does not match your string / regex it will fail with status code 403. Another way of using this is by adding multiple options in your regex, example: `"CN=(option1|option2|myvalidclient)"`. In this case, as long as one of the options in the brackets matches the certificate CN then you will receive a 200 status code. 

The following headers are sent to the upstream service according to the `auth-tls-*` annotations:

* `ssl-client-issuer-dn`: The issuer information of the client certificate. Example: "CN=My CA"
* `ssl-client-subject-dn`: The subject information of the client certificate. Example: "CN=My Client"
* `ssl-client-verify`: The result of the client verification. Possible values: "SUCCESS", "FAILED: <description, why the verification failed>"
* `ssl-client-cert`: The full client certificate in PEM format. Will only be sent when `nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream` is set to "true". Example: `-----BEGIN%20CERTIFICATE-----%0A...---END%20CERTIFICATE-----%0A`

!!! example
    Please check the [client-certs](../../examples/auth/client-certs/README.md) example.

!!! attention
    TLS with Client Authentication is **not** possible in Cloudflare and might result in unexpected behavior.

    Cloudflare only allows Authenticated Origin Pulls and is required to use their own certificate: [https://blog.cloudflare.com/protecting-the-origin-with-tls-authenticated-origin-pulls/](https://blog.cloudflare.com/protecting-the-origin-with-tls-authenticated-origin-pulls/)

    Only Authenticated Origin Pulls are allowed and can be configured by following their tutorial: [https://support.cloudflare.com/hc/en-us/articles/204494148-Setting-up-NGINX-to-use-TLS-Authenticated-Origin-Pulls](https://web.archive.org/web/20200907143649/https://support.cloudflare.com/hc/en-us/articles/204899617-Setting-up-NGINX-to-use-TLS-Authenticated-Origin-Pulls#section5)

### Backend Certificate Authentication

It is possible to authenticate to a proxied HTTPS backend with certificate using additional annotations in Ingress Rule.

* `nginx.ingress.kubernetes.io/proxy-ssl-secret: secretName`:
  Specifies a Secret with the certificate `tls.crt`, key `tls.key` in PEM format used for authentication to a proxied HTTPS server. It should also contain trusted CA certificates `ca.crt` in PEM format used to verify the certificate of the proxied HTTPS server.
  This annotation expects the Secret name in the form "namespace/secretName".
* `nginx.ingress.kubernetes.io/proxy-ssl-verify`:
  Enables or disables verification of the proxied HTTPS server certificate. (default: off)
* `nginx.ingress.kubernetes.io/proxy-ssl-verify-depth`:
  Sets the verification depth in the proxied HTTPS server certificates chain. (default: 1)
* `nginx.ingress.kubernetes.io/proxy-ssl-ciphers`:
  Specifies the enabled [ciphers](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_ssl_ciphers) for requests to a proxied HTTPS server. The ciphers are specified in the format understood by the OpenSSL library.
* `nginx.ingress.kubernetes.io/proxy-ssl-name`:
  Allows to set [proxy_ssl_name](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_ssl_name). This allows overriding the server name used to verify the certificate of the proxied HTTPS server. This value is also passed through SNI when a connection is established to the proxied HTTPS server.
* `nginx.ingress.kubernetes.io/proxy-ssl-protocols`:
  Enables the specified [protocols](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_ssl_protocols) for requests to a proxied HTTPS server.
* `nginx.ingress.kubernetes.io/proxy-ssl-server-name`:
  Enables passing of the server name through TLS Server Name Indication extension (SNI, RFC 6066) when establishing a connection with the proxied HTTPS server.

### Configuration snippet

Using this annotation you can add additional configuration to the NGINX location. For example:

```yaml
nginx.ingress.kubernetes.io/configuration-snippet: |
  more_set_headers "Request-Id: $req_id";
```

Be aware this can be dangerous in multi-tenant clusters, as it can lead to people with otherwise limited permissions being able to retrieve all secrets on the cluster. The recommended mitigation for this threat is to disable this feature, so it may not work for you. See CVE-2021-25742 and the [related issue on github](https://github.com/kubernetes/ingress-nginx/issues/7837) for more information.

### Custom HTTP Errors

Like the [`custom-http-errors`](./configmap.md#custom-http-errors) value in the ConfigMap, this annotation will set NGINX `proxy-intercept-errors`, but only for the NGINX location associated with this ingress. If a [default backend annotation](#default-backend) is specified on the ingress, the errors will be routed to that annotation's default backend service (instead of the global default backend).
Different ingresses can specify different sets of error codes. Even if multiple ingress objects share the same hostname, this annotation can be used to intercept different error codes for each ingress (for example, different error codes to be intercepted for different paths on the same hostname, if each path is on a different ingress).
If `custom-http-errors` is also specified globally, the error values specified in this annotation will override the global value for the given ingress' hostname and path.

Example usage:
```
nginx.ingress.kubernetes.io/custom-http-errors: "404,415"
```

### Custom Headers
This annotation is of the form `nginx.ingress.kubernetes.io/custom-headers: <namespace>/<custom headers configmap>` to specify a namespace and configmap name that contains custom headers. This annotation uses `more_set_headers` nginx directive.

Example annotation for following example configmap:

```yaml
nginx.ingress.kubernetes.io/custom-headers: default/custom-headers-configmap
```

Example configmap:
```yaml
apiVersion: v1
data:
  Content-Type: application/json
kind: ConfigMap
metadata:
  name: custom-headers-configmap
  namespace: default
```

!!! attention
  First define the allowed response headers in [global-allowed-response-headers](https://github.com/kubernetes/ingress-nginx/blob/main/docs/user-guide/nginx-configuration/configmap.md#global-allowed-response-headers).

### Default Backend

This annotation is of the form `nginx.ingress.kubernetes.io/default-backend: <svc name>` to specify a custom default backend.  This `<svc name>` is a reference to a service inside of the same namespace in which you are applying this annotation. This annotation overrides the global default backend. In case the service has [multiple ports](https://kubernetes.io/docs/concepts/services-networking/service/#multi-port-services), the first one is the one which will receive the backend traffic. 

This service will be used to handle the response when the configured service in the Ingress rule does not have any active endpoints. It will also be used to handle the error responses if both this annotation and the [custom-http-errors annotation](#custom-http-errors) are set.

### Enable CORS

To enable Cross-Origin Resource Sharing (CORS) in an Ingress rule, add the annotation
`nginx.ingress.kubernetes.io/enable-cors: "true"`. This will add a section in the server
location enabling this functionality.

CORS can be controlled with the following annotations:

* `nginx.ingress.kubernetes.io/cors-allow-methods`: Controls which methods are accepted.

    This is a multi-valued field, separated by ',' and accepts only letters (upper and lower case).

    - Default: `GET, PUT, POST, DELETE, PATCH, OPTIONS`
    - Example: `nginx.ingress.kubernetes.io/cors-allow-methods: "PUT, GET, POST, OPTIONS"`

* `nginx.ingress.kubernetes.io/cors-allow-headers`: Controls which headers are accepted.

    This is a multi-valued field, separated by ',' and accepts letters, numbers, _ and -.

    - Default: `DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization`
    - Example: `nginx.ingress.kubernetes.io/cors-allow-headers: "X-Forwarded-For, X-app123-XPTO"`

* `nginx.ingress.kubernetes.io/cors-expose-headers`: Controls which headers are exposed to response.

    This is a multi-valued field, separated by ',' and accepts letters, numbers, _, - and *.

    - Default: *empty*
    - Example: `nginx.ingress.kubernetes.io/cors-expose-headers: "*, X-CustomResponseHeader"`

* `nginx.ingress.kubernetes.io/cors-allow-origin`: Controls what's the accepted Origin for CORS.

    This is a multi-valued field, separated by ','. It must follow this format: `protocol://origin-site.com` or `protocol://origin-site.com:port`

    - Default: `*`
    - Example: `nginx.ingress.kubernetes.io/cors-allow-origin: "https://origin-site.com:4443, http://origin-site.com, myprotocol://example.org:1199"`

    It also supports single level wildcard subdomains and follows this format: `protocol://*.foo.bar`, `protocol://*.bar.foo:8080` or `protocol://*.abc.bar.foo:9000`
    - Example: `nginx.ingress.kubernetes.io/cors-allow-origin: "https://*.origin-site.com:4443, http://*.origin-site.com, myprotocol://example.org:1199"`

* `nginx.ingress.kubernetes.io/cors-allow-credentials`: Controls if credentials can be passed during CORS operations.

    - Default: `true`
    - Example: `nginx.ingress.kubernetes.io/cors-allow-credentials: "false"`

* `nginx.ingress.kubernetes.io/cors-max-age`: Controls how long preflight requests can be cached.

    - Default: `1728000`
    - Example: `nginx.ingress.kubernetes.io/cors-max-age: 600`

!!! note
    For more information please see [https://enable-cors.org](https://enable-cors.org/server_nginx.html)

### HTTP2 Push Preload.

Enables automatic conversion of preload links specified in the “Link” response header fields into push requests.

!!! example

    * `nginx.ingress.kubernetes.io/http2-push-preload: "true"`

### Server Alias

Allows the definition of one or more aliases in the server definition of the NGINX configuration using the annotation `nginx.ingress.kubernetes.io/server-alias: "<alias 1>,<alias 2>"`.
This will create a server with the same configuration, but adding new values to the `server_name` directive.

!!! note
	  A server-alias name cannot conflict with the hostname of an existing server. If it does, the server-alias annotation will be ignored.
    If a server-alias is created and later a new server with the same hostname is created, the new server configuration will take
    place over the alias configuration.

For more information please see [the `server_name` documentation](https://nginx.org/en/docs/http/ngx_http_core_module.html#server_name).

### Server snippet

Using the annotation `nginx.ingress.kubernetes.io/server-snippet` it is possible to add custom configuration in the server configuration block.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/server-snippet: |
        set $agentflag 0;

        if ($http_user_agent ~* "(Mobile)" ){
          set $agentflag 1;
        }

        if ( $agentflag = 1 ) {
          return 301 https://m.example.com;
        }
```

!!! attention
    This annotation can be used only once per host.

### Client Body Buffer Size

Sets buffer size for reading client request body per location. In case the request body is larger than the buffer,
the whole body or only its part is written to a temporary file. By default, buffer size is equal to two memory pages.
This is 8K on x86, other 32-bit platforms, and x86-64. It is usually 16K on other 64-bit platforms. This annotation is
applied to each location provided in the ingress rule.

!!! note
    The annotation value must be given in a format understood by Nginx.

!!! example

    * `nginx.ingress.kubernetes.io/client-body-buffer-size: "1000"` # 1000 bytes
    * `nginx.ingress.kubernetes.io/client-body-buffer-size: 1k` # 1 kilobyte
    * `nginx.ingress.kubernetes.io/client-body-buffer-size: 1K` # 1 kilobyte
    * `nginx.ingress.kubernetes.io/client-body-buffer-size: 1m` # 1 megabyte
    * `nginx.ingress.kubernetes.io/client-body-buffer-size: 1M` # 1 megabyte

For more information please see [https://nginx.org](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size)

### External Authentication

To use an existing service that provides authentication the Ingress rule can be annotated with `nginx.ingress.kubernetes.io/auth-url` to indicate the URL where the HTTP request should be sent.

```yaml
nginx.ingress.kubernetes.io/auth-url: "URL to the authentication service"
```

Additionally it is possible to set:

* `nginx.ingress.kubernetes.io/auth-keepalive`:
  `<Connections>` to specify the maximum number of keepalive connections to `auth-url`. Only takes effect
   when no variables are used in the host part of the URL. Defaults to `0` (keepalive disabled).

> Note: does not work with HTTP/2 listener because of a limitation in Lua [subrequests](https://github.com/openresty/lua-nginx-module#spdy-mode-not-fully-supported).
> [UseHTTP2](./configmap.md#use-http2) configuration should be disabled!

* `nginx.ingress.kubernetes.io/auth-keepalive-share-vars`:
  Whether to share Nginx variables among the current request and the auth request. Example use case is to track requests: when set to "true" X-Request-ID HTTP header will be the same for the backend and the auth request.
  Defaults to "false".
* `nginx.ingress.kubernetes.io/auth-keepalive-requests`:
  `<Requests>` to specify the maximum number of requests that can be served through one keepalive connection.
  Defaults to `1000` and only applied if `auth-keepalive` is set to higher than `0`.
* `nginx.ingress.kubernetes.io/auth-keepalive-timeout`:
  `<Timeout>` to specify a duration in seconds which an idle keepalive connection to an upstream server will stay open.
  Defaults to `60` and only applied if `auth-keepalive` is set to higher than `0`.
* `nginx.ingress.kubernetes.io/auth-method`:
  `<Method>` to specify the HTTP method to use.
* `nginx.ingress.kubernetes.io/auth-signin`:
  `<SignIn_URL>` to specify the location of the error page.
* `nginx.ingress.kubernetes.io/auth-signin-redirect-param`:
  `<SignIn_URL>` to specify the URL parameter in the error page which should contain the original URL for a failed signin request.
* `nginx.ingress.kubernetes.io/auth-response-headers`:
  `<Response_Header_1, ..., Response_Header_n>` to specify headers to pass to backend once authentication request completes.
* `nginx.ingress.kubernetes.io/auth-proxy-set-headers`:
  `<ConfigMap>` the name of a ConfigMap that specifies headers to pass to the authentication service
* `nginx.ingress.kubernetes.io/auth-request-redirect`:
  `<Request_Redirect_URL>`  to specify the X-Auth-Request-Redirect header value.
* `nginx.ingress.kubernetes.io/auth-cache-key`:
  `<Cache_Key>` this enables caching for auth requests. specify a lookup key for auth responses. e.g. `$remote_user$http_authorization`. Each server and location has it's own keyspace. Hence a cached response is only valid on a per-server and per-location basis.
* `nginx.ingress.kubernetes.io/auth-cache-duration`:
  `<Cache_duration>` to specify a caching time for auth responses based on their response codes, e.g. `200 202 30m`. See [proxy_cache_valid](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_valid) for details. You may specify multiple, comma-separated values: `200 202 10m, 401 5m`. defaults to `200 202 401 5m`.
* `nginx.ingress.kubernetes.io/auth-always-set-cookie`:
  `<Boolean_Flag>` to set a cookie returned by auth request. By default, the cookie will be set only if an upstream reports with the code 200, 201, 204, 206, 301, 302, 303, 304, 307, or 308.
* `nginx.ingress.kubernetes.io/auth-snippet`:
  `<Auth_Snippet>` to specify a custom snippet to use with external authentication, e.g.

```yaml
nginx.ingress.kubernetes.io/auth-url: http://foo.com/external-auth
nginx.ingress.kubernetes.io/auth-snippet: |
    proxy_set_header Foo-Header 42;
```
> Note: `nginx.ingress.kubernetes.io/auth-snippet` is an optional annotation. However, it may only be used in conjunction with `nginx.ingress.kubernetes.io/auth-url` and will be ignored if `nginx.ingress.kubernetes.io/auth-url` is not set

!!! example
    Please check the [external-auth](../../examples/auth/external-auth/README.md) example.

#### Global External Authentication

By default the controller redirects all requests to an existing service that provides authentication if `global-auth-url` is set in the Ingress NGINX ConfigMap. If you want to disable this behavior for that Ingress, you can use the `nginx.ingress.kubernetes.io/enable-global-auth: "false"` annotation.

- `nginx.ingress.kubernetes.io/enable-global-auth`:
   indicates if GlobalExternalAuth configuration should be applied or not to this Ingress rule. Default values is set to `"true"`.

!!! note
    For more information please see [global-auth-url](./configmap.md#global-auth-url).

### Rate Limiting

These annotations define limits on connections and transmission rates.  These can be used to mitigate [DDoS Attacks](https://www.nginx.com/blog/mitigating-ddos-attacks-with-nginx-and-nginx-plus).

!!! attention
    Rate limits are applied per Ingress NGINX controller replica. 
    If you're running multiple replicas or using a horizontal pod autoscaler (HPA), the effective rate limit will be multiplied by the number of replicas.
    When using HPA, the exact rate limit becomes dynamic as the number of replicas may change based on load.

* `nginx.ingress.kubernetes.io/limit-connections`: number of concurrent connections allowed from a single IP address per controller replica. A 503 error is returned when exceeding this limit.
* `nginx.ingress.kubernetes.io/limit-rps`: number of requests accepted from a given IP each second per controller replica. The burst limit is set to this limit multiplied by the burst multiplier, the default multiplier is 5. When clients exceed this limit, [limit-req-status-code](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#limit-req-status-code) ***default:*** 503 is returned.
* `nginx.ingress.kubernetes.io/limit-rpm`: number of requests accepted from a given IP each minute per controller replica. The burst limit is set to this limit multiplied by the burst multiplier, the default multiplier is 5. When clients exceed this limit, [limit-req-status-code](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#limit-req-status-code) ***default:*** 503 is returned.
* `nginx.ingress.kubernetes.io/limit-burst-multiplier`: multiplier of the limit rate for burst size. The default burst multiplier is 5, this annotation override the default multiplier. When clients exceed this limit, [limit-req-status-code](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#limit-req-status-code) ***default:*** 503 is returned.
* `nginx.ingress.kubernetes.io/limit-rate-after`: initial number of kilobytes after which the further transmission of a response to a given connection will be rate limited. This feature must be used with [proxy-buffering](#proxy-buffering) enabled.
* `nginx.ingress.kubernetes.io/limit-rate`: number of kilobytes per second allowed to send to a given connection.  The zero value disables rate limiting. This feature must be used with [proxy-buffering](#proxy-buffering) enabled.
* `nginx.ingress.kubernetes.io/limit-whitelist`: client IP source ranges to be excluded from rate-limiting. The value is a comma separated list of CIDRs.

If you specify multiple annotations in a single Ingress rule, limits are applied in the order `limit-connections`, `limit-rpm`, `limit-rps`.

To configure settings globally for all Ingress rules, the `limit-rate-after` and `limit-rate` values may be set in the [NGINX ConfigMap](./configmap.md#limit-rate).  The value set in an Ingress annotation will override the global setting.

The client IP address will be set based on the use of [PROXY protocol](./configmap.md#use-proxy-protocol) or from the `X-Forwarded-For` header value when [use-forwarded-headers](./configmap.md#use-forwarded-headers) is enabled.

### Permanent Redirect

This annotation allows to return a permanent redirect (Return Code 301) instead of sending data to the upstream.  For example `nginx.ingress.kubernetes.io/permanent-redirect: https://www.google.com` would redirect everything to Google.

### Permanent Redirect Code

This annotation allows you to modify the status code used for permanent redirects.  For example `nginx.ingress.kubernetes.io/permanent-redirect-code: '308'` would return your permanent-redirect with a 308.

### Temporal Redirect
This annotation allows you to return a temporal redirect (Return Code 302) instead of sending data to the upstream. For example `nginx.ingress.kubernetes.io/temporal-redirect: https://www.google.com` would redirect everything to Google with a Return Code of 302 (Moved Temporarily)

### Temporal Redirect Code

This annotation allows you to modify the status code used for temporal redirects.  For example `nginx.ingress.kubernetes.io/temporal-redirect-code: '307'` would return your temporal-redirect with a 307.

### SSL Passthrough

The annotation `nginx.ingress.kubernetes.io/ssl-passthrough` instructs the controller to send TLS connections directly
to the backend instead of letting NGINX decrypt the communication. See also [TLS/HTTPS](../tls.md#ssl-passthrough) in
the User guide.

!!! note
    SSL Passthrough is **disabled by default** and requires starting the controller with the
    [`--enable-ssl-passthrough`](../cli-arguments.md) flag.

!!! attention
    Because SSL Passthrough works on layer 4 of the OSI model (TCP) and not on the layer 7 (HTTP), using SSL Passthrough
    invalidates all the other annotations set on an Ingress object.

### Service Upstream

By default the Ingress-Nginx Controller uses a list of all endpoints (Pod IP/port) in the NGINX upstream configuration.

The `nginx.ingress.kubernetes.io/service-upstream` annotation disables that behavior and instead uses a single upstream in NGINX, the service's Cluster IP and port.

This can be desirable for things like zero-downtime deployments . See issue [#257](https://github.com/kubernetes/ingress-nginx/issues/257).

#### Known Issues

If the `service-upstream` annotation is specified the following things should be taken into consideration:

* Sticky Sessions will not work as only round-robin load balancing is supported.
* The `proxy_next_upstream` directive will not have any effect meaning on error the request will not be dispatched to another upstream.

### Server-side HTTPS enforcement through redirect

By default the controller redirects (308) to HTTPS if TLS is enabled for that ingress.
If you want to disable this behavior globally, you can use `ssl-redirect: "false"` in the NGINX [ConfigMap](./configmap.md#ssl-redirect).

To configure this feature for specific ingress resources, you can use the `nginx.ingress.kubernetes.io/ssl-redirect: "false"`
annotation in the particular resource.

When using SSL offloading outside of cluster (e.g. AWS ELB) it may be useful to enforce a redirect to HTTPS
even when there is no TLS certificate available.
This can be achieved by using the `nginx.ingress.kubernetes.io/force-ssl-redirect: "true"` annotation in the particular resource.

To preserve the trailing slash in the URI with `ssl-redirect`, set `nginx.ingress.kubernetes.io/preserve-trailing-slash: "true"` annotation for that particular resource.

### Redirect from/to www

In some scenarios, it is required to redirect from `www.domain.com` to `domain.com` or vice versa, which way the redirect is performed depends on the configured `host` value in the Ingress object.

For example, if `.spec.rules.host` is configured with a value like `www.example.com`, then this annotation will redirect from `example.com` to `www.example.com`. If `.spec.rules.host` is configured with a value like `example.com`, so without a `www`, then this annotation will redirect from `www.example.com` to `example.com` instead.

To enable this feature use the annotation `nginx.ingress.kubernetes.io/from-to-www-redirect: "true"`

!!! attention
    If at some point a new Ingress is created with a host equal to one of the options (like `domain.com`) the annotation will be omitted.

!!! attention
    For HTTPS to HTTPS redirects is mandatory the SSL Certificate defined in the Secret, located in the TLS section of Ingress, contains both FQDN in the common name of the certificate.

### Denylist source range

You can specify blocked client IP source ranges through the `nginx.ingress.kubernetes.io/denylist-source-range` annotation.
The value is a comma separated list of [CIDRs](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), e.g.  `10.0.0.0/24,172.10.0.1`.

To configure this setting globally for all Ingress rules, the `denylist-source-range` value may be set in the [NGINX ConfigMap](./configmap.md#denylist-source-range).

!!! note
    Adding an annotation to an Ingress rule overrides any global restriction.


### Whitelist source range

You can specify allowed client IP source ranges through the `nginx.ingress.kubernetes.io/whitelist-source-range` annotation.
The value is a comma separated list of [CIDRs](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing), e.g.  `10.0.0.0/24,172.10.0.1`.

To configure this setting globally for all Ingress rules, the `whitelist-source-range` value may be set in the [NGINX ConfigMap](./configmap.md#whitelist-source-range).

!!! note
    Adding an annotation to an Ingress rule overrides any global restriction.

### Custom timeouts

Using the configuration configmap it is possible to set the default global timeout for connections to the upstream servers.
In some scenarios is required to have different values. To allow this we provide annotations that allows this customization:

- `nginx.ingress.kubernetes.io/proxy-connect-timeout`
- `nginx.ingress.kubernetes.io/proxy-send-timeout`
- `nginx.ingress.kubernetes.io/proxy-read-timeout`
- `nginx.ingress.kubernetes.io/proxy-next-upstream`
- `nginx.ingress.kubernetes.io/proxy-next-upstream-timeout`
- `nginx.ingress.kubernetes.io/proxy-next-upstream-tries`
- `nginx.ingress.kubernetes.io/proxy-request-buffering`

If you indicate [Backend Protocol](#backend-protocol) as `GRPC` or `GRPCS`, the following grpc values will be set and inherited from proxy timeouts:

- [`grpc_connect_timeout=5s`](https://nginx.org/en/docs/http/ngx_http_grpc_module.html#grpc_connect_timeout), from `nginx.ingress.kubernetes.io/proxy-connect-timeout`
- [`grpc_send_timeout=60s`](https://nginx.org/en/docs/http/ngx_http_grpc_module.html#grpc_send_timeout), from `nginx.ingress.kubernetes.io/proxy-send-timeout`
- [`grpc_read_timeout=60s`](https://nginx.org/en/docs/http/ngx_http_grpc_module.html#grpc_read_timeout), from `nginx.ingress.kubernetes.io/proxy-read-timeout`

Note: All timeout values are unitless and in seconds e.g. `nginx.ingress.kubernetes.io/proxy-read-timeout: "120"` sets a valid 120 seconds proxy read timeout.

### Proxy redirect

The annotations `nginx.ingress.kubernetes.io/proxy-redirect-from` and `nginx.ingress.kubernetes.io/proxy-redirect-to` will set the first and second parameters of NGINX's proxy_redirect directive respectively. It is possible to
set the text that should be changed in the `Location` and `Refresh` header fields of a [proxied server response](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_redirect)

Setting "off" or "default" in the annotation `nginx.ingress.kubernetes.io/proxy-redirect-from` disables `nginx.ingress.kubernetes.io/proxy-redirect-to`,
otherwise, both annotations must be used in unison. Note that each annotation must be a string without spaces.

By default the value of each annotation is "off".

### Custom max body size

For NGINX, an 413 error will be returned to the client when the size in a request exceeds the maximum allowed size of the client request body. This size can be configured by the parameter [`client_max_body_size`](https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size).

To configure this setting globally for all Ingress rules, the `proxy-body-size` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-body-size).
To use custom values in an Ingress rule define these annotation:

```yaml
nginx.ingress.kubernetes.io/proxy-body-size: 8m
```

### Proxy cookie domain

Sets a text that [should be changed in the domain attribute](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_domain) of the "Set-Cookie" header fields of a proxied server response.

To configure this setting globally for all Ingress rules, the `proxy-cookie-domain` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-cookie-domain).

### Proxy cookie path

Sets a text that [should be changed in the path attribute](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_path) of the "Set-Cookie" header fields of a proxied server response.

To configure this setting globally for all Ingress rules, the `proxy-cookie-path` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-cookie-path).

### Proxy buffering

Enable or disable proxy buffering [`proxy_buffering`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering).
By default proxy buffering is disabled in the NGINX config.

To configure this setting globally for all Ingress rules, the `proxy-buffering` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-buffering).
To use custom values in an Ingress rule define these annotation:

```yaml
nginx.ingress.kubernetes.io/proxy-buffering: "on"
```

### Proxy buffers Number

Sets the number of the buffers in [`proxy_buffers`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffers) used for reading the first part of the response received from the proxied server.
By default proxy buffers number is set as 4

To configure this setting globally, set `proxy-buffers-number` in [NGINX ConfigMap](./configmap.md#proxy-buffers-number). To use custom values in an Ingress rule, define this annotation:
```yaml
nginx.ingress.kubernetes.io/proxy-buffers-number: "4"
```

### Proxy buffer size

Sets the size of the buffer [`proxy_buffer_size`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size) used for reading the first part of the response received from the proxied server.
By default proxy buffer size is set as "4k"

To configure this setting globally, set `proxy-buffer-size` in [NGINX ConfigMap](./configmap.md#proxy-buffer-size). To use custom values in an Ingress rule, define this annotation:
```yaml
nginx.ingress.kubernetes.io/proxy-buffer-size: "8k"
```

### Proxy busy buffers size

[Limits the total size of buffers that can be busy](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_busy_buffers_size) sending a response to the client while the response is not yet fully read.

By default proxy busy buffers size is set as "8k".

To configure this setting globally, set `proxy-busy-buffers-size` in the [ConfigMap](./configmap.md#proxy-busy-buffers-size). To use custom values in an Ingress rule, define this annotation:

```yaml
nginx.ingress.kubernetes.io/proxy-busy-buffers-size: "16k"
```

### Proxy max temp file size

When [`buffering`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering) of responses from the proxied server is enabled, and the whole response does not fit into the buffers set by the [`proxy_buffer_size`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size) and [`proxy_buffers`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffers) directives, a part of the response can be saved to a temporary file. This directive sets the maximum `size` of the temporary file setting the [`proxy_max_temp_file_size`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_max_temp_file_size). The size of data written to the temporary file at a time is set by the [`proxy_temp_file_write_size`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_temp_file_write_size) directive.

The zero value disables buffering of responses to temporary files.

To use custom values in an Ingress rule, define this annotation:
```yaml
nginx.ingress.kubernetes.io/proxy-max-temp-file-size: "1024m"
```

### Proxy HTTP version

Using this annotation sets the [`proxy_http_version`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_http_version) that the Nginx reverse proxy will use to communicate with the backend.
By default this is set to "1.1".

```yaml
nginx.ingress.kubernetes.io/proxy-http-version: "1.0"
```

### SSL ciphers

Specifies the [enabled ciphers](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers).

Using this annotation will set the `ssl_ciphers` directive at the server level. This configuration is active for all the paths in the host.

```yaml
nginx.ingress.kubernetes.io/ssl-ciphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"
```

The following annotation will set the `ssl_prefer_server_ciphers` directive at the server level. This configuration specifies that server ciphers should be preferred over client ciphers when using the SSLv3 and TLS protocols.

```yaml
nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers: "true"
```

### Connection proxy header

Using this annotation will override the default connection header set by NGINX.
To use custom values in an Ingress rule, define the annotation:

```yaml
nginx.ingress.kubernetes.io/connection-proxy-header: "keep-alive"
```

### Enable Access Log

Access logs are enabled by default, but in some scenarios access logs might be required to be disabled for a given
ingress. To do this, use the annotation:

```yaml
nginx.ingress.kubernetes.io/enable-access-log: "false"
```

### Enable Rewrite Log

Rewrite logs are not enabled by default. In some scenarios it could be required to enable NGINX rewrite logs.
Note that rewrite logs are sent to the error_log file at the notice level. To enable this feature use the annotation:

```yaml
nginx.ingress.kubernetes.io/enable-rewrite-log: "true"
```

### Enable Opentelemetry

Opentelemetry can be enabled or disabled globally through the ConfigMap but this will sometimes need to be overridden
to enable it or disable it for a specific ingress (e.g. to turn off telemetry of external health check endpoints)

```yaml
nginx.ingress.kubernetes.io/enable-opentelemetry: "true"
```

### Opentelemetry Trust Incoming Span

The option to trust incoming trace spans can be enabled or disabled globally through the ConfigMap but this will
sometimes need to be overridden to enable it or disable it for a specific ingress (e.g. only enable on a private endpoint)

```yaml
nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-spans: "true"
```

### X-Forwarded-Prefix Header
To add the non-standard `X-Forwarded-Prefix` header to the upstream request with a string value, the following annotation can be used:

```yaml
nginx.ingress.kubernetes.io/x-forwarded-prefix: "/path"
```

### ModSecurity

[ModSecurity](http://modsecurity.org/) is an OpenSource Web Application firewall. It can be enabled for a particular set
of ingress locations. The ModSecurity module must first be enabled by enabling ModSecurity in the
[ConfigMap](./configmap.md#enable-modsecurity). Note this will enable ModSecurity for all paths, and each path
must be disabled manually.

It can be enabled using the following annotation:
```yaml
nginx.ingress.kubernetes.io/enable-modsecurity: "true"
```
ModSecurity will run in "Detection-Only" mode using the [recommended configuration](https://github.com/SpiderLabs/ModSecurity/blob/v3/master/modsecurity.conf-recommended).

You can enable the [OWASP Core Rule Set](https://www.modsecurity.org/CRS/Documentation/) by
setting the following annotation:
```yaml
nginx.ingress.kubernetes.io/enable-owasp-core-rules: "true"
```

You can pass transactionIDs from nginx by setting up the following:
```yaml
nginx.ingress.kubernetes.io/modsecurity-transaction-id: "$request_id"
```

You can also add your own set of modsecurity rules via a snippet:
```yaml
nginx.ingress.kubernetes.io/modsecurity-snippet: |
SecRuleEngine On
SecDebugLog /tmp/modsec_debug.log
```

Note: If you use both `enable-owasp-core-rules` and `modsecurity-snippet` annotations together, only the
`modsecurity-snippet` will take effect. If you wish to include the [OWASP Core Rule Set](https://www.modsecurity.org/CRS/Documentation/) or
[recommended configuration](https://github.com/SpiderLabs/ModSecurity/blob/v3/master/modsecurity.conf-recommended) simply use the include
statement:

nginx 0.24.1 and below
```yaml
nginx.ingress.kubernetes.io/modsecurity-snippet: |
Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
Include /etc/nginx/modsecurity/modsecurity.conf
```
nginx 0.25.0 and above
```yaml
nginx.ingress.kubernetes.io/modsecurity-snippet: |
Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
```

### Backend Protocol

Using `backend-protocol` annotations is possible to indicate how NGINX should communicate with the backend service. (Replaces `secure-backends` in older versions)
Valid Values: HTTP, HTTPS, AUTO_HTTP, GRPC, GRPCS and FCGI

By default NGINX uses `HTTP`.

Example:

```yaml
nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
```

### Use Regex

!!! attention
    When using this annotation with the NGINX annotation `nginx.ingress.kubernetes.io/affinity` of type `cookie`,  `nginx.ingress.kubernetes.io/session-cookie-path` must be also set; Session cookie paths do not support regex.

Using the `nginx.ingress.kubernetes.io/use-regex` annotation will indicate whether or not the paths defined on an Ingress use regular expressions.  The default value is `false`.

The following will indicate that regular expression paths are being used:
```yaml
nginx.ingress.kubernetes.io/use-regex: "true"
```

The following will indicate that regular expression paths are __not__ being used:
```yaml
nginx.ingress.kubernetes.io/use-regex: "false"
```

When this annotation is set to `true`, the case insensitive regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) will be enforced on ALL paths for a given host regardless of what Ingress they are defined on.

Additionally, if the [`rewrite-target` annotation](#rewrite) is used on any Ingress for a given host, then the case insensitive regular expression [location modifier](https://nginx.org/en/docs/http/ngx_http_core_module.html#location) will be enforced on ALL paths for a given host regardless of what Ingress they are defined on.

Please read about [ingress path matching](../ingress-path-matching.md) before using this modifier.

### Satisfy

By default, a request would need to satisfy all authentication requirements in order to be allowed. By using this annotation, requests that satisfy either any or all authentication requirements are allowed, based on the configuration value.

```yaml
nginx.ingress.kubernetes.io/satisfy: "any"
```

### Mirror

Enables a request to be mirrored to a mirror backend. Responses by mirror backends are ignored. This feature is useful, to see how requests will react in "test" backends.

The mirror backend can be set by applying:

```yaml
nginx.ingress.kubernetes.io/mirror-target: https://test.env.com$request_uri
```

By default the request-body is sent to the mirror backend, but can be turned off by applying:

```yaml
nginx.ingress.kubernetes.io/mirror-request-body: "off"
```

Also by default header Host for mirrored requests will be set the same as a host part of uri in the "mirror-target" annotation. You can override it by "mirror-host" annotation:

```yaml
nginx.ingress.kubernetes.io/mirror-target: https://1.2.3.4$request_uri
nginx.ingress.kubernetes.io/mirror-host: "test.env.com"
```

**Note:** The mirror directive will be applied to all paths within the ingress resource.

The request sent to the mirror is linked to the original request. If you have a slow mirror backend, then the original request will throttle.

For more information on the mirror module see [ngx_http_mirror_module](https://nginx.org/en/docs/http/ngx_http_mirror_module.html)


### Stream snippet

Using the annotation `nginx.ingress.kubernetes.io/stream-snippet` it is possible to add custom stream configuration.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/stream-snippet: |
      server {
        listen 8000;
        proxy_pass 127.0.0.1:80;
      }
```
