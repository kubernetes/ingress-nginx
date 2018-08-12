/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

package v1alpha1

import (
	"net"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration represents the NGINX Ingress controller configuration.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired configuration
	// +optional
	Spec ConfigurationSpec `json:"spec,omitempty"`

	// Current status of the configuration.
	// +optional
	Status ConfigurationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigurationList is a collection of configurations.
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// items is the list of Configuration.
	Items []Configuration `json:"items"`
}

// ConfigurationStatus represents the current state of a configuration.
type ConfigurationStatus struct {
	// Checksum contains a checksum of the configuration
	Checksum string `json:"checksum"`
}

// ConfigurationSpec describes how the job execution will look like and when it will actually run.
type ConfigurationSpec struct {
	Global      Global      `json:"global"`
	Log         Log         `json:"log"`
	Metrics     Metrics     `json:"metrics"`
	Snippets    Snippet     `json:"snippets"`
	SSL         SSL         `json:"ssl"`
	Opentracing Opentracing `json:"opentracing"`
	Upstream    Upstream    `json:"upstream"`
	WAF         WAF         `json:"waf"`
}

type LoadBalanceAlgorithm string

var (
	RoundRobin LoadBalanceAlgorithm = "round-robin"
	LeastConn  LoadBalanceAlgorithm = "least_conn"
	IPHash     LoadBalanceAlgorithm = "ip_hash"
	EWMA       LoadBalanceAlgorithm = "ewma"
)

