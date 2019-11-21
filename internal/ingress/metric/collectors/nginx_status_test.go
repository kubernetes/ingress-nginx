/*
Copyright 2018 The Kubernetes Authors.

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

package collectors

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/ingress-nginx/internal/nginx"
)

func TestStatusCollector(t *testing.T) {
	cases := []struct {
		name    string
		mock    string
		metrics []string
		want    string
	}{
		{
			name: "should return empty metrics",
			mock: `
			`,
			want: `
				# HELP nginx_ingress_controller_nginx_process_connections_total total number of connections with state {accepted, handled}
				# TYPE nginx_ingress_controller_nginx_process_connections_total counter
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="accepted"} 0
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="handled"} 0
			`,
			metrics: []string{"nginx_ingress_controller_nginx_process_connections_total"},
		},
		{
			name: "should return metrics for total connections",
			mock: `
				Active connections: 15
				server accepts handled requests
				1 2 3
				Reading: 4 Writing: 5 Waiting: 6
			`,
			want: `
				# HELP nginx_ingress_controller_nginx_process_connections_total total number of connections with state {accepted, handled}
				# TYPE nginx_ingress_controller_nginx_process_connections_total counter
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="accepted"} 1
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="handled"} 2
			`,
			metrics: []string{"nginx_ingress_controller_nginx_process_connections_total"},
		},
		{
			name: "should return nginx metrics all available metrics",
			mock: `
				Active connections: 15
				server accepts handled requests
				1 2 3
				Reading: 4 Writing: 5 Waiting: 6
			`,
			want: `
				# HELP nginx_ingress_controller_nginx_process_connections current number of client connections with state {active, reading, writing, waiting}
				# TYPE nginx_ingress_controller_nginx_process_connections gauge
				nginx_ingress_controller_nginx_process_connections{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="active"} 15
				nginx_ingress_controller_nginx_process_connections{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="reading"} 4
				nginx_ingress_controller_nginx_process_connections{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="waiting"} 6
				nginx_ingress_controller_nginx_process_connections{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="writing"} 5
				# HELP nginx_ingress_controller_nginx_process_connections_total total number of connections with state {accepted, handled}
				# TYPE nginx_ingress_controller_nginx_process_connections_total counter
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="accepted"} 1
				nginx_ingress_controller_nginx_process_connections_total{controller_class="nginx",controller_namespace="default",controller_pod="pod",state="handled"} 2
				# HELP nginx_ingress_controller_nginx_process_requests_total total number of client requests
				# TYPE nginx_ingress_controller_nginx_process_requests_total counter
				nginx_ingress_controller_nginx_process_requests_total{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 3
			`,
			metrics: []string{
				"nginx_ingress_controller_nginx_process_connections_total",
				"nginx_ingress_controller_nginx_process_requests_total",
				"nginx_ingress_controller_nginx_process_connections",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%v", nginx.StatusPort))
			if err != nil {
				t.Fatalf("crating unix listener: %s", err)
			}

			server := &httptest.Server{
				Listener: listener,
				Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)

					if r.URL.Path == "/nginx_status" {
						_, err := fmt.Fprintf(w, c.mock)
						if err != nil {
							t.Fatal(err)
						}

						return
					}

					fmt.Fprintf(w, "OK")
				})},
			}
			server.Start()

			time.Sleep(1 * time.Second)

			cm, err := NewNGINXStatus("pod", "default", "nginx")
			if err != nil {
				t.Errorf("unexpected error creating nginx status collector: %v", err)
			}

			go cm.Start()

			reg := prometheus.NewPedanticRegistry()
			if err := reg.Register(cm); err != nil {
				t.Errorf("registering collector failed: %s", err)
			}

			if err := GatherAndCompare(cm, c.want, c.metrics, reg); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}

			reg.Unregister(cm)

			server.Close()
			cm.Stop()

			listener.Close()
		})
	}
}
