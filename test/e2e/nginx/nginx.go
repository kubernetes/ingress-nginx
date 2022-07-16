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
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var (
	cfgOK = `#
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
	}
	`

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
				alias /www/html;
			}
		}
	}
	`

	cfgRoot = `#
	events {
		worker_connections  1024;
		multi_accept on;
	}
	
	http {
		default_type 'text/plain';
		client_max_body_size 0;
		root /srv/www;
		server {
			access_log on;
			access_log /dev/stdout;
	
			listen 80;
	
		}
	}
	`
)

var _ = framework.DescribeSetting("nginx-configuration", func() {
	f := framework.NewSimpleFramework("nginxconfiguration")

	ginkgo.It("start nginx with default configuration", func() {

		f.NGINXWithConfigDeployment("default-nginx", cfgOK)
		f.WaitForPod("app=default-nginx", 60*time.Second, false)
		framework.Sleep(5 * time.Second)

		f.HTTPDumbTestClient().
			GET("/").
			WithURL(fmt.Sprintf("http://default-nginx.%s", f.Namespace)).
			Expect().
			Status(http.StatusOK).Body().Contains("ok")
	})

	ginkgo.It("fails when using alias directive", func() {

		f.NGINXDeployment("alias-nginx", cfgAlias, false)
		// This should fail with a crashloopback because our NGINX does not have
		// alias directive!
		f.WaitForPod("app=alias-nginx", 60*time.Second, true)

	})

	ginkgo.It("fails when using root directive", func() {

		f.NGINXDeployment("root-nginx", cfgRoot, false)
		// This should fail with a crashloopback because our NGINX does not have
		// root directive!
		f.WaitForPod("app=root-nginx", 60*time.Second, true)

	})
})
