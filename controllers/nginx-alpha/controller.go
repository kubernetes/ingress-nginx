/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package main

import (
	"log"
	"os"
	"os/exec"
	"reflect"
	"text/template"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/flowcontrol"
)

const (
	nginxConf = `
events {
  worker_connections 1024;
}
http {
  # http://nginx.org/en/docs/http/ngx_http_core_module.html
  types_hash_max_size 2048;
  server_names_hash_max_size 512;
  server_names_hash_bucket_size 64;

{{range $ing := .Items}}
{{range $rule := $ing.Spec.Rules}}
  server {
    listen 80;
    server_name {{$rule.Host}};
{{ range $path := $rule.HTTP.Paths }}
    location {{$path.Path}} {
      proxy_set_header Host $host;
      proxy_pass http://{{$path.Backend.ServiceName}}.{{$ing.Namespace}}.svc.cluster.local:{{$path.Backend.ServicePort}};
    }{{end}}
  }{{end}}{{end}}
}`
)

func shellOut(cmd string) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to execute %v: %v, err: %v", cmd, string(out), err)
	}
}

func main() {
	var ingClient client.IngressInterface
	if kubeClient, err := client.NewInCluster(); err != nil {
		log.Fatalf("Failed to create client: %v.", err)
	} else {
		ingClient = kubeClient.Extensions().Ingress(api.NamespaceAll)
	}
	tmpl, _ := template.New("nginx").Parse(nginxConf)
	rateLimiter := flowcontrol.NewTokenBucketRateLimiter(0.1, 1)
	known := &extensions.IngressList{}

	// Controller loop
	shellOut("nginx")
	for {
		rateLimiter.Accept()
		ingresses, err := ingClient.List(api.ListOptions{})
		if err != nil {
			log.Printf("Error retrieving ingresses: %v", err)
			continue
		}
		if reflect.DeepEqual(ingresses.Items, known.Items) {
			continue
		}
		known = ingresses
		if w, err := os.Create("/etc/nginx/nginx.conf"); err != nil {
			log.Fatalf("Failed to open %v: %v", nginxConf, err)
		} else if err := tmpl.Execute(w, ingresses); err != nil {
			log.Fatalf("Failed to write template %v", err)
		}
		shellOut("nginx -s reload")
	}
}
