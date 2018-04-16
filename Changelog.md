# Changelog

### 0.13.0

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.13.0`

*New Features:*

- NGINX 1.13.12
- Support for gRPC:
  - The annotation `nginx.ingress.kubernetes.io/grpc-backend: "true"` enable this feature
  - If the gRPC service requires TLS `nginx.ingress.kubernetes.io/secure-backends: "true"`
- Configurable load balancing with EWMA
- Support for [lua-resty-waf](https://github.com/p0pr0ck5/lua-resty-waf) as alternative to ModSecurity. [Check configuration guide](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/annotations.md#lua-resty-waf)
- Support for session affinity when dynamic configuration is enabled.
- Add NoAuthLocations and default it to "/.well-known/acme-challenge"

*Changes:*

- [X] [#2078](https://github.com/kubernetes/ingress-nginx/pull/2078) Expose SSL client cert data to external auth provider.
- [X] [#2187](https://github.com/kubernetes/ingress-nginx/pull/2187) Managing a whitelist for _/nginx_status
- [X] [#2208](https://github.com/kubernetes/ingress-nginx/pull/2208) Add missing lua bindata change
- [X] [#2209](https://github.com/kubernetes/ingress-nginx/pull/2209) fix go test TestSkipEnqueue error, move queue.Run
- [X] [#2210](https://github.com/kubernetes/ingress-nginx/pull/2210) allow ipv6 localhost when enabled
- [X] [#2212](https://github.com/kubernetes/ingress-nginx/pull/2212) Fix dynamic configuration when custom errors are enabled
- [X] [#2215](https://github.com/kubernetes/ingress-nginx/pull/2215) fix wrong config generation when upstream-hash-by is set
- [X] [#2220](https://github.com/kubernetes/ingress-nginx/pull/2220) fix: cannot set $service_name if use rewrite
- [X] [#2221](https://github.com/kubernetes/ingress-nginx/pull/2221) Update nginx to 1.13.10 and enable gRPC
- [X] [#2223](https://github.com/kubernetes/ingress-nginx/pull/2223) Add support for gRPC
- [X] [#2227](https://github.com/kubernetes/ingress-nginx/pull/2227) do not hardcode keepalive for upstream_balancer
- [X] [#2228](https://github.com/kubernetes/ingress-nginx/pull/2228) Fix broken links in multi-tls
- [X] [#2229](https://github.com/kubernetes/ingress-nginx/pull/2229) Configurable load balancing with EWMA
- [X] [#2232](https://github.com/kubernetes/ingress-nginx/pull/2232) Make proxy_next_upstream_tries configurable
- [X] [#2233](https://github.com/kubernetes/ingress-nginx/pull/2233) clean backends data before sending to Lua endpoint
- [X] [#2234](https://github.com/kubernetes/ingress-nginx/pull/2234) Update go dependencies
- [X] [#2235](https://github.com/kubernetes/ingress-nginx/pull/2235) add proxy header ssl-client-issuer-dn, fix #2178
- [X] [#2241](https://github.com/kubernetes/ingress-nginx/pull/2241) Revert "Get file max from fs/file-max. (#2050)"
- [X] [#2243](https://github.com/kubernetes/ingress-nginx/pull/2243) Add NoAuthLocations and default it to "/.well-known/acme-challenge"
- [X] [#2244](https://github.com/kubernetes/ingress-nginx/pull/2244) fix: empty ingress path
- [X] [#2246](https://github.com/kubernetes/ingress-nginx/pull/2246) Fix grpc json tag name
- [X] [#2254](https://github.com/kubernetes/ingress-nginx/pull/2254) e2e tests for dynamic configuration and Lua features and a bug fix
- [X] [#2263](https://github.com/kubernetes/ingress-nginx/pull/2263) clean up tmpl
- [X] [#2270](https://github.com/kubernetes/ingress-nginx/pull/2270) Revert deleted code in #2146
- [X] [#2271](https://github.com/kubernetes/ingress-nginx/pull/2271) Use SharedIndexInformers in place of Informers
- [X] [#2272](https://github.com/kubernetes/ingress-nginx/pull/2272) Disable opentracing for nginx internal urls
- [X] [#2273](https://github.com/kubernetes/ingress-nginx/pull/2273) Update go to 1.10.1
- [X] [#2280](https://github.com/kubernetes/ingress-nginx/pull/2280) Fix bug when auth req is enabled(external authentication)
- [X] [#2283](https://github.com/kubernetes/ingress-nginx/pull/2283) Fix flaky e2e tests
- [X] [#2285](https://github.com/kubernetes/ingress-nginx/pull/2285) Update controller.go
- [X] [#2290](https://github.com/kubernetes/ingress-nginx/pull/2290) Update nginx to 1.13.11
- [X] [#2294](https://github.com/kubernetes/ingress-nginx/pull/2294) Fix HSTS without preload
- [X] [#2296](https://github.com/kubernetes/ingress-nginx/pull/2296) Improve indentation of generated nginx.conf
- [X] [#2298](https://github.com/kubernetes/ingress-nginx/pull/2298) Disable dynamic configuration in s390x and ppc64le
- [X] [#2300](https://github.com/kubernetes/ingress-nginx/pull/2300) Fix race condition when Ingress does not contains a secret
- [X] [#2301](https://github.com/kubernetes/ingress-nginx/pull/2301) include lua-resty-waf and its dependencies in the base Nginx image
- [X] [#2303](https://github.com/kubernetes/ingress-nginx/pull/2303) More lua dependencies
- [X] [#2304](https://github.com/kubernetes/ingress-nginx/pull/2304) Lua resty waf controller
- [X] [#2305](https://github.com/kubernetes/ingress-nginx/pull/2305) Fix issues building nginx image in different platforms
- [X] [#2306](https://github.com/kubernetes/ingress-nginx/pull/2306) Disable lua waf where luajit is not available
- [X] [#2308](https://github.com/kubernetes/ingress-nginx/pull/2308) Add verification of lua load balancer to health check
- [X] [#2309](https://github.com/kubernetes/ingress-nginx/pull/2309) Configure upload limits for setup of lua load balancer
- [X] [#2314](https://github.com/kubernetes/ingress-nginx/pull/2314) annotation to ignore given list of WAF rulesets
- [X] [#2315](https://github.com/kubernetes/ingress-nginx/pull/2315) extra waf rules per ingress
- [X] [#2317](https://github.com/kubernetes/ingress-nginx/pull/2317) run lua-resty-waf in different modes
- [X] [#2327](https://github.com/kubernetes/ingress-nginx/pull/2327) Update nginx to 1.13.12
- [X] [#2328](https://github.com/kubernetes/ingress-nginx/pull/2328) Update nginx image
- [X] [#2331](https://github.com/kubernetes/ingress-nginx/pull/2331) fix nil pointer when ssl with ca.crt
- [X] [#2333](https://github.com/kubernetes/ingress-nginx/pull/2333) disable lua for arch s390x and ppc64le
- [X] [#2340](https://github.com/kubernetes/ingress-nginx/pull/2340) Fix buildupstream name to work with dynamic session affinity
- [X] [#2341](https://github.com/kubernetes/ingress-nginx/pull/2341) Add session affinity to custom load balancing
- [X] [#2342](https://github.com/kubernetes/ingress-nginx/pull/2342) Sync SSL certificates on events

*Documentation:*

- [X] [#2236](https://github.com/kubernetes/ingress-nginx/pull/2236) Add missing configuration in #2235
- [X] [#1785](https://github.com/kubernetes/ingress-nginx/pull/1785) Add deployment docs for AWS NLB
- [X] [#2213](https://github.com/kubernetes/ingress-nginx/pull/2213) Update cli-arguments.md
- [X] [#2219](https://github.com/kubernetes/ingress-nginx/pull/2219) Fix log format documentation
- [X] [#2238](https://github.com/kubernetes/ingress-nginx/pull/2238) Correct typo
- [X] [#2239](https://github.com/kubernetes/ingress-nginx/pull/2239) fix-link
- [X] [#2240](https://github.com/kubernetes/ingress-nginx/pull/2240) fix:"any value other" should be "any other value"
- [X] [#2255](https://github.com/kubernetes/ingress-nginx/pull/2255) Update annotations.md
- [X] [#2267](https://github.com/kubernetes/ingress-nginx/pull/2267) Update README.md
- [X] [#2274](https://github.com/kubernetes/ingress-nginx/pull/2274) Typo fixes in modsecurity.md
- [X] [#2276](https://github.com/kubernetes/ingress-nginx/pull/2276) Update README.md
- [X] [#2282](https://github.com/kubernetes/ingress-nginx/pull/2282) Fix nlb instructions

### 0.12.0

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.12.0`

*New Features:*

- Live NGINX configuration update without reloading using the flag `--enable-dynamic-configuration` (disabled by default).
- New flag `--publish-status-address` to manually set the Ingress status IP address.
- Add worker-cpu-affinity NGINX option.
- Enable remote logging using syslog.
- Do not redirect `/.well-known/acme-challenge` to HTTPS.

*Changes:*

