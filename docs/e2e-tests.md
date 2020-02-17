

# e2e test suite for [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx/tree/master/)



### [[Default Backend] change default settings](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/with_hosts.go#L33)

- [should apply the annotation to the default backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/with_hosts.go#L41)

### [[Default Backend]](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/default_backend.go#L30)

- [should return 404 sending requests when only a default backend is running](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/default_backend.go#L33)
- [enables access logging for default backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/default_backend.go#L94)
- [disables access logging for default backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/default_backend.go#L110)

### [[Default Backend] custom service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/custom_default_backend.go#L35)

- [uses custom default backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/custom_default_backend.go#L59)

### [[Default Backend] SSL](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/ssl.go#L29)

- [should return a self generated SSL certificate](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/defaultbackend/ssl.go#L32)

### [[TCP] tcp-services](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/tcpudp/tcp.go#L37)

- [should expose a TCP service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/tcpudp/tcp.go#L40)
- [should expose an ExternalName TCP service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/tcpudp/tcp.go#L93)

### [auth-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L36)

- [should return status code 200 when no authentication is configured](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L46)
- [should return status code 503 when authentication is configured with an invalid secret](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L68)
- [should return status code 401 when authentication is configured but Authorization header is not configured](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L95)
- [should return status code 401 when authentication is configured and Authorization header is sent with invalid credentials](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L125)
- [should return status code 200 when authentication is configured and Authorization header is sent](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L156)
- [should return status code 200 when authentication is configured with a map and Authorization header is sent](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L186)
- [should return status code 401 when authentication is configured with invalid content and Authorization header is sent](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L217)
- [proxy_set_header My-Custom-Header 42;](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L259)
- [proxy_set_header My-Custom-Header 42;](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L277)
- [proxy_set_header 'My-Custom-Header' '42';](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L294)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L315)
- [retains cookie set by external authentication server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L335)
- [should return status code 200 when signed in](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L410)
- [should redirect to signin url when not signed in](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L424)
- [should return status code 200 when signed in after auth backend is deleted ](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L488)
- [should deny login for different location on same server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L517)
- [should deny login for different servers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L559)
- [should redirect to signin url when not signed in](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/auth.go#L603)

### [proxy-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L28)

- [should set proxy_redirect to off](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L36)
- [should set proxy_redirect to default](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L51)
- [should set proxy_redirect to hello.com goodbye.com](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L66)
- [should set proxy client-max-body-size to 8m](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L81)
- [should not set proxy client-max-body-size to incorrect value](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L95)
- [should set valid proxy timeouts](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L109)
- [should not set invalid proxy timeouts](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L127)
- [should turn on proxy-buffering](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L145)
- [should turn off proxy-request-buffering](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L164)
- [should build proxy next upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L178)
- [should build proxy next upstream using configmap values](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L196)
- [should setup proxy cookies](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L215)
- [should change the default proxy HTTP version](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxy.go#L231)

### [affinity session-cookie-name](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L36)

- [should set sticky cookie SERVERID](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L43)
- [should change cookie name on ingress definition change](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L68)
- [should set the path to /something on the generated cookie](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L105)
- [does not set the path to / on the generated cookie if there's more than one rule referring to the same backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L130)
- [should set cookie with expires](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L196)
- [should work with use-regex annotation and session-cookie-path](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L231)
- [should warn user when use-regex is true and session-cookie-path is not set](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L259)
- [should not set affinity across all server locations when using separate ingresses](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L288)
- [should set sticky cookie without host](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/affinity.go#L324)

### [mirror-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/mirror.go#L28)

- [should set mirror-target to http://localhost/mirror](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/mirror.go#L36)
- [should set mirror-target to https://test.env.com/$request_uri](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/mirror.go#L51)
- [should disable mirror-request-body](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/mirror.go#L67)

### [canary-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L34)

- [should response with a 200 status from the mainline upstream when requests are made to the mainline ingress](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L46)
- [should return 404 status for requests to the canary if no matching ingress is found](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L79)
- [should return the correct status codes when endpoints are unavailable](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L108)
- [should route requests to the correct upstream if mainline ingress is created before the canary ingress](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L162)
- [should route requests to the correct upstream if mainline ingress is created after the canary ingress](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L210)
- [should route requests to the correct upstream if the mainline ingress is modified](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L259)
- [should route requests to the correct upstream if the canary ingress is modified](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L321)
- [should route requests to the correct upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L380)
- [should route requests to the correct upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L444)
- [should route requests to the correct upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L522)
- [should route requests to the correct upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L562)
- [should route requests to the correct upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L626)
- [should not use canary as a catch-all server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L685)
- [should not use canary with domain as a server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L710)
- [does not crash when canary ingress has multiple paths to the same non-matching backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/canary.go#L732)

### [force-ssl-redirect](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/forcesslredirect.go#L29)

- [should redirect to https](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/forcesslredirect.go#L36)

### [http2-push-preload](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/http2pushpreload.go#L25)

- [enable the http2-push-preload directive](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/http2pushpreload.go#L32)

### [proxy-ssl-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxyssl.go#L28)

- [should set valid proxy-ssl-secret](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxyssl.go#L35)
- [should set valid proxy-ssl-secret, proxy-ssl-verify to on, and proxy-ssl-verify-depth to 2](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxyssl.go#L50)
- [should set valid proxy-ssl-secret, proxy-ssl-ciphers to HIGH:!AES](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxyssl.go#L67)
- [should set valid proxy-ssl-secret, proxy-ssl-protocols](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/proxyssl.go#L83)

### [modsecurity owasp](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/modsecurity.go#L26)

- [should enable modsecurity](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/modsecurity.go#L33)
- [should enable modsecurity with transaction ID and OWASP rules](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/modsecurity.go#L51)
- [should disable modsecurity](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/modsecurity.go#L72)
- [should enable modsecurity with snippet](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/modsecurity.go#L89)

### [backend-protocol - GRPC](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/grpc.go#L39)

- [should use grpc_pass in the configuration file](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/grpc.go#L42)
- [should return OK for service with backend protocol GRPC](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/grpc.go#L67)
- [should return OK for service with backend protocol GRPCS](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/grpc.go#L125)

### [cors-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L29)

- [should enable cors](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L36)
- [should set cors methods to only allow POST, GET](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L79)
- [should set cors max-age](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L95)
- [should disable cors allow credentials](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L111)
- [should allow origin for cors](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L127)
- [should allow headers for cors](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/cors.go#L143)

### [influxdb-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/influxdb.go#L39)

- [should send the request metric to the influxdb server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/influxdb.go#L48)

### [client-body-buffer-size](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L25)

- [should set client_body_buffer_size to 1000](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L32)
- [should set client_body_buffer_size to 1K](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L47)
- [should set client_body_buffer_size to 1k](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L62)
- [should set client_body_buffer_size to 1m](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L77)
- [should set client_body_buffer_size to 1M](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L92)
- [should not set client_body_buffer_size to invalid 1b](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/clientbodybuffersize.go#L107)

### [default-backend](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/default_backend.go#L30)

- [should use a custom default backend as upstream](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/default_backend.go#L38)

### [connection-proxy-header](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/connection.go#L30)

- [set connection header to keep-alive](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/connection.go#L37)

### [upstream-vhost](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/upstreamvhost.go#L25)

- [set host to upstreamvhost.bar.com](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/upstreamvhost.go#L32)

### [custom-http-errors](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/customhttperrors.go#L34)

- [configures Nginx correctly](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/customhttperrors.go#L41)

### [server-snippet](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/serversnippet.go#L27)

- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/serversnippet.go#L34)

### [rewrite-target use-regex enable-rewrite-log](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L32)

- [should write rewrite logs](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L39)
- [should use correct longest path match](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L70)
- [should use ~* location modifier if regex annotation is present](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L115)
- [should fail to use longest match for documented warning](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L161)
- [should allow for custom rewrite parameters](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/rewrite.go#L192)

### [app-root](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/approot.go#L29)

- [should redirect to /foo](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/approot.go#L36)

### [whitelist-source-range](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/ipwhitelist.go#L27)

- [should set valid ip whitelist range](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/ipwhitelist.go#L34)

### [enable-access-log enable-rewrite-log](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/log.go#L26)

- [set access_log off](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/log.go#L33)
- [set rewrite_log on](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/log.go#L48)

### [x-forwarded-prefix](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/xforwardedprefix.go#L28)

- [should set the X-Forwarded-Prefix to the annotation value](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/xforwardedprefix.go#L35)
- [should not add X-Forwarded-Prefix if the annotation value is empty](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/xforwardedprefix.go#L59)

### [configuration-snippet](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/snippet.go#L25)

- [ in all locations](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/snippet.go#L32)

### [backend-protocol - FastCGI](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fastcgi.go#L31)

- [should use fastcgi_pass in the configuration file](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fastcgi.go#L38)
- [should add fastcgi_index in the configuration file](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fastcgi.go#L55)
- [should add fastcgi_param in the configuration file](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fastcgi.go#L72)
- [should return OK for service with backend protocol FastCGI](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fastcgi.go#L105)

### [from-to-www-redirect](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fromtowwwredirect.go#L32)

- [should redirect from www HTTP to HTTP](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fromtowwwredirect.go#L39)
- [should redirect from www HTTPS to HTTPS](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/fromtowwwredirect.go#L70)

### [permanen-redirect permanen-redirect-code](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/redirect.go#L37)

- [should respond with a standard redirect code](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/redirect.go#L40)
- [should respond with a custom redirect code](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/redirect.go#L72)

### [upstream-hash-by-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/upstreamhashby.go#L75)

- [should connect to the same pod](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/upstreamhashby.go#L82)
- [should connect to the same subset of pods](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/upstreamhashby.go#L92)

### [backend-protocol](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L25)

- [should set backend protocol to https:// and use proxy_pass](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L32)
- [should set backend protocol to grpc:// and use grpc_pass](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L47)
- [should set backend protocol to grpcs:// and use grpc_pass](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L62)
- [should set backend protocol to '' and use fastcgi_pass](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L77)
- [should set backend protocol to '' and use ajp_pass](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/backendprotocol.go#L92)

### [satisfy](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/satisfy.go#L33)

- [should configure satisfy directive correctly](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/satisfy.go#L40)
- [should allow multiple auth with satisfy any](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/satisfy.go#L85)

### [server-alias](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/alias.go#L30)

- [should return status code 200 for host 'foo' and 404 for 'bar'](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/alias.go#L37)
- [should return status code 200 for host 'foo' and 'bar'](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/alias.go#L68)

### [ssl-ciphers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/sslciphers.go#L26)

- [should change ssl ciphers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/sslciphers.go#L33)

### [auth-tls-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/authtls.go#L31)

- [should set valid auth-tls-secret](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/authtls.go#L38)
- [should set valid auth-tls-secret, sslVerify to off, and sslVerifyDepth to 2](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/authtls.go#L78)
- [should set valid auth-tls-secret, pass certificate to upstream, and error page](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/annotations/authtls.go#L111)

### [[Status] status update](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/status/update.go#L37)

- [should update status field after client-go reconnection](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/status/update.go#L42)

### [Debug CLI](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/dbg/main.go#L29)

- [should list the backend servers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/dbg/main.go#L37)
- [should get information for a specific backend server](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/dbg/main.go#L57)
- [should produce valid JSON for /dbg general](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/dbg/main.go#L86)

### [[Memory Leak] Dynamic Certificates](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/leaks/lua_ssl.go#L36)

- [should not leak memory from ingress SSL certificates or configuration updates](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/leaks/lua_ssl.go#L43)

### [[Security] request smuggling](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/security/request_smuggling.go#L32)

- [should not return body content from error_page](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/security/request_smuggling.go#L39)

### [[SSL] [Flag] default-ssl-certificate](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/default_ssl_certificate.go#L32)

- [uses default ssl certificate for catch-all ingress](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/default_ssl_certificate.go#L64)
- [uses default ssl certificate for host based ingress when configured certificate does not match host](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/default_ssl_certificate.go#L80)

### [[Lua] lua-shared-dicts](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/lua_shared_dicts.go#L26)

- [configures lua shared dicts](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/lua_shared_dicts.go#L34)

### [server-tokens](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/server_tokens.go#L30)

- [should not exists Server header in the response](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/server_tokens.go#L38)
- [should exists Server header in the response when is enabled](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/server_tokens.go#L50)

### [use-proxy-protocol](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_protocol.go#L31)

- [should respect port passed by the PROXY Protocol](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_protocol.go#L41)
- [should respect proto passed by the PROXY Protocol server port](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_protocol.go#L73)

### [[Flag] custom HTTP and HTTPS ports](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/listen_nondefault_ports.go#L33)

- [should set X-Forwarded-Port headers accordingly when listening on a non-default HTTP port](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/listen_nondefault_ports.go#L49)
- [should set X-Forwarded-Port header to 443](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/listen_nondefault_ports.go#L72)
- [should set the X-Forwarded-Port header to 443](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/listen_nondefault_ports.go#L102)

### [[Security] no-auth-locations](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/no_auth_locations.go#L35)

- [should return status code 401 when accessing '/' unauthentication](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/no_auth_locations.go#L56)
- [should return status code 200 when accessing '/' authentication](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/no_auth_locations.go#L72)
- [should return status code 200 when accessing '/noauth' unauthenticated](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/no_auth_locations.go#L88)

### [Dynamic $proxy_host](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_host.go#L32)

- [should exist a proxy_host](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_host.go#L40)
- [should exist a proxy_host using the upstream-vhost annotation value](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/proxy_host.go#L64)

### [[Security] Pod Security Policies](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/pod_security_policy.go#L41)

- [should be running with a Pod Security Policy](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/pod_security_policy.go#L80)

### [Geoip2](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/geoip2.go#L30)

- [should only allow requests from specific countries](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/geoip2.go#L39)

### [[Security] Pod Security Policies with volumes](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/pod_security_policy_volumes.go#L37)

- [should be running with a Pod Security Policy](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/pod_security_policy_volumes.go#L40)

### [enable-multi-accept](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/multi_accept.go#L27)

- [should be enabled by default](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/multi_accept.go#L31)
- [should be enabled when set to true](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/multi_accept.go#L39)
- [should be disabled when set to false](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/multi_accept.go#L49)

### [[Flag] ingress-class](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/ingress_class.go#L33)

- [should ignore Ingress with class](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/ingress_class.go#L42)
- [should ignore Ingress with no class](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/ingress_class.go#L89)
- [should delete Ingress when class is removed](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/ingress_class.go#L125)

### [[Security] global-auth-url](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L31)

- [should return status code 401 when request any protected service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L82)
- [should return status code 200 when request whitelisted (via no-auth-locations) service and 401 when request protected service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L99)
- [should return status code 200 when request whitelisted (via ingress annotation) service and 401 when request protected service](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L119)
- [should still return status code 200 after auth backend is deleted using cache ](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L147)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L193)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L206)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L219)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L233)
- [](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_external_auth.go#L246)

### [[Security] block-*](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_access_block.go#L30)

- [should block CIDRs defined in the ConfigMap](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_access_block.go#L40)
- [should block User-Agents defined in the ConfigMap](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_access_block.go#L58)
- [should block Referers defined in the ConfigMap](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/global_access_block.go#L94)

### [use-forwarded-headers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/forwarded_headers.go#L31)

- [should trust X-Forwarded headers when setting is true](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/forwarded_headers.go#L41)
- [should not trust X-Forwarded headers when setting is false](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/forwarded_headers.go#L88)

### [Add custom headers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/custom_header.go#L31)

- [Add a custom header](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/custom_header.go#L44)
- [Add multiple custom headers](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/custom_header.go#L72)

### [[Flag] disable-catch-all](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/disable_catch_all.go#L34)

- [should ignore catch all Ingress](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/disable_catch_all.go#L52)
- [should delete Ingress updated to catch-all](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/disable_catch_all.go#L71)
- [should allow Ingress with both a default backend and rules](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/disable_catch_all.go#L111)

### [main-snippet](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/main_snippet.go#L27)

- [should add value of main-snippet setting to nginx config](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/main_snippet.go#L31)

### [[SSL] TLS protocols, ciphers and headers)](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/tls.go#L37)

- [should configure TLS protocol](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/tls.go#L46)
- [should configure HSTS policy header](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/tls.go#L102)
- [should not use ports during the HTTP to HTTPS redirection](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/tls.go#L168)
- [should not use ports or X-Forwarded-Host during the HTTP to HTTPS redirection](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/tls.go#L190)

### [Configmap change](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/configmap_change.go#L29)

- [should reload after an update in the configuration](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/configmap_change.go#L36)

### [[Security] modsecurity-snippet](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/modsecurity_snippet.go#L27)

- [should add value of modsecurity-snippet setting to nginx config](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/settings/modsecurity_snippet.go#L30)

### [[Shutdown] Graceful shutdown with pending request](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/slow_requests.go#L30)

- [should let slow requests finish before shutting down](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/slow_requests.go#L38)

### [[Shutdown] ingress controller](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/shutdown.go#L31)

- [should shutdown in less than 60 secons without pending connections](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/shutdown.go#L42)
- [should shutdown after waiting 60 seconds for pending connections to be closed](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/shutdown.go#L68)
- [should shutdown after waiting 150 seconds for pending connections to be closed](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/gracefulshutdown/shutdown.go#L127)

### [[Service] backend status code 503](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_backend.go#L35)

- [should return 503 when backend service does not exist](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_backend.go#L38)
- [should return 503 when all backend service endpoints are unavailable](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_backend.go#L57)

### [[Service] Type ExternalName](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L35)

- [works with external name set to incomplete fdqn](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L38)
- [should return 200 for service type=ExternalName without a port defined](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L72)
- [should return 200 for service type=ExternalName with a port defined](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L104)
- [should return status 502 for service type=ExternalName with an invalid host](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L143)
- [should return 200 for service type=ExternalName using a port name](https://github.com/kubernetes/ingress-nginx/tree/master/test/e2e/servicebackend/service_externalname.go#L175)
