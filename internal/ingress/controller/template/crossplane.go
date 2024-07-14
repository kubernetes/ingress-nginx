package template

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	crossplane "github.com/nginxinc/nginx-go-crossplane"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

/*
Unsupported directives:
- opentelemetry
- modsecurity
- any stream directive (TCP/UDP forwarding)
- geoip2
*/

// On this case we will try to use the go crossplane to write the template instead of the template renderer

type crossplaneTemplate struct {
	options   *crossplane.BuildOptions
	config    *crossplane.Config
	tplConfig *config.TemplateConfig
}

func NewCrossplaneTemplate() *crossplaneTemplate {
	lua := crossplane.Lua{}
	return &crossplaneTemplate{
		options: &crossplane.BuildOptions{
			Builders: []crossplane.RegisterBuilder{
				lua.RegisterBuilder(),
			},
		},
	}
}

func (c *crossplaneTemplate) Write(conf *config.TemplateConfig) ([]byte, error) {
	c.tplConfig = conf
	// Write basic directives
	config := &crossplane.Config{
		Parsed: crossplane.Directives{
			buildDirective("pid", c.tplConfig.PID),
			buildDirective("daemon", "off"),
			buildDirective("worker_processes", c.tplConfig.Cfg.WorkerProcesses),
			buildDirective("worker_rlimit_nofile", c.tplConfig.Cfg.MaxWorkerOpenFiles),
			buildDirective("worker_shutdown_timeout", c.tplConfig.Cfg.WorkerShutdownTimeout),
		},
	}
	if c.tplConfig.Cfg.WorkerCPUAffinity != "" {
		config.Parsed = append(config.Parsed, buildDirective("worker_cpu_affinity", c.tplConfig.Cfg.WorkerCPUAffinity))
	}
	c.config = config

	// build events directive
	c.buildEvents()

	// build http directive
	c.buildHTTP()

	var buf bytes.Buffer

	err := crossplane.Build(&buf, *c.config, &crossplane.BuildOptions{})
	return buf.Bytes(), err
}

func (c *crossplaneTemplate) buildEvents() {
	events := &crossplane.Directive{
		Directive: "events",
		Block: crossplane.Directives{
			buildDirective("worker_connections", c.tplConfig.Cfg.MaxWorkerConnections),
			buildDirective("use", "epool"),
			buildDirective("multi_accept", c.tplConfig.Cfg.EnableMultiAccept),
		},
	}
	for k := range c.tplConfig.Cfg.DebugConnections {
		events.Block = append(events.Block, buildDirective("debug_connection", c.tplConfig.Cfg.DebugConnections[k]))
	}
	c.config.Parsed = append(c.config.Parsed, events)
}

