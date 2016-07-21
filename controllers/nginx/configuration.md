## Contents
* [Customizing NGINX](#customizing-nginx)
* [Custom NGINX configuration](#custom-nginx-configuration)
* [Custom NGINX template](#custom-nginx-template)
* [Annotations](#annotations)
* [Custom NGINX upstream checks](#custom-nginx-upstream-checks)
* [Authentication](#authentication)
* [Rewrite](#rewrite)
* [Rate limiting](#rate-limiting)
* [Secure backends](#secure-backends)
* [Whitelist source range](#whitelist-source-range)
* [Allowed parameters in configuration config map](#allowed-parameters-in-configuration-configmap)
* [Default configuration options](#default-configuration-options)
* [Websockets](#websockets)
* [Optimizing TLS Time To First Byte (TTTFB)](#optimizing-tls-time-to-first-byte-tttfb)
* [Retries in no idempotent methods](#retries-in-no-idempotent-methods)


### Customizing nginx

there are 3 ways to customize nginx

1. config map: create a stand alone config map, use this if you want a different global configuration
2. annoations: [annotate the ingress](#annotations), use this if you want a specific configuration for the site defined in the ingress rule
3. custom template: when is required a specific setting like [open_file_cache](http://nginx.org/en/docs/http/ngx_http_core_module.html#open_file_cache), custom [log_format](http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format), adjust [listen](http://nginx.org/en/docs/http/ngx_http_core_module.html#listen) options as `rcvbuf` or when is not possible to change an through the config map


#### Custom NGINX configuration

It's possible to customize the defaults in NGINX using a config map.

Please check the [custom configuration](examples/custom-configuration/README.md) example

#### Annotations

The following annotations are supported:

|Name                 |type|
|---------------------------|------|
|[ingress.kubernetes.io/rewrite-target](#rewrite)|URI|
|[ingress.kubernetes.io/add-base-url](#rewrite)|true or false|
|[ingress.kubernetes.io/limit-connections](#rate-limiting)|number|
|[ingress.kubernetes.io/limit-rps](#rate-limiting)|number|
|[ingress.kubernetes.io/auth-type](#authentication)|basic or digest|
|[ingress.kubernetes.io/auth-secret](#authentication)|string|
|[ingress.kubernetes.io/auth-realm](#authentication)|string|
|[ingress.kubernetes.io/ssl-redirect](#server-side-https-enforcement-through-redirect)|true or false|
|[ingress.kubernetes.io/upstream-max-fails](#custom-nginx-upstream-checks)|number|
|[ingress.kubernetes.io/upstream-fail-timeout](#custom-nginx-upstream-checks)|number|
|[ingress.kubernetes.io/secure-backends](#secure-backends)|true or false|
|[ingress.kubernetes.io/whitelist-source-range](#whitelist-source-range)|CIDR|



#### Custom NGINX template

The NGINX template is located in the file `/etc/nginx/template/nginx.tmpl`. Mounting a volume is possible to use a custom version.
Use the [custom-template](examples/custom-template/README.md) example as a guide

**Please note the template is tied to the go code. Be sure to no change names in the variable `$cfg`**


### Custom NGINX upstream checks

NGINX exposes some flags in the [upstream configuration](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) that enables the configuration of each server in the upstream. The ingress controller allows custom `max_fails` and `fail_timeout` parameters in a global context using `upstream-max-fails` or `upstream-fail-timeout` in the NGINX config map or in a particular Ingress rule. It defaults to 0. This means NGINX will respect the `readinessProbe`, if is defined. If there is no probe, NGINX will not mark a server inside an upstream down.

**With the default values NGINX will not health check your backends, and whenever the endpoints controller notices a readiness probe failure that pod's ip will be removed from the list of endpoints, causing nginx to also remove it from the upstreams.**

To use custom values in an Ingress rule define this annotations:

`ingress.kubernetes.io/upstream-max-fails`: number of unsuccessful attempts to communicate with the server that should happen in the duration set by the fail_timeout parameter to consider the server unavailable

`ingress.kubernetes.io/upstream-fail-timeout`: time in seconds during which the specified number of unsuccessful attempts to communicate with the server should happen to consider the server unavailable. Also the period of time the server will be considered unavailable.

**Important:** 
The upstreams are shared. i.e. Ingress rule using the same service will use the same upstream. 
This means only one of the rules should define annotations to configure the upstream servers


Please check the [custom upstream check](examples/custom-upstream-check/README.md) example


### Authentication

Is possible to add authentication adding additional annotations in the Ingress rule. The source of the authentication is a secret that contains usernames and passwords inside the the key `auth`

The annotations are:

```
ingress.kubernetes.io/auth-type:[basic|digest]
```

Indicates the [HTTP Authentication Type: Basic or Digest Access Authentication](https://tools.ietf.org/html/rfc2617).

```
ingress.kubernetes.io/auth-secret:secretName
```

Name of the secret that contains the usernames and passwords with access to the `path/s` defined in the Ingress Rule.
The secret must be created in the same namespace than the Ingress rule

```
ingress.kubernetes.io/auth-realm:"realm string"
```

Please check the [auth](examples/custom-upstream-check/README.md) example


### Rewrite

In some scenarios the exposed URL in the backend service differs from the specified path in the Ingress rule. Without a rewrite any request will return 404.
Set the annotation `ingress.kubernetes.io/rewrite-target` to the path expected by the service.

If the application contains relative links is possible to add an additional annotation `ingress.kubernetes.io/add-base-url` that will append a `base` tag in the header of the returned HTML from the backend.


Please check the [rewrite](examples/rewrite/README.md) example


### Rate limiting

The annotations `ingress.kubernetes.io/limit-connections` and `ingress.kubernetes.io/limit-rps` allows the creation of a limit in the connections that can be opened by a single client IP address. This can be use to mitigate [DDoS Attacks](https://www.nginx.com/blog/mitigating-ddos-attacks-with-nginx-and-nginx-plus)

`ingress.kubernetes.io/limit-connections`: number of concurrent allowed connections from a single IP address

`ingress.kubernetes.io/limit-rps`: number of allowed connections per second from a single IP address


Is possible to specify both annotation in the same Ingress rule. If you specify both annotations in a single Ingress rule, limit-rps takes precedence


### Secure upstreams

By default NGINX uses `http` to reach the services. Adding the annotation `ingress.kubernetes.io/secure-backends: "true"` in the ingress rule changes the protocol to `https`.


### Whitelist source range

You can specify the allowed client ip source ranges through the `ingress.kubernetes.io/whitelist-source-range` annotation, eg;  `10.0.0.0/24,172.10.0.1`
For a global restriction (any URL) is possible to use `whitelist-source-range` in the NGINX config map

*Note:* adding an annotation overrides any global restriction

Please check the [whitelist](examples/whitelist/README.md) example



### **Allowed parameters in configuration config map:**

**body-size:** Sets the maximum allowed size of the client request body. See NGINX [client_max_body_size](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size)


**custom-http-errors:** Enables which HTTP codes should be passed for processing with the [error_page directive](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page)
Setting at least one code this also enables [proxy_intercept_errors](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors) (required to process error_page)
For instance setting `custom-http-errors: 404,415` 


**enable-sticky-sessions:**  Enables sticky sessions using cookies. This is provided by [nginx-sticky-module-ng](https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng) module


**enable-vts-status:** Allows the replacement of the default status page with a third party module named [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)


**error-log-level:** Configures the logging level of errors. Log levels above are listed in the order of increasing severity
http://nginx.org/en/docs/ngx_core_module.html#error_log


**retry-non-idempotent:** Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error in the upstream server. 
The previous behavior can be restored using the value "true"


**hsts:** Enables or disables the header HSTS in servers running SSL.
HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header) that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP. It provides protection against protocol downgrade attacks and cookie theft.
https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
https://blog.qualys.com/securitylabs/2016/03/28/the-importance-of-a-proper-http-strict-transport-security-implementation-on-your-web-server


**hsts-include-subdomains:** Enables or disables the use of HSTS in all the subdomains of the servername


**hsts-max-age:** Sets the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.


**keep-alive:** Sets the time during which a keep-alive client connection will stay open on the server side.
The zero value disables keep-alive client connections
http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout


**max-worker-connections:** Sets the maximum number of simultaneous connections that can be opened by each [worker process](http://nginx.org/en/docs/ngx_core_module.html#worker_connections)


**proxy-connect-timeout:** Sets the timeout for [establishing a connection with a proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout). It should be noted that this timeout cannot usually exceed 75 seconds.


**proxy-read-timeout:** Sets the timeout in seconds for [reading a response from the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). The timeout is set only between two successive read operations, not for the transmission of the whole response 


**proxy-send-timeout:** Sets the timeout in seconds for [transmitting a request to the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout). The timeout is set only between two successive write operations, not for the transmission of the whole request.


**resolver:** Configures name servers used to [resolve](http://nginx.org/en/docs/http/ngx_http_core_module.html#resolver) names of upstream servers into addresses


**server-name-hash-max-size:** Sets the maximum size of the [server names hash tables](http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size) used in server names, map directive’s values, MIME types, names of request header strings, etc.
http://nginx.org/en/docs/hash.html


**server-name-hash-bucket-size:** Sets the size of the bucker for the server names hash tables
http://nginx.org/en/docs/hash.html
http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size

**ssl-buffer-size:** Sets the size of the [SSL buffer](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) used for sending data.
4k helps NGINX to improve TLS Time To First Byte (TTTFB)
https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/

**ssl-ciphers:** Sets the [ciphers](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers) list to enable. The ciphers are specified in the format understood by the OpenSSL library
The default cipher list is: `ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA`


The ordering of a ciphersuite is very important because it decides which algorithms are going to be selected in priority. 
The recommendation above prioritizes algorithms that provide perfect [forward secrecy](https://wiki.mozilla.org/Security/Server_Side_TLS#Forward_Secrecy)

Please check the [Mozilla SSL Configuration Generator](https://mozilla.github.io/server-side-tls/ssl-config-generator/)


**ssl-protocols:** Sets the [SSL protocols](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols) to use.
The default is: `TLSv1 TLSv1.1 TLSv1.2`

TLSv1 is enabled to allow old clients like:
- [IE 8-10 / Win 7](https://www.ssllabs.com/ssltest/viewClient.html?name=IE&version=8-10&platform=Win%207&key=113)
- [Java 7u25](https://www.ssllabs.com/ssltest/viewClient.html?name=Java&version=7u25&key=26)

If you dont need to support this clients please remove TLSv1


Please check the result of the configuration using `https://ssllabs.com/ssltest/analyze.html` or `https://testssl.sh`


**ssl-dh-param:** sets the Base64 string that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
https://www.openssl.org/docs/manmaster/apps/dhparam.html
https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam


**ssl-session-cache:** Enables or disables the use of shared [SSL cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) among worker processes.


**ssl-session-cache-size:** Sets the size of the [SSL shared session cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) between all worker processes.


**ssl-session-tickets:** Enables or disables session resumption through [TLS session tickets](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets)


**ssl-session-timeout:** Sets the time during which a client may [reuse the session](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout) parameters stored in a cache.


**ssl-redirect:** Sets the global value of redirects (301) to HTTPS if the server has a TLS certificate (defined in an Ingress rule)
Default is true


**upstream-max-fails:** Sets the number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) that should happen in the duration set by the fail_timeout parameter to consider the server unavailable


**upstream-fail-timeout:** Sets the time during which the specified number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) should happen to consider the server unavailable


**use-proxy-protocol:** Enables or disables the use of the [PROXY protocol](https://www.nginx.com/resources/admin-guide/proxy-protocol/) to receive client connection (real IP address) information passed through proxy servers and load balancers such as HAproxy and Amazon Elastic Load Balancer (ELB).


**use-gzip:** Enables or disables the use of the nginx module that compresses responses using the ["gzip" module](http://nginx.org/en/docs/http/ngx_http_gzip_module.html)
The default mime type list to compress is: `application/atom+xml application/javascript aplication/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component`

**use-http2:** Enables or disables the [HTTP/2](http://nginx.org/en/docs/http/ngx_http_v2_module.html) support in secure connections 


**gzip-types:** Sets the MIME types in addition to "text/html" to compress. The special value "*"" matches any MIME type.
Responses with the "text/html" type are always compressed if `use-gzip` is enabled


**worker-processes:** Sets the number of [worker processes](http://nginx.org/en/docs/ngx_core_module.html#worker_processes). By default "auto" means number of available CPU cores


### Default configuration options

Running `/nginx-ingress-controller --dump-nginx-configuration` is possible to get the value of the options that can be changed.
The next table shows the options, the default value and a description

|name                 |default|
|---------------------------|------|
|body-size|1m|
|custom-http-errors|" "|
|enable-sticky-sessions|"false"|
|enable-vts-status|"false"|
|error-log-level|notice|
|gzip-types||
|hsts|"true"|
|hsts-include-subdomains|"true"|
|hsts-max-age|"15724800"|
|keep-alive|"75"|
|max-worker-connections|"16384"|
|proxy-connect-timeout|"5"|
|proxy-read-timeout|"60"|
|proxy-real-ip-cidr|0.0.0.0/0|
|proxy-send-timeout|"60"|
|retry-non-idempotent|"false"|
|server-name-hash-bucket-size|"64"|
|server-name-hash-max-size|"512"|
|ssl-buffer-size|4k|
|ssl-ciphers||
|ssl-protocols|TLSv1 TLSv1.1 TLSv1.2|
|ssl-session-cache|"true"|
|ssl-session-cache-size|10m|
|ssl-session-tickets|"true"|
|ssl-session-timeout|10m|
|use-gzip|"true"|
|use-http2|"true"|
|vts-status-zone-size|10m|
|worker-processes|<number of CPUs>|


### Websockets

Support for websockets is provided by NGINX OOTB. No special configuration required.

The only requirement to avoid the close of connections is the increase of the values of `proxy-read-timeout` and `proxy-send-timeout`. The default value of this settings is `30 seconds`. 
A more adequate value to support websockets is a value higher than one hour (`3600`)


#### Optimizing TLS Time To First Byte (TTTFB)

NGINX provides the configuration option [ssl_buffer_size](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) to allow the optimization of the TLS record size. This improves the [Time To First Byte](https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/) (TTTFB). The default value in the Ingress controller is `4k` (nginx default is `16k`);

#### Retries in no idempotent methods

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error.
The previous behavior can be restored using `retry-non-idempotent=true` in the configuration config map
