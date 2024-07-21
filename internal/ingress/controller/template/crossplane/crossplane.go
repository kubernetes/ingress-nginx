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
	"bytes"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

/*
Unsupported directives:
- opentelemetry
- modsecurity
- any stream directive (TCP/UDP forwarding)
- geoip2
*/

// On this case we will try to use the go ngx_crossplane to write the template instead of the template renderer

type CrossplaneTemplate struct {
	options   *ngx_crossplane.BuildOptions
	config    *ngx_crossplane.Config
	tplConfig *config.TemplateConfig
	mimeFile  string
}

func NewCrossplaneTemplate() *CrossplaneTemplate {
	lua := ngx_crossplane.Lua{}
	return &CrossplaneTemplate{
		mimeFile: "/etc/nginx/mime.types",
		options: &ngx_crossplane.BuildOptions{
			Builders: []ngx_crossplane.RegisterBuilder{
				lua.RegisterBuilder(),
			},
		},
	}
}

func (c *CrossplaneTemplate) SetMimeFile(file string) {
	c.mimeFile = file
}

func (c *CrossplaneTemplate) Write(conf *config.TemplateConfig) ([]byte, error) {
	c.tplConfig = conf

	// build root directives
	c.buildConfig()

	// build events directive
	c.buildEvents()

	// build http directive
	c.buildHTTP()

	var buf bytes.Buffer

	err := ngx_crossplane.Build(&buf, *c.config, &ngx_crossplane.BuildOptions{})
	return buf.Bytes(), err
}
