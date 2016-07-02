/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"runtime"
	"strconv"

	"github.com/golang/glog"
)

const (
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	bodySize = "1m"

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	errorLevel = "notice"

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.
	hstsMaxAge = "15724800"

	// If UseProxyProtocol is enabled defIPCIDR defines the default the IP/network address of your external load balancer
	defIPCIDR = "0.0.0.0/0"

	gzipTypes = "application/atom+xml application/javascript aplication/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA"

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1 TLSv1.1 TLSv1.2"

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"
)

var (
	// SSLDirectory contains the mounted secrets with SSL certificates, keys and
	SSLDirectory = "/etc/nginx-ssl"
)

// Configuration represents the content of nginx.conf file
type Configuration struct {
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	BodySize string `structs:"body-size,omitempty"`

	// EnableDynamicTLSRecords enables dynamic TLS record sizes
	// https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency
	// By default this is enabled
	EnableDynamicTLSRecords bool `structs:"enable-dynamic-tls-records"`

	// EnableSPDY enables spdy and use ALPN and NPN to advertise the availability of the two protocols
	// https://blog.cloudflare.com/open-sourcing-our-nginx-http-2-spdy-code
	// By default this is enabled
	EnableSPDY bool `structs:"enable-spdy"`

	// EnableStickySessions enabled sticky sessions using cookies
	// https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng
	// By default this is disabled
	EnableStickySessions bool `structs:"enable-sticky-sessions,omitempty"`

	// EnableVtsStatus allows the replacement of the default status page with a third party module named
	// nginx-module-vts - https://github.com/vozlt/nginx-module-vts
	// By default this is disabled
	EnableVtsStatus bool `structs:"enable-vts-status,omitempty"`

	VtsStatusZoneSize string `structs:"vts-status-zone-size,omitempty"`

	// RetryNonIdempotent since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH)
	// in case of an error. The previous behavior can be restored using the value true
	RetryNonIdempotent bool `structs:"retry-non-idempotent"`

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `structs:"error-log-level,omitempty"`

	// Enables or disables the header HSTS in servers running SSL
	HSTS bool `structs:"hsts,omitempty"`

	// Enables or disables the use of HSTS in all the subdomains of the servername
	// Default: true
	HSTSIncludeSubdomains bool `structs:"hsts-include-subdomains,omitempty"`

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be
	// accessed using HTTPS.
	HSTSMaxAge string `structs:"hsts-max-age,omitempty"`

	// enables which HTTP codes should be passed for processing with the error_page directive
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page
	// By default this is disabled
	CustomHTTPErrors []int `structs:"custom-http-errors,-"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `structs:"keep-alive,omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// http://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `structs:"max-worker-connections,omitempty"`

	// Defines a timeout for establishing a connection with a proxied server.
	// It should be noted that this timeout cannot usually exceed 75 seconds.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout
	ProxyConnectTimeout int `structs:"proxy-connect-timeout,omitempty"`

	// If UseProxyProtocol is enabled ProxyRealIPCIDR defines the default the IP/network address
	// of your external load balancer
	ProxyRealIPCIDR string `structs:"proxy-real-ip-cidr,omitempty"`

	// Timeout in seconds for reading a response from the proxied server. The timeout is set only between
	// two successive read operations, not for the transmission of the whole response
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout
	ProxyReadTimeout int `structs:"proxy-read-timeout,omitempty"`

	// Timeout in seconds for transmitting a request to the proxied server. The timeout is set only between
	// two successive write operations, not for the transmission of the whole request.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout
	ProxySendTimeout int `structs:"proxy-send-timeout,omitempty"`

	// Configures name servers used to resolve names of upstream servers into addresses
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#resolver
	Resolver string `structs:"resolver,omitempty"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `structs:"server-name-hash-max-size,omitempty"`

	// Size of the bucker for the server names hash tables
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `structs:"server-name-hash-bucket-size,omitempty"`

	// SkipAccessLogURLs sets a list of URLs that should not appear in the NGINX access log
	// This is useful with urls like `/health` or `health-check` that make "complex" reading the logs
	// By default this list is empty
	SkipAccessLogURLs []string `structs:"skip-access-log-urls,-"`

	// Enables or disables the redirect (301) to the HTTPS port
	SSLRedirect bool `structs:"ssl-redirect,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	SSLBufferSize string `structs:"ssl-buffer-size,omitempty"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	SSLCiphers string `structs:"ssl-ciphers,omitempty"`

	// Base64 string that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
	// https://www.openssl.org/docs/manmaster/apps/dhparam.html
	// https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
	SSLDHParam string `structs:"ssl-dh-param,omitempty"`

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	SSLProtocols string `structs:"ssl-protocols,omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCache bool `structs:"ssl-session-cache,omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCacheSize string `structs:"ssl-session-cache-size,omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SSLSessionTickets bool `structs:"ssl-session-tickets,omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SSLSessionTimeout string `structs:"ssl-session-timeout,omitempty"`

	// Number of unsuccessful attempts to communicate with the server that should happen in the
	// duration set by the fail_timeout parameter to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	UpstreamMaxFails int `structs:"upstream-max-fails,omitempty"`

	// Time during which the specified number of unsuccessful attempts to communicate with
	// the server should happen to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	UpstreamFailTimeout int `structs:"upstream-fail-timeout,omitempty"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	UseProxyProtocol bool `structs:"use-proxy-protocol,omitempty"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// http://nginx.org/en/docs/http/ngx_http_gzip_module.html
	UseGzip bool `structs:"use-gzip,omitempty"`

	// Enables or disables the HTTP/2 support in secure connections
	// http://nginx.org/en/docs/http/ngx_http_v2_module.html
	// Default: true
	UseHTTP2 bool `structs:"use-http2,omitempty"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `structs:"gzip-types,omitempty"`

	// WhitelistSourceRange allows limiting access to certain client addresses
	// http://nginx.org/en/docs/http/ngx_http_access_module.html
	WhitelistSourceRange []string `structs:"whitelist-source-range,omitempty"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// http://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses string `structs:"worker-processes,omitempty"`
}

// NewDefault returns the default configuration contained
// in the file default-conf.json
func NewDefault() Configuration {
	cfg := Configuration{
		BodySize:                bodySize,
		EnableDynamicTLSRecords: true,
		EnableSPDY:              true,
		ErrorLogLevel:           errorLevel,
		HSTS:                    true,
		HSTSIncludeSubdomains:    true,
		HSTSMaxAge:               hstsMaxAge,
		GzipTypes:                gzipTypes,
		KeepAlive:                75,
		MaxWorkerConnections:     16384,
		ProxyConnectTimeout:      5,
		ProxyRealIPCIDR:          defIPCIDR,
		ProxyReadTimeout:         60,
		ProxySendTimeout:         60,
		ServerNameHashMaxSize:    512,
		ServerNameHashBucketSize: 64,
		SSLRedirect:              true,
		SSLBufferSize:            sslBufferSize,
		SSLCiphers:               sslCiphers,
		SSLProtocols:             sslProtocols,
		SSLSessionCache:          true,
		SSLSessionCacheSize:      sslSessionCacheSize,
		SSLSessionTickets:        true,
		SSLSessionTimeout:        sslSessionTimeout,
		UseProxyProtocol:         false,
		UseGzip:                  true,
		WorkerProcesses:          strconv.Itoa(runtime.NumCPU()),
		VtsStatusZoneSize:        "10m",
		UseHTTP2:                 true,
		CustomHTTPErrors:         make([]int, 0),
		WhitelistSourceRange:     make([]string, 0),
		SkipAccessLogURLs:        make([]string, 0),
	}

	if glog.V(5) {
		cfg.ErrorLogLevel = "debug"
	}

	return cfg
}
