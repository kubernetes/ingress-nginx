package template

import (
	"strconv"
	"strings"

	crossplane "github.com/nginxinc/nginx-go-crossplane"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

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

	// Load Modules
	c.loadModules()

	// build events directive
	c.buildEvents()

	// build http directive
	c.buildHTTP()

	return nil, nil
}

func (c *crossplaneTemplate) loadModules() {
	if c.tplConfig.Cfg.EnableBrotli {
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_brotli_filter_module.so"))
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_brotli_static_module.so"))
	}

	if c.tplConfig.Cfg.UseGeoIP2 {
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_geoip2_module.so"))
	}

	if shouldLoadAuthDigestModule(c.tplConfig.Servers) {
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_auth_digest_module.so"))
	}

	if shouldLoadModSecurityModule(c.tplConfig.Cfg, c.tplConfig.Servers) {
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_modsecurity_module.so"))
	}

	if shouldLoadOpentelemetryModule(c.tplConfig.Cfg, c.tplConfig.Servers) {
		c.config.Parsed = append(c.config.Parsed, buildDirective("load_module", "/etc/nginx/modules/otel_ngx_module.so"))
	}
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
		buildDirective("proxy_cache_path", strings.Split("/tmp/nginx/nginx-cache-auth levels=1:2 keys_zone=auth_cache:10m max_size=128m inactive=30m use_temp_path=off", " ")),

		// add grpc_buffer_size
	}

	c.config.Parsed = append(c.config.Parsed, &crossplane.Directive{
		Directive: "http",
		Block:     httpBlock,
	})
}

type seconds int

func buildDirective(directive string, args ...any) *crossplane.Directive {
	argsVal := make([]string, len(args))
	for k := range args {
		switch v := args[k].(type) {
		case string:
			argsVal[k] = v
		case int:
			argsVal[k] = strconv.Itoa(v)
		case bool:
			argsVal[k] = boolToStr(v)
		case seconds:
			argsVal[k] = strconv.Itoa(int(v)) + "s"
		default:
			argsVal[k] = ""
		}
	}
	return &crossplane.Directive{
		Directive: directive,
		Args:      argsVal,
	}
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
