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
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/ingress-nginx/internal/ingress"
)

func TestControllerCounters(t *testing.T) {
	const metadata = `
		# HELP nginx_ingress_controller_config_last_reload_successful Whether the last configuration reload attempt was successful
		# TYPE nginx_ingress_controller_config_last_reload_successful gauge
		# HELP nginx_ingress_controller_success Cumulative number of Ingress controller reload operations
		# TYPE nginx_ingress_controller_success counter
	`
	cases := []struct {
		name    string
		test    func(*Controller)
		metrics []string
		want    string
	}{
		{
			name: "should return not increment in metrics if no operations are invoked",
			test: func(cm *Controller) {
			},
			want: metadata + `
				nginx_ingress_controller_config_last_reload_successful{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 0
			`,
			metrics: []string{"nginx_ingress_controller_config_last_reload_successful", "nginx_ingress_controller_success"},
		},
		{
			name: "single increase in reload count should return 1",
			test: func(cm *Controller) {
				cm.IncReloadCount()
				cm.ConfigSuccess(0, true)
			},
			want: metadata + `
				nginx_ingress_controller_config_last_reload_successful{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 1
				nginx_ingress_controller_success{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 1
			`,
			metrics: []string{"nginx_ingress_controller_config_last_reload_successful", "nginx_ingress_controller_success"},
		},
		{
			name: "single increase in error reload count should return 1",
			test: func(cm *Controller) {
				cm.IncReloadErrorCount()
			},
			want: `
				# HELP nginx_ingress_controller_errors Cumulative number of Ingress controller errors during reload operations
				# TYPE nginx_ingress_controller_errors counter
				nginx_ingress_controller_errors{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 1
			`,
			metrics: []string{"nginx_ingress_controller_errors"},
		},
		{
			name: "should set SSL certificates metrics",
			test: func(cm *Controller) {
				t1, _ := time.Parse(
					time.RFC3339,
					"2012-11-01T22:08:41+00:00")

				servers := []*ingress.Server{
					{
						Hostname: "demo",
						SSLCert: &ingress.SSLCert{
							ExpireTime: t1,
						},
					},
					{
						Hostname: "invalid",
						SSLCert: &ingress.SSLCert{
							ExpireTime: time.Unix(0, 0),
						},
					},
				}
				cm.SetSSLExpireTime(servers)
			},
			want: `
				# HELP nginx_ingress_controller_ssl_expire_time_seconds Number of seconds since 1970 to the SSL Certificate expire.\n			An example to check if this certificate will expire in 10 days is: "nginx_ingress_controller_ssl_expire_time_seconds < (time() + (10 * 24 * 3600))"
				# TYPE nginx_ingress_controller_ssl_expire_time_seconds gauge
				nginx_ingress_controller_ssl_expire_time_seconds{class="nginx",host="demo",namespace="default"} 1.351807721e+09
			`,
			metrics: []string{"nginx_ingress_controller_ssl_expire_time_seconds"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cm := NewController("pod", "default", "nginx")
			reg := prometheus.NewPedanticRegistry()
			if err := reg.Register(cm); err != nil {
				t.Errorf("registering collector failed: %s", err)
			}

			c.test(cm)

			if err := GatherAndCompare(cm, c.want, c.metrics, reg); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}

			reg.Unregister(cm)
		})
	}
}

func TestRemoveMetrics(t *testing.T) {
	cm := NewController("pod", "default", "nginx")
	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(cm); err != nil {
		t.Errorf("registering collector failed: %s", err)
	}

	t1, _ := time.Parse(
		time.RFC3339,
		"2012-11-01T22:08:41+00:00")

	servers := []*ingress.Server{
		{
			Hostname: "demo",
			SSLCert: &ingress.SSLCert{
				ExpireTime: t1,
			},
		},
		{
			Hostname: "invalid",
			SSLCert: &ingress.SSLCert{
				ExpireTime: time.Unix(0, 0),
			},
		},
	}
	cm.SetSSLExpireTime(servers)

	cm.RemoveMetrics([]string{"demo"}, reg)

	if err := GatherAndCompare(cm, "", []string{"nginx_ingress_controller_ssl_expire_time_seconds"}, reg); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	reg.Unregister(cm)
}
