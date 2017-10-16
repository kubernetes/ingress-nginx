# NGINX Ingress controller configuration ConfigMap

---

### Logs

#### disable-access-log

Disables the Access Log from the entire Ingress Controller. This is 'false' by default.

#### access-log-path

Access log path. Goes to '/var/log/nginx/access.log' by default.

_References:_

- http://nginx.org/en/docs/http/ngx_http_log_module.html#access_log

#### error-log-level

Configures the logging level of errors. Log levels above are listed in the order of increasing severity.

_References:_

- http://nginx.org/en/docs/ngx_core_module.html#error_log

#### error-log-path

Error log path. Goes to '/var/log/nginx/error.log' by default.

_References:_

- http://nginx.org/en/docs/ngx_core_module.html#error_log

#### log-format-stream

Sets the nginx [stream format](https://nginx.org/en/docs/stream/ngx_stream_log_module.html#log_format).

#### log-format-upstream

Sets the nginx [log format](http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format).
Example for json output:

```console
log-format-upstream: '{ "time": "$time_iso8601", "remote_addr": "$proxy_protocol_addr",
    "x-forward-for": "$proxy_add_x_forwarded_for", "request_id": "$request_id", "remote_user":
    "$remote_user", "bytes_sent": $bytes_sent, "request_time": $request_time, "status":
    $status, "vhost": "$host", "request_proto": "$server_protocol", "path": "$uri",
    "request_query": "$args", "request_length": $request_length, "duration": $request_time,
    "method": "$request_method", "http_referrer": "$http_referer", "http_user_agent":
    "$http_user_agent" }'
  ```

Please check [log-format](log-format.md) for definition of each field.

### Proxy configuration

#### load-balance

Sets the algorithm to use for load balancing.
The value can either be:

- round_robin: to use the default round robin loadbalancer
- least_conn: to use the least connected method
- ip_hash: to use a hash of the server for routing.

The default is least_conn.

_References:_

- http://nginx.org/en/docs/http/load_balancing.html.

#### proxy-body-size

Sets the maximum allowed size of the client request body. 
See NGINX [client_max_body_size](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size).

#### proxy-buffer-size

Sets the size of the buffer used for [reading the first part of the response](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size) received from the proxied server. This part usually contains a small response header.

#### proxy-connect-timeout

Sets the timeout for [establishing a connection with a proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout). It should be noted that this timeout cannot usually exceed 75 seconds.

#### proxy-cookie-domain

Sets a text that [should be changed in the domain attribute](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_domain) of the “Set-Cookie” header fields of a proxied server response.

#### proxy-cookie-path

Sets a text that [should be changed in the path attribute](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_path) of the “Set-Cookie” header fields of a proxied server response.

#### proxy-next-upstream

Specifies in [which cases](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream) a request should be passed to the next server.

#### proxy-read-timeout

Sets the timeout in seconds for [reading a response from the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). The timeout is set only between two successive read operations, not for the transmission of the whole response.

#### proxy-send-timeout

Sets the timeout in seconds for [transmitting a request to the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout). The timeout is set only between two successive write operations, not for the transmission of the whole request.

#### proxy-request-buffering

Enables or disables [buffering of a client request body](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_request_buffering).

#### custom-http-errors

Enables which HTTP codes should be passed for processing with the [error_page directive](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page).
Setting at least one code also enables [proxy_intercept_errors](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors) which are required to process error_page.

Example usage: `custom-http-errors: 404,415`

#### enable-modsecurity

Enables the modsecurity module for NGINX
By default this is disabled.

#### enable-owasp-modsecurity-crs

Eenables the OWASP ModSecurity Core Rule Set (CRS)
By default this is disabled.

#### disable-ipv6

Disable listening on IPV6.
By default this is disabled.

#### enable-dynamic-tls-records

Enables dynamically sized TLS records to improve time-to-first-byte.
By default this is enabled.
See [CloudFlare's blog](https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency) for more information.

#### enable-underscores-in-headers

Enables underscores in header names.
By default this is disabled.

#### enable-vts-status

Allows the replacement of the default status page with a third party module named [nginx-module-vts](https://github.com/vozlt/nginx-module-vts).
By default this is disabled.

#### gzip-types

Sets the MIME types in addition to "text/html" to compress. The special value "\*" matches any MIME type.
Responses with the "text/html" type are always compressed if `use-gzip` is enabled.

#### hsts

Enables or disables the header HSTS in servers running SSL.
HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header) that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP. It provides protection against protocol downgrade attacks and cookie theft.

_References:_

- https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
- https://blog.qualys.com/securitylabs/2016/03/28/the-importance-of-a-proper-http-strict-transport-security-implementation-on-your-web-server

#### hsts-include-subdomains

Enables or disables the use of HSTS in all the subdomains of the server-name.

#### hsts-max-age

Sets the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.

#### hsts-preload

Enables or disables the preload attribute in the HSTS feature (when it is enabled)

#### ignore-invalid-headers

Set if header fields with invalid names should be ignored.
By default this is enabled.

#### keep-alive

Sets the time during which a keep-alive client connection will stay open on the server side.
The zero value disables keep-alive client connections.

_References:_

- http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout

#### max-worker-connections

Sets the maximum number of simultaneous connections that can be opened by each [worker process](http://nginx.org/en/docs/ngx_core_module.html#worker_connections)

#### retry-non-idempotent

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error in the upstream server.
The previous behavior can be restored using the value "true".

#### server-name-hash-bucket-size

Sets the size of the bucket for the server names hash tables.

_References:_

- http://nginx.org/en/docs/hash.html
- http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size

#### server-name-hash-max-size

Sets the maximum size of the [server names hash tables](http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size) used in server names,map directive’s values, MIME types, names of request header strings, etc.

_References:_

- http://nginx.org/en/docs/hash.html

#### proxy-headers-hash-bucket-size

Sets the size of the bucket for the proxy headers hash tables.

_References:_

- http://nginx.org/en/docs/hash.html
- https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_bucket_size

#### proxy-headers-hash-max-size

Sets the maximum size of the proxy headers hash tables.

_References:_

- http://nginx.org/en/docs/hash.html
- https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_max_size

#### server-tokens

Send NGINX Server header in responses and display NGINX version in error pages.
By default this is enabled.

#### map-hash-bucket-size

Sets the bucket size for the [map variables hash tables](http://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size). 
The details of setting up hash tables are provided in a separate [document](http://nginx.org/en/docs/hash.html).

#### ssl-buffer-size

Sets the size of the [SSL buffer](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) used for sending data.
The default of 4k helps NGINX to improve TLS Time To First Byte (TTTFB).
https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/

#### ssl-ciphers

Sets the [ciphers](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers) list to enable. 
The ciphers are specified in the format understood by the OpenSSL library.

The default cipher list is:
 `ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256`.

The ordering of a ciphersuite is very important because it decides which algorithms are going to be selected in priority.
The recommendation above prioritizes algorithms that provide perfect [forward secrecy](https://wiki.mozilla.org/Security/Server_Side_TLS#Forward_Secrecy).

Please check the [Mozilla SSL Configuration Generator](https://mozilla.github.io/server-side-tls/ssl-config-generator/).

#### ssl-dh-param

Sets the name of the secret that contains Diffie-Hellman key to help with "Perfect Forward Secrecy".

_References:_

- https://www.openssl.org/docs/manmaster/apps/dhparam.html
- https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
- http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam

#### ssl-protocols

Sets the [SSL protocols](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols) to use.
The default is: `TLSv1.2`.

Please check the result of the configuration using `https://ssllabs.com/ssltest/analyze.html` or `https://testssl.sh`.

#### ssl-redirect

Sets the global value of redirects (301) to HTTPS if the server has a TLS certificate (defined in an Ingress rule).

Default is "true".

#### ssl-session-cache

Enables or disables the use of shared [SSL cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) among worker processes.

#### ssl-session-cache-size

Sets the size of the [SSL shared session cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) between all worker processes.

#### ssl-session-tickets

Enables or disables session resumption through [TLS session tickets](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets).

#### ssl-session-ticket-key

Sets the secret key used to encrypt and decrypt TLS session tickets. The value must be a valid base64 string.
http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
By default, a randomly generated key is used.

To create a ticket: `openssl rand 80 | base64 -w0`

#### ssl-session-timeout

Sets the time during which a client may [reuse the session](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout) parameters stored in a cache.

#### upstream-max-fails

Sets the number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) that should happen in the duration set by the `fail_timeout` parameter to consider the server unavailable.

#### upstream-fail-timeout

Sets the time during which the specified number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) should happen to consider the server unavailable.

#### use-gzip

Enables or disables compression of HTTP responses using the ["gzip" module](http://nginx.org/en/docs/http/ngx_http_gzip_module.html).

The default mime type list to compress is: `application/atom+xml application/javascript aplication/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component`.

#### use-http2

Enables or disables [HTTP/2](http://nginx.org/en/docs/http/ngx_http_v2_module.html) support in secure connections.

#### use-proxy-protocol

Enables or disables the [PROXY protocol](https://www.nginx.com/resources/admin-guide/proxy-protocol/) to receive client connection (real IP address) information passed through proxy servers and load balancers such as HAProxy and Amazon Elastic Load Balancer (ELB).

#### whitelist-source-range

Sets the default whitelisted IPs for each `server` block.
This can be overwritten by an annotation on an Ingress rule.
See [ngx_http_access_module](http://nginx.org/en/docs/http/ngx_http_access_module.html).

#### worker-processes

Sets the number of [worker processes](http://nginx.org/en/docs/ngx_core_module.html#worker_processes). 
The default of "auto" means number of available CPU cores.

#### worker-shutdown-timeout

Sets a timeout for Nginx to [wait for worker to gracefully shutdown](http://nginx.org/en/docs/ngx_core_module.html#worker_shutdown_timeout).
The default is "10s".

#### limit-conn-zone-variable

Sets parameters for a shared memory zone that will keep states for various keys of [limit_conn_zone](http://nginx.org/en/docs/http/ngx_http_limit_conn_module.html#limit_conn_zone). The default of "$binary_remote_addr" variable’s size is always 4 bytes for IPv4 addresses or 16 bytes for IPv6 addresses.

#### proxy-set-headers

Sets custom headers from a configmap before sending traffic to backends. See [example](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-headers)

#### add-headers

Sets custom headers from a configmap before sending traffic to the client. See `proxy-set-headers` [example](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-headers)

#### bind-address

Sets the addresses on which the server will accept requests instead of *.
It should be noted that these addresses must exist in the runtime environment or the controller will crash loop.

#### http-snippet

Adds custom configuration to the http section of the nginx configuration
Default: ""

#### server-snippet

Adds custom configuration to all the servers in the nginx configuration
Default: ""

#### location-snippet

Adds custom configuration to all the locations in the nginx configuration
Default: ""

### Opentracing

#### enable-opentracing

Enables the nginx Opentracing extension https://github.com/rnburn/nginx-opentracing
By default this is disabled

#### zipkin-collector-host

Specifies the host to use when uploading traces. It must be a valid URL

#### zipkin-collector-port

Specifies the port to use when uploading traces
Default: 9411

#### zipkin-service-name

Specifies the service name to use for any traces created
Default: nginx

### Default configuration options

The following table shows the options, the default value and a description.

|name                                   | default |
|:---                                   |:-------|
|body-size|1m|
|custom-http-errors|" "|
|enable-dynamic-tls-records|"true"|
|enable-sticky-sessions|"false"|
|enable-underscores-in-headers|"false"|
|enable-vts-status|"false"|
|error-log-level|notice|
|forwarded-for-header|X-Forwarded-For|
|gzip-types|see use-gzip description above|
|hsts|"true"|
|hsts-include-subdomains|"true"|
|hsts-max-age|"15724800"|
|hsts-preload|"false"|
|ignore-invalid-headers|"true"|
|keep-alive|"75"|
|log-format-stream|[$time_local] $protocol $status $bytes_sent $bytes_received $session_time|
|log-format-upstream|[$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status|
|map-hash-bucket-size|"64"|
|max-worker-connections|"16384"|
|proxy-body-size|same as body-size|
|proxy-buffer-size|"4k"|
|proxy-request-buffering|"on"|
|proxy-connect-timeout|"5"|
|proxy-cookie-domain|"off"|
|proxy-cookie-path|"off"|
|proxy-read-timeout|"60"|
|proxy-real-ip-cidr|0.0.0.0/0|
|proxy-send-timeout|"60"|
|proxy-stream-timeout|"600s"|
|retry-non-idempotent|"false"|
|server-name-hash-bucket-size|"64"|
|server-name-hash-max-size|"512"|
|server-tokens|"true"|
|ssl-buffer-size|4k|
|ssl-ciphers|ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256|
|ssl-dh-param|value from openssl|
|ssl-protocols|TLSv1.2|
|ssl-session-cache|"true"|
|ssl-session-cache-size|10m|
|ssl-session-tickets|"true"|
|ssl-session-timeout|10m|
|use-gzip|"true"|
|use-http2|"true"|
|upstream-keepalive-connections|"0" (disabled)|
|variables-hash-bucket-size|64|
|variables-hash-max-size|2048|
|vts-status-zone-size|10m|
|vts-default-filter-key|$geoip_country_code country::*|
|whitelist-source-range|permit all|
|worker-processes|number of CPUs|
|limit-conn-zone-variable|$binary_remote_addr|
|bind-address||
