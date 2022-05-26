/*
Copyright 2022 The Kubernetes Authors.

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

package nginx

import (
	"net/http"
	"strings"

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var (
	cfgOk = `#
events {
	worker_connections  1024;
	multi_accept on;
}

http {
	default_type 'text/plain';
	client_max_body_size 0;

	server {
		access_log on;
		access_log /dev/stdout;

		listen 80;

		location / {
			content_by_lua_block {
				ngx.print("ok")
			}
		}
	}
}`

	cfgRoot = `#
events {
	worker_connections  1024;
	multi_accept on;
}

http {
	default_type 'text/plain';
	client_max_body_size 0;

	server {
		access_log on;
		access_log /dev/stdout;

		listen 80;

		location / {
			content_by_lua_block {
				ngx.print("ok")
			}
		}

		location  /hello {
			root /www/data;
		}
	}
}`

	cfgAlias = `#
events {
	worker_connections  1024;
	multi_accept on;
}

http {
	default_type 'text/plain';
	client_max_body_size 0;

	server {
		access_log on;
		access_log /dev/stdout;

		listen 80;

		location / {
			content_by_lua_block {
				ngx.print("ok")
			}
		}

		location  /testalias {
			alias /www/data;
		}
	}
}`
)

var _ = framework.IngressNginxDescribe("[NGINX] NGINX Default configuration tests", func() {
	f := framework.NewDefaultFramework("nginx-configurations")

	ginkgo.It("should start nginx with default configuration", func() {
		host := "default-nginx"
		f.NGINXWithConfigDeployment("working-nginx", cfgOk)

		f.WaitForNginxServer(host,
			func(conf string) bool {
				return strings.Contains(conf, "access_log on")
			})

		f.HTTPTestClient().
			GET("/").
			Expect().
			Status(http.StatusOK).Body().Contains("ok")

		f.DeleteNGINXPod(60)

	})
})