type Global struct {
	// WorkerCpuAffinity bind nginx worker processes to CPUs this will improve response latency
	// http://nginx.org/en/docs/ngx_core_module.html#worker_cpu_affinity
	// By default this is disabled
	WorkerCpuAffinity string `json:"worker-cpu-affinity,omitempty"`

	// ClientHeaderBufferSize allows to configure a custom buffer
	// size for reading client request header
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_buffer_size
	ClientHeaderBufferSize string `json:"client-header-buffer-size"`

	// Defines a timeout for reading client request header, in seconds
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
	ClientHeaderTimeout int `json:"client-header-timeout,omitempty"`

	// Sets buffer size for reading client request body
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size
	ClientBodyBufferSize string `json:"client-body-buffer-size,omitempty"`

	// Defines a timeout for reading client request body, in seconds
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout
	ClientBodyTimeout int `json:"client-body-timeout,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_field_size
	// HTTP2MaxFieldSize Limits the maximum size of an HPACK-compressed request header field
	HTTP2MaxFieldSize string `json:"http2-max-field-size,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_header_size
	// HTTP2MaxHeaderSize Limits the maximum size of the entire request header list after HPACK decompression
	HTTP2MaxHeaderSize string `json:"http2-max-header-size,omitempty"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `json:"keep-alive,omitempty"`

	// Sets the maximum number of requests that can be served through one keep-alive connection.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests
	KeepAliveRequests int `json:"keep-alive-requests,omitempty"`

	// LargeClientHeaderBuffers Sets the maximum number and size of buffers used for reading
	// large client request header.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#large_client_header_buffers
	// Default: 4 8k
	LargeClientHeaderBuffers string `json:"large-client-header-buffers"`

	// If disabled, a worker process will accept one new connection at a time.
	// Otherwise, a worker process will accept all new connections at a time.
	// http://nginx.org/en/docs/ngx_core_module.html#multi_accept
	// Default: true
	EnableMultiAccept bool `json:"enable-multi-accept,omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// http://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `json:"max-worker-connections,omitempty"`

	// Sets the bucket size for the map variables hash tables.
	// Default value depends on the processor’s cache line size.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size
	MapHashBucketSize int `json:"map-hash-bucket-size,omitempty"`

	// EnableIPV6DNS disables IPv6 for nginx resolver
	EnableIPV6DNS bool `json:"enable-ipv6-dns"`

	// EnableIPV6 disable listening on ipv6 address
	EnableIPV6 bool `json:"enable-ipv6,omitempty"`

	// EnableUnderscoresInHeaders enables underscores in header names
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#underscores_in_headers
	// By default this is disabled
	EnableUnderscoresInHeaders bool `json:"enable-underscores-in-headers"`

	// IgnoreInvalidHeaders set if header fields with invalid names should be ignored
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#ignore_invalid_headers
	// By default this is enabled
	IgnoreInvalidHeaders bool `json:"ignore-invalid-headers"`

	// RetryNonIdempotent since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH)
	// in case of an error. The previous behavior can be restored using the value true
	RetryNonIdempotent bool `json:"retry-non-idempotent"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `json:"server-name-hash-max-size,omitempty"`

	// Size of the bucket for the server names hash tables
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `json:"server-name-hash-bucket-size,omitempty"`

	// Size of the bucket for the proxy headers hash tables
	// http://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_max_size
	ProxyHeadersHashMaxSize int `json:"proxy-headers-hash-max-size,omitempty"`

	// Maximum size of the bucket for the proxy headers hash tables
	// http://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_bucket_size
	ProxyHeadersHashBucketSize int `json:"proxy-headers-hash-bucket-size,omitempty"`

	// Enables or disables emitting nginx version in error messages and in the “Server” response header field.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens
	// Default: true
	ShowServerTokens bool `json:"server-tokens"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	EnableProxyProtocol bool `json:"enable-proxy-protocol"`

	// When use-proxy-protocol is enabled, sets the maximum time the connection handler will wait
	// to receive proxy headers.
	// Example '60s'
	ProxyProtocolHeaderTimeout time.Duration `json:"proxy-protocol-header-timeout"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// http://nginx.org/en/docs/http/ngx_http_gzip_module.html
	EnableGzip bool `json:"enable-gzip"`

	// Enables or disables the use of the nginx geoip module that creates variables with values depending on the client IP
	// http://nginx.org/en/docs/http/ngx_http_geoip_module.html
	EnableGeoIP bool `json:"enable-geoip"`

	// Enables or disables the use of the NGINX Brotli Module for compression
	// https://github.com/google/ngx_brotli
	EnableBrotli bool `json:"enable-brotli"`

	// Brotli Compression Level that will be used
	BrotliLevel int `json:"brotli-level"`

	// MIME Types that will be compressed on-the-fly using Brotli module
	BrotliTypes string `json:"brotli-types"`

	// Enables or disables the HTTP/2 support in secure connections
	// http://nginx.org/en/docs/http/ngx_http_v2_module.html
	// Default: true
	EnableHTTP2 bool `json:"enable-http2"`

	// gzip Compression Level that will be used
	GzipLevel int `json:"gzip-level"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `json:"gzip-types"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// http://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses string `json:"worker-processes"`

	// Defines a timeout for a graceful shutdown of worker processes
	// http://nginx.org/en/docs/ngx_core_module.html#worker_shutdown_timeout
	WorkerShutdownTimeout string `json:"worker-shutdown-timeout"`

	// Defines the load balancing algorithm to use. The default is round-robin
	LoadBalanceAlgorithm LoadBalanceAlgorithm `json:"load-balance"`

	// Sets the bucket size for the variables hash table.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#variables_hash_bucket_size
	VariablesHashBucketSize int `json:"variables-hash-bucket-size"`

	// Sets the maximum size of the variables hash table.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#variables_hash_max_size
	VariablesHashMaxSize int `json:"variables-hash-max-size"`

	// NginxStatusIpv4Whitelist has the list of cidr that are allowed to access
	// the /nginx_status endpoint of the "_" server
	NginxStatusIpv4Whitelist []string `json:"nginx-status-ipv4-whitelist,omitempty"`
	NginxStatusIpv6Whitelist []string `json:"nginx-status-ipv6-whitelist,omitempty"`

	// If UseProxyProtocol is enabled ProxyRealIPCIDR defines the default the IP/network address
	// of your external load balancer
	ProxyRealIPCIDR []string `json:"proxy-real-ip-cidr,omitempty"`

	// Sets the maximum size of the variables hash table.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#variables_hash_max_size
	LimitConnZoneVariable string `json:"limit-conn-zone-variable,omitempty"`

	// Sets the ipv4 addresses on which the server will accept requests.
	BindAddressIpv4 []net.IP `json:"bind-address-ipv4,omitempty"`

	// Sets the ipv6 addresses on which the server will accept requests.
	BindAddressIpv6 []net.IP `json:"bind-address-ipv6,omitempty"`

	// Sets the header field for identifying the originating IP address of a client
	// Default is X-Forwarded-For
	ForwardedForHeader string `json:"forwarded-for-header,omitempty"`

	// Append the remote address to the X-Forwarded-For header instead of replacing it
	// Default: false
	ComputeFullForwardedFor bool `json:"compute-full-forwarded-for,omitempty"`

	// If the request does not have a request-id, should we generate a random value?
	// Default: true
	EnableRequestId bool `json:"enable-request-id,omitempty"`

	// Adds an X-Original-Uri header with the original request URI to the backend request
	// Default: true
	ProxyAddOriginalUriHeader bool `json:"proxy-add-original-uri-header"`

	// HTTPRedirectCode sets the HTTP status code to be used in redirects.
	// Supported codes are 301,302,307 and 308
	// Default: 308
	HTTPRedirectCode int `json:"http-redirect-code"`

	// ReusePort instructs NGINX to create an individual listening socket for
	// each worker process (using the SO_REUSEPORT socket option), allowing a
	// kernel to distribute incoming connections between worker processes
	// Default: true
	ReusePort bool `json:"reuse-port"`

	// LimitReqStatusCode Sets the status code to return in response to rejected requests.
	// http://nginx.org/en/docs/http/ngx_http_limit_req_module.html#limit_req_status
	// Default: 503
	LimitReqStatusCode int `json:"limit-req-status-code"`

	// NoAuthLocations is a comma-separated list of locations that
	// should not get authenticated
	NoAuthLocations string `json:"no-auth-locations"`

	// EnableInfluxDB enables the nginx InfluxDB extension
	// http://github.com/influxdata/nginx-influxdb-module/
	// By default this is disabled
	EnableInfluxDB bool `json:"enable-influxdb"`

	// CustomHTTPErrors enables which HTTP codes should be passed for processing with the error_page directive
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page
	// By default this is disabled
	CustomHTTPErrors []int `json:"custom-http-errors,-"`

	// Name server/s used to resolve names of upstream servers into IP addresses.
	// The file /etc/resolv.conf is used as DNS resolution configuration.
	Resolver []net.IP

	// Enables or disables the specification of port in redirects
	// Default: false
	PortInRedirects bool `json:"port-in-redirects"`

	// Let's us choose a load balancing algorithm per ingress
	LoadBalancing string `json:"load-balance"`

	// WhitelistSourceRange allows limiting access to certain client addresses
	// http://nginx.org/en/docs/http/ngx_http_access_module.html
	WhitelistSourceRange []net.IPNet `json:"whitelist-source-range"`

	// Limits the rate of response transmission to a client.
	// The rate is specified in bytes per second. The zero value disables rate limiting.
	// The limit is set per a request, and so if a client simultaneously opens two connections,
	// the overall rate will be twice as much as the specified limit.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#limit_rate
	LimitRate int `json:"limit-rate"`

	// Sets the initial amount after which the further transmission of a response to a client will be rate limited.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#limit_rate_after
	LimitRateAfter int `json:"limit-rate-after"`
}

// Log ...
type Log struct {
	// EnableAccessLog enables the Access Log globally from NGINX ingress controller
	//http://nginx.org/en/docs/http/ngx_http_log_module.html
	EnableAccessLog bool `json:"enable-access-log"`

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `json:"error-log-level"`

	// FormatEscapeJSON enables json escaping
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	FormatEscapeJSON bool `json:"format-escape-json"`

	// FormatUpstream customize upstream log_format
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	FormatUpstream string `json:"format-upstream"`

	// FormatStream customize stream log_format
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	FormatStream string `json:"format-stream"`

	// SkipAccessLogURLs sets a list of URLs that should not appear in the NGINX access log
	// This is useful with urls like `/health` or `health-check` that make "complex" reading the logs
	// By default this list is empty
	SkipAccessLogURLs []string `json:"skip-access-log-urls"`

	File   LogFileConfiguration `json:"file"`
	Syslog SyslogConfiguration  `json:"syslog"`
}

// LogFileConfiguration ...
type LogFileConfiguration struct {
	// AccessLogPath sets the path of the access logs if enabled
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#access_log
	// By default access logs go to /var/log/nginx/access.log
	AccessLogPath string `json:"access-log-path"`

	// ErrorLogPath sets the path of the error logs
	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// By default error logs go to /var/log/nginx/error.log
	ErrorLogPath string `json:"error-log-path"`
}

/*
Latency = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
ResponseLength = []float64{100, 1000, 10000, 100000, 1000000}
RequestLength = []float64{1000, 10000, 100000, 1000000, 10000000}
*/

type Metrics struct {
	Enabled bool

	Latency        []float64
	ResponseLength []float64
	RequestLength  []float64
}

// SyslogConfiguration ...
type SyslogConfiguration struct {
	// Enabled enables the configuration for remote logging in NGINX
	Enabled bool `json:"enabled"`
	// Host FQDN or IP address where the logs should be sent
	Host string `json:"host"`
	// SyslogPort port
	Port int `json:"port"`
}

// Snippet ...
type Snippet struct {
	// Main adds custom configuration to the main section of the nginx configuration
	Main string `json:"main"`

	// HTTP adds custom configuration to the http section of the nginx configuration
	HTTP string `json:"http"`

	// Server adds custom configuration to all the servers in the nginx configuration
	Server string `json:"server"`

	// Location adds custom configuration to all the locations in the nginx configuration
	Location string `json:"location"`
}

// Opentracing ...
type Opentracing struct {
	// Enabled enables the nginx Opentracing extension
	// https://github.com/rnburn/nginx-opentracing
	// By default this is disabled
	Enabled bool                `json:"enabled"`
	Jaeger  JaegerConfiguration `json:"jaeger"`
	Zipkin  ZipkinConfiguration `json:"zipkin"`
}

// JaegerConfiguration ...
type JaegerConfiguration struct {
	// CollectorHost specifies the host to use when uploading traces
	CollectorHost string `json:"collector-host"`

	// CollectorPort specifies the port to use when uploading traces
	CollectorPort int `json:"collector-port"`

	// ServiceName specifies the service name to use for any traces created
	// Default: nginx
	ServiceName string `json:"service-name"`

	// SamplerType specifies the sampler to be used when sampling traces.
	// The available samplers are: const, probabilistic, ratelimiting, remote
	// Default: const
	SamplerType string `json:"sampler-type"`

	// SamplerParam specifies the argument to be passed to the sampler constructor
	// Default: 1
	SamplerParam string `json:"sampler-param"`
}

// ZipkinConfiguration ...
type ZipkinConfiguration struct {
	// CollectorHost specifies the host to use when uploading traces
	CollectorHost string `json:"collector-host"`

	// CollectorPort specifies the port to use when uploading traces
	CollectorPort int `json:"collector-port"`

	// ServiceName specifies the service name to use for any traces created
	// Default: nginx
	ServiceName string `json:"service-name"`

	// SampleRate specifies sampling rate for traces
	// Default: 1.0
	SampleRate float32 `json:"sample-rate"`
}

// WAF ...
type WAF struct {
	// EnableModsecurity enables the modsecurity module for NGINX
	// By default this is disabled
	EnableModsecurity bool `json:"enable-modsecurity"`

	// EnableModsecurity enables the OWASP ModSecurity Core Rule Set (CRS)
	// By default this is disabled
	EnableOWASPCoreRules bool `json:"enable-owasp-modsecurity-crs"`

	// EnableLuaRestyWAF disables lua-resty-waf globally regardless
	// of whether there's an ingress that has enabled the WAF using annotation
	EnableLuaRestyWAF bool `json:"enable-lua-resty-waf"`
}

type HSTS struct {
	// Enables or disables the header HSTS in servers running SSL
	Enabled bool `json:"enabled,omitempty"`

	// Enables or disables the use of HSTS in all the subdomains of the servername
	// Default: true
	IncludeSubdomains bool `json:"include-subdomains,omitempty"`

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be
	// accessed using HTTPS.
	MaxAge string `json:"max-age,omitempty"`

	// Enables or disables the preload attribute in HSTS feature
	Preload bool `json:"preload,omitempty"`
}

// SSL ...
type SSL struct {
	HSTS HSTS

	// Enables or disables the redirect (308) to the HTTPS port
	SSLRedirect bool `json:"ssl-redirect"`

	// Enables or disables the redirect (308) to the HTTPS port even without TLS cert
	// This is useful if doing SSL offloading outside of cluster eg AWS ELB
	ForceSSLRedirect bool `json:"force-ssl-redirect"`

	// NoTLSRedirectLocations is a comma-separated list of locations
	// that should not get redirected to TLS
	NoTLSRedirectLocations string `json:"no-tls-redirect-locations"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	Ciphers string `json:"ciphers,omitempty"`

	// Specifies a curve for ECDHE ciphers.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ecdh_curve
	ECDHCurve string `json:"ecdh-curve,omitempty"`

	// The secret that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
	// https://wiki.openssl.org/index.php/Diffie-Hellman_parameters
	// https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
	DHParam string `json:"dh-param,omitempty"`

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	Protocols string `json:"protocols,omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SessionCache bool `json:"session-cache,omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SessionCacheSize string `json:"session-cache-size,omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SessionTickets bool `json:"session-tickets,omitempty"`

	// Sets the secret key used to encrypt and decrypt TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	// By default, a randomly generated key is used.
	// Example: openssl rand 80 | openssl enc -A -base64
	SessionTicketKey string `json:"session-ticket-key,omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SessionTimeout string `json:"session-timeout,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	BufferSize string `json:"buffer-size,omitempty"`
}