func (c *crossplaneTemplate) buildHTTP() {
	cfg := c.tplConfig.Cfg
	httpBlock := crossplane.Directives{
		buildDirective("lua_package_path", "/etc/nginx/lua/?.lua;;"),
		buildDirective("include", "/etc/nginx/mime.types"),
		buildDirective("default_type", cfg.DefaultType),
		buildDirective("real_ip_recursive", "on"),
		buildDirective("aio", "threads"),
		buildDirective("aio_write", cfg.EnableAioWrite),
		buildDirective("server_tokens", cfg.ShowServerTokens),
		buildDirective("resolver", buildResolversInternal(cfg.Resolver, cfg.DisableIpv6DNS)),
		buildDirective("tcp_nopush", "on"),
		buildDirective("tcp_nodelay", "on"),
		buildDirective("log_subrequest", "on"),
		buildDirective("reset_timedout_connection", "on"),
		buildDirective("keepalive_timeout", seconds(cfg.KeepAlive)),
		buildDirective("keepalive_requests", cfg.KeepAliveRequests),
		buildDirective("client_body_temp_path", "/tmp/nginx/client-body"),
		buildDirective("fastcgi_temp_path", "/tmp/nginx/fastcgi-temp"),
		buildDirective("proxy_temp_path", "/tmp/nginx/proxy-temp"),
		buildDirective("client_header_buffer_size", cfg.ClientHeaderBufferSize),
		buildDirective("client_header_timeout", seconds(cfg.ClientHeaderTimeout)),
		buildDirective("large_client_header_buffers", cfg.LargeClientHeaderBuffers),
		buildDirective("client_body_buffer_size", cfg.ClientBodyBufferSize),
		buildDirective("client_body_timeout", seconds(cfg.ClientBodyTimeout)),
		buildDirective("types_hash_max_size", "2048"),
		buildDirective("server_names_hash_max_size", cfg.ServerNameHashMaxSize),
		buildDirective("server_names_hash_bucket_size", cfg.ServerNameHashBucketSize),
		buildDirective("map_hash_bucket_size", cfg.MapHashBucketSize),
		buildDirective("proxy_headers_hash_max_size", cfg.ProxyHeadersHashMaxSize),
		buildDirective("proxy_headers_hash_bucket_size", cfg.ProxyHeadersHashBucketSize),
		buildDirective("variables_hash_bucket_size", cfg.VariablesHashBucketSize),
		buildDirective("variables_hash_max_size", cfg.VariablesHashMaxSize),
		buildDirective("underscores_in_headers", cfg.EnableUnderscoresInHeaders),
		buildDirective("ignore_invalid_headers", cfg.IgnoreInvalidHeaders),
		buildDirective("limit_req_status", cfg.LimitReqStatusCode),
		buildDirective("limit_conn_status", cfg.LimitConnStatusCode),
		buildDirective("uninitialized_variable_warn", "off"),
		buildDirective("server_name_in_redirect", "off"),
		buildDirective("port_in_redirect", "off"),
		buildDirective("ssl_protocols", strings.Split(cfg.SSLProtocols, " ")),
		buildDirective("ssl_early_data", cfg.SSLEarlyData),
		buildDirective("ssl_session_tickets", cfg.SSLSessionTickets),
		buildDirective("ssl_buffer_size", cfg.SSLBufferSize),
		buildDirective("ssl_ecdh_curve", cfg.SSLECDHCurve),
		buildDirective("ssl_certificate", cfg.DefaultSSLCertificate.PemFileName),
		buildDirective("ssl_certificate_key", cfg.DefaultSSLCertificate.PemFileName),
		buildDirective("proxy_ssl_session_reuse", "on"),
		buildDirective("proxy_cache_path", []string{
			"/tmp/nginx/nginx-cache-auth", "levels=1:2", "keys_zone=auth_cache:10m",
			"max_size=128m", "inactive=30m", "use_temp_path=off",
		}),
	}

	httpBlock = append(httpBlock, buildLuaSharedDictionariesForCrossplane(c.tplConfig.Cfg)...)

	// Real IP dealing
	if (cfg.UseForwardedHeaders || cfg.UseProxyProtocol) || cfg.EnableRealIP {
		if cfg.UseProxyProtocol {
			httpBlock = append(httpBlock, buildDirective("real_ip_header", "proxy_protocol"))
		} else {
			httpBlock = append(httpBlock, buildDirective("real_ip_header", cfg.ForwardedForHeader))
		}

		for k := range cfg.ProxyRealIPCIDR {
			httpBlock = append(httpBlock, buildDirective("set_real_ip_from", cfg.ProxyRealIPCIDR[k]))
		}
	}

	if cfg.GRPCBufferSizeKb > 0 {
		httpBlock = append(httpBlock, buildDirective("grpc_buffer_size", strconv.Itoa(cfg.GRPCBufferSizeKb)+"k"))
	}

	// HTTP2 Configuration
	if cfg.HTTP2MaxHeaderSize != "" && cfg.HTTP2MaxFieldSize != "" {
		httpBlock = append(httpBlock, buildDirective("http2_max_field_size", cfg.HTTP2MaxFieldSize))
		httpBlock = append(httpBlock, buildDirective("http2_max_header_size", cfg.HTTP2MaxHeaderSize))
		if cfg.HTTP2MaxRequests > 0 {
			httpBlock = append(httpBlock, buildDirective("http2_max_requests", cfg.HTTP2MaxRequests))
		}
	}

	if cfg.UseGzip {
		httpBlock = append(httpBlock, buildDirective("gzip", "on"))
		httpBlock = append(httpBlock, buildDirective("gzip_comp_level", cfg.GzipLevel))
		httpBlock = append(httpBlock, buildDirective("gzip_http_version", "1.1"))
		httpBlock = append(httpBlock, buildDirective("gzip_min_length", cfg.GzipMinLength))
		httpBlock = append(httpBlock, buildDirective("gzip_types", cfg.GzipTypes))
		httpBlock = append(httpBlock, buildDirective("gzip_proxied", "any"))
		httpBlock = append(httpBlock, buildDirective("gzip_vary", "on"))

		if cfg.GzipDisable != "" {
			httpBlock = append(httpBlock, buildDirective("gzip_disable", strings.Split(cfg.GzipDisable, "")))
		}
	}

	if !cfg.ShowServerTokens {
		httpBlock = append(httpBlock, buildDirective("more_clear_headers", "Server"))
	}

	if len(c.tplConfig.AddHeaders) > 0 {
		additionalHeaders := make([]string, 0)
		for headerName, headerValue := range c.tplConfig.AddHeaders {
			additionalHeaders = append(additionalHeaders, fmt.Sprintf("%s: %s", headerName, headerValue))
		}
		httpBlock = append(httpBlock, buildDirective("more_set_headers", additionalHeaders))
	}

	escape := ""
	if cfg.LogFormatEscapeNone {
		escape = "escape=none"
	} else if cfg.LogFormatEscapeJSON {
		escape = "escape=json"
	}

	httpBlock = append(httpBlock, buildDirective("log_format", "upstreaminfo", escape, cfg.LogFormatUpstream))

	// buildMap directive
	mapLogDirective := &crossplane.Directive{
		Directive: "map",
		Args:      []string{"$request_uri", "$loggable"},
		Block:     make(crossplane.Directives, 0),
	}
	for k := range cfg.SkipAccessLogURLs {
		mapLogDirective.Block = append(mapLogDirective.Block, buildDirective(cfg.SkipAccessLogURLs[k], "0"))
	}
	mapLogDirective.Block = append(mapLogDirective.Block, buildDirective("default", "1"))
	httpBlock = append(httpBlock, mapLogDirective)
	// end of build mapLog

	if cfg.DisableAccessLog || cfg.DisableHTTPAccessLog {
		httpBlock = append(httpBlock, buildDirective("access_log", "off"))
	} else {
		logDirectives := []string{"upstreaminfo", "if=$loggable"}
		if cfg.EnableSyslog {
			httpBlock = append(httpBlock, buildDirective("access_log", fmt.Sprintf("syslog:server%s:%d", cfg.SyslogHost, cfg.SyslogPort), logDirectives))
		} else {
			accessLog := cfg.AccessLogPath
			if cfg.HTTPAccessLogPath != "" {
				accessLog = cfg.HTTPAccessLogPath
			}
			httpBlock = append(httpBlock, buildDirective("access_log", accessLog, logDirectives))
		}
	}

	if cfg.EnableSyslog {
		httpBlock = append(httpBlock, buildDirective("error_log", fmt.Sprintf("syslog:server%s:%d", cfg.SyslogHost, cfg.SyslogPort), cfg.ErrorLogLevel))
	} else {
		httpBlock = append(httpBlock, buildDirective("error_log", cfg.ErrorLogPath, cfg.ErrorLogLevel))
	}

	c.config.Parsed = append(c.config.Parsed, &crossplane.Directive{
		Directive: "http",
		Block:     httpBlock,
	})
}

type seconds int

// TODO: This conversion doesn't work with array
func buildDirective(directive string, args ...any) *crossplane.Directive {
	argsVal := make([]string, 0)
	for k := range args {
		switch v := args[k].(type) {
		case string:
			argsVal = append(argsVal, v)
		case []string:
			argsVal = append(argsVal, v...)
		case int:
			argsVal = append(argsVal, strconv.Itoa(v))
		case bool:
			argsVal = append(argsVal, boolToStr(v))
		case seconds:
			argsVal = append(argsVal, strconv.Itoa(int(v))+"s")
		}
	}
	return &crossplane.Directive{
		Directive: directive,
		Args:      argsVal,
	}
}

func buildLuaSharedDictionariesForCrossplane(cfg config.Configuration) []*crossplane.Directive {
	out := make([]*crossplane.Directive, 0, len(cfg.LuaSharedDicts))
	for name, size := range cfg.LuaSharedDicts {
		sizeStr := dictKbToStr(size)
		out = append(out, buildDirective("lua_shared_dict", name, sizeStr))
	}

	return out
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
