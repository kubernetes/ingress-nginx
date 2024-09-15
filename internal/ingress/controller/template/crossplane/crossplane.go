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
	"os"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/template/crossplane/extramodules"
)

/*
Unsupported directives:
- opentelemetry
- modsecurity
- any stream directive (TCP/UDP forwarding)
- geoip2
*/

// On this case we will try to use the go ngx_crossplane to write the template instead of the template renderer

type Template struct {
	options   *ngx_crossplane.BuildOptions
	config    *ngx_crossplane.Config
	tplConfig *config.TemplateConfig
	mimeFile  string
}

func NewTemplate() (*Template, error) {
	lua := ngx_crossplane.Lua{}
	return &Template{
		mimeFile: "/etc/nginx/mime.types",
		options: &ngx_crossplane.BuildOptions{
			Builders: []ngx_crossplane.RegisterBuilder{
				lua.RegisterBuilder(),
			},
		},
	}, nil
}

func (c *Template) SetMimeFile(file string) {
	c.mimeFile = file
}

func (c *Template) Write(conf *config.TemplateConfig) ([]byte, error) {
	c.tplConfig = conf

	// build root directives
	c.buildConfig()

	// build events directive
	c.buildEvents()

	// build http directive
	c.buildHTTP()

	var buf bytes.Buffer

	err := ngx_crossplane.Build(&buf, *c.config, &ngx_crossplane.BuildOptions{})
	if err != nil {
		return nil, err
	}

	lua := ngx_crossplane.Lua{}
	options := ngx_crossplane.ParseOptions{
		ParseComments:            true,
		ErrorOnUnknownDirectives: true,
		StopParsingOnError:       true,
		DirectiveSources: []ngx_crossplane.MatchFunc{
			ngx_crossplane.DefaultDirectivesMatchFunc,
			ngx_crossplane.MatchLuaLatest,
			ngx_crossplane.MatchHeadersMoreLatest,
			extramodules.BrotliMatchFn,
			extramodules.OpentelemetryMatchFn,
			ngx_crossplane.MatchGeoip2Latest,
		},
		LexOptions: ngx_crossplane.LexOptions{
			Lexers: []ngx_crossplane.RegisterLexer{lua.RegisterLexer()},
		},
		// Modules that needs to be ported:
		// // https://github.com/openresty/set-misc-nginx-module?tab=readme-ov-file#set_escape_uri
		IgnoreDirectives: []string{"set_escape_uri"},
	}

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
		_ = tmpFile.Close()
	}()

	_, err = tmpFile.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	_, err = ngx_crossplane.Parse(tmpFile.Name(), &options)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}