- [X] [#2125](https://github.com/kubernetes/ingress-nginx/pull/2125) Add GCB config to build defaultbackend
- [X] [#2127](https://github.com/kubernetes/ingress-nginx/pull/2127) Revert deletion of dependency version override
- [X] [#2137](https://github.com/kubernetes/ingress-nginx/pull/2137) Updated log level to v2 for sysctlFSFileMax.
- [X] [#2140](https://github.com/kubernetes/ingress-nginx/pull/2140) Cors header should always be returned
- [X] [#2141](https://github.com/kubernetes/ingress-nginx/pull/2141) Fix error loading modules
- [X] [#2143](https://github.com/kubernetes/ingress-nginx/pull/2143) Only add HSTS headers in HTTPS
- [X] [#2144](https://github.com/kubernetes/ingress-nginx/pull/2144) Add annotation to disable logs in a location
- [X] [#2145](https://github.com/kubernetes/ingress-nginx/pull/2145) Add option in the configuration configmap to enable remote logging
- [X] [#2146](https://github.com/kubernetes/ingress-nginx/pull/2146) In case of TLS errors do not allow traffic
- [X] [#2148](https://github.com/kubernetes/ingress-nginx/pull/2148) Add publish-status-address flag
- [X] [#2155](https://github.com/kubernetes/ingress-nginx/pull/2155) Update nginx with new modules
- [X] [#2162](https://github.com/kubernetes/ingress-nginx/pull/2162) Remove duplicated BuildConfigFromFlags func
- [X] [#2163](https://github.com/kubernetes/ingress-nginx/pull/2163) include lua-upstream-nginx-module in Nginx build
- [X] [#2164](https://github.com/kubernetes/ingress-nginx/pull/2164) use the correct error channel
- [X] [#2167](https://github.com/kubernetes/ingress-nginx/pull/2167) configuring load balancing per ingress
- [X] [#2172](https://github.com/kubernetes/ingress-nginx/pull/2172) include lua-resty-lock in nginx image
- [X] [#2174](https://github.com/kubernetes/ingress-nginx/pull/2174) Live Nginx configuration update without reloading
- [X] [#2180](https://github.com/kubernetes/ingress-nginx/pull/2180) Include tests in golint checks, fix warnings
- [X] [#2181](https://github.com/kubernetes/ingress-nginx/pull/2181) change nginx process pgid
- [X] [#2185](https://github.com/kubernetes/ingress-nginx/pull/2185) Remove ProxyPassParams setting
- [X] [#2191](https://github.com/kubernetes/ingress-nginx/pull/2191) Add checker test for bad pid
- [X] [#2193](https://github.com/kubernetes/ingress-nginx/pull/2193) fix wrong json tag
- [X] [#2201](https://github.com/kubernetes/ingress-nginx/pull/2201) Add worker-cpu-affinity nginx option
- [X] [#2202](https://github.com/kubernetes/ingress-nginx/pull/2202) Allow config to disable geoip
- [X] [#2205](https://github.com/kubernetes/ingress-nginx/pull/2205) add luacheck to lint lua files

*Documentation:*

- [X] [#2124](https://github.com/kubernetes/ingress-nginx/pull/2124) Document how to provide list types in configmap
- [X] [#2133](https://github.com/kubernetes/ingress-nginx/pull/2133) fix limit-req-status-code doc
- [X] [#2139](https://github.com/kubernetes/ingress-nginx/pull/2139) Update documentation for nginx-ingress-role RBAC.
- [X] [#2165](https://github.com/kubernetes/ingress-nginx/pull/2165) Typo fix "api server " -> "API server"
- [X] [#2169](https://github.com/kubernetes/ingress-nginx/pull/2169) Add documentation about secure-verify-ca-secret
- [X] [#2200](https://github.com/kubernetes/ingress-nginx/pull/2200) fix grammer mistake

### 0.11.0

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.11.0`

*New Features:*

- NGINX 1.13.9

*Changes:*

- [X] [#1992](https://github.com/kubernetes/ingress-nginx/pull/1992) Added configmap option to disable IPv6 in nginx DNS resolver
- [X] [#1993](https://github.com/kubernetes/ingress-nginx/pull/1993) Enable Customization of Auth Request Redirect
- [X] [#1996](https://github.com/kubernetes/ingress-nginx/pull/1996) Use v3/dev/performance of ModSecurity because of performance
- [X] [#1997](https://github.com/kubernetes/ingress-nginx/pull/1997) fix var checked
- [X] [#1998](https://github.com/kubernetes/ingress-nginx/pull/1998) Add support to enable/disable proxy buffering
- [X] [#1999](https://github.com/kubernetes/ingress-nginx/pull/1999) Add connection-proxy-header annotation
- [X] [#2001](https://github.com/kubernetes/ingress-nginx/pull/2001) Add limit-request-status-code option
- [X] [#2005](https://github.com/kubernetes/ingress-nginx/pull/2005) fix typo error for server name _
- [X] [#2006](https://github.com/kubernetes/ingress-nginx/pull/2006) Add support for enabling ssl_ciphers per host
- [X] [#2019](https://github.com/kubernetes/ingress-nginx/pull/2019) Update nginx image
- [X] [#2021](https://github.com/kubernetes/ingress-nginx/pull/2021) Add nginx_cookie_flag_module
- [X] [#2026](https://github.com/kubernetes/ingress-nginx/pull/2026) update KUBERNETES from v1.8.0 to 1.9.0
- [X] [#2027](https://github.com/kubernetes/ingress-nginx/pull/2027) Show pod information in http-svc example
- [X] [#2030](https://github.com/kubernetes/ingress-nginx/pull/2030) do not ignore $http_host and $http_x_forwarded_host
- [X] [#2031](https://github.com/kubernetes/ingress-nginx/pull/2031) The maximum number of open file descriptors should be maxOpenFiles.
- [X] [#2036](https://github.com/kubernetes/ingress-nginx/pull/2036) add matchLabels in Deployment yaml, that both API extensions/v1beta1 …
- [X] [#2050](https://github.com/kubernetes/ingress-nginx/pull/2050) Get file max from fs/file-max.
- [X] [#2063](https://github.com/kubernetes/ingress-nginx/pull/2063) Run one test at a time
- [X] [#2065](https://github.com/kubernetes/ingress-nginx/pull/2065) Always return an IP address
- [X] [#2069](https://github.com/kubernetes/ingress-nginx/pull/2069) Do not cancel the synchronization of secrets
- [X] [#2071](https://github.com/kubernetes/ingress-nginx/pull/2071) Update Go to 1.9.4
- [X] [#2082](https://github.com/kubernetes/ingress-nginx/pull/2082) Use a ring channel to avoid blocking write of events
- [X] [#2089](https://github.com/kubernetes/ingress-nginx/pull/2089) Retry initial connection to the Kubernetes cluster
- [X] [#2093](https://github.com/kubernetes/ingress-nginx/pull/2093) Only pods in running phase are vallid for status
- [X] [#2099](https://github.com/kubernetes/ingress-nginx/pull/2099) Added GeoIP Organisational data
- [X] [#2107](https://github.com/kubernetes/ingress-nginx/pull/2107) Enabled the dynamic reload of GeoIP data
- [X] [#2119](https://github.com/kubernetes/ingress-nginx/pull/2119) Remove deprecated flag disable-node-list
- [X] [#2120](https://github.com/kubernetes/ingress-nginx/pull/2120) Migrate to codecov.io

*Documentation:*

- [X] [#1987](https://github.com/kubernetes/ingress-nginx/pull/1987) add kube-system namespace for oauth2-proxy example
- [X] [#1991](https://github.com/kubernetes/ingress-nginx/pull/1991) Add comment about bolean and number values
- [X] [#2009](https://github.com/kubernetes/ingress-nginx/pull/2009) docs/user-guide/tls: remove duplicated section
- [X] [#2011](https://github.com/kubernetes/ingress-nginx/pull/2011) broken link for sticky-ingress.yaml
- [X] [#2014](https://github.com/kubernetes/ingress-nginx/pull/2014) Add document for connection-proxy-header annotation
- [X] [#2016](https://github.com/kubernetes/ingress-nginx/pull/2016) Minor link fix in deployment docs
- [X] [#2018](https://github.com/kubernetes/ingress-nginx/pull/2018) Added documentation for Permanent Redirect
- [X] [#2035](https://github.com/kubernetes/ingress-nginx/pull/2035) fix broken links in static-ip readme
- [X] [#2038](https://github.com/kubernetes/ingress-nginx/pull/2038) fix typo: appropiate -> [appropriate]
- [X] [#2039](https://github.com/kubernetes/ingress-nginx/pull/2039) fix typo stickyness to stickiness
- [X] [#2040](https://github.com/kubernetes/ingress-nginx/pull/2040) fix wrong annotation
- [X] [#2041](https://github.com/kubernetes/ingress-nginx/pull/2041) fix spell error reslover -> resolver
- [X] [#2046](https://github.com/kubernetes/ingress-nginx/pull/2046) Fix typos
- [X] [#2054](https://github.com/kubernetes/ingress-nginx/pull/2054) Adding documentation for helm with RBAC enabled
- [X] [#2075](https://github.com/kubernetes/ingress-nginx/pull/2075) Fix opentracing configuration when multiple options are configured
- [X] [#2076](https://github.com/kubernetes/ingress-nginx/pull/2076) Fix spelling errors
- [X] [#2077](https://github.com/kubernetes/ingress-nginx/pull/2077) Remove initContainer from default deployment

### 0.10.2

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.10.2`

*Changes:*

- [X] [#1978](https://github.com/kubernetes/ingress-nginx/pull/1978) Fix chain completion and default certificate flag issues

### 0.10.1

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.10.1`

*Changes:*

- [X] [#1945](https://github.com/kubernetes/ingress-nginx/pull/1945) When a secret is updated read ingress annotations (again)
- [X] [#1948](https://github.com/kubernetes/ingress-nginx/pull/1948) Update go to 1.9.3
- [X] [#1953](https://github.com/kubernetes/ingress-nginx/pull/1953) Added annotation for upstream-vhost
- [X] [#1960](https://github.com/kubernetes/ingress-nginx/pull/1960) Adjust sysctl values to improve nginx performance
- [X] [#1963](https://github.com/kubernetes/ingress-nginx/pull/1963) Fix tests
- [X] [#1969](https://github.com/kubernetes/ingress-nginx/pull/1969) Rollback #1854
- [X] [#1970](https://github.com/kubernetes/ingress-nginx/pull/1970) By default brotli is disabled

### 0.10.0

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.10.0`

*Breaking changes:*

Changed the names of default Nginx ingress prometheus metrics.
If you are scraping default Nginx ingress metrics with prometheus the metrics changes are as follows:

```
nginx_active_connections_total          -> nginx_connections_total{state="active"}
nginx_accepted_connections_total        -> nginx_connections_total{state="accepted"}
nginx_handled_connections_total         -> nginx_connections_total{state="handled"}
nginx_current_reading_connections_total -> nginx_connections{state="reading"}
nginx_current_writing_connections_total -> nginx_connections{state="writing"}
current_waiting_connections_total       -> nginx_connections{state="waiting"}
```

*New Features:*

- NGINX 1.13.8
- Support to hide headers from upstream servers
- Support for Jaeger
- CORS max age annotation

*Changes:*

- [X] [#1782](https://github.com/kubernetes/ingress-nginx/pull/1782) auth-tls-pass-certificate-to-upstream should be bool
- [X] [#1787](https://github.com/kubernetes/ingress-nginx/pull/1787) force external_auth requests to http/1.1
- [X] [#1800](https://github.com/kubernetes/ingress-nginx/pull/1800) Add control of the configuration refresh interval
- [X] [#1805](https://github.com/kubernetes/ingress-nginx/pull/1805) Add X-Forwarded-Prefix on rewrites
- [X] [#1844](https://github.com/kubernetes/ingress-nginx/pull/1844) Validate x-forwarded-proto and connection scheme before redirect to https
- [X] [#1852](https://github.com/kubernetes/ingress-nginx/pull/1852) Update nginx to v1.13.8 and update modules
- [X] [#1854](https://github.com/kubernetes/ingress-nginx/pull/1854) Fix redirect to ssl
- [X] [#1858](https://github.com/kubernetes/ingress-nginx/pull/1858) When upstream-hash-by annotation is used do not configure a lb algorithm
- [X] [#1861](https://github.com/kubernetes/ingress-nginx/pull/1861) Improve speed of tests execution
- [X] [#1869](https://github.com/kubernetes/ingress-nginx/pull/1869) "proxy_redirect default" should be placed after the "proxy_pass"
- [X] [#1870](https://github.com/kubernetes/ingress-nginx/pull/1870) Fix SSL Passthrough template issue and custom ports in redirect to HTTPS
- [X] [#1871](https://github.com/kubernetes/ingress-nginx/pull/1871) Update nginx image to 0.31
- [X] [#1872](https://github.com/kubernetes/ingress-nginx/pull/1872) Fix data race updating ingress status
- [X] [#1880](https://github.com/kubernetes/ingress-nginx/pull/1880) Update go dependencies and cleanup deprecated packages
- [X] [#1888](https://github.com/kubernetes/ingress-nginx/pull/1888) Add CORS max age annotation
- [X] [#1891](https://github.com/kubernetes/ingress-nginx/pull/1891) Refactor initial synchronization of ingress objects
- [X] [#1903](https://github.com/kubernetes/ingress-nginx/pull/1903) If server_tokens is disabled remove the Server header
- [X] [#1906](https://github.com/kubernetes/ingress-nginx/pull/1906) Random string function should only contains letters
- [X] [#1907](https://github.com/kubernetes/ingress-nginx/pull/1907) Fix custom port in redirects
- [X] [#1909](https://github.com/kubernetes/ingress-nginx/pull/1909) Release nginx 0.32
- [X] [#1910](https://github.com/kubernetes/ingress-nginx/pull/1910) updating prometheus metrics names according to naming best practices
- [X] [#1912](https://github.com/kubernetes/ingress-nginx/pull/1912) removing _total prefix from nginx guage metrics
- [X] [#1914](https://github.com/kubernetes/ingress-nginx/pull/1914) Add --with-http_secure_link_module for the Nginx build configuration
- [X] [#1916](https://github.com/kubernetes/ingress-nginx/pull/1916) Add support for jaeger backend
- [X] [#1918](https://github.com/kubernetes/ingress-nginx/pull/1918) Update nginx image to 0.32
- [X] [#1919](https://github.com/kubernetes/ingress-nginx/pull/1919) Add option for reuseport in nginx listen section
- [X] [#1926](https://github.com/kubernetes/ingress-nginx/pull/1926) Do not use port from host header
- [X] [#1927](https://github.com/kubernetes/ingress-nginx/pull/1927) Remove sendfile configuration
- [X] [#1928](https://github.com/kubernetes/ingress-nginx/pull/1928) Add support to hide headers from upstream servers
- [X] [#1929](https://github.com/kubernetes/ingress-nginx/pull/1929) Refactoring of kubernetes informers and local caches
- [X] [#1933](https://github.com/kubernetes/ingress-nginx/pull/1933) Remove deploy of ingress controller from the example

*Documentation:*

- [X] [#1786](https://github.com/kubernetes/ingress-nginx/pull/1786) fix: some typo.
- [X] [#1792](https://github.com/kubernetes/ingress-nginx/pull/1792) Add note about annotation values
- [X] [#1814](https://github.com/kubernetes/ingress-nginx/pull/1814) Fix link to custom configuration
- [X] [#1826](https://github.com/kubernetes/ingress-nginx/pull/1826) Add note about websocket and load balancers
- [X] [#1840](https://github.com/kubernetes/ingress-nginx/pull/1840) Add note about default log files
- [X] [#1853](https://github.com/kubernetes/ingress-nginx/pull/1853) Clarify docs for add-headers and proxy-set-headers
- [X] [#1864](https://github.com/kubernetes/ingress-nginx/pull/1864) configmap.md: Convert hyphens in name column to non-breaking-hyphens
- [X] [#1865](https://github.com/kubernetes/ingress-nginx/pull/1865) Add docs for legacy TLS version and ciphers
- [X] [#1867](https://github.com/kubernetes/ingress-nginx/pull/1867) Fix publish-service patch and update README
- [X] [#1913](https://github.com/kubernetes/ingress-nginx/pull/1913) Missing r
- [X] [#1925](https://github.com/kubernetes/ingress-nginx/pull/1925) Fix doc links


### 0.9.0

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0`

*Changes:*

- [X] [#1731](https://github.com/kubernetes/ingress-nginx/pull/1731) Allow configuration of proxy_responses value for tcp/udp configmaps
- [X] [#1766](https://github.com/kubernetes/ingress-nginx/pull/1766) Fix ingress typo
- [X] [#1768](https://github.com/kubernetes/ingress-nginx/pull/1768) Custom default backend must use annotations if present
- [X] [#1769](https://github.com/kubernetes/ingress-nginx/pull/1769) Use custom https port in redirects
- [X] [#1771](https://github.com/kubernetes/ingress-nginx/pull/1771) Add additional check for old SSL certificates
- [X] [#1776](https://github.com/kubernetes/ingress-nginx/pull/1776) Add option to configure the redirect code

### 0.9-beta.19

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0-beta.19`

*Changes:*

- Fix regression with ingress.class annotation introduced in 0.9-beta.18

### 0.9-beta.18

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0-beta.18`

*Breaking changes:*

- The NGINX ingress annotations contains a new prefix: **nginx.ingress.kubernetes.io**. This change is behind a flag to avoid breaking running deployments.
  To avoid breaking a running NGINX ingress controller add the flag **--annotations-prefix=ingress.kubernetes.io** to the nginx ingress controller deployment.
  There is one exception, the annotation `kubernetes.io/ingress.class` remains unchanged (this annotation is used in multiple ingress controllers)

*New Features:*

- NGINX 1.13.7
- Support for s390x
- e2e tests

*Changes:*

- [X] [#1648](https://github.com/kubernetes/ingress-nginx/pull/1648) Remove GenericController and add tests
- [X] [#1650](https://github.com/kubernetes/ingress-nginx/pull/1650) Fix misspell errors
- [X] [#1651](https://github.com/kubernetes/ingress-nginx/pull/1651) Remove node lister
- [X] [#1652](https://github.com/kubernetes/ingress-nginx/pull/1652) Remove node lister
- [X] [#1653](https://github.com/kubernetes/ingress-nginx/pull/1653) Fix diff execution
- [X] [#1654](https://github.com/kubernetes/ingress-nginx/pull/1654) Fix travis script and update kubernetes to 1.8.0
- [X] [#1658](https://github.com/kubernetes/ingress-nginx/pull/1658) Tests
- [X] [#1659](https://github.com/kubernetes/ingress-nginx/pull/1659) Add nginx helper tests
- [X] [#1662](https://github.com/kubernetes/ingress-nginx/pull/1662) Refactor annotations
- [X] [#1665](https://github.com/kubernetes/ingress-nginx/pull/1665) Add the original http request method to the auth request
- [X] [#1687](https://github.com/kubernetes/ingress-nginx/pull/1687) Fix use merge of annotations
- [X] [#1689](https://github.com/kubernetes/ingress-nginx/pull/1689) Enable s390x
- [X] [#1693](https://github.com/kubernetes/ingress-nginx/pull/1693) Fix docker build
- [X] [#1695](https://github.com/kubernetes/ingress-nginx/pull/1695) Update nginx to v0.29
- [X] [#1696](https://github.com/kubernetes/ingress-nginx/pull/1696) Always add cors headers when enabled
- [X] [#1697](https://github.com/kubernetes/ingress-nginx/pull/1697) Disable features not availables in some platforms
- [X] [#1698](https://github.com/kubernetes/ingress-nginx/pull/1698) Auth e2e tests
- [X] [#1699](https://github.com/kubernetes/ingress-nginx/pull/1699) Refactor SSL intermediate CA certificate check
- [X] [#1700](https://github.com/kubernetes/ingress-nginx/pull/1700) Add patch command to append publish-service flag
- [X] [#1701](https://github.com/kubernetes/ingress-nginx/pull/1701) fix: Core() is deprecated use CoreV1() instead.
- [X] [#1702](https://github.com/kubernetes/ingress-nginx/pull/1702) Fix TLS example [ci skip]
- [X] [#1704](https://github.com/kubernetes/ingress-nginx/pull/1704) Add e2e tests to verify the correct source IP address
- [X] [#1705](https://github.com/kubernetes/ingress-nginx/pull/1705) Add annotation for setting proxy_redirect
- [X] [#1706](https://github.com/kubernetes/ingress-nginx/pull/1706) Increase ELB idle timeouts [ci skip]
- [X] [#1710](https://github.com/kubernetes/ingress-nginx/pull/1710) Do not update a secret not referenced by ingress rules
- [X] [#1713](https://github.com/kubernetes/ingress-nginx/pull/1713) add --report-node-internal-ip-address describe to cli-arguments.md
- [X] [#1717](https://github.com/kubernetes/ingress-nginx/pull/1717) Fix command used to detect version
- [X] [#1720](https://github.com/kubernetes/ingress-nginx/pull/1720) Add docker-registry example [ci skip]
- [X] [#1722](https://github.com/kubernetes/ingress-nginx/pull/1722) Add annotation to enable passing the certificate to the upstream server
- [X] [#1723](https://github.com/kubernetes/ingress-nginx/pull/1723) Add timeouts to http server and additional pprof routes
- [X] [#1724](https://github.com/kubernetes/ingress-nginx/pull/1724) Cleanup main
- [X] [#1725](https://github.com/kubernetes/ingress-nginx/pull/1725) Enable all e2e tests
- [X] [#1726](https://github.com/kubernetes/ingress-nginx/pull/1726) fix: replace deprecated methods.
- [X] [#1734](https://github.com/kubernetes/ingress-nginx/pull/1734) Changes ssl-client-cert header
- [X] [#1737](https://github.com/kubernetes/ingress-nginx/pull/1737) Update nginx v1.13.7
- [X] [#1738](https://github.com/kubernetes/ingress-nginx/pull/1738) Cleanup
- [X] [#1739](https://github.com/kubernetes/ingress-nginx/pull/1739) Improve e2e checks
- [X] [#1740](https://github.com/kubernetes/ingress-nginx/pull/1740) Update nginx
- [X] [#1745](https://github.com/kubernetes/ingress-nginx/pull/1745) Simplify annotations
- [X] [#1746](https://github.com/kubernetes/ingress-nginx/pull/1746) Cleanup of e2e helpers

*Documentation:*

- [X] [#1657](https://github.com/kubernetes/ingress-nginx/pull/1657) Add better documentation for deploying for dev
- [X] [#1680](https://github.com/kubernetes/ingress-nginx/pull/1680) Add doc for log-format-escape-json [ci skip]
- [X] [#1685](https://github.com/kubernetes/ingress-nginx/pull/1685) Fix default SSL certificate flag docs [ci skip]
- [X] [#1686](https://github.com/kubernetes/ingress-nginx/pull/1686) Fix development doc [ci skip]
- [X] [#1727](https://github.com/kubernetes/ingress-nginx/pull/1727) fix: fix typos in docs.
- [X] [#1747](https://github.com/kubernetes/ingress-nginx/pull/1747) Add config-map usage and options to Documentation


### 0.9-beta.17

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0-beta.17`

*Changes:*

- Fix regression with annotations introduced in 0.9-beta.16 (thanks @tomlanyon)

### 0.9-beta.16

**Image:**  `quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0-beta.16`

*New Features:*

- Images are published to [quay.io](https://quay.io/repository/kubernetes-ingress-controller)
- NGINX 1.13.6
- OpenTracing Jaeger support inNGINX
- [ModSecurity support](https://github.com/SpiderLabs/ModSecurity-nginx)
- Support for [brotli compression in NGINX](https://certsimple.com/blog/nginx-brotli)
- Return 503 error instead of 404 when no endpoint is available

*Breaking changes:*

- The default SSL configuration was updated to use `TLSv1.2` and the default cipher list is `ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256`

*Known issues:*

- When ModSecurity is enabled a segfault could occur - [ModSecurity#1590](https://github.com/SpiderLabs/ModSecurity/issues/1590)

*Changes:*

- [X] [#1489](https://github.com/kubernetes/ingress-nginx/pull/1489) Compute a real `X-Forwarded-For` header
- [X] [#1490](https://github.com/kubernetes/ingress-nginx/pull/1490) Introduce an upstream-hash-by annotation to support consistent hashing by nginx variable or text
- [X] [#1498](https://github.com/kubernetes/ingress-nginx/pull/1498) Add modsecurity module
- [X] [#1500](https://github.com/kubernetes/ingress-nginx/pull/1500) Enable modsecurity feature
- [X] [#1501](https://github.com/kubernetes/ingress-nginx/pull/1501) Request ingress controller version in issue template
- [X] [#1502](https://github.com/kubernetes/ingress-nginx/pull/1502) Force reload on template change
- [X] [#1503](https://github.com/kubernetes/ingress-nginx/pull/1503) Add falg to report node internal IP address in ingress status
- [X] [#1505](https://github.com/kubernetes/ingress-nginx/pull/1505) Increase size of variable hash bucket
- [X] [#1506](https://github.com/kubernetes/ingress-nginx/pull/1506) Update nginx ssl configuration
- [X] [#1507](https://github.com/kubernetes/ingress-nginx/pull/1507) Add tls session ticket key setting
- [X] [#1511](https://github.com/kubernetes/ingress-nginx/pull/1511) fix deprecated ssl_client_cert. add ssl_client_verify header
- [X] [#1513](https://github.com/kubernetes/ingress-nginx/pull/1513) Return 503 by default when no endpoint is available
- [X] [#1520](https://github.com/kubernetes/ingress-nginx/pull/1520) Change alias behaviour not to create new server section needlessly
- [X] [#1523](https://github.com/kubernetes/ingress-nginx/pull/1523) Include the serversnippet from the config map in server blocks
- [X] [#1533](https://github.com/kubernetes/ingress-nginx/pull/1533) Remove authentication send body annotation
- [X] [#1535](https://github.com/kubernetes/ingress-nginx/pull/1535) Remove auth-send-body [ci skip]
- [X] [#1538](https://github.com/kubernetes/ingress-nginx/pull/1538) Rename service-nodeport.yml to service-nodeport.yaml
- [X] [#1543](https://github.com/kubernetes/ingress-nginx/pull/1543) Fix glog initialization error
- [X] [#1544](https://github.com/kubernetes/ingress-nginx/pull/1544) Fix `make container` for OSX.
- [X] [#1547](https://github.com/kubernetes/ingress-nginx/pull/1547) fix broken GCE-GKE service descriptor
- [X] [#1550](https://github.com/kubernetes/ingress-nginx/pull/1550) Add e2e tests - default backend
- [X] [#1553](https://github.com/kubernetes/ingress-nginx/pull/1553) Cors features improvements
- [X] [#1554](https://github.com/kubernetes/ingress-nginx/pull/1554) Add missing unit test for nextPowerOf2 function
- [X] [#1556](https://github.com/kubernetes/ingress-nginx/pull/1556) fixed https port forwarding in Azure LB service
- [X] [#1566](https://github.com/kubernetes/ingress-nginx/pull/1566) Release nginx-slim 0.27
- [X] [#1568](https://github.com/kubernetes/ingress-nginx/pull/1568) update defaultbackend tag
- [X] [#1569](https://github.com/kubernetes/ingress-nginx/pull/1569) Update 404 server image
- [X] [#1570](https://github.com/kubernetes/ingress-nginx/pull/1570) Update nginx version
- [X] [#1571](https://github.com/kubernetes/ingress-nginx/pull/1571) Fix cors tests
- [X] [#1572](https://github.com/kubernetes/ingress-nginx/pull/1572) Certificate Auth Bugfix
- [X] [#1577](https://github.com/kubernetes/ingress-nginx/pull/1577) Do not use relative urls for yaml files
- [X] [#1580](https://github.com/kubernetes/ingress-nginx/pull/1580) Upgrade to use the latest version of nginx-opentracing.
- [X] [#1581](https://github.com/kubernetes/ingress-nginx/pull/1581) Fix Makefile to work in OSX.
- [X] [#1582](https://github.com/kubernetes/ingress-nginx/pull/1582) Add scripts to release from travis-ci
- [X] [#1584](https://github.com/kubernetes/ingress-nginx/pull/1584) Add missing probes in deployments
- [X] [#1585](https://github.com/kubernetes/ingress-nginx/pull/1585) Add version flag
- [X] [#1587](https://github.com/kubernetes/ingress-nginx/pull/1587) Use pass access scheme in signin url
- [X] [#1589](https://github.com/kubernetes/ingress-nginx/pull/1589) Fix upstream vhost Equal comparison
- [X] [#1590](https://github.com/kubernetes/ingress-nginx/pull/1590) Fix Equals Comparison for CORS annotation
- [X] [#1592](https://github.com/kubernetes/ingress-nginx/pull/1592) Update opentracing module and release image to  quay.io
- [X] [#1593](https://github.com/kubernetes/ingress-nginx/pull/1593) Fix makefile default task
- [X] [#1605](https://github.com/kubernetes/ingress-nginx/pull/1605) Fix ExternalName services
- [X] [#1607](https://github.com/kubernetes/ingress-nginx/pull/1607) Add support for named ports with service-upstream. #1459
- [X] [#1608](https://github.com/kubernetes/ingress-nginx/pull/1608) Fix issue with clusterIP detection on service upstream. #1534
- [X] [#1610](https://github.com/kubernetes/ingress-nginx/pull/1610) Only set alias if not already set
- [X] [#1618](https://github.com/kubernetes/ingress-nginx/pull/1618) Fix full XFF with PROXY
- [X] [#1620](https://github.com/kubernetes/ingress-nginx/pull/1620) Add gzip_vary
- [X] [#1621](https://github.com/kubernetes/ingress-nginx/pull/1621) Fix path to ELB listener image
- [X] [#1627](https://github.com/kubernetes/ingress-nginx/pull/1627) Add brotli support
- [X] [#1629](https://github.com/kubernetes/ingress-nginx/pull/1629) Add ssl-client-dn header
- [X] [#1632](https://github.com/kubernetes/ingress-nginx/pull/1632) Rename OWNERS assignees: to approvers:
- [X] [#1635](https://github.com/kubernetes/ingress-nginx/pull/1635) Install dumb-init using apt-get
- [X] [#1636](https://github.com/kubernetes/ingress-nginx/pull/1636) Update go to 1.9.2
- [X] [#1640](https://github.com/kubernetes/ingress-nginx/pull/1640) Update nginx to 0.28 and enable brotli

*Documentation:*

- [X] [#1491](https://github.com/kubernetes/ingress-nginx/pull/1491) Note that GCE has moved to a new repo
- [X] [#1492](https://github.com/kubernetes/ingress-nginx/pull/1492) Cleanup readme.md
- [X] [#1494](https://github.com/kubernetes/ingress-nginx/pull/1494) Cleanup
- [X] [#1497](https://github.com/kubernetes/ingress-nginx/pull/1497) Cleanup examples directory
- [X] [#1504](https://github.com/kubernetes/ingress-nginx/pull/1504) Clean readme
- [X] [#1508](https://github.com/kubernetes/ingress-nginx/pull/1508) Fixed link in prometheus example
- [X] [#1527](https://github.com/kubernetes/ingress-nginx/pull/1527) Split documentation
- [X] [#1536](https://github.com/kubernetes/ingress-nginx/pull/1536) Update documentation and examples [ci skip]
- [X] [#1541](https://github.com/kubernetes/ingress-nginx/pull/1541) fix(documentation): Fix some typos
- [X] [#1548](https://github.com/kubernetes/ingress-nginx/pull/1548) link to prometheus docs
- [X] [#1562](https://github.com/kubernetes/ingress-nginx/pull/1562) Fix development guide link
- [X] [#1563](https://github.com/kubernetes/ingress-nginx/pull/1563) Add task to verify markdown links
- [X] [#1583](https://github.com/kubernetes/ingress-nginx/pull/1583) Add note for certificate authentication in Cloudflare
- [X] [#1617](https://github.com/kubernetes/ingress-nginx/pull/1617) fix typo in user-guide/annotations.md

### 0.9-beta.15

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.15`

*New Features:*

- Add OCSP support
- Configurable ssl_verify_client

*Changes:*

- [X] [#1468](https://github.com/kubernetes/ingress/pull/1468) Add the original URL to the auth request
- [X] [#1469](https://github.com/kubernetes/ingress/pull/1469) Typo: Add missing {{ }}
- [X] [#1472](https://github.com/kubernetes/ingress/pull/1472) Fix X-Auth-Request-Redirect value to reflect the request uri
- [X] [#1473](https://github.com/kubernetes/ingress/pull/1473) Fix proxy protocol check
- [X] [#1475](https://github.com/kubernetes/ingress/pull/1475) Add OCSP support
- [X] [#1477](https://github.com/kubernetes/ingress/pull/1477) Fix semicolons in global configuration
- [X] [#1478](https://github.com/kubernetes/ingress/pull/1478) Pass redirect field in login page to get a proper redirect
- [X] [#1480](https://github.com/kubernetes/ingress/pull/1480) configurable ssl_verify_client
- [X] [#1485](https://github.com/kubernetes/ingress/pull/1485) Fix source IP address
- [X] [#1486](https://github.com/kubernetes/ingress/pull/1486) Fix overwrite of custom configuration

*Documentation:*

- [X] [#1460](https://github.com/kubernetes/ingress/pull/1460) Expose UDP port in UDP ingress example
- [X] [#1465](https://github.com/kubernetes/ingress/pull/1465) review prometheus docs

### 0.9-beta.14

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.14`

*New Features:*

- Opentracing support for NGINX
- Setting upstream vhost for nginx
- Allow custom global configuration at multiple levels
- Add support for proxy protocol decoding and encoding in TCP services

*Changes:*

- [X] [#719](https://github.com/kubernetes/ingress/pull/719) Setting upstream vhost for nginx.
- [X] [#1321](https://github.com/kubernetes/ingress/pull/1321) Enable keepalive in upstreams
- [X] [#1322](https://github.com/kubernetes/ingress/pull/1322) parse real ip
- [X] [#1323](https://github.com/kubernetes/ingress/pull/1323) use $the_real_ip for rate limit whitelist
- [X] [#1326](https://github.com/kubernetes/ingress/pull/1326) Pass headers from the custom error backend
- [X] [#1328](https://github.com/kubernetes/ingress/pull/1328) update deprecated interface
- [X] [#1329](https://github.com/kubernetes/ingress/pull/1329) add example for nginx-ingress
- [X] [#1330](https://github.com/kubernetes/ingress/pull/1330) Increase coverage in template.go for nginx controller
- [X] [#1335](https://github.com/kubernetes/ingress/pull/1335) Configurable proxy_request_buffering per location..
- [X] [#1338](https://github.com/kubernetes/ingress/pull/1338) Fix multiple leader election
- [X] [#1339](https://github.com/kubernetes/ingress/pull/1339) Enable status port listening in all interfaces
- [X] [#1340](https://github.com/kubernetes/ingress/pull/1340) Update sha256sum of nginx substitutions
- [X] [#1341](https://github.com/kubernetes/ingress/pull/1341) Fix typos
- [X] [#1345](https://github.com/kubernetes/ingress/pull/1345) refactor controllers.go
- [X] [#1349](https://github.com/kubernetes/ingress/pull/1349) Force reload if a secret is updated
- [X] [#1363](https://github.com/kubernetes/ingress/pull/1363) Fix proxy request buffering default configuration
- [X] [#1365](https://github.com/kubernetes/ingress/pull/1365) Fix equals comparsion returing False if both objects have nil Targets or Services.
- [X] [#1367](https://github.com/kubernetes/ingress/pull/1367) Fix typos
- [X] [#1379](https://github.com/kubernetes/ingress/pull/1379) Fix catch all  upstream server
- [X] [#1380](https://github.com/kubernetes/ingress/pull/1380) Cleanup
- [X] [#1381](https://github.com/kubernetes/ingress/pull/1381) Refactor X-Forwarded-* headers
- [X] [#1382](https://github.com/kubernetes/ingress/pull/1382) Cleanup
- [X] [#1387](https://github.com/kubernetes/ingress/pull/1387) Improve resource usage in nginx controller
- [X] [#1392](https://github.com/kubernetes/ingress/pull/1392) Avoid issues with goroutines updating fields
- [X] [#1393](https://github.com/kubernetes/ingress/pull/1393) Limit the number of goroutines used for the update of  ingress status
- [X] [#1394](https://github.com/kubernetes/ingress/pull/1394) Improve equals
- [X] [#1402](https://github.com/kubernetes/ingress/pull/1402) fix error when cert or key is nil
- [X] [#1403](https://github.com/kubernetes/ingress/pull/1403) Added tls ports to rbac nginx ingress controller and service
- [X] [#1404](https://github.com/kubernetes/ingress/pull/1404) Use nginx default value for SSLECDHCurve
- [X] [#1411](https://github.com/kubernetes/ingress/pull/1411) Add more descriptive logging in certificate loading
- [X] [#1412](https://github.com/kubernetes/ingress/pull/1412) Correct Error Handling to avoid panics and add more logging to template
- [X] [#1413](https://github.com/kubernetes/ingress/pull/1413) Validate external names
- [X] [#1418](https://github.com/kubernetes/ingress/pull/1418) Fix links after design proposals move
- [X] [#1419](https://github.com/kubernetes/ingress/pull/1419) Remove duplicated ingress check code
- [X] [#1420](https://github.com/kubernetes/ingress/pull/1420) Process queue items by time window
- [X] [#1423](https://github.com/kubernetes/ingress/pull/1423) Fix cast error
- [X] [#1424](https://github.com/kubernetes/ingress/pull/1424) Allow overriding the tag and registry
- [X] [#1426](https://github.com/kubernetes/ingress/pull/1426) Enhance Certificate Logging and Clearup Mutual Auth Docs
- [X] [#1430](https://github.com/kubernetes/ingress/pull/1430) Add support for proxy protocol decoding and encoding in TCP services
- [X] [#1434](https://github.com/kubernetes/ingress/pull/1434) Fix exec of readSecrets
- [X] [#1435](https://github.com/kubernetes/ingress/pull/1435) Add header to upstream server for external authentication
- [X] [#1438](https://github.com/kubernetes/ingress/pull/1438) Do not intercept errors from the custom error service
- [X] [#1439](https://github.com/kubernetes/ingress/pull/1439) Nginx master process killed thus no futher reloads
- [X] [#1440](https://github.com/kubernetes/ingress/pull/1440) Kill worker processes to allow the restart of nginx
- [X] [#1445](https://github.com/kubernetes/ingress/pull/1445) Updated godeps
- [X] [#1450](https://github.com/kubernetes/ingress/pull/1450) Fix links
- [X] [#1451](https://github.com/kubernetes/ingress/pull/1451) Add example of server-snippet
- [X] [#1452](https://github.com/kubernetes/ingress/pull/1452) Fix sync of secrets  (kube lego)
- [X] [#1454](https://github.com/kubernetes/ingress/pull/1454) Allow custom global configuration at multiple levels

*Documentation:*

- [X] [#1400](https://github.com/kubernetes/ingress/pull/1400) Fix ConfigMap link in doc
- [X] [#1422](https://github.com/kubernetes/ingress/pull/1422) Add docs for opentracing
- [X] [#1441](https://github.com/kubernetes/ingress/pull/1441) Improve custom error pages doc
- [X] [#1442](https://github.com/kubernetes/ingress/pull/1442) Opentracing docs
- [X] [#1446](https://github.com/kubernetes/ingress/pull/1446) Add custom timeout annotations doc


### 0.9-beta.13

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.13`

*New Features:*

- NGINX 1.3.5
- New flag to disable node listing
- Custom X-Forwarder-Header (CloudFlare uses `CF-Connecting-IP` as header)
- Custom error page in Client Certificate Authentication

*Changes:*

- [X] [#1272](https://github.com/kubernetes/ingress/pull/1272) Delete useless statement
- [X] [#1277](https://github.com/kubernetes/ingress/pull/1277) Add indent for nginx.conf
- [X] [#1278](https://github.com/kubernetes/ingress/pull/1278) Add proxy-pass-params annotation and Backend field
- [X] [#1282](https://github.com/kubernetes/ingress/pull/1282) Fix nginx stats
- [X] [#1288](https://github.com/kubernetes/ingress/pull/1288) Allow PATCH in enable-cors
- [X] [#1290](https://github.com/kubernetes/ingress/pull/1290) Add flag to disabling node listing
- [X] [#1293](https://github.com/kubernetes/ingress/pull/1293) Adds support for error page in Client Certificate Authentication
- [X] [#1308](https://github.com/kubernetes/ingress/pull/1308) A trivial typo in config
- [X] [#1310](https://github.com/kubernetes/ingress/pull/1310) Refactoring nginx configuration configmap
- [X] [#1311](https://github.com/kubernetes/ingress/pull/1311) Enable nginx async writes
- [X] [#1312](https://github.com/kubernetes/ingress/pull/1312) Allow custom forwarded for header
- [X] [#1313](https://github.com/kubernetes/ingress/pull/1313) Fix eol in nginx template
- [X] [#1315](https://github.com/kubernetes/ingress/pull/1315) Fix nginx custom error pages


*Documentation:*

- [X] [#1270](https://github.com/kubernetes/ingress/pull/1270) add missing yamls in controllers/nginx
- [X] [#1276](https://github.com/kubernetes/ingress/pull/1276) Link rbac sample from deployment docs
- [X] [#1291](https://github.com/kubernetes/ingress/pull/1291) fix link to conformance suite
- [X] [#1295](https://github.com/kubernetes/ingress/pull/1295) fix README of nginx-ingress-controller
- [X] [#1299](https://github.com/kubernetes/ingress/pull/1299) fix two doc issues in nginx/README
- [X] [#1306](https://github.com/kubernetes/ingress/pull/1306) Fix kubeconfig example for nginx deployment


### 0.9-beta.12

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.12`

*Breaking changes:*

- SSL passthrough is disabled by default. To enable the feature use `--enable-ssl-passthrough`

*New Features:*

- Support for arm64
- New flags to customize listen ports
- Per minute rate limiting
- Rate limit whitelist
- Configuration of nginx worker timeout (to avoid zombie nginx workers processes)
- Redirects from non-www to www
- Custom default backend (per Ingress)
- Graceful shutdown for NGINX

*Changes:*

- [X] [#977](https://github.com/kubernetes/ingress/pull/977) Add sort-backends command line option
- [X] [#981](https://github.com/kubernetes/ingress/pull/981) Add annotation to allow use of service ClusterIP for NGINX upstream.
- [X] [#991](https://github.com/kubernetes/ingress/pull/991) Remove secret sync loop
- [X] [#992](https://github.com/kubernetes/ingress/pull/992) Check errors generating pem files
- [X] [#993](https://github.com/kubernetes/ingress/pull/993) Fix the sed command to work on macOS
- [X] [#1013](https://github.com/kubernetes/ingress/pull/1013) The fields of vtsDate are unified in the form of plural
- [X] [#1025](https://github.com/kubernetes/ingress/pull/1025) Fix file watch
- [X] [#1027](https://github.com/kubernetes/ingress/pull/1027) Lint code
- [X] [#1031](https://github.com/kubernetes/ingress/pull/1031) Change missing secret name log level to V(3)
- [X] [#1032](https://github.com/kubernetes/ingress/pull/1032) Alternative syncSecret approach #1030
- [X] [#1042](https://github.com/kubernetes/ingress/pull/1042) Add function to allow custom values in Ingress status
- [X] [#1043](https://github.com/kubernetes/ingress/pull/1043) Return reference to object providing Endpoint
- [X] [#1046](https://github.com/kubernetes/ingress/pull/1046) Add field FileSHA in BasicDigest struct
- [X] [#1058](https://github.com/kubernetes/ingress/pull/1058) add per minute rate limiting
- [X] [#1060](https://github.com/kubernetes/ingress/pull/1060) Update fsnotify dependency to fix arm64 issue
- [X] [#1065](https://github.com/kubernetes/ingress/pull/1065) Add more descriptive steps in Dev Documentation
- [X] [#1073](https://github.com/kubernetes/ingress/pull/1073) Release nginx-slim 0.22
- [X] [#1074](https://github.com/kubernetes/ingress/pull/1074) Remove lua and use fastcgi to render errors
- [X] [#1075](https://github.com/kubernetes/ingress/pull/1075) (feat/ #374) support proxy timeout
- [X] [#1076](https://github.com/kubernetes/ingress/pull/1076) Add more ssl test cases
- [X] [#1078](https://github.com/kubernetes/ingress/pull/1078) fix the same udp port and tcp port, update nginx.conf error
- [X] [#1080](https://github.com/kubernetes/ingress/pull/1080) Disable platform s390x
- [X] [#1081](https://github.com/kubernetes/ingress/pull/1081) Spit Static check and Coverage in diff Stages of Travis CI
- [X] [#1082](https://github.com/kubernetes/ingress/pull/1082) Fix build tasks
- [X] [#1087](https://github.com/kubernetes/ingress/pull/1087) Release nginx-slim 0.23
- [X] [#1088](https://github.com/kubernetes/ingress/pull/1088) Configure nginx worker timeout
- [X] [#1089](https://github.com/kubernetes/ingress/pull/1089) Update nginx to 1.13.4
- [X] [#1098](https://github.com/kubernetes/ingress/pull/1098) Exposing the event recorder to allow other controllers to create events
- [X] [#1102](https://github.com/kubernetes/ingress/pull/1102) Fix lose SSL Passthrough
- [X] [#1104](https://github.com/kubernetes/ingress/pull/1104) Simplify verification of hostname in ssl certificates
- [X] [#1109](https://github.com/kubernetes/ingress/pull/1109) Cleanup remote address in nginx template
- [X] [#1110](https://github.com/kubernetes/ingress/pull/1110) Fix Endpoint comparison
- [X] [#1118](https://github.com/kubernetes/ingress/pull/1118) feat(#733)Support nginx bandwidth control
- [X] [#1124](https://github.com/kubernetes/ingress/pull/1124) check fields len in dns.go
- [X] [#1130](https://github.com/kubernetes/ingress/pull/1130) Update nginx.go
- [X] [#1134](https://github.com/kubernetes/ingress/pull/1134) replace deprecated interface with versioned ones
- [X] [#1136](https://github.com/kubernetes/ingress/pull/1136) Fix status update - changed in #1074
- [X] [#1138](https://github.com/kubernetes/ingress/pull/1138) update nginx.go: preformance improve
- [X] [#1139](https://github.com/kubernetes/ingress/pull/1139) Fix Todo:convert sequence to table
- [X] [#1162](https://github.com/kubernetes/ingress/pull/1162) Optimize CI build time
- [X] [#1164](https://github.com/kubernetes/ingress/pull/1164) Use variable request_uri as redirect after auth
- [X] [#1179](https://github.com/kubernetes/ingress/pull/1179) Fix sticky upstream not used when enable rewrite
- [X] [#1184](https://github.com/kubernetes/ingress/pull/1184) Add support for temporal and permanent redirects
- [X] [#1185](https://github.com/kubernetes/ingress/pull/1185) Add more info about Server-Alias usage
- [X] [#1186](https://github.com/kubernetes/ingress/pull/1186) Add annotation for client-body-buffer-size per location
- [X] [#1190](https://github.com/kubernetes/ingress/pull/1190) Add flag to disable SSL passthrough
- [X] [#1193](https://github.com/kubernetes/ingress/pull/1193) fix broken link
- [X] [#1198](https://github.com/kubernetes/ingress/pull/1198) Add option for specific scheme for base url
- [X] [#1202](https://github.com/kubernetes/ingress/pull/1202) formatIP issue
- [X] [#1203](https://github.com/kubernetes/ingress/pull/1203) NGINX not reloading correctly
- [X] [#1204](https://github.com/kubernetes/ingress/pull/1204) Fix template error
- [X] [#1205](https://github.com/kubernetes/ingress/pull/1205) Add initial sync of secrets
- [X] [#1206](https://github.com/kubernetes/ingress/pull/1206) Update ssl-passthrough docs
- [X] [#1207](https://github.com/kubernetes/ingress/pull/1207) delete broken link
- [X] [#1208](https://github.com/kubernetes/ingress/pull/1208) fix some typo
- [X] [#1210](https://github.com/kubernetes/ingress/pull/1210) add rate limit whitelist
- [X] [#1215](https://github.com/kubernetes/ingress/pull/1215) Replace base64 encoding with random uuid
- [X] [#1218](https://github.com/kubernetes/ingress/pull/1218) Trivial fixes in core/pkg/net
- [X] [#1219](https://github.com/kubernetes/ingress/pull/1219) keep zones unique per ingress resource
- [X] [#1221](https://github.com/kubernetes/ingress/pull/1221) Move certificate authentication from location to server
- [X] [#1223](https://github.com/kubernetes/ingress/pull/1223) Add doc for non-www to www annotation
- [X] [#1224](https://github.com/kubernetes/ingress/pull/1224) refactor rate limit whitelist
- [X] [#1226](https://github.com/kubernetes/ingress/pull/1226) Remove useless variable in nginx.tmpl
- [X] [#1227](https://github.com/kubernetes/ingress/pull/1227) Update annotations doc with base-url-scheme
- [X] [#1233](https://github.com/kubernetes/ingress/pull/1233) Fix ClientBodyBufferSize annotation
- [X] [#1234](https://github.com/kubernetes/ingress/pull/1234) Lint code
- [X] [#1235](https://github.com/kubernetes/ingress/pull/1235) Fix Equal comparison
- [X] [#1236](https://github.com/kubernetes/ingress/pull/1236) Add Validation for Client Body Buffer Size
- [X] [#1238](https://github.com/kubernetes/ingress/pull/1238) Add support for 'client_body_timeout' and 'client_header_timeout'
- [X] [#1239](https://github.com/kubernetes/ingress/pull/1239) Add flags to customize listen ports and detect port collisions
- [X] [#1243](https://github.com/kubernetes/ingress/pull/1243) Add support for access-log-path and error-log-path
- [X] [#1244](https://github.com/kubernetes/ingress/pull/1244) Add custom default backend annotation
- [X] [#1246](https://github.com/kubernetes/ingress/pull/1246) Add additional headers when custom default backend is used
- [X] [#1247](https://github.com/kubernetes/ingress/pull/1247) Make Ingress annotations available in template
- [X] [#1248](https://github.com/kubernetes/ingress/pull/1248) Improve nginx controller performance
- [X] [#1254](https://github.com/kubernetes/ingress/pull/1254) fix Type transform panic
- [X] [#1257](https://github.com/kubernetes/ingress/pull/1257) Graceful shutdown for Nginx
- [X] [#1261](https://github.com/kubernetes/ingress/pull/1261) Add support for 'worker-shutdown-timeout'


*Documentation:*

- [X] [#976](https://github.com/kubernetes/ingress/pull/976) Update annotations doc
- [X] [#979](https://github.com/kubernetes/ingress/pull/979) Missing auth example
- [X] [#980](https://github.com/kubernetes/ingress/pull/980) Add nginx basic auth example
- [X] [#1001](https://github.com/kubernetes/ingress/pull/1001) examples/nginx/rbac: Give access to own namespace
- [X] [#1005](https://github.com/kubernetes/ingress/pull/1005) Update configuration.md
- [X] [#1018](https://github.com/kubernetes/ingress/pull/1018) add docs for `proxy-set-headers` and `add-headers`
- [X] [#1038](https://github.com/kubernetes/ingress/pull/1038) typo / spelling in README.md
- [X] [#1039](https://github.com/kubernetes/ingress/pull/1039) typo in examples/tcp/nginx/README.md
- [X] [#1049](https://github.com/kubernetes/ingress/pull/1049) Fix config name in the example.
- [X] [#1054](https://github.com/kubernetes/ingress/pull/1054) Fix link to UDP example
- [X] [#1084](https://github.com/kubernetes/ingress/pull/1084) (issue #310)Fix some broken link
- [X] [#1103](https://github.com/kubernetes/ingress/pull/1103) Add GoDoc Widget
- [X] [#1105](https://github.com/kubernetes/ingress/pull/1105) Make Readme file more readable
- [X] [#1106](https://github.com/kubernetes/ingress/pull/1106) Update annotations.md
- [X] [#1107](https://github.com/kubernetes/ingress/pull/1107) Fix Broken Link
- [X] [#1119](https://github.com/kubernetes/ingress/pull/1119) fix typos in controllers/nginx/README.md
- [X] [#1122](https://github.com/kubernetes/ingress/pull/1122) Fix broken link
- [X] [#1131](https://github.com/kubernetes/ingress/pull/1131) Add short help doc in configuration for nginx limit rate
- [X] [#1143](https://github.com/kubernetes/ingress/pull/1143) Minor Typo Fix
- [X] [#1144](https://github.com/kubernetes/ingress/pull/1144) Minor Typo fix
- [X] [#1145](https://github.com/kubernetes/ingress/pull/1145) Minor Typo fix
- [X] [#1146](https://github.com/kubernetes/ingress/pull/1146) Fix Minor Typo in Readme
- [X] [#1147](https://github.com/kubernetes/ingress/pull/1147) Minor Typo Fix
- [X] [#1148](https://github.com/kubernetes/ingress/pull/1148) Minor Typo Fix in Getting-Started.md
- [X] [#1149](https://github.com/kubernetes/ingress/pull/1149) Fix Minor Typo in TLS authentication
- [X] [#1150](https://github.com/kubernetes/ingress/pull/1150) Fix Minor Typo in Customize the HAProxy configuration
- [X] [#1151](https://github.com/kubernetes/ingress/pull/1151) Fix Minor Typo in customization custom-template
- [X] [#1152](https://github.com/kubernetes/ingress/pull/1152) Fix minor typo in HAProxy Multi TLS certificate termination
- [X] [#1153](https://github.com/kubernetes/ingress/pull/1153) Fix minor typo in Multi TLS certificate termination
- [X] [#1154](https://github.com/kubernetes/ingress/pull/1154) Fix minor typo in Role Based Access Control
- [X] [#1155](https://github.com/kubernetes/ingress/pull/1155) Fix minor typo in TCP loadbalancing
- [X] [#1156](https://github.com/kubernetes/ingress/pull/1156) Fix minor typo in UDP loadbalancing
- [X] [#1157](https://github.com/kubernetes/ingress/pull/1157) Fix minor typos in Prerequisites
- [X] [#1158](https://github.com/kubernetes/ingress/pull/1158) Fix minor typo in Ingress examples
- [X] [#1159](https://github.com/kubernetes/ingress/pull/1159) Fix minor typos in Ingress admin guide
- [X] [#1160](https://github.com/kubernetes/ingress/pull/1160) Fix a broken href and typo in Ingress FAQ
- [X] [#1165](https://github.com/kubernetes/ingress/pull/1165) Update CONTRIBUTING.md
- [X] [#1168](https://github.com/kubernetes/ingress/pull/1168) finx link to running-locally.md
- [X] [#1170](https://github.com/kubernetes/ingress/pull/1170) Update dead link in nginx/HTTPS section
- [X] [#1172](https://github.com/kubernetes/ingress/pull/1172) Update README.md
- [X] [#1173](https://github.com/kubernetes/ingress/pull/1173) Update admin.md
- [X] [#1174](https://github.com/kubernetes/ingress/pull/1174) fix several titles
- [X] [#1177](https://github.com/kubernetes/ingress/pull/1177) fix typos
- [X] [#1188](https://github.com/kubernetes/ingress/pull/1188) Fix minor typo
- [X] [#1189](https://github.com/kubernetes/ingress/pull/1189) Fix sign in URL redirect parameter
- [X] [#1192](https://github.com/kubernetes/ingress/pull/1192) Update README.md
- [X] [#1195](https://github.com/kubernetes/ingress/pull/1195) Update troubleshooting.md
- [X] [#1196](https://github.com/kubernetes/ingress/pull/1196) Update README.md
- [X] [#1209](https://github.com/kubernetes/ingress/pull/1209) Update README.md
- [X] [#1085](https://github.com/kubernetes/ingress/pull/1085) Fix ConfigMap's namespace in custom configuration example for nginx
- [X] [#1142](https://github.com/kubernetes/ingress/pull/1142) Fix typo in multiple docs
- [X] [#1228](https://github.com/kubernetes/ingress/pull/1228) Update release doc in getting-started.md
- [X] [#1230](https://github.com/kubernetes/ingress/pull/1230) Update godep guide link


### 0.9-beta.11

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.11`

Fixes NGINX [CVE-2017-7529](http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-7529)

*Changes:*

- [X] [#659](https://github.com/kubernetes/ingress/pull/659) [nginx] TCP configmap should allow listen proxy_protocol per service
- [X] [#730](https://github.com/kubernetes/ingress/pull/730) Add support for add_headers
- [X] [#808](https://github.com/kubernetes/ingress/pull/808) HTTP->HTTPS redirect does not work with use-proxy-protocol: "true"
- [X] [#921](https://github.com/kubernetes/ingress/pull/921) Make proxy-real-ip-cidr a comma separated list
- [X] [#930](https://github.com/kubernetes/ingress/pull/930) Add support for proxy protocol in TCP services
- [X] [#933](https://github.com/kubernetes/ingress/pull/933) Lint code
- [X] [#937](https://github.com/kubernetes/ingress/pull/937) Fix lint code errors
- [X] [#940](https://github.com/kubernetes/ingress/pull/940) Sets parameters for a shared memory zone of limit_conn_zone
- [X] [#949](https://github.com/kubernetes/ingress/pull/949) fix nginx version to 1.13.3 to fix integer overflow
- [X] [#956](https://github.com/kubernetes/ingress/pull/956) Simplify handling of ssl certificates
- [X] [#958](https://github.com/kubernetes/ingress/pull/958) Release ubuntu-slim:0.13
- [X] [#959](https://github.com/kubernetes/ingress/pull/959) Release nginx-slim 0.21
- [X] [#960](https://github.com/kubernetes/ingress/pull/960) Update nginx in ingress controller
- [X] [#964](https://github.com/kubernetes/ingress/pull/964) Support for proxy_headers_hash_bucket_size and proxy_headers_hash_max_size
- [X] [#966](https://github.com/kubernetes/ingress/pull/966) Fix error checking for pod name & NS
- [X] [#967](https://github.com/kubernetes/ingress/pull/967) Fix runningAddresses typo
- [X] [#968](https://github.com/kubernetes/ingress/pull/968) Fix missing hyphen in yaml for nginx RBAC example
- [X] [#973](https://github.com/kubernetes/ingress/pull/973) check number of servers in configuration comparator


### 0.9-beta.10

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.10`

Fix release 0.9-beta.9

### 0.9-beta.9

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.9`

*New Features:*

- Add support for arm and ppc64le


*Changes:*

- [X] [#548](https://github.com/kubernetes/ingress/pull/548) nginx: support multidomain certificates
- [X] [#620](https://github.com/kubernetes/ingress/pull/620) [nginx] Listening ports are not configurable, so ingress can't be run multiple times per node when using CNI
- [X] [#648](https://github.com/kubernetes/ingress/pull/648) publish-service argument isn't honored when ELB is internal only facing.
- [X] [#833](https://github.com/kubernetes/ingress/pull/833) WIP: Avoid reloads implementing Equals in structs
- [X] [#838](https://github.com/kubernetes/ingress/pull/838) Feature request: Add ingress annotation to enable upstream "keepalive" option
- [X] [#844](https://github.com/kubernetes/ingress/pull/844) ingress annotations affinity is not working
- [X] [#862](https://github.com/kubernetes/ingress/pull/862) Avoid reloads implementing Equaler interface
- [X] [#864](https://github.com/kubernetes/ingress/pull/864) Remove dead code
- [X] [#868](https://github.com/kubernetes/ingress/pull/868) Lint nginx code
- [X] [#871](https://github.com/kubernetes/ingress/pull/871) Add feature to allow sticky sessions per location
- [X] [#873](https://github.com/kubernetes/ingress/pull/873) Update README.md
- [X] [#876](https://github.com/kubernetes/ingress/pull/876) Add information about nginx controller flags
- [X] [#878](https://github.com/kubernetes/ingress/pull/878) Update go to 1.8.3
- [X] [#881](https://github.com/kubernetes/ingress/pull/881) Option to not remove loadBalancer status record?
- [X] [#882](https://github.com/kubernetes/ingress/pull/882) Add flag to skip the update of Ingress status on shutdown
- [X] [#885](https://github.com/kubernetes/ingress/pull/885) Don't use $proxy_protocol var which may be undefined.
- [X] [#886](https://github.com/kubernetes/ingress/pull/886) Add support for SubjectAltName in SSL certificates
- [X] [#888](https://github.com/kubernetes/ingress/pull/888) Update nginx-slim to 0.19
- [X] [#889](https://github.com/kubernetes/ingress/pull/889) Add PHOST to backend
- [X] [#890](https://github.com/kubernetes/ingress/pull/890) Improve variable configuration for source IP address
- [X] [#892](https://github.com/kubernetes/ingress/pull/892) Add upstream keepalive connections cache
- [X] [#897](https://github.com/kubernetes/ingress/pull/897) Update outdated ingress resource link
- [X] [#898](https://github.com/kubernetes/ingress/pull/898) add error check right when reload nginx fail
- [X] [#899](https://github.com/kubernetes/ingress/pull/899) Fix nginx error check
- [X] [#900](https://github.com/kubernetes/ingress/pull/900) After #862 changes in the configmap do not trigger a reload
- [X] [#901](https://github.com/kubernetes/ingress/pull/901) [doc] Update NGinX status port to 18080
- [X] [#902](https://github.com/kubernetes/ingress/pull/902) Always reload after a change in the configuration
- [X] [#904](https://github.com/kubernetes/ingress/pull/904) Fix nginx sticky sessions
- [X] [#906](https://github.com/kubernetes/ingress/pull/906) Fix race condition with closed channels
- [X] [#907](https://github.com/kubernetes/ingress/pull/907) nginx/proxy: allow specifying next upstream behaviour
- [X] [#910](https://github.com/kubernetes/ingress/pull/910) Feature request: use `X-Forwarded-Host` from the reverse proxy before
- [X] [#911](https://github.com/kubernetes/ingress/pull/911) Improve X-Forwarded-Host support
- [X] [#915](https://github.com/kubernetes/ingress/pull/915) Release nginx-slim 0.20
- [X] [#916](https://github.com/kubernetes/ingress/pull/916) Add arm and ppc64le support
- [X] [#919](https://github.com/kubernetes/ingress/pull/919) Apply the 'ssl-redirect' annotation per-location
- [X] [#922](https://github.com/kubernetes/ingress/pull/922) Add example of TLS termination using a classic ELB

### 0.9-beta.8

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.8`

*Changes:*

- [X] [#761](https://github.com/kubernetes/ingress/pull/761) NGINX TCP Ingresses do not bind on IPv6
- [X] [#850](https://github.com/kubernetes/ingress/pull/850) Fix IPv6 UDP stream section
- [X] [#851](https://github.com/kubernetes/ingress/pull/851) ensure private key and certificate match
- [X] [#852](https://github.com/kubernetes/ingress/pull/852) Don't expose certificate metrics for default server
- [X] [#846](https://github.com/kubernetes/ingress/pull/846) Match ServicePort to Endpoints by Name
- [X] [#854](https://github.com/kubernetes/ingress/pull/854) Document log-format-stream and log-format-upstream
- [X] [#847](https://github.com/kubernetes/ingress/pull/847) fix semicolon
- [X] [#848](https://github.com/kubernetes/ingress/pull/848) Add metric "ssl certificate expiration"
- [X] [#839](https://github.com/kubernetes/ingress/pull/839) "No endpoints" issue
- [X] [#845](https://github.com/kubernetes/ingress/pull/845) Fix no endpoints issue when named ports are used
- [X] [#822](https://github.com/kubernetes/ingress/pull/822) Release ubuntu-slim 0.11
- [X] [#824](https://github.com/kubernetes/ingress/pull/824) Update nginx-slim to 0.18
- [X] [#823](https://github.com/kubernetes/ingress/pull/823) Release nginx-slim 0.18
- [X] [#827](https://github.com/kubernetes/ingress/pull/827) Introduce working example of nginx controller with rbac
- [X] [#835](https://github.com/kubernetes/ingress/pull/835) Make log format json escaping configurable
- [X] [#843](https://github.com/kubernetes/ingress/pull/843) Avoid setting maximum number of open file descriptors lower than 1024
- [X] [#837](https://github.com/kubernetes/ingress/pull/837) Cleanup interface
- [X] [#836](https://github.com/kubernetes/ingress/pull/836) Make log format json escaping configurable
- [X] [#828](https://github.com/kubernetes/ingress/pull/828) Wrap IPv6 endpoints in []
- [X] [#821](https://github.com/kubernetes/ingress/pull/821) nginx-ingress: occasional 503 Service Temporarily Unavailable
- [X] [#829](https://github.com/kubernetes/ingress/pull/829) feat(template): wrap IPv6 addresses in []
- [X] [#786](https://github.com/kubernetes/ingress/pull/786) Update echoserver image version in examples
- [X] [#825](https://github.com/kubernetes/ingress/pull/825) Create or delete ingress based on class annotation
- [X] [#790](https://github.com/kubernetes/ingress/pull/790) #789 removing duplicate X-Real-IP header
- [X] [#792](https://github.com/kubernetes/ingress/pull/792) Avoid checking if the controllers are synced
- [X] [#798](https://github.com/kubernetes/ingress/pull/798) nginx: RBAC for leader election
- [X] [#799](https://github.com/kubernetes/ingress/pull/799) could not build variables_hash
- [X] [#809](https://github.com/kubernetes/ingress/pull/809) Fix dynamic variable name
- [X] [#804](https://github.com/kubernetes/ingress/pull/804) Fix #798 - RBAC for leader election
- [X] [#806](https://github.com/kubernetes/ingress/pull/806) fix ingress rbac roles
- [X] [#811](https://github.com/kubernetes/ingress/pull/811) external auth - proxy_pass_request_body off + big bodies give 500/413
- [X] [#785](https://github.com/kubernetes/ingress/pull/785) Publish echoheader image
- [X] [#813](https://github.com/kubernetes/ingress/pull/813) Added client_max_body_size to authPath location
- [X] [#814](https://github.com/kubernetes/ingress/pull/814) rbac-nginx: resourceNames cannot filter create verb
- [X] [#774](https://github.com/kubernetes/ingress/pull/774) Add IPv6 support in TCP and UDP stream section
- [X] [#784](https://github.com/kubernetes/ingress/pull/784) Allow customization of variables hash tables
- [X] [#782](https://github.com/kubernetes/ingress/pull/782) Set "proxy_pass_header Server;"
- [X] [#783](https://github.com/kubernetes/ingress/pull/783) nginx/README.md: clarify app-root and fix example hyperlink
- [X] [#787](https://github.com/kubernetes/ingress/pull/787) Add setting to allow returning the Server header from the backend

### 0.9-beta.7

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.7`

*Changes:*

- [X] [#777](https://github.com/kubernetes/ingress/pull/777) Update sniff parser to fix index out of bound error

### 0.9-beta.6

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.6`

*Changes:*

- [X] [#647](https://github.com/kubernetes/ingress/pull/647) ingress.class enhancement for debugging.
- [X] [#708](https://github.com/kubernetes/ingress/pull/708) ingress losing real source IP when tls enabled
- [X] [#760](https://github.com/kubernetes/ingress/pull/760) Change recorder event scheme
- [X] [#704](https://github.com/kubernetes/ingress/pull/704) fix nginx reload flags '-c'
- [X] [#757](https://github.com/kubernetes/ingress/pull/757) Replace use of endpoints as locks with configmap
- [X] [#752](https://github.com/kubernetes/ingress/pull/752) nginx ingress header config backwards
- [X] [#756](https://github.com/kubernetes/ingress/pull/756) Fix bad variable assignment in template nginx
- [X] [#729](https://github.com/kubernetes/ingress/pull/729) Release nginx-slim 0.17
- [X] [#755](https://github.com/kubernetes/ingress/pull/755) Fix server name hash maxSize default value
- [X] [#741](https://github.com/kubernetes/ingress/pull/741) Update golang dependencies
- [X] [#749](https://github.com/kubernetes/ingress/pull/749) Remove service annotation for namedPorts
- [X] [#740](https://github.com/kubernetes/ingress/pull/740) Refactoring whitelist source IP verification
- [X] [#734](https://github.com/kubernetes/ingress/pull/734) Specify nginx image arch
- [X] [#728](https://github.com/kubernetes/ingress/pull/728) Update nginx image
- [X] [#723](https://github.com/kubernetes/ingress/pull/723) update readme about vts metrics
- [X] [#726](https://github.com/kubernetes/ingress/pull/726) Release ubuntu-slim 0.10
- [X] [#727](https://github.com/kubernetes/ingress/pull/727) [nginx] whitelist-source-range doesn’t work on ssl port
- [X] [#709](https://github.com/kubernetes/ingress/pull/709) Add config for X-Forwarded-For trust
- [X] [#679](https://github.com/kubernetes/ingress/pull/679) add getenv
- [X] [#680](https://github.com/kubernetes/ingress/pull/680) nginx/pkg/config: delete unuseful variable
- [X] [#716](https://github.com/kubernetes/ingress/pull/716) Add secure-verify-ca-secret annotation
- [X] [#722](https://github.com/kubernetes/ingress/pull/722) Remove go-reap and use tini as process reaper
- [X] [#725](https://github.com/kubernetes/ingress/pull/725) Add keepalive_requests and client_body_buffer_size options
- [X] [#724](https://github.com/kubernetes/ingress/pull/724) change the directory of default-backend.yaml
- [X] [#656](https://github.com/kubernetes/ingress/pull/656) Nginx Ingress Controller - Specify load balancing method
- [X] [#717](https://github.com/kubernetes/ingress/pull/717) delete unuseful variable
- [X] [#712](https://github.com/kubernetes/ingress/pull/712) Set $proxy_upstream_name before location directive
- [X] [#715](https://github.com/kubernetes/ingress/pull/715) Corrected annotation ex `signin-url` to `auth-url`
- [X] [#718](https://github.com/kubernetes/ingress/pull/718) nodeController sync
- [X] [#694](https://github.com/kubernetes/ingress/pull/694) SSL-Passthrough broken in beta.5
- [X] [#678](https://github.com/kubernetes/ingress/pull/678) Convert CN SSL Certificate to lowercase before comparison
- [X] [#690](https://github.com/kubernetes/ingress/pull/690) Fix IP in logs for https traffic
- [X] [#673](https://github.com/kubernetes/ingress/pull/673) Override load balancer alg view config map
- [X] [#675](https://github.com/kubernetes/ingress/pull/675) Use proxy-protocol to pass through source IP to nginx
- [X] [#707](https://github.com/kubernetes/ingress/pull/707) use nginx vts module version 0.1.14
- [X] [#702](https://github.com/kubernetes/ingress/pull/702) Document passing of ssl_client_cert to backend
- [X] [#688](https://github.com/kubernetes/ingress/pull/688) Add example of UDP loadbalancing
- [X] [#696](https://github.com/kubernetes/ingress/pull/696) [nginx] pass non-SNI TLS hello to default backend, Fixes #693
- [X] [#685](https://github.com/kubernetes/ingress/pull/685) Fix error in generated nginx.conf for optional hsts-preload


### 0.9-beta.5

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.5`

*Changes:*

- [X] [#663](https://github.com/kubernetes/ingress/pull/663) Remove helper required in go < 1.8
- [X] [#662](https://github.com/kubernetes/ingress/pull/662) Add debug information about ingress class
- [X] [#661](https://github.com/kubernetes/ingress/pull/661) Avoid running nginx if the configuration file is empty
- [X] [#660](https://github.com/kubernetes/ingress/pull/660) Rollback queue refactoring
- [X] [#654](https://github.com/kubernetes/ingress/pull/654) Update go version to 1.8


### 0.9-beta.4

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.4`

*New Features:*

- Add support for services of type ExternalName


*Changes:*

- [X] [#635](https://github.com/kubernetes/ingress/pull/635) Allow configuration of features underscores_in_headers and ignore_invalid_headers
- [X] [#633](https://github.com/kubernetes/ingress/pull/633) Fix lint errors
- [X] [#630](https://github.com/kubernetes/ingress/pull/630) Add example of TCP loadbalancing
- [X] [#629](https://github.com/kubernetes/ingress/pull/629) Add support for services of type ExternalName
- [X] [#624](https://github.com/kubernetes/ingress/pull/624) Compute server_names_hash_bucket_size correctly
- [X] [#615](https://github.com/kubernetes/ingress/pull/615) Process exited cleanly before we hit wait4
- [X] [#614](https://github.com/kubernetes/ingress/pull/614) Refactor nginx ssl passthrough
- [X] [#613](https://github.com/kubernetes/ingress/pull/613) Status leader election must consired the ingress class
- [X] [#607](https://github.com/kubernetes/ingress/pull/607) Allow custom server_names_hash_max_size & server_names_hash_bucket_size
- [X] [#601](https://github.com/kubernetes/ingress/pull/601) add a judgment
- [X] [#601](https://github.com/kubernetes/ingress/pull/600) Replace custom child reap code with go-reap
- [X] [#597](https://github.com/kubernetes/ingress/pull/599) Add flag to force namespace isolation
- [X] [#595](https://github.com/kubernetes/ingress/pull/595) Remove Host header from auth_request proxy configuration
- [X] [#588](https://github.com/kubernetes/ingress/pull/588) Read resolv.conf file just once
- [X] [#586](https://github.com/kubernetes/ingress/pull/586) Updated instructions to create an ingress controller build
- [X] [#583](https://github.com/kubernetes/ingress/pull/583) fixed lua_package_path in nginx.tmpl
- [X] [#580](https://github.com/kubernetes/ingress/pull/580) Updated faq for running multiple ingress controller
- [X] [#579](https://github.com/kubernetes/ingress/pull/579) Detect if the ingress controller is running with multiple replicas
- [X] [#578](https://github.com/kubernetes/ingress/pull/578) Set different listeners per protocol version
- [X] [#577](https://github.com/kubernetes/ingress/pull/577) Avoid zombie child processes
- [X] [#576](https://github.com/kubernetes/ingress/pull/576) Replace secret workqueue
- [X] [#568](https://github.com/kubernetes/ingress/pull/568) Revert merge annotations to the implicit root context
- [X] [#563](https://github.com/kubernetes/ingress/pull/563) Add option to disable hsts preload
- [X] [#560](https://github.com/kubernetes/ingress/pull/560) Fix intermittent misconfiguration of backend.secure and SessionAffinity
- [X] [#556](https://github.com/kubernetes/ingress/pull/556) Update nginx version and remove dumb-init
- [X] [#551](https://github.com/kubernetes/ingress/pull/551) Build namespace and ingress class as label
- [X] [#546](https://github.com/kubernetes/ingress/pull/546) Fix a couple of 'does not contains' typos
- [X] [#542](https://github.com/kubernetes/ingress/pull/542) Fix lint errors
- [X] [#540](https://github.com/kubernetes/ingress/pull/540) Add Backends.SSLPassthrough attribute
- [X] [#539](https://github.com/kubernetes/ingress/pull/539) Migrate to client-go
- [X] [#536](https://github.com/kubernetes/ingress/pull/536) add unit test cases for core/pkg/ingress/controller/backend_ssl
- [X] [#535](https://github.com/kubernetes/ingress/pull/535) Add test for ingress status update
- [X] [#532](https://github.com/kubernetes/ingress/pull/532) Add setting to configure ecdh curve
- [X] [#531](https://github.com/kubernetes/ingress/pull/531) Fix link to examples
- [X] [#530](https://github.com/kubernetes/ingress/pull/530) Fix link to custom nginx configuration
- [X] [#528](https://github.com/kubernetes/ingress/pull/528) Add reference to apiserver-host flag
- [X] [#527](https://github.com/kubernetes/ingress/pull/527) Add annotations to location of default backend (root context)
- [X] [#525](https://github.com/kubernetes/ingress/pull/525) Avoid negative values configuring the max number of open files
- [X] [#523](https://github.com/kubernetes/ingress/pull/523) Fix a typo in an error message
- [X] [#521](https://github.com/kubernetes/ingress/pull/521) nginx-ingress-controller is built twice by docker-build target
- [X] [#517](https://github.com/kubernetes/ingress/pull/517) Use whitelist-source-range from configmap when no annotation on ingress
- [X] [#516](https://github.com/kubernetes/ingress/pull/516) Convert WorkerProcesses setting to string to allow the value auto
- [X] [#512](https://github.com/kubernetes/ingress/pull/512) Fix typos regarding the ssl-passthrough annotation documentation
- [X] [#505](https://github.com/kubernetes/ingress/pull/505) add unit test cases for core/pkg/ingress/controller/annotations
- [X] [#503](https://github.com/kubernetes/ingress/pull/503) Add example for nginx in aws
- [X] [#502](https://github.com/kubernetes/ingress/pull/502) Add information about SSL Passthrough annotation
- [X] [#500](https://github.com/kubernetes/ingress/pull/500) Improve TLS secret configuration
- [X] [#498](https://github.com/kubernetes/ingress/pull/498) Proper enqueue a secret on the secret queue
- [X] [#493](https://github.com/kubernetes/ingress/pull/493) Update nginx and vts module
- [X] [#490](https://github.com/kubernetes/ingress/pull/490) Add unit test case for named_port
- [X] [#488](https://github.com/kubernetes/ingress/pull/488) Adds support for CORS on error responses and Authorization header
- [X] [#485](https://github.com/kubernetes/ingress/pull/485) Fix typo nginx configMap vts metrics customization
- [X] [#481](https://github.com/kubernetes/ingress/pull/481) Remove unnecessary quote in nginx log format
- [X] [#471](https://github.com/kubernetes/ingress/pull/471) prometheus scrape annotations
- [X] [#460](https://github.com/kubernetes/ingress/pull/460) add example of 'run multiple haproxy ingress controllers as a deployment'
- [X] [#459](https://github.com/kubernetes/ingress/pull/459) Add information about SSL certificates in the default log level
- [X] [#456](https://github.com/kubernetes/ingress/pull/456) Avoid upstreams with multiple servers with the same port
- [X] [#454](https://github.com/kubernetes/ingress/pull/454) Pass request port to real server
- [X] [#450](https://github.com/kubernetes/ingress/pull/450) fix nginx-tcp-and-udp on same port
- [X] [#446](https://github.com/kubernetes/ingress/pull/446) remove configmap validations
- [X] [#445](https://github.com/kubernetes/ingress/pull/445) Remove snakeoil certificate generation
- [X] [#442](https://github.com/kubernetes/ingress/pull/442) Fix a few bugs in the nginx-ingress-controller Makefile
- [X] [#441](https://github.com/kubernetes/ingress/pull/441) skip validation when configmap is empty
- [X] [#439](https://github.com/kubernetes/ingress/pull/439) Avoid a nil-reference when the temporary file cannot be created
- [X] [#438](https://github.com/kubernetes/ingress/pull/438) Improve English in error messages
- [X] [#437](https://github.com/kubernetes/ingress/pull/437) Reference constant


### 0.9-beta.3

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.3`

*New Features:*

- Custom log formats using `log-format-upstream` directive in the configuration configmap.
- Force redirect to SSL using the annotation `ingress.kubernetes.io/force-ssl-redirect`
- Prometheus metric for VTS status module (transparent, just enable vts stats)
- Improved external authentication adding `ingress.kubernetes.io/auth-signin` annotation. Please check this [example](https://github.com/kubernetes/ingress/tree/master/examples/external-auth/nginx)


*Breaking changes:*

- `ssl-dh-param` configuration in configmap is now the name of a secret that contains the Diffie-Hellman key

*Changes:*

- [X] [#433](https://github.com/kubernetes/ingress/pull/433) close over the ingress variable or the last assignment will be used
- [X] [#424](https://github.com/kubernetes/ingress/pull/424) Manually sync secrets from certificate authentication annotations
- [X] [#423](https://github.com/kubernetes/ingress/pull/423) Scrap json metrics from nginx vts module when enabled
- [X] [#418](https://github.com/kubernetes/ingress/pull/418) Only update Ingress status for the configured class
- [X] [#415](https://github.com/kubernetes/ingress/pull/415) Improve external authentication docs
- [X] [#410](https://github.com/kubernetes/ingress/pull/410) Add support for "signin url"
- [X] [#409](https://github.com/kubernetes/ingress/pull/409) Allow custom http2 header sizes
- [X] [#408](https://github.com/kubernetes/ingress/pull/408) Review docs
- [X] [#406](https://github.com/kubernetes/ingress/pull/406) Add debug info and fix spelling
- [X] [#402](https://github.com/kubernetes/ingress/pull/402) allow specifying custom dh param
- [X] [#397](https://github.com/kubernetes/ingress/pull/397) Fix external auth
- [X] [#394](https://github.com/kubernetes/ingress/pull/394) Update README.md
- [X] [#392](https://github.com/kubernetes/ingress/pull/392) Fix http2 header size
- [X] [#391](https://github.com/kubernetes/ingress/pull/391) remove tmp nginx-diff files
- [X] [#390](https://github.com/kubernetes/ingress/pull/390) Fix RateLimit comment
- [X] [#385](https://github.com/kubernetes/ingress/pull/385) add Copyright
- [X] [#382](https://github.com/kubernetes/ingress/pull/382) Ingress Fake Certificate generation
- [X] [#380](https://github.com/kubernetes/ingress/pull/380) Fix custom log format
- [X] [#373](https://github.com/kubernetes/ingress/pull/373) Cleanup
- [X] [#371](https://github.com/kubernetes/ingress/pull/371) add configuration to disable listening on ipv6
- [X] [#370](https://github.com/kubernetes/ingress/pull/270) Add documentation for ingress.kubernetes.io/force-ssl-redirect
- [X] [#369](https://github.com/kubernetes/ingress/pull/369) Minor text fix for "ApiServer"
- [X] [#367](https://github.com/kubernetes/ingress/pull/367) BuildLogFormatUpstream was always using the default log-format
- [X] [#366](https://github.com/kubernetes/ingress/pull/366) add_judgment
- [X] [#365](https://github.com/kubernetes/ingress/pull/365) add ForceSSLRedirect ingress annotation
- [X] [#364](https://github.com/kubernetes/ingress/pull/364) Fix error caused by increasing proxy_buffer_size (#363)
- [X] [#362](https://github.com/kubernetes/ingress/pull/362) Fix ingress class
- [X] [#360](https://github.com/kubernetes/ingress/pull/360) add example of 'run multiple nginx ingress controllers as a deployment'
- [X] [#358](https://github.com/kubernetes/ingress/pull/358) Checks if the TLS secret contains a valid keypair structure
- [X] [#356](https://github.com/kubernetes/ingress/pull/356) Disable listen only on ipv6 and fix proxy_protocol
- [X] [#354](https://github.com/kubernetes/ingress/pull/354) add judgment
- [X] [#352](https://github.com/kubernetes/ingress/pull/352) Add ability to customize upstream and stream log format
- [X] [#351](https://github.com/kubernetes/ingress/pull/351) Enable custom election id for status sync.
- [X] [#347](https://github.com/kubernetes/ingress/pull/347) Fix client source IP address
- [X] [#345](https://github.com/kubernetes/ingress/pull/345) Fix lint error
- [X] [#344](https://github.com/kubernetes/ingress/pull/344) Refactoring of TCP and UDP services
- [X] [#343](https://github.com/kubernetes/ingress/pull/343) Fix node lister when --watch-namespace is used
- [X] [#341](https://github.com/kubernetes/ingress/pull/341) Do not run coverage check in the default target.
- [X] [#340](https://github.com/kubernetes/ingress/pull/340) Add support for specify proxy cookie path/domain
- [X] [#337](https://github.com/kubernetes/ingress/pull/337) Fix for formatting error introduced in #304
- [X] [#335](https://github.com/kubernetes/ingress/pull/335) Fix for vet complaints:
- [X] [#332](https://github.com/kubernetes/ingress/pull/332) Add annotation to customize nginx configuration
- [X] [#331](https://github.com/kubernetes/ingress/pull/331) Correct spelling mistake
- [X] [#328](https://github.com/kubernetes/ingress/pull/328) fix misspell "affinity" in main.go
- [X] [#326](https://github.com/kubernetes/ingress/pull/326) add nginx daemonset example
- [X] [#311](https://github.com/kubernetes/ingress/pull/311) Sort stream service ports to avoid extra reloads
- [X] [#307](https://github.com/kubernetes/ingress/pull/307) Add docs for body-size annotation
- [X] [#306](https://github.com/kubernetes/ingress/pull/306) modify nginx readme
- [X] [#304](https://github.com/kubernetes/ingress/pull/304) change 'buildSSPassthrouthUpstreams' to 'buildSSLPassthroughUpstreams'


### 0.9-beta.2

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.2`

*New Features:*

- New configuration flag `proxy-set-headers` to allow set custom headers before send traffic to backends. [Example here](https://github.com/kubernetes/ingress/tree/master/examples/customization/custom-headers/nginx)
- Disable directive access_log globally using `disable-access-log: "true"` in the configuration ConfigMap.
- Sticky session per Ingress rule using the annotation `ingress.kubernetes.io/affinity`. [Example here](https://github.com/kubernetes/ingress/tree/master/examples/affinity/cookie/nginx)

*Changes:*

- [X] [#300](https://github.com/kubernetes/ingress/pull/300) Change nginx variable to use in filter of access_log
- [X] [#296](https://github.com/kubernetes/ingress/pull/296) Fix rewrite regex to match the start of the URL and not a substring
- [X] [#293](https://github.com/kubernetes/ingress/pull/293) Update makefile gcloud docker command
- [X] [#290](https://github.com/kubernetes/ingress/pull/290) Update nginx version in ingress controller to 1.11.10
- [X] [#286](https://github.com/kubernetes/ingress/pull/286) Add logs to help debugging and simplify default upstream configuration
- [X] [#285](https://github.com/kubernetes/ingress/pull/285) Added a Node StoreLister type
- [X] [#281](https://github.com/kubernetes/ingress/pull/281) Add chmod up directory tree for world read/execute on directories
- [X] [#279](https://github.com/kubernetes/ingress/pull/279) fix wrong link in the file of examples/README.md
- [X] [#275](https://github.com/kubernetes/ingress/pull/275) Pass headers to custom error backend
- [X] [#272](https://github.com/kubernetes/ingress/pull/272) Fix error getting class information from Ingress annotations
- [X] [#268](https://github.com/kubernetes/ingress/pull/268) minor: Fix typo in nginx README
- [X] [#265](https://github.com/kubernetes/ingress/pull/265) Fix rewrite annotation parser
- [X] [#262](https://github.com/kubernetes/ingress/pull/262) Add nginx README and configuration docs back
- [X] [#261](https://github.com/kubernetes/ingress/pull/261) types.go: fix typo in godoc
- [X] [#258](https://github.com/kubernetes/ingress/pull/258) Nginx sticky annotations
- [X] [#255](https://github.com/kubernetes/ingress/pull/255) Adds support for disabling access_log globally
- [X] [#247](https://github.com/kubernetes/ingress/pull/247) Fix wrong URL in nginx ingress configuration
- [X] [#246](https://github.com/kubernetes/ingress/pull/246) Add support for custom proxy headers using a ConfigMap
- [X] [#244](https://github.com/kubernetes/ingress/pull/244) Add information about cors annotation
- [X] [#241](https://github.com/kubernetes/ingress/pull/241) correct a spell mistake
- [X] [#232](https://github.com/kubernetes/ingress/pull/232) Change searchs with searches
- [X] [#231](https://github.com/kubernetes/ingress/pull/231) Add information about proxy_protocol in port 442
- [X] [#228](https://github.com/kubernetes/ingress/pull/228) Fix worker check issue
- [X] [#227](https://github.com/kubernetes/ingress/pull/227) proxy_protocol on ssl_passthrough listener
- [X] [#223](https://github.com/kubernetes/ingress/pull/223) Fix panic if a tempfile cannot be created
- [X] [#220](https://github.com/kubernetes/ingress/pull/220) Fixes for minikube usage instructions.
- [X] [#219](https://github.com/kubernetes/ingress/pull/219) Fix typo, add a couple of links.
- [X] [#218](https://github.com/kubernetes/ingress/pull/218) Improve links from CONTRIBUTING.
- [X] [#217](https://github.com/kubernetes/ingress/pull/217) Fix an e2e link.
- [X] [#212](https://github.com/kubernetes/ingress/pull/212) Simplify code to obtain TCP or UDP services
- [X] [#208](https://github.com/kubernetes/ingress/pull/208) Fix nil HTTP field
- [X] [#198](https://github.com/kubernetes/ingress/pull/198) Add an example for static-ip and deployment


### 0.9-beta.1

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.1`

*New Features:*

- SSL Passthrough
- New Flag `--publish-service` that set the Service fronting the ingress controllers
- Ingress status shows the correct IP/hostname address without duplicates
- Custom body sizes per Ingress
- Prometheus metrics


*Breaking changes:*

- Flag `--nginx-configmap` was replaced with `--configmap`
- Configmap field `body-size` was replaced with `proxy-body-size`

*Changes:*

- [X] [#184](https://github.com/kubernetes/ingress/pull/184) Fix template error
- [X] [#179](https://github.com/kubernetes/ingress/pull/179) Allows the usage of Default SSL Cert
- [X] [#178](https://github.com/kubernetes/ingress/pull/178) Add initialization of proxy variable
- [X] [#177](https://github.com/kubernetes/ingress/pull/177) Refactoring sysctlFSFileMax helper
- [X] [#176](https://github.com/kubernetes/ingress/pull/176) Fix TLS does not get updated when changed
- [X] [#174](https://github.com/kubernetes/ingress/pull/174) Update nginx to 1.11.9
- [X] [#172](https://github.com/kubernetes/ingress/pull/172) add some unit test cases for some packages under folder "core.pkg.ingress"
- [X] [#168](https://github.com/kubernetes/ingress/pull/168) Changes the SSL Temp file to something inside the same SSL Directory
- [X] [#165](https://github.com/kubernetes/ingress/pull/165) Fix rate limit issue when more than 2 servers enabled in ingress
- [X] [#161](https://github.com/kubernetes/ingress/pull/161) Document some missing parameters and their defaults for NGINX controller
- [X] [#158](https://github.com/kubernetes/ingress/pull/158) prefect unit test cases for annotation.proxy
- [X] [#156](https://github.com/kubernetes/ingress/pull/156) Fix issue for ratelimit
- [X] [#154](https://github.com/kubernetes/ingress/pull/154) add unit test cases for core.pkg.ingress.annotations.cors
- [X] [#151](https://github.com/kubernetes/ingress/pull/151) Port in redirect
- [X] [#150](https://github.com/kubernetes/ingress/pull/150) Add support for custom header sizes
- [X] [#149](https://github.com/kubernetes/ingress/pull/149) Add flag to allow switch off the update of Ingress status
- [X] [#148](https://github.com/kubernetes/ingress/pull/148) Add annotation to allow custom body sizes
- [X] [#145](https://github.com/kubernetes/ingress/pull/145) fix wrong links and punctuations
- [X] [#144](https://github.com/kubernetes/ingress/pull/144) add unit test cases for core.pkg.k8s
- [X] [#143](https://github.com/kubernetes/ingress/pull/143) Use protobuf instead of rest to connect to apiserver host and add troubleshooting doc
- [X] [#142](https://github.com/kubernetes/ingress/pull/142) Use system fs.max-files as limits instead of hard-coded value
- [X] [#141](https://github.com/kubernetes/ingress/pull/141) Add reuse port and backlog to port 80 and 443
- [X] [#138](https://github.com/kubernetes/ingress/pull/138) reference to const
- [X] [#136](https://github.com/kubernetes/ingress/pull/136) Add content and descriptions about nginx's configuration
- [X] [#135](https://github.com/kubernetes/ingress/pull/135) correct improper punctuation
- [X] [#134](https://github.com/kubernetes/ingress/pull/134) fix typo
- [X] [#133](https://github.com/kubernetes/ingress/pull/133) Add TCP and UDP services removed in migration
- [X] [#132](https://github.com/kubernetes/ingress/pull/132) Document nginx controller configuration tweaks
- [X] [#128](https://github.com/kubernetes/ingress/pull/128) Add tests and godebug to compare structs
- [X] [#126](https://github.com/kubernetes/ingress/pull/126) change the type of imagePullPolicy
- [X] [#123](https://github.com/kubernetes/ingress/pull/123) Add resolver configuration to nginx
- [X] [#119](https://github.com/kubernetes/ingress/pull/119) add unit test case for annotations.service
- [X] [#115](https://github.com/kubernetes/ingress/pull/115) add default_server to listen statement for default backend
- [X] [#114](https://github.com/kubernetes/ingress/pull/114) fix typo
- [X] [#113](https://github.com/kubernetes/ingress/pull/113) Add condition of enqueue and unit test cases for task.Queue
- [X] [#108](https://github.com/kubernetes/ingress/pull/108) annotations: print error and skip if malformed
- [X] [#107](https://github.com/kubernetes/ingress/pull/107) fix some wrong links of examples which to be used for nginx
- [X] [#103](https://github.com/kubernetes/ingress/pull/103) Update the nginx controller manifests
- [X] [#101](https://github.com/kubernetes/ingress/pull/101) Add unit test for strings.StringInSlice
- [X] [#99](https://github.com/kubernetes/ingress/pull/99) Update nginx to 1.11.8
- [X] [#97](https://github.com/kubernetes/ingress/pull/97) Fix gofmt
- [X] [#96](https://github.com/kubernetes/ingress/pull/96) Fix typo PassthrougBackends -> PassthroughBackends
- [X] [#95](https://github.com/kubernetes/ingress/pull/95) Deny location mapping in case of specific errors
- [X] [#94](https://github.com/kubernetes/ingress/pull/94) Add support to disable server_tokens directive
- [X] [#93](https://github.com/kubernetes/ingress/pull/93) Fix sort for catch all server
- [X] [#92](https://github.com/kubernetes/ingress/pull/92) Refactoring of nginx configuration deserialization
- [X] [#91](https://github.com/kubernetes/ingress/pull/91) Fix x-forwarded-port mapping
- [X] [#90](https://github.com/kubernetes/ingress/pull/90) fix the wrong link to build/test/release
- [X] [#89](https://github.com/kubernetes/ingress/pull/89) fix the wrong links to the examples and developer documentation
- [X] [#88](https://github.com/kubernetes/ingress/pull/88) Fix multiple tls hosts sharing the same secretName
- [X] [#86](https://github.com/kubernetes/ingress/pull/86) Update X-Forwarded-Port
- [X] [#82](https://github.com/kubernetes/ingress/pull/82) Fix incorrect X-Forwarded-Port for TLS
- [X] [#81](https://github.com/kubernetes/ingress/pull/81) Do not push containers to remote repo as part of test-e2e
- [X] [#78](https://github.com/kubernetes/ingress/pull/78) Fix #76: hardcode X-Forwarded-Port due to SSL Passthrough
- [X] [#77](https://github.com/kubernetes/ingress/pull/77) Add support for IPV6 in dns resolvers
- [X] [#66](https://github.com/kubernetes/ingress/pull/66) Start FAQ docs
- [X] [#65](https://github.com/kubernetes/ingress/pull/65) Support hostnames in Ingress status
- [X] [#64](https://github.com/kubernetes/ingress/pull/64) Sort whitelist list to avoid random orders
- [X] [#62](https://github.com/kubernetes/ingress/pull/62) Fix e2e make targets
- [X] [#61](https://github.com/kubernetes/ingress/pull/61) Ignore coverage profile files
- [X] [#58](https://github.com/kubernetes/ingress/pull/58) Fix "invalid port in upstream" on nginx controller
- [X] [#57](https://github.com/kubernetes/ingress/pull/57) Fix invalid port in upstream
- [X] [#54](https://github.com/kubernetes/ingress/pull/54) Expand developer docs
- [X] [#52](https://github.com/kubernetes/ingress/pull/52) fix typo in variable ProxyRealIPCIDR
- [X] [#44](https://github.com/kubernetes/ingress/pull/44) Bump nginx version to one higher than that in contrib
- [X] [#36](https://github.com/kubernetes/ingress/pull/36) Add nginx metrics to prometheus
- [X] [#34](https://github.com/kubernetes/ingress/pull/34) nginx: also listen on ipv6
- [X] [#32](https://github.com/kubernetes/ingress/pull/32) Restart nginx if master process dies
- [X] [#31](https://github.com/kubernetes/ingress/pull/31) Add healthz checker
- [X] [#25](https://github.com/kubernetes/ingress/pull/25) Fix a data race in TestFileWatcher
- [X] [#12](https://github.com/kubernetes/ingress/pull/12) Split implementations from generic code
- [X] [#10](https://github.com/kubernetes/ingress/pull/10) Copy Ingress history from kubernetes/contrib
- [X] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1571](https://github.com/kubernetes/contrib/pull/1571) use POD_NAMESPACE as a namespace in cli parameters
- [X] [#1591](https://github.com/kubernetes/contrib/pull/1591) Always listen on port 443, even without ingress rules
- [X] [#1596](https://github.com/kubernetes/contrib/pull/1596) Adapt nginx hash sizes to the number of ingress
- [X] [#1653](https://github.com/kubernetes/contrib/pull/1653) Update image version
- [X] [#1672](https://github.com/kubernetes/contrib/pull/1672) Add firewall rules and ing class clarifications
- [X] [#1711](https://github.com/kubernetes/contrib/pull/1711) Add function helpers to nginx template
- [X] [#1743](https://github.com/kubernetes/contrib/pull/1743) Allow customisation of the nginx proxy_buffer_size directive via ConfigMap
- [X] [#1749](https://github.com/kubernetes/contrib/pull/1749) Readiness probe that works behind a CP lb
- [X] [#1751](https://github.com/kubernetes/contrib/pull/1751) Add the name of the upstream in the log
- [X] [#1758](https://github.com/kubernetes/contrib/pull/1758) Update nginx to 1.11.4
- [X] [#1759](https://github.com/kubernetes/contrib/pull/1759) Add support for default backend in Ingress rule
- [X] [#1762](https://github.com/kubernetes/contrib/pull/1762) Add cloud detection
- [X] [#1766](https://github.com/kubernetes/contrib/pull/1766) Clarify the controller uses endpoints and not services
- [X] [#1767](https://github.com/kubernetes/contrib/pull/1767) Update godeps
- [X] [#1772](https://github.com/kubernetes/contrib/pull/1772) Avoid replacing nginx.conf file if the new configuration is invalid
- [X] [#1773](https://github.com/kubernetes/contrib/pull/1773) Add annotation to add CORS support
- [X] [#1786](https://github.com/kubernetes/contrib/pull/1786) Add docs about go template
- [X] [#1796](https://github.com/kubernetes/contrib/pull/1796) Add external authentication support using auth_request
- [X] [#1802](https://github.com/kubernetes/contrib/pull/1802) Initialize proxy_upstream_name variable
- [X] [#1806](https://github.com/kubernetes/contrib/pull/1806) Add docs about the log format
- [X] [#1808](https://github.com/kubernetes/contrib/pull/1808) WebSocket documentation
- [X] [#1847](https://github.com/kubernetes/contrib/pull/1847) Change structure of packages
- [X] Add annotation for custom upstream timeouts
- [X] Mutual TLS auth (https://github.com/kubernetes/contrib/issues/1870)

### 0.8.3

- [X] [#1450](https://github.com/kubernetes/contrib/pull/1450) Check for errors in nginx template
- [ ] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1467](https://github.com/kubernetes/contrib/pull/1467) Use ClientConfig to configure connection
- [X] [#1575](https://github.com/kubernetes/contrib/pull/1575) Update nginx to 1.11.3

### 0.8.2

- [X] [#1336](https://github.com/kubernetes/contrib/pull/1336) Add annotation to skip ingress rule
- [X] [#1338](https://github.com/kubernetes/contrib/pull/1338) Add HTTPS default backend
- [X] [#1351](https://github.com/kubernetes/contrib/pull/1351) Avoid generation of invalid ssl certificates
- [X] [#1379](https://github.com/kubernetes/contrib/pull/1379) improve nginx performance
- [X] [#1350](https://github.com/kubernetes/contrib/pull/1350) Improve performance (listen backlog=net.core.somaxconn)
- [X] [#1384](https://github.com/kubernetes/contrib/pull/1384) Unset Authorization header when proxying
- [X] [#1398](https://github.com/kubernetes/contrib/pull/1398) Mitigate HTTPoxy Vulnerability

### 0.8.1

- [X] [#1317](https://github.com/kubernetes/contrib/pull/1317) Fix duplicated real_ip_header
- [X] [#1315](https://github.com/kubernetes/contrib/pull/1315) Addresses #1314

### 0.8

- [X] [#1063](https://github.com/kubernetes/contrib/pull/1063) watches referenced tls secrets
- [X] [#850](https://github.com/kubernetes/contrib/pull/850) adds configurable SSL redirect nginx controller
- [X] [#1136](https://github.com/kubernetes/contrib/pull/1136) Fix nginx rewrite rule order
- [X] [#1144](https://github.com/kubernetes/contrib/pull/1144) Add cidr whitelist support
- [X] [#1230](https://github.com/kubernetes/contrib/pull/1130) Improve docs and examples
- [X] [#1258](https://github.com/kubernetes/contrib/pull/1258) Avoid sync without a reachable
- [X] [#1235](https://github.com/kubernetes/contrib/pull/1235) Fix stats by country in nginx status page
- [X] [#1236](https://github.com/kubernetes/contrib/pull/1236) Update nginx to add dynamic TLS records and spdy
- [X] [#1238](https://github.com/kubernetes/contrib/pull/1238) Add support for dynamic TLS records and spdy
- [X] [#1239](https://github.com/kubernetes/contrib/pull/1239) Add support for conditional log of urls
- [X] [#1253](https://github.com/kubernetes/contrib/pull/1253) Use delayed queue
- [X] [#1296](https://github.com/kubernetes/contrib/pull/1296) Fix formatting
- [X] [#1299](https://github.com/kubernetes/contrib/pull/1299) Fix formatting

### 0.7

- [X] [#898](https://github.com/kubernetes/contrib/pull/898) reorder locations. Location / must be the last one to avoid errors routing to subroutes
- [X] [#946](https://github.com/kubernetes/contrib/pull/946) Add custom authentication (Basic or Digest) to ingress rules
- [X] [#926](https://github.com/kubernetes/contrib/pull/926) Custom errors should be optional
- [X] [#1002](https://github.com/kubernetes/contrib/pull/1002) Use k8s probes (disable NGINX checks)
- [X] [#962](https://github.com/kubernetes/contrib/pull/962) Make optional http2
- [X] [#1054](https://github.com/kubernetes/contrib/pull/1054) force reload if some certificate change
- [X] [#958](https://github.com/kubernetes/contrib/pull/958) update NGINX to 1.11.0 and add digest module
- [X] [#960](https://github.com/kubernetes/contrib/issues/960) https://trac.nginx.org/nginx/changeset/ce94f07d50826fcc8d48f046fe19d59329420fdb/nginx
- [X] [#1057](https://github.com/kubernetes/contrib/pull/1057) Remove loadBalancer ip on shutdown
- [X] [#1079](https://github.com/kubernetes/contrib/pull/1079) path rewrite
- [X] [#1093](https://github.com/kubernetes/contrib/pull/1093) rate limiting
- [X] [#1102](https://github.com/kubernetes/contrib/pull/1102) geolocation of traffic in stats
- [X] [#884](https://github.com/kubernetes/contrib/issues/884) support services running ssl
- [X] [#930](https://github.com/kubernetes/contrib/issues/930) detect changes in configuration configmaps
