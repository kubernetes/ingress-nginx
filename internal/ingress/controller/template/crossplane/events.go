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
	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
)

func (c *Template) buildEvents() {
	events := &ngx_crossplane.Directive{
		Directive: "events",
		Block: ngx_crossplane.Directives{
			buildDirective("worker_connections", c.tplConfig.Cfg.MaxWorkerConnections),
			buildDirective("use", "epoll"),
			buildDirective("multi_accept", c.tplConfig.Cfg.EnableMultiAccept),
		},
	}
	for k := range c.tplConfig.Cfg.DebugConnections {
		events.Block = append(events.Block, buildDirective("debug_connection", c.tplConfig.Cfg.DebugConnections[k]))
	}
	c.config.Parsed = append(c.config.Parsed, events)
}
