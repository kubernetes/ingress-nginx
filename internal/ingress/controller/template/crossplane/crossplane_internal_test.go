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
	"testing"

	ngx_crossplane "github.com/nginxinc/nginx-go-crossplane"
	"github.com/stretchr/testify/require"

	"k8s.io/ingress-nginx/internal/ingress/controller/config"
)

// THIS FILE SHOULD BE USED JUST FOR INTERNAL TESTS - Private functions

func Test_Internal_buildEvents(t *testing.T) {
	t.Run("should fill correctly events directives with defaults", func(t *testing.T) {
		c := ngx_crossplane.Config{}
		tplConfig := &config.TemplateConfig{
			Cfg: config.NewDefault(),
		}

		expectedEvents := &ngx_crossplane.Config{
			File: "",
			Parsed: ngx_crossplane.Directives{
				{
					Directive: "events",
					Block: ngx_crossplane.Directives{
						buildDirective("worker_connections", 16384),
						buildDirective("use", "epoll"),
						buildDirective("multi_accept", true),
					},
				},
			},
		}

		cplane, err := NewTemplate()
		require.NoError(t, err)
		cplane.config = &c
		cplane.tplConfig = tplConfig
		cplane.buildEvents()
		require.Equal(t, expectedEvents, cplane.config)
	})

	t.Run("should fill correctly events directives with specific values", func(t *testing.T) {
		c := ngx_crossplane.Config{}
		tplConfig := &config.TemplateConfig{
			Cfg: config.Configuration{
				MaxWorkerConnections: 50,
				EnableMultiAccept:    false,
				DebugConnections:     []string{"127.0.0.1/32", "192.168.0.10"},
			},
		}

		expectedEvents := &ngx_crossplane.Config{
			File: "",
			Parsed: ngx_crossplane.Directives{
				{
					Directive: "events",
					Block: ngx_crossplane.Directives{
						buildDirective("worker_connections", 50),
						buildDirective("use", "epoll"),
						buildDirective("multi_accept", false),
						buildDirective("debug_connection", "127.0.0.1/32"),
						buildDirective("debug_connection", "192.168.0.10"),
					},
				},
			},
		}

		cplane, err := NewTemplate()
		require.NoError(t, err)
		cplane.config = &c
		cplane.tplConfig = tplConfig
		cplane.buildEvents()
		require.Equal(t, expectedEvents, cplane.config)
	})
}
