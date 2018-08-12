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

package v1alpha1

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

	gzipTypes = "application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"

	brotliTypes = "application/xml+rss application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"

	logFormatUpstream = `%v - [$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`

	logFormatStream = `[$time_local] $protocol $status $bytes_sent $bytes_received $session_time`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256"

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1.2"

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"

	// Default setting for load balancer algorithm
	defaultLoadBalancerAlgorithm = ""

	// Parameters for a shared memory zone that will keep states for various keys.
	// http://nginx.org/en/docs/http/ngx_http_limit_conn_module.html#limit_conn_zone
	defaultLimitConnZoneVariable = "$binary_remote_addr"
)

/*
// NewDefaultConfiguration returns the default nginx configuration
func NewDefaultConfiguration() ConfigurationSpec {
	defIPCIDR := make([]string, 0)
	defBindAddress := make([]string, 0)
	defNginxStatusIpv4Whitelist := make([]string, 0)
	defNginxStatusIpv6Whitelist := make([]string, 0)

	defIPCIDR = append(defIPCIDR, "0.0.0.0/0")
	defNginxStatusIpv4Whitelist = append(defNginxStatusIpv4Whitelist, "127.0.0.1")
	defNginxStatusIpv6Whitelist = append(defNginxStatusIpv6Whitelist, "::1")
	defProxyDeadlineDuration := time.Duration(5) * time.Second

	cfg := Configuration{
		AllowBackendServerHeader:   false,
		AccessLogPath:              "/var/log/nginx/access.log",
		WorkerCpuAffinity:          "",
		ErrorLogPath:               "/var/log/nginx/error.log",
		BrotliLevel:                4,
		BrotliTypes:                brotliTypes,
		ClientHeaderBufferSize:     "1k",
		ClientHeaderTimeout:        60,
		ClientBodyBufferSize:       "8k",
		ClientBodyTimeout:          60,
		EnableDynamicTLSRecords:    true,
		EnableUnderscoresInHeaders: false,
		ErrorLogLevel:              errorLevel,
		ForwardedForHeader:         "X-Forwarded-For",
		ComputeFullForwardedFor:    false,
		ProxyAddOriginalUriHeader:  true,
		GenerateRequestId:          true,
		HTTP2MaxFieldSize:          "4k",
		HTTP2MaxHeaderSize:         "16k",
		HTTPRedirectCode:           308,
		HSTS:                       true,
		HSTSIncludeSubdomains:      true,
		HSTSMaxAge:                 hstsMaxAge,
		HSTSPreload:                false,
		IgnoreInvalidHeaders:       true,
		GzipLevel:                  5,
		GzipTypes:                  gzipTypes,
		KeepAlive:                  75,
		KeepAliveRequests:          100,
		LargeClientHeaderBuffers:   "4 8k",
		LogFormatEscapeJSON:        false,
		LogFormatStream:            logFormatStream,
		LogFormatUpstream:          logFormatUpstream,
		EnableMultiAccept:          true,
		MaxWorkerConnections:       16384,
		MapHashBucketSize:          64,
		NginxStatusIpv4Whitelist:   defNginxStatusIpv4Whitelist,
		NginxStatusIpv6Whitelist:   defNginxStatusIpv6Whitelist,
		ProxyRealIPCIDR:            defIPCIDR,
		ProxyProtocolHeaderTimeout: defProxyDeadlineDuration,
		ServerNameHashMaxSize:      1024,
		ProxyHeadersHashMaxSize:    512,
		ProxyHeadersHashBucketSize: 64,
		ProxyStreamResponses:       1,
		ReusePort:                  true,
		ShowServerTokens:           true,
		SSLBufferSize:              sslBufferSize,
		SSLCiphers:                 sslCiphers,
		SSLECDHCurve:               "auto",
		SSLProtocols:               sslProtocols,
		SSLSessionCache:            true,
		SSLSessionCacheSize:        sslSessionCacheSize,
		SSLSessionTickets:          true,
		SSLSessionTimeout:          sslSessionTimeout,
		EnableBrotli:               false,
		UseGzip:                    true,
		UseGeoIP:                   true,
		WorkerProcesses:            strconv.Itoa(runtime.NumCPU()),
		WorkerShutdownTimeout:      "10s",
		LoadBalanceAlgorithm:       defaultLoadBalancerAlgorithm,
		VariablesHashBucketSize:    128,
		VariablesHashMaxSize:       2048,
		UseHTTP2:                   true,
		ProxyStreamTimeout:         "600s",
		Backend: defaults.Backend{
			ProxyBodySize:          bodySize,
			ProxyConnectTimeout:    5,
			ProxyReadTimeout:       60,
			ProxySendTimeout:       60,
			ProxyBufferSize:        "4k",
			ProxyCookieDomain:      "off",
			ProxyCookiePath:        "off",
			ProxyNextUpstream:      "error timeout",
			ProxyNextUpstreamTries: 3,
			ProxyRequestBuffering:  "on",
			ProxyRedirectFrom:      "off",
			ProxyRedirectTo:        "off",
			SSLRedirect:            true,
			CustomHTTPErrors:       []int{},
			WhitelistSourceRange:   []string{},
			SkipAccessLogURLs:      []string{},
			LimitRate:              0,
			LimitRateAfter:         0,
			ProxyBuffering:         "off",
		},
		UpstreamKeepaliveConnections: 32,
		LimitConnZoneVariable:        defaultLimitConnZoneVariable,
		BindAddressIpv4:              defBindAddress,
		BindAddressIpv6:              defBindAddress,
		ZipkinCollectorPort:          9411,
		ZipkinServiceName:            "nginx",
		ZipkinSampleRate:             1.0,
		JaegerCollectorPort:          6831,
		JaegerServiceName:            "nginx",
		JaegerSamplerType:            "const",
		JaegerSamplerParam:           "1",
		LimitReqStatusCode:           503,
		SyslogPort:                   514,
		NoTLSRedirectLocations:       "/.well-known/acme-challenge",
		NoAuthLocations:              "/.well-known/acme-challenge",
	}

	if glog.V(5) {
		cfg.ErrorLogLevel = "debug"
	}

	return cfg
}
*/
