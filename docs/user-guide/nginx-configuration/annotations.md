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
|[nginx.ingress.kubernetes.io/auth-realm](#authentication)|string|
|[nginx.ingress.kubernetes.io/auth-secret](#authentication)|string|
|[nginx.ingress.kubernetes.io/auth-type](#authentication)|basic or digest|
|[nginx.ingress.kubernetes.io/auth-tls-secret](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-verify-depth](#client-certificate-authentication)|number|
|[nginx.ingress.kubernetes.io/auth-tls-verify-client](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-error-page](#client-certificate-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream](#client-certificate-authentication)|"true" or "false"|
|[nginx.ingress.kubernetes.io/auth-url](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/auth-snippet](#external-authentication)|string|
|[nginx.ingress.kubernetes.io/backend-protocol](#backend-protocol)|string|HTTP,HTTPS,GRPC,GRPCS,AJP|
|[nginx.ingress.kubernetes.io/canary](#canary)|"true" or "false"|
|[nginx.ingress.kubernetes.io/canary-by-header](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-by-header-value](#canary)|string
|[nginx.ingress.kubernetes.io/canary-by-cookie](#canary)|string|
|[nginx.ingress.kubernetes.io/canary-weight](#canary)|number|
|[nginx.ingress.kubernetes.io/client-body-buffer-size](#client-body-buffer-size)|string|
|[nginx.ingress.kubernetes.io/configuration-snippet](#configuration-snippet)|string|
|[nginx.ingress.kubernetes.io/custom-http-errors](#custom-http-errors)|[]int|
|[nginx.ingress.kubernetes.io/default-backend](#default-backend)|string|
|[nginx.ingress.kubernetes.io/enable-cors](#enable-cors)|"true" or "false"|
|[nginx.ingress.kubernetes.io/cors-allow-origin](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-methods](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-headers](#enable-cors)|string|
|[nginx.ingress.kubernetes.io/cors-allow-credentials](#enable-cors)|"true" or "false"|
|[nginx.ingress.kubernetes.io/cors-max-age](#enable-cors)|number|
|[nginx.ingress.kubernetes.io/force-ssl-redirect](#server-side-https-enforcement-through-redirect)|"true" or "false"|
|[nginx.ingress.kubernetes.io/from-to-www-redirect](#redirect-from-to-www)|"true" or "false"|
|[nginx.ingress.kubernetes.io/http2-push-preload](#http2-push-preload)|"true" or "false"|
|[nginx.ingress.kubernetes.io/limit-connections](#rate-limiting)|number|
|[nginx.ingress.kubernetes.io/limit-rps](#rate-limiting)|number|
|[nginx.ingress.kubernetes.io/permanent-redirect](#permanent-redirect)|string|
|[nginx.ingress.kubernetes.io/permanent-redirect-code](#permanent-redirect-code)|number|
|[nginx.ingress.kubernetes.io/temporal-redirect](#temporal-redirect)|string|
|[nginx.ingress.kubernetes.io/proxy-body-size](#custom-max-body-size)|string|
|[nginx.ingress.kubernetes.io/proxy-cookie-domain](#proxy-cookie-domain)|string|
|[nginx.ingress.kubernetes.io/proxy-cookie-path](#proxy-cookie-path)|string|
|[nginx.ingress.kubernetes.io/proxy-connect-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-send-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-read-timeout](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-next-upstream](#custom-timeouts)|string|
|[nginx.ingress.kubernetes.io/proxy-next-upstream-tries](#custom-timeouts)|number|
|[nginx.ingress.kubernetes.io/proxy-request-buffering](#custom-timeouts)|string|
|[nginx.ingress.kubernetes.io/proxy-redirect-from](#proxy-redirect)|string|
|[nginx.ingress.kubernetes.io/proxy-redirect-to](#proxy-redirect)|string|
|[nginx.ingress.kubernetes.io/enable-rewrite-log](#enable-rewrite-log)|"true" or "false"|
|[nginx.ingress.kubernetes.io/rewrite-target](#rewrite)|URI|
|[nginx.ingress.kubernetes.io/secure-verify-ca-secret](#secure-backends)|string|
|[nginx.ingress.kubernetes.io/server-alias](#server-alias)|string|
|[nginx.ingress.kubernetes.io/server-snippet](#server-snippet)|string|
|[nginx.ingress.kubernetes.io/service-upstream](#service-upstream)|"true" or "false"|
|[nginx.ingress.kubernetes.io/session-cookie-name](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-hash](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/session-cookie-path](#cookie-affinity)|string|
|[nginx.ingress.kubernetes.io/ssl-redirect](#server-side-https-enforcement-through-redirect)|"true" or "false"|
|[nginx.ingress.kubernetes.io/ssl-passthrough](#ssl-passthrough)|"true" or "false"|
|[nginx.ingress.kubernetes.io/upstream-hash-by](#custom-nginx-upstream-hashing)|string|
|[nginx.ingress.kubernetes.io/x-forwarded-prefix](#x-forwarded-prefix-header)|string|
|[nginx.ingress.kubernetes.io/load-balance](#custom-nginx-load-balancing)|string|
|[nginx.ingress.kubernetes.io/upstream-vhost](#custom-nginx-upstream-vhost)|string|
|[nginx.ingress.kubernetes.io/whitelist-source-range](#whitelist-source-range)|CIDR|
|[nginx.ingress.kubernetes.io/proxy-buffering](#proxy-buffering)|string|
|[nginx.ingress.kubernetes.io/proxy-buffer-size](#proxy-buffer-size)|string|
|[nginx.ingress.kubernetes.io/ssl-ciphers](#ssl-ciphers)|string|
|[nginx.ingress.kubernetes.io/connection-proxy-header](#connection-proxy-header)|string|
|[nginx.ingress.kubernetes.io/enable-access-log](#enable-access-log)|"true" or "false"|
|[nginx.ingress.kubernetes.io/lua-resty-waf](#lua-resty-waf)|string|
|[nginx.ingress.kubernetes.io/lua-resty-waf-debug](#lua-resty-waf)|"true" or "false"|
|[nginx.ingress.kubernetes.io/lua-resty-waf-ignore-rulesets](#lua-resty-waf)|string|
|[nginx.ingress.kubernetes.io/lua-resty-waf-extra-rules](#lua-resty-waf)|string|
|[nginx.ingress.kubernetes.io/lua-resty-waf-allow-unknown-content-types](#lua-resty-waf)|"true" or "false"|
|[nginx.ingress.kubernetes.io/lua-resty-waf-score-threshold](#lua-resty-waf)|number|
|[nginx.ingress.kubernetes.io/lua-resty-waf-process-multipart-body](#lua-resty-waf)|"true" or "false"|
|[nginx.ingress.kubernetes.io/enable-influxdb](#influxdb)|"true" or "false"|
|[nginx.ingress.kubernetes.io/influxdb-measurement](#influxdb)|string|
|[nginx.ingress.kubernetes.io/influxdb-port](#influxdb)|string|
|[nginx.ingress.kubernetes.io/influxdb-host](#influxdb)|string|
|[nginx.ingress.kubernetes.io/influxdb-server-name](#influxdb)|string|
|[nginx.ingress.kubernetes.io/use-regex](#use-regex)|bool|
|[nginx.ingress.kubernetes.io/enable-modsecurity](#modsecurity)|bool|
|[nginx.ingress.kubernetes.io/enable-owasp-core-rules](#modsecurity)|bool|
|[nginx.ingress.kubernetes.io/modsecurity-transaction-id](#modsecurity)|string|
|[nginx.ingress.kubernetes.io/modsecurity-snippet](#modsecurity)|string|

### Canary

In some cases, you may want to "canary" a new set of changes by sending a small number of requests to a different service than the production service. The canary annotation enables the Ingress spec to act as an alternative service for requests to route to depending on the rules applied. The following annotations to configure canary can be enabled after `nginx.ingress.kubernetes.io/canary: "true"` is set:

* `nginx.ingress.kubernetes.io/canary-by-header`: The header to use for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the request header is set to `always`, it will be routed to the canary. When the header is set to `never`, it will never be routed to the canary. For any other value, the header will be ignored and the request compared against the other canary rules by precedence.

* `nginx.ingress.kubernetes.io/canary-by-header-value`: The header value to match for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the request header is set to this value, it will be routed to the canary. For any other header value, the header will be ignored and the request compared against the other canary rules by precedence. This annotation has to be used together with . The annotation is an extension of the `nginx.ingress.kubernetes.io/canary-by-header` to allow customizing the header value instead of using hardcoded values. It doesn't have any effect if the `nginx.ingress.kubernetes.io/canary-by-header` annotation is not defined.

* `nginx.ingress.kubernetes.io/canary-by-cookie`: The cookie to use for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the cookie value is set to `always`, it will be routed to the canary. When the cookie is set to `never`, it will never be routed to the canary. For any other value, the cookie will be ingored and the request compared against the other canary rules by precedence. 

* `nginx.ingress.kubernetes.io/canary-weight`: The integer based (0 - 100) percent of random requests that should be routed to the service specified in the canary Ingress. A weight of 0 implies that no requests will be sent to the service in the Canary ingress by this canary rule. A weight of 100 means implies all requests will be sent to the alternative service specified in the Ingress.   

Canary rules are evaluated in order of precedence. Precedence is as follows: 
`canary-by-header -> canary-by-cookie -> canary-weight` 

**Note** that when you mark an ingress as canary, then all the other non-canary annotations will be ignored (inherited from the corresponding main ingress) except `nginx.ingress.kubernetes.io/load-balance` and `nginx.ingress.kubernetes.io/upstream-hash-by`.

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

!!! attention
    If more than one Ingress is defined for a host and at least one Ingress uses `nginx.ingress.kubernetes.io/affinity: cookie`, then only paths on the Ingress using `nginx.ingress.kubernetes.io/affinity` will use session cookie affinity. All paths defined on other Ingresses for the host will be load balanced through the random selection of a backend server. 

!!! example
    Please check the [affinity](../../examples/affinity/cookie/README.md) example.

#### Cookie affinity

If you use the ``cookie`` affinity type you can also specify the name of the cookie that will be used to route the requests with the annotation `nginx.ingress.kubernetes.io/session-cookie-name`. The default is to create a cookie named 'INGRESSCOOKIE'.

In case of NGINX the annotation `nginx.ingress.kubernetes.io/session-cookie-hash` defines which algorithm will be used to hash the used upstream. Default value is `md5` and possible values are `md5`, `sha1` and `index`.

The NGINX annotation `nginx.ingress.kubernetes.io/session-cookie-path` defines the path that will be set on the cookie. This is optional unless the annotation `nginx.ingress.kubernetes.io/use-regex` is set to true; Session cookie paths do not support regex.  

!!! attention
    The `index` option is not an actual hash; an in-memory index is used instead, which has less overhead.
    However, with `index`, matching against a changing upstream server list is inconsistent.
    So, at reload, if upstream servers have changed, index values are not guaranteed to correspond to the same server as before!
    **Use `index` with caution** and only if you need to!

In NGINX this feature is implemented by the third party module [nginx-sticky-module-ng](https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng). The workflow used to define which upstream server will be used is explained [here](https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng/raw/08a395c66e425540982c00482f55034e1fee67b6/docs/sticky.pdf)


### Authentication

Is possible to add authentication adding additional annotations in the Ingress rule. The source of the authentication is a secret that contains usernames and passwords inside the key `auth`.

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
nginx.ingress.kubernetes.io/auth-realm: "realm string"
```

!!! example
    Please check the [auth](../../examples/auth/basic/README.md) example.

### Custom NGINX upstream hashing

NGINX supports load balancing by client-server mapping based on [consistent hashing](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#hash) for a given key. The key can contain text, variables or any combination thereof. This feature allows for request stickiness other than client IP or cookies. The [ketama](http://www.last.fm/user/RJ/journal/2007/04/10/392555/) consistent hashing method will be used which ensures only a few keys would be remapped to different servers on upstream group changes.

There is a special mode of upstream hashing called subset. In this mode, upstream servers are grouped into subsets, and stickiness works by mapping keys to a subset instead of individual upstream servers. Specific server is chosen uniformly at random from the selected sticky subset. It provides a balance between stickiness and load distribution.

To enable consistent hashing for a backend:

`nginx.ingress.kubernetes.io/upstream-hash-by`: the nginx variable, text value or any combination thereof to use for consistent hashing. For example `nginx.ingress.kubernetes.io/upstream-hash-by: "$request_uri"` to consistently hash upstream requests by the current request URI.

"subset" hashing can be enabled setting `nginx.ingress.kubernetes.io/upstream-hash-by-subset`: "true". This maps requests to subset of nodes instead of a single one. `upstream-hash-by-subset-size` determines the size of each subset (default 3).

Please check the [chashsubset](../../examples/chashsubset/deployment.yaml) example.

### Custom NGINX load balancing

This is similar to [`load-balance` in ConfigMap](./configmap.md#load-balance), but configures load balancing algorithm per ingress.
>Note that `nginx.ingress.kubernetes.io/upstream-hash-by` takes preference over this. If this and `nginx.ingress.kubernetes.io/upstream-hash-by` are not set then we fallback to using globally configured load balancing algorithm.

### Custom NGINX upstream vhost

This configuration setting allows you to control the value for host in the following statement: `proxy_set_header Host $host`, which forms part of the location block.  This is useful if you need to call the upstream server by something other than `$host`.

### Client Certificate Authentication

It is possible to enable Client Certificate Authentication using additional annotations in Ingress Rule.

The annotations are:

* `nginx.ingress.kubernetes.io/auth-tls-secret: secretName`:
  The name of the Secret that contains the full Certificate Authority chain `ca.crt` that is enabled to authenticate against this Ingress.
  This annotation also accepts the alternative form "namespace/secretName", in which case the Secret lookup is performed in the referenced namespace instead of the Ingress namespace.
* `nginx.ingress.kubernetes.io/auth-tls-verify-depth`:
  The validation depth between the provided client certificate and the Certification Authority chain.
* `nginx.ingress.kubernetes.io/auth-tls-verify-client`:
  Enables verification of client certificates.
* `nginx.ingress.kubernetes.io/auth-tls-error-page`:
  The URL/Page that user should be redirected in case of a Certificate Authentication Error
* `nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream`:
  Indicates if the received certificates should be passed or not to the upstream server.  By default this is disabled.

!!! example
    Please check the [client-certs](../../examples/auth/client-certs/README.md) example.

!!! attention
    TLS with Client Authentication is **not** possible in Cloudflare and might result in unexpected behavior.

    Cloudflare only allows Authenticated Origin Pulls and is required to use their own certificate: [https://blog.cloudflare.com/protecting-the-origin-with-tls-authenticated-origin-pulls/](https://blog.cloudflare.com/protecting-the-origin-with-tls-authenticated-origin-pulls/)

    Only Authenticated Origin Pulls are allowed and can be configured by following their tutorial: [https://support.cloudflare.com/hc/en-us/articles/204494148-Setting-up-NGINX-to-use-TLS-Authenticated-Origin-Pulls](https://support.cloudflare.com/hc/en-us/articles/204494148-Setting-up-NGINX-to-use-TLS-Authenticated-Origin-Pulls)


### Configuration snippet

Using this annotation you can add additional configuration to the NGINX location. For example:

```yaml
nginx.ingress.kubernetes.io/configuration-snippet: |
  more_set_headers "Request-Id: $req_id";
```

### Default Backend

The ingress controller requires a [default backend](../default-backend.md).
This service handles the response when the service in the Ingress rule does not have endpoints.
This is a global configuration for the ingress controller. In some cases could be required to return a custom content or format. In this scenario we can use the annotation `nginx.ingress.kubernetes.io/default-backend: <svc name>` to specify a custom default backend.  This `<svc name>` is a reference to a service inside of the same namespace in which you are applying this annotation.

### Custom HTTP Errors

Like the [`custom-http-errors`](./configmap.md#custom-http-errors) value in the ConfigMap, this annotation will set NGINX `proxy-intercept-errors`, but only for the NGINX location associated with this ingress.
Different ingresses can specify different sets of error codes. Even if multiple ingress objects share the same hostname, this annotation can be used to intercept different error codes for each ingress (for example, different error codes to be intercepted for different paths on the same hostname, if each path is on a different ingress).
If `custom-http-errors` is also specified globally, the error values specified in this annotation will override the global value for the given ingress' hostname and path.

Example usage:
```
custom-http-errors: "404,415"
```

### Enable CORS

To enable Cross-Origin Resource Sharing (CORS) in an Ingress rule, add the annotation
`nginx.ingress.kubernetes.io/enable-cors: "true"`. This will add a section in the server
location enabling this functionality.

CORS can be controlled with the following annotations:

* `nginx.ingress.kubernetes.io/cors-allow-methods`
  controls which methods are accepted. This is a multi-valued field, separated by ',' and
  accepts only letters (upper and lower case).
  - Default: `GET, PUT, POST, DELETE, PATCH, OPTIONS`
  - Example: `nginx.ingress.kubernetes.io/cors-allow-methods: "PUT, GET, POST, OPTIONS"`

* `nginx.ingress.kubernetes.io/cors-allow-headers`
  controls which headers are accepted. This is a multi-valued field, separated by ',' and accepts letters,
  numbers, _ and -.
  - Default: `DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization`
  - Example: `nginx.ingress.kubernetes.io/cors-allow-headers: "X-Forwarded-For, X-app123-XPTO"`

* `nginx.ingress.kubernetes.io/cors-allow-origin`
  controls what's the accepted Origin for CORS.
  This is a single field value, with the following format: `http(s)://origin-site.com` or `http(s)://origin-site.com:port`
  - Default: `*`
  - Example: `nginx.ingress.kubernetes.io/cors-allow-origin: "https://origin-site.com:4443"`

* `nginx.ingress.kubernetes.io/cors-allow-credentials`
  controls if credentials can be passed during CORS operations.
  - Default: `true`
  - Example: `nginx.ingress.kubernetes.io/cors-allow-credentials: "false"`

* `nginx.ingress.kubernetes.io/cors-max-age`
  controls how long preflight requests can be cached.
  Default: `1728000`
  Example: `nginx.ingress.kubernetes.io/cors-max-age: 600`

!!! note
    For more information please see [https://enable-cors.org](https://enable-cors.org/server_nginx.html) 

### HTTP2 Push Preload.

Enables automatic conversion of preload links specified in the “Link” response header fields into push requests.

!!! example

    * `nginx.ingress.kubernetes.io/http2-push-preload: "true"`

### Server Alias

To add Server Aliases to an Ingress rule add the annotation `nginx.ingress.kubernetes.io/server-alias: "<alias>"`.
This will create a server with the same configuration, but a different `server_name` as the provided host.

!!! Note
	A server-alias name cannot conflict with the hostname of an existing server. If it does the server-alias annotation will be ignored.
    If a server-alias is created and later a new server with the same hostname is created,
    the new server configuration will take place over the alias configuration.

For more information please see [the `server_name` documentation](http://nginx.org/en/docs/http/ngx_http_core_module.html#server_name).

### Server snippet

Using the annotation `nginx.ingress.kubernetes.io/server-snippet` it is possible to add custom configuration in the server configuration block.

```yaml
apiVersion: extensions/v1beta1
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

For more information please see [http://nginx.org](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size)

### External Authentication

To use an existing service that provides authentication the Ingress rule can be annotated with `nginx.ingress.kubernetes.io/auth-url` to indicate the URL where the HTTP request should be sent.

```yaml
nginx.ingress.kubernetes.io/auth-url: "URL to the authentication service"
```

Additionally it is possible to set:

* `nginx.ingress.kubernetes.io/auth-method`:
  `<Method>` to specify the HTTP method to use.
* `nginx.ingress.kubernetes.io/auth-signin`:
  `<SignIn_URL>` to specify the location of the error page.
* `nginx.ingress.kubernetes.io/auth-response-headers`:
  `<Response_Header_1, ..., Response_Header_n>` to specify headers to pass to backend once authentication request completes.
* `nginx.ingress.kubernetes.io/auth-request-redirect`:
  `<Request_Redirect_URL>`  to specify the X-Auth-Request-Redirect header value.
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

### Rate limiting

These annotations define a limit on the connections that can be opened by a single client IP address.
This can be used to mitigate [DDoS Attacks](https://www.nginx.com/blog/mitigating-ddos-attacks-with-nginx-and-nginx-plus).

* `nginx.ingress.kubernetes.io/limit-connections`: number of concurrent connections allowed from a single IP address.
* `nginx.ingress.kubernetes.io/limit-rps`: number of connections that may be accepted from a given IP each second.
* `nginx.ingress.kubernetes.io/limit-rpm`: number of connections that may be accepted from a given IP each minute.
* `nginx.ingress.kubernetes.io/limit-rate-after`: sets the initial amount after which the further transmission of a response to a client will be rate limited.
* `nginx.ingress.kubernetes.io/limit-rate`: rate of request that accepted from a client each second.

You can specify the client IP source ranges to be excluded from rate-limiting through the `nginx.ingress.kubernetes.io/limit-whitelist` annotation. The value is a comma separated list of CIDRs.

If you specify multiple annotations in a single Ingress rule, `limit-rpm`, and then `limit-rps` takes precedence.

The annotation `nginx.ingress.kubernetes.io/limit-rate`, `nginx.ingress.kubernetes.io/limit-rate-after` define a limit the rate of response transmission to a client. The rate is specified in bytes per second. The zero value disables rate limiting. The limit is set per a request, and so if a client simultaneously opens two connections, the overall rate will be twice as much as the specified limit.

To configure this setting globally for all Ingress rules, the `limit-rate-after` and `limit-rate` value may be set in the [NGINX ConfigMap](./configmap.md#limit-rate). if you set the value in ingress annotation will cover global setting.

### Permanent Redirect

This annotation allows to return a permanent redirect instead of sending data to the upstream.  For example `nginx.ingress.kubernetes.io/permanent-redirect: https://www.google.com` would redirect everything to Google.

### Permanent Redirect Code

This annotation allows you to modify the status code used for permanent redirects.  For example `nginx.ingress.kubernetes.io/permanent-redirect-code: '308'` would return your permanent-redirect with a 308.

### Temporal Redirect
This annotation allows you to return a temporal redirect (Return Code 302) instead of sending data to the upstream. For example `nginx.ingress.kubernetes.io/temporal-redirect: https://www.google.com` would redirect everything to Google with a Return Code of 302 (Moved Temporarily)

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

By default the NGINX ingress controller uses a list of all endpoints (Pod IP/port) in the NGINX upstream configuration.

The `nginx.ingress.kubernetes.io/service-upstream` annotation disables that behavior and instead uses a single upstream in NGINX, the service's Cluster IP and port.

This can be desirable for things like zero-downtime deployments as it reduces the need to reload NGINX configuration when Pods come up and down. See issue [#257](https://github.com/kubernetes/ingress-nginx/issues/257).

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

### Redirect from/to www

In some scenarios is required to redirect from `www.domain.com` to `domain.com` or vice versa.
To enable this feature use the annotation `nginx.ingress.kubernetes.io/from-to-www-redirect: "true"`

!!! attention
    If at some point a new Ingress is created with a host equal to one of the options (like `domain.com`) the annotation will be omitted.

!!! attention
    For HTTPS to HTTPS redirects is mandatory the SSL Certificate defined in the Secret, located in the TLS section of Ingress, contains both FQDN in the common name of the certificate.

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
- `nginx.ingress.kubernetes.io/proxy-next-upstream-tries`
- `nginx.ingress.kubernetes.io/proxy-request-buffering`

### Proxy redirect

With the annotations `nginx.ingress.kubernetes.io/proxy-redirect-from` and `nginx.ingress.kubernetes.io/proxy-redirect-to` it is possible to
set the text that should be changed in the `Location` and `Refresh` header fields of a [proxied server response](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_redirect)

Setting "off" or "default" in the annotation `nginx.ingress.kubernetes.io/proxy-redirect-from` disables `nginx.ingress.kubernetes.io/proxy-redirect-to`,
otherwise, both annotations must be used in unison. Note that each annotation must be a string without spaces.

By default the value of each annotation is "off".

### Custom max body size

For NGINX, an 413 error will be returned to the client when the size in a request exceeds the maximum allowed size of the client request body. This size can be configured by the parameter [`client_max_body_size`](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size).

To configure this setting globally for all Ingress rules, the `proxy-body-size` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-body-size).
To use custom values in an Ingress rule define these annotation:

```yaml
nginx.ingress.kubernetes.io/proxy-body-size: 8m
```

### Proxy cookie domain

Sets a text that [should be changed in the domain attribute](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_domain) of the "Set-Cookie" header fields of a proxied server response.

To configure this setting globally for all Ingress rules, the `proxy-cookie-domain` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-cookie-domain).

### Proxy cookie path

Sets a text that [should be changed in the path attribute](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_path) of the "Set-Cookie" header fields of a proxied server response.

To configure this setting globally for all Ingress rules, the `proxy-cookie-path` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-cookie-path).

### Proxy buffering

Enable or disable proxy buffering [`proxy_buffering`](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering).
By default proxy buffering is disabled in the NGINX config.

To configure this setting globally for all Ingress rules, the `proxy-buffering` value may be set in the [NGINX ConfigMap](./configmap.md#proxy-buffering).
To use custom values in an Ingress rule define these annotation:

```yaml
nginx.ingress.kubernetes.io/proxy-buffering: "on"
```

### Proxy buffer size

Sets the size of the buffer [`proxy_buffer_size`](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size) used for reading the first part of the response received from the proxied server.
By default proxy buffer size is set as "4k"

To configure this setting globally, set `proxy-buffer-size` in [NGINX ConfigMap](./configmap.md#proxy-buffer-size). To use custom values in an Ingress rule, define this annotation:
```yaml
nginx.ingress.kubernetes.io/proxy-buffer-size: "8k"
```

### SSL ciphers

Specifies the [enabled ciphers](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers).

Using this annotation will set the `ssl_ciphers` directive at the server level. This configuration is active for all the paths in the host.

```yaml
nginx.ingress.kubernetes.io/ssl-ciphers: "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP"
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

### X-Forwarded-Prefix Header
To add the non-standard `X-Forwarded-Prefix` header to the upstream request with a string value, the following annotation can be used:

```yaml
nginx.ingress.kubernetes.io/x-forwarded-prefix: "/path"
```

### Lua Resty WAF

Using `lua-resty-waf-*` annotations we can enable and control the [lua-resty-waf](https://github.com/p0pr0ck5/lua-resty-waf)
Web Application Firewall per location.

Following configuration will enable the WAF for the paths defined in the corresponding ingress:

```yaml
nginx.ingress.kubernetes.io/lua-resty-waf: "active"
```

In order to run it in debugging mode you can set `nginx.ingress.kubernetes.io/lua-resty-waf-debug` to `"true"` in addition to the above configuration.
The other possible values for `nginx.ingress.kubernetes.io/lua-resty-waf` are `inactive` and `simulate`.
In `inactive` mode WAF won't do anything, whereas in `simulate` mode it will log a warning message if there's a matching WAF rule for given request. This is useful to debug a rule and eliminate possible false positives before fully deploying it.

`lua-resty-waf` comes with predefined set of rules [https://github.com/p0pr0ck5/lua-resty-waf/tree/84b4f40362500dd0cb98b9e71b5875cb1a40f1ad/rules](https://github.com/p0pr0ck5/lua-resty-waf/tree/84b4f40362500dd0cb98b9e71b5875cb1a40f1ad/rules) that covers ModSecurity CRS.
You can use `nginx.ingress.kubernetes.io/lua-resty-waf-ignore-rulesets` to ignore a subset of those rulesets. For an example:

```yaml
nginx.ingress.kubernetes.io/lua-resty-waf-ignore-rulesets: "41000_sqli, 42000_xss"
```

will ignore the two mentioned rulesets.

It is also possible to configure custom WAF rules per ingress using the `nginx.ingress.kubernetes.io/lua-resty-waf-extra-rules` annotation. For an example the following snippet will configure a WAF rule to deny requests with query string value that contains word `foo`:


```yaml
nginx.ingress.kubernetes.io/lua-resty-waf-extra-rules: '[=[ { "access": [ { "actions": { "disrupt" : "DENY" }, "id": 10001, "msg": "my custom rule", "operator": "STR_CONTAINS", "pattern": "foo", "vars": [ { "parse": [ "values", 1 ], "type": "REQUEST_ARGS" } ] } ], "body_filter": [], "header_filter":[] } ]=]'
```

Since the default allowed contents were `"text/html", "text/json", "application/json"`
We can enable the following annotation for allow all contents type:


```yaml
nginx.ingress.kubernetes.io/lua-resty-waf-allow-unknown-content-types: "true"
```

The default score of lua-resty-waf is 5, which usually triggered if hitting 2 default rules, you can modify the score threshold with following annotation:


```yaml
nginx.ingress.kubernetes.io/lua-resty-waf-score-threshold: "10"
```

When you enabled HTTPS in the endpoint and since resty-lua will return 500 error when processing "multipart" contents
Reference for this [issue](https://github.com/p0pr0ck5/lua-resty-waf/issues/166)

By default, it will be "true"

You may enable the following annotation for work around:

```yaml
nginx.ingress.kubernetes.io/lua-resty-waf-process-multipart-body: "false"
```

For details on how to write WAF rules, please refer to [https://github.com/p0pr0ck5/lua-resty-waf](https://github.com/p0pr0ck5/lua-resty-waf).

[configmap]: ./configmap.md

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
```yaml
nginx.ingress.kubernetes.io/modsecurity-snippet: |
Include /etc/nginx/owasp-modsecurity-crs/nginx-modsecurity.conf
Include /etc/nginx/modsecurity/modsecurity.conf
```

### InfluxDB

Using `influxdb-*` annotations we can monitor requests passing through a Location by sending them to an InfluxDB backend exposing the UDP socket
using the [nginx-influxdb-module](https://github.com/influxdata/nginx-influxdb-module/).

```yaml
nginx.ingress.kubernetes.io/enable-influxdb: "true"
nginx.ingress.kubernetes.io/influxdb-measurement: "nginx-reqs"
nginx.ingress.kubernetes.io/influxdb-port: "8089"
nginx.ingress.kubernetes.io/influxdb-host: "127.0.0.1"
nginx.ingress.kubernetes.io/influxdb-server-name: "nginx-ingress"
```

For the `influxdb-host` parameter you have two options:

- Use an InfluxDB server configured with the  [UDP protocol](https://docs.influxdata.com/influxdb/v1.5/supported_protocols/udp/) enabled. 
- Deploy Telegraf as a sidecar proxy to the Ingress controller configured to listen UDP with the [socket listener input](https://github.com/influxdata/telegraf/tree/release-1.6/plugins/inputs/socket_listener) and to write using
anyone of the [outputs plugins](https://github.com/influxdata/telegraf/tree/release-1.7/plugins/outputs) like InfluxDB, Apache Kafka,
Prometheus, etc.. (recommended)

It's important to remember that there's no DNS resolver at this stage so you will have to configure
an ip address to `nginx.ingress.kubernetes.io/influxdb-host`. If you deploy Influx or Telegraf as sidecar (another container in the same pod) this becomes straightforward since you can directly use `127.0.0.1`.

### Backend Protocol

Using `backend-protocol` annotations is possible to indicate how NGINX should communicate with the backend service.
Valid Values: HTTP, HTTPS, GRPC, GRPCS and AJP

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