type Upstream struct {
	// Sets the additional headers to pass to the backend.
	SetHeaders map[string]string `json:"proxy-set-headers,omitempty"`

	// HideHeaders sets additional header that will not be passed from the upstream
	// server to the client response
	// Default: empty
	HideHeaders []string `json:"hide-headers"`

	// EnableServerHeaderFromBackend enables the return of the header Server from
	// the backend instead of the generic nginx string.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_hide_header
	// By default this is disabled
	EnableServerHeaderFromBackend bool `json:"enable-server-header-from-backend"`

	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	BodySize string `json:"body-size"`

	// Enables or disables buffering of responses from the proxied server.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering
	Buffering string `json:"buffering"`

	// Sets the size of the buffer used for reading the first part of the response received from the
	// proxied server. This part usually contains a small response header.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size)
	BufferSize string `json:"buffer-size"`

	// Sets a text that should be changed in the domain attribute of the “Set-Cookie” header fields
	// of a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_domain
	CookieDomain string `json:"cookie-domain"`

	// Sets a text that should be changed in the path attribute of the “Set-Cookie” header fields of
	// a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_path
	CookiePath string `json:"cookie-path"`

	// Defines a timeout for establishing a connection with a proxied server.
	// It should be noted that this timeout cannot usually exceed 75 seconds.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout
	ConnectTimeout int `json:"connect-timeout"`

	// Time during which the specified number of unsuccessful attempts to communicate with
	// the server should happen to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	FailTimeout int `json:"fail-timeout"`

	// Enable stickiness by client-server mapping based on a NGINX variable, text or a combination of both.
	// A consistent hashing method will be used which ensures only a few keys would be remapped to different
	// servers on upstream group changes
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#hash
	HashBy string `json:"hash-by"`

	// Activates the cache for connections to upstream servers.
	// The connections parameter sets the maximum number of idle keepalive connections to
	// upstream servers that are preserved in the cache of each worker process. When this
	// number is exceeded, the least recently used connections are closed.
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive
	// Default: 32
	KeepaliveConnections int `json:"upstream-keepalive-connections,omitempty"`

	// Number of unsuccessful attempts to communicate with the server that should happen in the
	// duration set by the fail_timeout parameter to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	MaxFails int `json:"max-fails"`

	// Specifies in which cases a request should be passed to the next server.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream
	NextUpstream string `json:"next-upstream"`
	// Limits the number of possible tries for passing a request to the next server.
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream_tries
	NextUpstreamTries int `json:"next-upstream-tries"`

	// Timeout in seconds for reading a response from the proxied server. The timeout is set only between
	// two successive read operations, not for the transmission of the whole response
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout
	ReadTimeout int `json:"read-timeout"`

	// Sets the original text that should be changed in the "Location" and "Refresh" header fields of a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_redirect
	// Default: off
	RedirectFrom string `json:"redirect-from"`

	// Sets the replacement text that should be changed in the "Location" and "Refresh" header fields of a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_redirect
	// Default: ""
	RedirectTo string `json:"redirect-to"`

	// Enables or disables buffering of a client request body.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_request_buffering
	RequestBuffering string `json:"request-buffering"`

	// Timeout in seconds for transmitting a request to the proxied server. The timeout is set only between
	// two successive write operations, not for the transmission of the whole request.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout
	SendTimeout int `json:"send-timeout"`
}
