package template

import (
	"strconv"

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
			buildDirective("worker_rlimit_nofile", strconv.Itoa(c.tplConfig.Cfg.MaxWorkerOpenFiles)),
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

func buildDirective(directive string, args ...string) *crossplane.Directive {
	return &crossplane.Directive{
		Directive: directive,
		Args:      args,
	}
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
			buildDirective("worker_connections", strconv.Itoa(c.tplConfig.Cfg.MaxWorkerConnections)),
			buildDirective("use", "epool"),
			buildDirective("use", "epool"),
			buildDirective("multi_accept", boolToStr(c.tplConfig.Cfg.EnableMultiAccept)),
		},
	}
	for _, v := range c.tplConfig.Cfg.DebugConnections {
		events.Block = append(events.Block, buildDirective("debug_connection", v))
	}
}

func (c *crossplaneTemplate) buildHTTP() {
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
