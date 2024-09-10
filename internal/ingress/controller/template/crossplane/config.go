/*
Copyright 2024 The Kubernetes Authors.

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

package crossplane

import (
	"strings"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
)

func (c *Template) buildConfig() {
	// Write basic directives
	config := &ngx_crossplane.Config{
		Parsed: ngx_crossplane.Directives{
			buildDirective("pid", c.tplConfig.PID),
			buildDirective("daemon", "off"),
			buildDirective("worker_processes", c.tplConfig.Cfg.WorkerProcesses),
			buildDirective("worker_rlimit_nofile", c.tplConfig.Cfg.MaxWorkerOpenFiles),
			buildDirective("worker_shutdown_timeout", c.tplConfig.Cfg.WorkerShutdownTimeout),
		},
	}
	if c.tplConfig.Cfg.WorkerCPUAffinity != "" {
		config.Parsed = append(config.Parsed, buildDirective("worker_cpu_affinity", strings.Split(c.tplConfig.Cfg.WorkerCPUAffinity, " ")))
	}

	if c.tplConfig.Cfg.EnableBrotli {
		config.Parsed = append(config.Parsed,
			buildDirective("load_module", "/etc/nginx/modules/ngx_http_brotli_filter_module.so"),
			buildDirective("load_module", "/etc/nginx/modules/ngx_http_brotli_static_module.so"),
		)
	}

	if shouldLoadAuthDigestModule(c.tplConfig.Servers) {
		config.Parsed = append(config.Parsed, buildDirective("load_module", "/etc/nginx/modules/ngx_http_auth_digest_module.so"))
	}

	if c.tplConfig.Cfg.EnableOpentelemetry || shouldLoadOpentelemetryModule(c.tplConfig.Servers) {
		config.Parsed = append(config.Parsed, buildDirective("load_module", "/etc/nginx/modules/otel_ngx_module.so"))
	}

	if c.tplConfig.Cfg.UseGeoIP2 {
		config.Parsed = append(config.Parsed,
			buildDirective("load_module", "/etc/nginx/modules/ngx_http_geoip2_module.so"),
		)
	}

	c.config = config
}
