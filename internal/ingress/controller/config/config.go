/*
Copyright 2016 The Kubernetes Authors.

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

package config

import (
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/ingress-nginx/pkg/util/runtime"
)

// EnableSSLChainCompletion Autocomplete SSL certificate chains with missing intermediate CA certificates.
var EnableSSLChainCompletion = false

const (
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	bodySize = "1m"

	// https://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	errorLevel = "notice"

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.
	hstsMaxAge = "31536000"

	gzipTypes = "application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/javascript text/plain text/x-component"

	brotliTypes = "application/xml+rss application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/javascript text/plain text/x-component"

	logFormatUpstream = `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] [$proxy_alternative_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`

	logFormatStream = `[$remote_addr] [$time_local] $protocol $status $bytes_sent $bytes_received $session_time`

	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256"

	// SSL enabled protocols to use
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1.2 TLSv1.3"

	// Disable TLS 1.3 early data
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_early_data
	sslEarlyData = false

	// Time during which a client may reuse the session parameters stored in a cache.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"

	// Parameters for a shared memory zone that will keep states for various keys.
	// https://nginx.org/en/docs/http/ngx_http_limit_conn_module.html#limit_conn_zone
	defaultLimitConnZoneVariable = "$binary_remote_addr"
)

// Configuration represents the content of nginx.conf file
type Configuration struct {
	defaults.Backend `json:",squash"` //nolint:staticcheck // Ignore unknown JSON option "squash" error

	// AllowSnippetAnnotations enable users to add their own snippets via ingress annotation.
	// If disabled, only snippets added via ConfigMap are added to ingress.
	AllowSnippetAnnotations bool `json:"allow-snippet-annotations"`

	// AllowCrossNamespaceResources enables users to consume cross namespace resource on annotations
	// Case disabled, attempts to use secrets or configmaps from a namespace different from Ingress will
	// be denied
	AllowCrossNamespaceResources bool `json:"allow-cross-namespace-resources"`

	// AnnotationsRiskLevel represents the risk accepted on an annotation. If the risk is, for instance `Medium`, annotations
	// with risk High and Critical will not be accepted.
	AnnotationsRiskLevel string `json:"annotations-risk-level"`

	// AnnotationValueWordBlocklist defines words that should not be part of an user annotation value
	// (can be used to run arbitrary code or configs, for example) and that should be dropped.
	// This list should be separated by "," character
	AnnotationValueWordBlocklist string `json:"annotation-value-word-blocklist"`

	// Sets the name of the configmap that contains the headers to pass to the client
	AddHeaders string `json:"add-headers,omitempty"`

	// AllowBackendServerHeader enables the return of the header Server from the backend
	// instead of the generic nginx string.
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_hide_header
	// By default this is disabled
	AllowBackendServerHeader bool `json:"allow-backend-server-header"`

	// AccessLogParams sets additional params for access_log
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#access_log
	// By default it's empty
	AccessLogParams string `json:"access-log-params,omitempty"`

	// EnableAccessLogForDefaultBackend enable access_log for default backend
	// By default this is disabled
	EnableAccessLogForDefaultBackend bool `json:"enable-access-log-for-default-backend"`

	// EnableAuthAccessLog enable auth access log
	// By default this is disabled
	EnableAuthAccessLog bool `json:"enable-auth-access-log"`

	// AccessLogPath sets the path of the access logs for both http and stream contexts if enabled
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#access_log
	// https://nginx.org/en/docs/stream/ngx_stream_log_module.html#access_log
	// By default access logs go to /var/log/nginx/access.log
	AccessLogPath string `json:"access-log-path,omitempty"`

	// HTTPAccessLogPath sets the path of the access logs for http context globally if enabled
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#access_log
	HTTPAccessLogPath string `json:"http-access-log-path,omitempty"`

	// StreamAccessLogPath sets the path of the access logs for stream context globally if enabled
	// https://nginx.org/en/docs/stream/ngx_stream_log_module.html#access_log
	StreamAccessLogPath string `json:"stream-access-log-path,omitempty"`

	// WorkerCPUAffinity bind nginx worker processes to CPUs this will improve response latency
	// https://nginx.org/en/docs/ngx_core_module.html#worker_cpu_affinity
	// By default this is disabled
	WorkerCPUAffinity string `json:"worker-cpu-affinity,omitempty"`
	// ErrorLogPath sets the path of the error logs
	// https://nginx.org/en/docs/ngx_core_module.html#error_log
	// By default error logs go to /var/log/nginx/error.log
	ErrorLogPath string `json:"error-log-path,omitempty"`

	// EnableModsecurity enables the modsecurity module for NGINX
	// By default this is disabled
	EnableModsecurity bool `json:"enable-modsecurity"`

	// EnableOCSP enables the OCSP support in SSL connections
	// By default this is disabled
	EnableOCSP bool `json:"enable-ocsp"`

	// EnableOWASPCoreRules enables the OWASP ModSecurity Core Rule Set (CRS)
	// By default this is disabled
	EnableOWASPCoreRules bool `json:"enable-owasp-modsecurity-crs"`

	// ModSecuritySnippet adds custom rules to modsecurity section of nginx configuration
	ModsecuritySnippet string `json:"modsecurity-snippet"`

	// ClientHeaderBufferSize allows to configure a custom buffer
	// size for reading client request header
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_buffer_size
	ClientHeaderBufferSize string `json:"client-header-buffer-size"`

	// Defines a timeout for reading client request header, in seconds
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
	ClientHeaderTimeout int `json:"client-header-timeout,omitempty"`

	// Sets buffer size for reading client request body
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size
	ClientBodyBufferSize string `json:"client-body-buffer-size,omitempty"`

	// Defines a timeout for reading client request body, in seconds
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout
	ClientBodyTimeout int `json:"client-body-timeout,omitempty"`

	// DisableAccessLog disables the Access Log globally for both HTTP and Stream contexts from NGINX ingress controller
	// https://nginx.org/en/docs/http/ngx_http_log_module.html
	// https://nginx.org/en/docs/stream/ngx_stream_log_module.html
	DisableAccessLog bool `json:"disable-access-log,omitempty"`

	// DisableHTTPAccessLog disables the Access Log for http context globally from NGINX ingress controller
	// https://nginx.org/en/docs/http/ngx_http_log_module.html
	DisableHTTPAccessLog bool `json:"disable-http-access-log,omitempty"`

	// DisableStreamAccessLog disables the Access Log for stream context globally from NGINX ingress controller
	// https://nginx.org/en/docs/stream/ngx_stream_log_module.html
	DisableStreamAccessLog bool `json:"disable-stream-access-log,omitempty"`

	// DisableIpv6DNS disables IPv6 for nginx resolver
	DisableIpv6DNS bool `json:"disable-ipv6-dns"`

	// DisableIpv6 disable listening on ipv6 address
	DisableIpv6 bool `json:"disable-ipv6,omitempty"`

	// EnableUnderscoresInHeaders enables underscores in header names
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#underscores_in_headers
	// By default this is disabled
	EnableUnderscoresInHeaders bool `json:"enable-underscores-in-headers"`

	// IgnoreInvalidHeaders set if header fields with invalid names should be ignored
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#ignore_invalid_headers
	// By default this is enabled
	IgnoreInvalidHeaders bool `json:"ignore-invalid-headers"`

	// RetryNonIdempotent since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH)
	// in case of an error. The previous behavior can be restored using the value true
	RetryNonIdempotent bool `json:"retry-non-idempotent"`

	// https://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `json:"error-log-level,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_field_size
	// HTTP2MaxFieldSize Limits the maximum size of an HPACK-compressed request header field
	// Deprecated: HTTP2MaxFieldSize is deprecated.
	HTTP2MaxFieldSize string `json:"http2-max-field-size,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_header_size
	// HTTP2MaxHeaderSize Limits the maximum size of the entire request header list after HPACK decompression
	// Deprecated: HTTP2MaxHeaderSize is deprecated.
	HTTP2MaxHeaderSize string `json:"http2-max-header-size,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_requests
	// HTTP2MaxRequests Sets the maximum number of requests (including push requests) that can be served
	// through one HTTP/2 connection, after which the next client request will lead to connection closing
	// and the need of establishing a new connection.
	// Deprecated: HTTP2MaxRequests is deprecated.
	HTTP2MaxRequests int `json:"http2-max-requests,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_concurrent_streams
	// Sets the maximum number of concurrent HTTP/2 streams in a connection.
	HTTP2MaxConcurrentStreams int `json:"http2-max-concurrent-streams,omitempty"`

	// Enables or disables the header HSTS in servers running SSL
	HSTS bool `json:"hsts,omitempty"`

	// Enables or disables the use of HSTS in all the subdomains of the servername
	// Default: true
	HSTSIncludeSubdomains bool `json:"hsts-include-subdomains,omitempty"`

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be
	// accessed using HTTPS.
	HSTSMaxAge string `json:"hsts-max-age,omitempty"`

	// Enables or disables the preload attribute in HSTS feature
	HSTSPreload bool `json:"hsts-preload,omitempty"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `json:"keep-alive,omitempty"`

	// Sets the maximum number of requests that can be served through one keep-alive connection.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests
	KeepAliveRequests int `json:"keep-alive-requests,omitempty"`

	// LargeClientHeaderBuffers Sets the maximum number and size of buffers used for reading
	// large client request header.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#large_client_header_buffers
	// Default: 4 8k
	LargeClientHeaderBuffers string `json:"large-client-header-buffers"`

	// Disable all escaping
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatEscapeNone bool `json:"log-format-escape-none,omitempty"`

	// Enable json escaping
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatEscapeJSON bool `json:"log-format-escape-json,omitempty"`

	// Customize upstream log_format
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatUpstream string `json:"log-format-upstream,omitempty"`

	// Customize stream log_format
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	LogFormatStream string `json:"log-format-stream,omitempty"`

	// If disabled, a worker process will accept one new connection at a time.
	// Otherwise, a worker process will accept all new connections at a time.
	// https://nginx.org/en/docs/ngx_core_module.html#multi_accept
	// Default: true
	EnableMultiAccept bool `json:"enable-multi-accept,omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// https://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `json:"max-worker-connections,omitempty"`

	// Maximum number of files that can be opened by each worker process.
	// https://nginx.org/en/docs/ngx_core_module.html#worker_rlimit_nofile
	MaxWorkerOpenFiles int `json:"max-worker-open-files,omitempty"`

	// Sets the bucket size for the map variables hash tables.
	// Default value depends on the processor’s cache line size.
	// https://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size
	MapHashBucketSize int `json:"map-hash-bucket-size,omitempty"`

	// NginxStatusIpv4Whitelist has the list of cidr that are allowed to access
	// the /nginx_status endpoint of the "_" server
	NginxStatusIpv4Whitelist []string `json:"nginx-status-ipv4-whitelist,omitempty"`
	NginxStatusIpv6Whitelist []string `json:"nginx-status-ipv6-whitelist,omitempty"`

	// If UseProxyProtocol is enabled ProxyRealIPCIDR defines the default the IP/network address
	// of your external load balancer
	ProxyRealIPCIDR []string `json:"proxy-real-ip-cidr,omitempty"`

	// Sets the name of the configmap that contains the headers to pass to the backend
	ProxySetHeaders string `json:"proxy-set-headers,omitempty"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// https://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `json:"server-name-hash-max-size,omitempty"`

	// Size of the bucket for the server names hash tables
	// https://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `json:"server-name-hash-bucket-size,omitempty"`

	// Size of the bucket for the proxy headers hash tables
	// https://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_max_size
	ProxyHeadersHashMaxSize int `json:"proxy-headers-hash-max-size,omitempty"`

	// Maximum size of the bucket for the proxy headers hash tables
	// https://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_bucket_size
	ProxyHeadersHashBucketSize int `json:"proxy-headers-hash-bucket-size,omitempty"`

	// Enables or disables emitting nginx version in error messages and in the “Server” response header field.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens
	// Default: false
	ShowServerTokens bool `json:"server-tokens"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	SSLCiphers string `json:"ssl-ciphers,omitempty"`

	// Specifies a curve for ECDHE ciphers.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ecdh_curve
	SSLECDHCurve string `json:"ssl-ecdh-curve,omitempty"`

	// The secret that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
	// https://wiki.openssl.org/index.php/Diffie-Hellman_parameters
	// https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
	SSLDHParam string `json:"ssl-dh-param,omitempty"`

	// SSL enabled protocols to use
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	SSLProtocols string `json:"ssl-protocols,omitempty"`

	// Enables or disable TLS 1.3 early data.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_early_data
	SSLEarlyData bool `json:"ssl-early-data,omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCache bool `json:"ssl-session-cache,omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCacheSize string `json:"ssl-session-cache-size,omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SSLSessionTickets bool `json:"ssl-session-tickets,omitempty"`

	// Sets the secret key used to encrypt and decrypt TLS session tickets.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	// By default, a randomly generated key is used.
	// Example: openssl rand 80 | openssl enc -A -base64
	SSLSessionTicketKey string `json:"ssl-session-ticket-key,omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SSLSessionTimeout string `json:"ssl-session-timeout,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	SSLBufferSize string `json:"ssl-buffer-size,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_reject_handshake
	// If enabled, SSL handshakes to an invalid virtualhost will be rejected
	// Default: false
	SSLRejectHandshake bool `json:"ssl-reject-handshake"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	UseProxyProtocol bool `json:"use-proxy-protocol,omitempty"`

	// When use-proxy-protocol is enabled, sets the maximum time the connection handler will wait
	// to receive proxy headers.
	// Example '60s'
	ProxyProtocolHeaderTimeout time.Duration `json:"proxy-protocol-header-timeout,omitempty"`

	// Enables or disables the directive aio_write that writes files asynchronously
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#aio_write
	EnableAioWrite bool `json:"enable-aio-write,omitempty"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// https://nginx.org/en/docs/http/ngx_http_gzip_module.html
	UseGzip bool `json:"use-gzip,omitempty"`

	// UseGeoIP2 enables the geoip2 module for NGINX
	// By default this is disabled
	UseGeoIP2 bool `json:"use-geoip2,omitempty"`

	// GeoIP2AutoReloadMinutes enables autoreload on geoip2 setting the interval in minutes
	// By default this is disabled using 0
	GeoIP2AutoReloadMinutes int `json:"geoip2-autoreload-in-minutes,omitempty"`

	// Enables or disables the use of the NGINX Brotli Module for compression
	// https://github.com/google/ngx_brotli
	EnableBrotli bool `json:"enable-brotli,omitempty"`

	// Brotli Compression Level that will be used
	BrotliLevel int `json:"brotli-level,omitempty"`

	// Minimum length of responses, in bytes, that will be eligible for brotli compression
	BrotliMinLength int `json:"brotli-min-length,omitempty"`

	// MIME Types that will be compressed on-the-fly using Brotli module
	BrotliTypes string `json:"brotli-types,omitempty"`

	// Enables or disables the HTTP/2 support in secure connections
	// https://nginx.org/en/docs/http/ngx_http_v2_module.html
	// Default: true
	UseHTTP2 bool `json:"use-http2,omitempty"`

	// Disables gzipping of responses for requests with "User-Agent" header fields matching any of
	// the specified regular expressions.
	// https://nginx.org/en/docs/http/ngx_http_gzip_module.html#gzip_disable
	GzipDisable string `json:"gzip-disable,omitempty"`

	// gzip Compression Level that will be used
	GzipLevel int `json:"gzip-level,omitempty"`

	// Minimum length of responses to be sent to the client before it is eligible
	// for gzip compression, in bytes.
	GzipMinLength int `json:"gzip-min-length,omitempty"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `json:"gzip-types,omitempty"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// https://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses string `json:"worker-processes,omitempty"`

	// Defines whether multiple concurrent reloads of worker processes should occur.
	// Set this to false to prevent more than n x 2 workers to exist at any time, to avoid potential OOM situations and high CPU load
	// With this setting on false, configuration changes in the queue will be re-queued with an exponential backoff, until the number of worker process is the expected value.
	// By default new worker processes are spawned every time there's a change that cannot be applied dynamically with no upper limit to the number of running workers
	// https://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerSerialReloads bool `json:"enable-serial-reloads,omitempty"`

	// Defines a timeout for a graceful shutdown of worker processes
	// https://nginx.org/en/docs/ngx_core_module.html#worker_shutdown_timeout
	WorkerShutdownTimeout string `json:"worker-shutdown-timeout,omitempty"`

	// Sets the bucket size for the variables hash table.
	// https://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size
	VariablesHashBucketSize int `json:"variables-hash-bucket-size,omitempty"`

	// Sets the maximum size of the variables hash table.
	// https://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_max_size
	VariablesHashMaxSize int `json:"variables-hash-max-size,omitempty"`

	// Activates the cache for connections to upstream servers.
	// The connections parameter sets the maximum number of idle keepalive connections to
	// upstream servers that are preserved in the cache of each worker process. When this
	// number is exceeded, the least recently used connections are closed.
	// https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive
	UpstreamKeepaliveConnections int `json:"upstream-keepalive-connections,omitempty"`

	// Sets the maximum time during which requests can be processed through one keepalive connection
	// https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_time
	UpstreamKeepaliveTime string `json:"upstream-keepalive-time,omitempty"`

	// Sets a timeout during which an idle keepalive connection to an upstream server will stay open.
	// https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_timeout
	UpstreamKeepaliveTimeout int `json:"upstream-keepalive-timeout,omitempty"`

	// Sets the maximum number of requests that can be served through one keepalive connection.
	// After the maximum number of requests is made, the connection is closed.
	// https://nginx.org/en/docs/http/ngx_http_upstream_module.html#keepalive_requests
	UpstreamKeepaliveRequests int `json:"upstream-keepalive-requests,omitempty"`

	// Sets the maximum size of the variables hash table.
	// https://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_max_size
	LimitConnZoneVariable string `json:"limit-conn-zone-variable,omitempty"`

	// Sets the timeout between two successive read or write operations on client or proxied server connections.
	// If no data is transmitted within this time, the connection is closed.
	// https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_timeout
	ProxyStreamTimeout string `json:"proxy-stream-timeout,omitempty"`

	// When a connection to the proxied server cannot be established, determines whether
	// a client connection will be passed to the next server.
	// https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_next_upstream
	ProxyStreamNextUpstream bool `json:"proxy-stream-next-upstream,omitempty"`

	// Limits the time allowed to pass a connection to the next server.
	// The 0 value turns off this limitation.
	// https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_next_upstream_timeout
	ProxyStreamNextUpstreamTimeout string `json:"proxy-stream-next-upstream-timeout,omitempty"`

	// Limits the number of possible tries a request should be passed to the next server.
	// The 0 value turns off this limitation.
	// https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_next_upstream_tries
	ProxyStreamNextUpstreamTries int `json:"proxy-stream-next-upstream-tries,omitempty"`

	// Sets the number of datagrams expected from the proxied server in response
	// to the client request if the UDP protocol is used.
	// https://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_responses
	// Default: 1
	ProxyStreamResponses int `json:"proxy-stream-responses,omitempty"`

	// Modifies the HTTP version the proxy uses to interact with the backend.
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_http_version
	ProxyHTTPVersion string `json:"proxy-http-version"`

	// Disables NGINX proxy-intercept-errors when error_page/custom-http-errors are set
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors
	DisableProxyInterceptErrors bool `json:"disable-proxy-intercept-errors,omitempty"`

	// Disable absolute redirects and enables relative redirects.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#absolute_redirect
	RelativeRedirects bool `json:"relative-redirects"`

	// Sets the ipv4 addresses on which the server will accept requests.
	BindAddressIpv4 []string `json:"bind-address-ipv4,omitempty"`

	// Sets the ipv6 addresses on which the server will accept requests.
	BindAddressIpv6 []string `json:"bind-address-ipv6,omitempty"`

	// Sets whether to use incoming X-Forwarded headers.
	UseForwardedHeaders bool `json:"use-forwarded-headers"`

	// Sets whether to enable the real ip module
	EnableRealIP bool `json:"enable-real-ip"`

	// Sets the header field for identifying the originating IP address of a client
	// Default is X-Forwarded-For
	ForwardedForHeader string `json:"forwarded-for-header,omitempty"`

	// Sets the name of the intermediate header used to determine the client's originating IP
	// when both use-proxy-protocol and use-forwarded-headers are enabled. This doesn't impact
	// functionality and should not typically be modified.
	// Default is X-Forwarded-For-Proxy-Protocol
	ForwardedForProxyProtocolHeader string `json:"forwarded-for-proxy-protocol-header,omitempty"`

	// Sets the header field for identifying the originating Host header of a client
	// Default is X-Forwarded-Host
	ForwardedHostHeader string `json:"forwarded-host-header,omitempty"`

	// Append the remote address to the X-Forwarded-For header instead of replacing it
	// Default: false
	ComputeFullForwardedFor bool `json:"compute-full-forwarded-for,omitempty"`

	// If the request does not have a request-id, should we generate a random value?
	// Default: true
	GenerateRequestID bool `json:"generate-request-id,omitempty"`

	// Adds an X-Original-Uri header with the original request URI to the backend request
	// Default: false
	ProxyAddOriginalURIHeader bool `json:"proxy-add-original-uri-header"`

	// EnableOpentelemetry enables the nginx Opentelemetry extension
	// By default this is disabled
	EnableOpentelemetry bool `json:"enable-opentelemetry"`

	// OpentelemetryConfig sets the opentelemetry config file
	// Default: /etc/ingress-controller/telemetry/opentelemetry.toml
	OpentelemetryConfig string `json:"opentelemetry-config"`

	// OpentelemetryOperationName specifies a custom name for the server span
	OpentelemetryOperationName string `json:"opentelemetry-operation-name"`

	// OpentelemetryTrustIncomingSpan sets whether or not to trust incoming trace spans
	// If false, incoming span headers will be rejected
	// Default: true
	OpentelemetryTrustIncomingSpan bool `json:"opentelemetry-trust-incoming-span"`

	// OtlpCollectorHost specifies the host to use when uploading traces
	OtlpCollectorHost string `json:"otlp-collector-host"`

	// OtlpCollectorPort specifies the port to use when uploading traces
	// Default: 4317
	OtlpCollectorPort string `json:"otlp-collector-port"`

	// OtelServiceName specifies the service name to use for any traces created
	// Default: nginx
	OtelServiceName string `json:"otel-service-name"`

	// OtelSampler specifies the sampler to use for any traces created
	// Default: AlwaysOn
	OtelSampler string `json:"otel-sampler"`

	// OtelSamplerRatio specifies the sampler ratio to use for any traces created
	// Default: 0.01
	OtelSamplerRatio float32 `json:"otel-sampler-ratio"`

	// OtelSamplerParentBased specifies the parent based sampler to be used for any traces created
	// Default: true
	OtelSamplerParentBased bool `json:"otel-sampler-parent-based"`

	// MaxQueueSize specifies the max queue size for uploading traces
	// Default: 2048
	OtelMaxQueueSize int32 `json:"otel-max-queuesize"`

	// ScheduleDelayMillis specifies the max delay between uploading traces
	// Default: 5000
	OtelScheduleDelayMillis int32 `json:"otel-schedule-delay-millis"`

	// MaxExportBatchSize specifies the max export batch size to used when uploading traces
	// Default: 512
	OtelMaxExportBatchSize int32 `json:"otel-max-export-batch-size"`

	// MainSnippet adds custom configuration to the main section of the nginx configuration
	MainSnippet string `json:"main-snippet"`

	// HTTPSnippet adds custom configuration to the http section of the nginx configuration
	HTTPSnippet string `json:"http-snippet"`

	// ServerSnippet adds custom configuration to all the servers in the nginx configuration
	ServerSnippet string `json:"server-snippet"`

	// StreamSnippet adds custom configuration to the stream section of the nginx configuration
	StreamSnippet string `json:"stream-snippet"`

	// LocationSnippet adds custom configuration to all the locations in the nginx configuration
	LocationSnippet string `json:"location-snippet"`

	// HTTPRedirectCode sets the HTTP status code to be used in redirects.
	// Supported codes are 301,302,307 and 308
	// Default: 308
	HTTPRedirectCode int `json:"http-redirect-code"`

	// ReusePort instructs NGINX to create an individual listening socket for
	// each worker process (using the SO_REUSEPORT socket option), allowing a
	// kernel to distribute incoming connections between worker processes
	// Default: true
	ReusePort bool `json:"reuse-port"`

	// HideHeaders sets additional header that will not be passed from the upstream
	// server to the client response
	// Default: empty
	HideHeaders []string `json:"hide-headers"`

	// LimitReqStatusCode Sets the status code to return in response to rejected requests.
	// https://nginx.org/en/docs/http/ngx_http_limit_req_module.html#limit_req_status
	// Default: 503
	LimitReqStatusCode int `json:"limit-req-status-code"`

	// LimitConnStatusCode Sets the status code to return in response to rejected connections.
	// https://nginx.org/en/docs/http/ngx_http_limit_conn_module.html#limit_conn_status
	// Default: 503
	LimitConnStatusCode int `json:"limit-conn-status-code"`

	// EnableSyslog enables the configuration for remote logging in NGINX
	EnableSyslog bool `json:"enable-syslog"`
	// SyslogHost FQDN or IP address where the logs should be sent
	SyslogHost string `json:"syslog-host"`
	// SyslogPort port
	SyslogPort int `json:"syslog-port"`

	// NoTLSRedirectLocations is a comma-separated list of locations
	// that should not get redirected to TLS
	NoTLSRedirectLocations string `json:"no-tls-redirect-locations"`

	// NoAuthLocations is a comma-separated list of locations that
	// should not get authenticated
	NoAuthLocations string `json:"no-auth-locations"`

	// GlobalExternalAuth indicates the access to all locations requires
	// authentication using an external provider
	// +optional
	GlobalExternalAuth GlobalExternalAuth `json:"global-external-auth"`

	// Checksum contains a checksum of the configmap configuration
	Checksum string `json:"-"`

	// Block all requests from given IPs
	BlockCIDRs []string `json:"block-cidrs"`

	// Block all requests with given User-Agent headers
	BlockUserAgents []string `json:"block-user-agents"`

	// Block all requests with given Referer headers
	BlockReferers []string `json:"block-referers"`

	// Lua shared dict configuration data / certificate data
	LuaSharedDicts map[string]int `json:"lua-shared-dicts"`

	// DefaultSSLCertificate holds the default SSL certificate to use in the configuration
	// It can be the fake certificate or the one behind the flag --default-ssl-certificate
	DefaultSSLCertificate *ingress.SSLCert `json:"-"`

	// ProxySSLLocationOnly controls whether the proxy-ssl parameters defined in the
	// proxy-ssl-* annotations are applied on location level only in the nginx.conf file
	// Default is that those are applied on server level, too
	ProxySSLLocationOnly bool `json:"proxy-ssl-location-only"`

	// DefaultType Sets the default MIME type of a response.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#default_type
	// Default: text/html
	DefaultType string `json:"default-type"`

	// DebugConnections Enables debugging log for selected client connections
	// https://nginx.org/en/docs/ngx_core_module.html#debug_connection
	// Default: ""
	DebugConnections []string `json:"debug-connections"`

	// StrictValidatePathType enable the strict validation of Ingress Paths
	// It enforces that pathType of type Exact or Prefix should start with / and contain only
	// alphanumeric chars, "-", "_", "/".In case of additional characters,
	// like used on Rewrite configurations the user should use pathType as ImplementationSpecific
	StrictValidatePathType bool `json:"strict-validate-path-type"`

	// GRPCBufferSizeKb Sets the size of the buffer used for reading the response received
	// from the gRPC server. The response is passed to the client synchronously,
	// as soon as it is received.
	GRPCBufferSizeKb int `json:"grpc-buffer-size-kb"`
}

// NewDefault returns the default nginx configuration
func NewDefault() Configuration {
	defIPCIDR := make([]string, 0)
	defBindAddress := make([]string, 0)
	defBlockEntity := make([]string, 0)
	defNginxStatusIpv4Whitelist := make([]string, 0)
	defNginxStatusIpv6Whitelist := make([]string, 0)
	defResponseHeaders := make([]string, 0)
	defIPCIDR = append(defIPCIDR, "0.0.0.0/0")
	defNginxStatusIpv4Whitelist = append(defNginxStatusIpv4Whitelist, "127.0.0.1")
	defNginxStatusIpv6Whitelist = append(defNginxStatusIpv6Whitelist, "::1")
	defProxyDeadlineDuration := time.Duration(5) * time.Second
	defGlobalExternalAuth := GlobalExternalAuth{"", "", "", "", "", append(defResponseHeaders, ""), "", "", "", []string{}, map[string]string{}, false}

	cfg := Configuration{
		AllowSnippetAnnotations:          false,
		AllowCrossNamespaceResources:     false,
		AllowBackendServerHeader:         false,
		AnnotationValueWordBlocklist:     "",
		AnnotationsRiskLevel:             "High",
		AccessLogPath:                    "/var/log/nginx/access.log",
		AccessLogParams:                  "",
		EnableAccessLogForDefaultBackend: false,
		EnableAuthAccessLog:              false,
		WorkerCPUAffinity:                "",
		ErrorLogPath:                     "/var/log/nginx/error.log",
		BlockCIDRs:                       defBlockEntity,
		BlockUserAgents:                  defBlockEntity,
		BlockReferers:                    defBlockEntity,
		BrotliLevel:                      4,
		BrotliMinLength:                  20,
		BrotliTypes:                      brotliTypes,
		ClientHeaderBufferSize:           "1k",
		ClientHeaderTimeout:              60,
		ClientBodyBufferSize:             "8k",
		ClientBodyTimeout:                60,
		EnableUnderscoresInHeaders:       false,
		ErrorLogLevel:                    errorLevel,
		UseForwardedHeaders:              false,
		EnableRealIP:                     false,
		ForwardedForHeader:               "X-Forwarded-For",
		ForwardedForProxyProtocolHeader:  "X-Forwarded-For-Proxy-Protocol",
		ForwardedHostHeader:              "X-Forwarded-Host",
		ComputeFullForwardedFor:          false,
		ProxyAddOriginalURIHeader:        false,
		GenerateRequestID:                true,
		HTTP2MaxFieldSize:                "",
		HTTP2MaxHeaderSize:               "",
		HTTP2MaxRequests:                 0,
		HTTP2MaxConcurrentStreams:        128,
		HTTPRedirectCode:                 308,
		HSTS:                             true,
		HSTSIncludeSubdomains:            true,
		HSTSMaxAge:                       hstsMaxAge,
		HSTSPreload:                      false,
		IgnoreInvalidHeaders:             true,
		GzipLevel:                        1,
		GzipMinLength:                    256,
		GzipTypes:                        gzipTypes,
		KeepAlive:                        75,
		KeepAliveRequests:                1000,
		LargeClientHeaderBuffers:         "4 8k",
		LogFormatEscapeJSON:              false,
		LogFormatStream:                  logFormatStream,
		LogFormatUpstream:                logFormatUpstream,
		EnableMultiAccept:                true,
		MaxWorkerConnections:             16384,
		MaxWorkerOpenFiles:               0,
		MapHashBucketSize:                64,
		NginxStatusIpv4Whitelist:         defNginxStatusIpv4Whitelist,
		NginxStatusIpv6Whitelist:         defNginxStatusIpv6Whitelist,
		ProxyRealIPCIDR:                  defIPCIDR,
		ProxyProtocolHeaderTimeout:       defProxyDeadlineDuration,
		ServerNameHashMaxSize:            1024,
		ProxyHeadersHashMaxSize:          512,
		ProxyHeadersHashBucketSize:       64,
		ProxyStreamResponses:             1,
		ReusePort:                        true,
		ShowServerTokens:                 false,
		SSLBufferSize:                    sslBufferSize,
		SSLCiphers:                       sslCiphers,
		SSLECDHCurve:                     "auto",
		SSLProtocols:                     sslProtocols,
		SSLEarlyData:                     sslEarlyData,
		SSLRejectHandshake:               false,
		SSLSessionCache:                  true,
		SSLSessionCacheSize:              sslSessionCacheSize,
		SSLSessionTickets:                false,
		SSLSessionTimeout:                sslSessionTimeout,
		EnableBrotli:                     false,
		EnableAioWrite:                   true,
		UseGzip:                          false,
		UseGeoIP2:                        false,
		GeoIP2AutoReloadMinutes:          0,
		WorkerProcesses:                  strconv.Itoa(runtime.NumCPU()),
		WorkerSerialReloads:              false,
		WorkerShutdownTimeout:            "240s",
		VariablesHashBucketSize:          256,
		VariablesHashMaxSize:             2048,
		UseHTTP2:                         true,
		DisableProxyInterceptErrors:      false,
		RelativeRedirects:                false,
		ProxyStreamTimeout:               "600s",
		ProxyStreamNextUpstream:          true,
		ProxyStreamNextUpstreamTimeout:   "600s",
		ProxyStreamNextUpstreamTries:     3,
		Backend: defaults.Backend{
			ProxyBodySize:               bodySize,
			ProxyConnectTimeout:         5,
			ProxyReadTimeout:            60,
			ProxySendTimeout:            60,
			ProxyBuffersNumber:          4,
			ProxyBufferSize:             "4k",
			ProxyBusyBuffersSize:        "",
			ProxyCookieDomain:           "off",
			ProxyCookiePath:             "off",
			ProxyNextUpstream:           "error timeout",
			ProxyNextUpstreamTimeout:    0,
			ProxyNextUpstreamTries:      3,
			ProxyRequestBuffering:       "on",
			ProxyRedirectFrom:           "off",
			ProxyRedirectTo:             "off",
			PreserveTrailingSlash:       false,
			SSLRedirect:                 true,
			CustomHTTPErrors:            []int{},
			DisableProxyInterceptErrors: false,
			RelativeRedirects:           false,
			DenylistSourceRange:         []string{},
			WhitelistSourceRange:        []string{},
			SkipAccessLogURLs:           []string{},
			LimitRate:                   0,
			LimitRateAfter:              0,
			ProxyBuffering:              "off",
			ProxyHTTPVersion:            "1.1",
			ProxyMaxTempFileSize:        "1024m",
			ServiceUpstream:             false,
			AllowedResponseHeaders:      []string{},
		},
		UpstreamKeepaliveConnections:   320,
		UpstreamKeepaliveTime:          "1h",
		UpstreamKeepaliveTimeout:       60,
		UpstreamKeepaliveRequests:      10000,
		LimitConnZoneVariable:          defaultLimitConnZoneVariable,
		BindAddressIpv4:                defBindAddress,
		BindAddressIpv6:                defBindAddress,
		OpentelemetryTrustIncomingSpan: true,
		OpentelemetryConfig:            "/etc/ingress-controller/telemetry/opentelemetry.toml",
		OtlpCollectorPort:              "4317",
		OtelServiceName:                "nginx",
		OtelSampler:                    "AlwaysOn",
		OtelSamplerRatio:               0.01,
		OtelSamplerParentBased:         true,
		OtelScheduleDelayMillis:        5000,
		OtelMaxExportBatchSize:         512,
		OtelMaxQueueSize:               2048,
		LimitReqStatusCode:             503,
		LimitConnStatusCode:            503,
		SyslogPort:                     514,
		NoTLSRedirectLocations:         "/.well-known/acme-challenge",
		NoAuthLocations:                "/.well-known/acme-challenge",
		GlobalExternalAuth:             defGlobalExternalAuth,
		ProxySSLLocationOnly:           false,
		DefaultType:                    "text/html",
		DebugConnections:               []string{},
		StrictValidatePathType:         true,
		GRPCBufferSizeKb:               0,
	}

	if klog.V(5).Enabled() {
		cfg.ErrorLogLevel = "debug"
	}

	return cfg
}

// TemplateConfig contains the nginx configuration to render the file nginx.conf
type TemplateConfig struct {
	ProxySetHeaders          map[string]string                `json:"ProxySetHeaders"`
	AddHeaders               map[string]string                `json:"AddHeaders"`
	BacklogSize              int                              `json:"BacklogSize"`
	Backends                 []*ingress.Backend               `json:"Backends"`
	PassthroughBackends      []*ingress.SSLPassthroughBackend `json:"PassthroughBackends"`
	Servers                  []*ingress.Server                `json:"Servers"`
	TCPBackends              []ingress.L4Service              `json:"TCPBackends"`
	UDPBackends              []ingress.L4Service              `json:"UDPBackends"`
	HealthzURI               string                           `json:"HealthzURI"`
	Cfg                      Configuration                    `json:"Cfg"`
	IsIPV6Enabled            bool                             `json:"IsIPV6Enabled"`
	IsSSLPassthroughEnabled  bool                             `json:"IsSSLPassthroughEnabled"`
	NginxStatusIpv4Whitelist []string                         `json:"NginxStatusIpv4Whitelist"`
	NginxStatusIpv6Whitelist []string                         `json:"NginxStatusIpv6Whitelist"`
	RedirectServers          interface{}                      `json:"RedirectServers"`
	ListenPorts              *ListenPorts                     `json:"ListenPorts"`
	PublishService           *apiv1.Service                   `json:"PublishService"`
	EnableMetrics            bool                             `json:"EnableMetrics"`
	MaxmindEditionFiles      *[]string                        `json:"MaxmindEditionFiles"`
	MonitorMaxBatchSize      int                              `json:"MonitorMaxBatchSize"`
	PID                      string                           `json:"PID"`
	StatusPath               string                           `json:"StatusPath"`
	StatusPort               int                              `json:"StatusPort"`
	StreamPort               int                              `json:"StreamPort"`
	StreamSnippets           []string                         `json:"StreamSnippets"`
}

// ListenPorts describe the ports required to run the
// NGINX Ingress controller
type ListenPorts struct {
	HTTP     int `json:"HTTP"`
	HTTPS    int `json:"HTTPS"`
	Health   int `json:"Health"`
	Default  int `json:"Default"`
	SSLProxy int `json:"SSLProxy"`
}

// GlobalExternalAuth describe external authentication configuration for the
// NGINX Ingress controller
type GlobalExternalAuth struct {
	URL string `json:"url"`
	// Host contains the hostname defined in the URL
	Host                   string            `json:"host"`
	SigninURL              string            `json:"signinUrl"`
	SigninURLRedirectParam string            `json:"signinUrlRedirectParam"`
	Method                 string            `json:"method"`
	ResponseHeaders        []string          `json:"responseHeaders,omitempty"`
	RequestRedirect        string            `json:"requestRedirect"`
	AuthSnippet            string            `json:"authSnippet"`
	AuthCacheKey           string            `json:"authCacheKey"`
	AuthCacheDuration      []string          `json:"authCacheDuration"`
	ProxySetHeaders        map[string]string `json:"proxySetHeaders,omitempty"`
	AlwaysSetCookie        bool              `json:"alwaysSetCookie,omitempty"`
}
