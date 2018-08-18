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

import (
	"net"
	"runtime"
	"time"
)

const (
	gzipTypes         = "application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"
	brotliTypes       = "application/xml+rss application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"
	logFormatUpstream = `%v - [$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256"
)

// NewDefaultConfiguration returns the default nginx configuration
func NewDefaultConfiguration() ConfigurationSpec {

	defIPCIDR := make([]net.IPAddr, 0)
	defBindAddress := make([]net.IPAddr, 0)
	localhost := net.IPAddr{IP: net.ParseIP("127.0.0.1")}
	defNginxStatusIpv4Whitelist := []net.IPAddr{localhost}

	loadBalancer := LoadBalanceAlgorithm(RoundRobin)
	jaegerSamplerType := JaegerSamplerType(ConstJaegerSampler)

	global := &Global{
		BindAddressIpv4:            defBindAddress,
		BindAddressIpv6:            defBindAddress,
		CustomHTTPErrors:           []int{},
		EnableBrotli:               false,
		BrotliLevel:                4,
		BrotliTypes:                brotliTypes,
		EnableGeoIP:                true,
		EnableGzip:                 true,
		GzipLevel:                  5,
		GzipTypes:                  gzipTypes,
		EnableInfluxDB:             false,
		EnableIPV6:                 true,
		EnableIPV6DNS:              true,
		EnableMultiAccept:          true,
		EnableProxyProtocol:        false,
		EnableRequestID:            true,
		EnableReusePort:            true,
		EnableUnderscoresInHeaders: false,
		IgnoreInvalidHeaders:       true,
		WorkerCPUAffinity:          "",
		KeepAlive:                  75,
		KeepAliveRequests:          100,
		LimitConnZoneVariable:      "$binary_remote_addr",
		LimitRequestStatusCode:     429,
		LoadBalanceAlgorithm:       &loadBalancer,
		MapHashBucketSize:          64,
		MaxWorkerConnections:       16384,
		ProxyHeadersHashBucketSize: 64,
		ProxyHeadersHashMaxSize:    512,
		ProxyProtocolHeaderTimeout: time.Duration(5) * time.Second,
		ProxyRealIPCIDR:            defIPCIDR,
		//Resolver: ,
		RetryNonIdempotent: false,
		//ServerNameHashBucketSize: ,
		ServerNameHashMaxSize: 1024,
		ShowServerTokens:      true,
		StatusIPV4Whitelist:   defNginxStatusIpv4Whitelist,
		//StatusIPV6Whitelist: defNginxStatusIpv6Whitelist,
		WorkerProcesses:       runtime.NumCPU(),
		WorkerShutdownTimeout: time.Duration(10) * time.Second,
		//VariablesHashBucketSize: ,
		VariablesHashMaxSize:    2048,
		ForwardedForHeader:      "X-Forwarded-For",
		ComputeFullForwardedFor: false,
		HTTPRedirectCode:        308,
		NoAuthLocations:         "",
		PortInRedirects:         false,
		//WhitelistSourceRange: ,
	}

	client := &Client{
		BodyBufferSize:           "8k",
		BodyTimeout:              60,
		HeaderBufferSize:         "1k",
		HeaderTimeout:            60,
		LargeClientHeaderBuffers: "4 8k",
	}

	http2 := &HTTP2{
		Enabled:       true,
		MaxFieldSize:  "4k",
		MaxHeaderSize: "16k",
	}

	file := &LogFileConfiguration{
		AccessLogPath: "/var/log/nginx/access.log",
		ErrorLogPath:  "/var/log/nginx/error.log",
	}

	syslog := &SyslogConfiguration{
		Enabled: false,
		Host:    "",
		Port:    514,
	}

	log := &Log{
		EnableAccessLog:   true,
		ErrorLogLevel:     "notice",
		FormatEscapeJSON:  false,
		FormatUpstream:    logFormatUpstream,
		SkipAccessLogURLs: []string{},
		File:              file,
		Syslog:            syslog,
	}

	metrics := &Metrics{
		Enabled:        false,
		Latency:        []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		ResponseLength: []float64{100, 1000, 10000, 100000, 1000000},
		RequestLength:  []float64{1000, 10000, 100000, 1000000, 10000000},
	}

	snippets := &Snippets{
		Main:     "",
		HTTP:     "",
		Server:   "",
		Location: "",
	}

	hsts := &HSTS{
		Enabled:           true,
		IncludeSubdomains: true,
		MaxAge:            15724800,
		Preload:           false,
	}

	ssl := &SSL{
		HSTS:        hsts,
		SSLRedirect: true,
		//ForceSSLRedirect: ,
		//NoTLSRedirectLocations:,
		Ciphers:   sslCiphers,
		ECDHCurve: "auto",
		//DHParam: ,
		Protocols:        "TLSv1.2",
		SessionCache:     true,
		SessionCacheSize: "10m",
		SessionTickets:   true,
		//SessionTicketKey: ,
		SessionTimeout: "10m",
		BufferSize:     "4k",
	}

	jaeger := &JaegerConfiguration{
		CollectorHost: "",
		CollectorPort: 6831,
		ServiceName:   "nginx",
		SamplerType:   jaegerSamplerType,
		SamplerParam:  "1",
	}

	zipkin := &ZipkinConfiguration{
		CollectorHost: "",
		CollectorPort: 9411,
		ServiceName:   "nginx",
		SampleRate:    1.0,
	}

	opentracing := &Opentracing{
		Enabled: false,
		Jaeger:  jaeger,
		Zipkin:  zipkin,
	}

	upstream := &Upstream{
		AddOriginalURIHeader:          true,
		SetHeaders:                    make(map[string]string),
		HideHeaders:                   []string{},
		EnableServerHeaderFromBackend: false,
		BodySize:                      "1m",
		Buffering:                     "off",
		BufferSize:                    "4k",
		CookieDomain:                  "off",
		CookiePath:                    "off",
		ConnectTimeout:                5,
		FailTimeout:                   0,
		//HashBy: "",
		KeepaliveConnections: 32,
		MaxFails:             0,
		NextUpstream:         "error timeout",
		NextUpstreamTries:    3,
		ReadTimeout:          60,
		RedirectFrom:         "off",
		RedirectTo:           "off",
		RequestBuffering:     "on",
		SendTimeout:          60,
	}

	waf := &WAF{
		EnableModsecurity:    false,
		EnableOWASPCoreRules: false,
		EnableLuaRestyWAF:    true,
	}

	cfg := ConfigurationSpec{
		Global:      global,
		Client:      client,
		HTTP2:       http2,
		Log:         log,
		Metrics:     metrics,
		Snippets:    snippets,
		SSL:         ssl,
		Opentracing: opentracing,
		Upstream:    upstream,
		WAF:         waf,
	}

	return cfg
}
