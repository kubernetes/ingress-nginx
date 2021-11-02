/*
Copyright 2021 The Kubernetes Authors.

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

	"github.com/prometheus/client_golang/prometheus"
)

func TestAdmissionCounters(t *testing.T) {
	const (
		metadataFirst = `
			# HELP nginx_ingress_controller_admission_config_size The size of the tested configuration
			# TYPE nginx_ingress_controller_admission_config_size gauge
			# HELP nginx_ingress_controller_admission_roundtrip_duration The complete duration of the admission controller at the time to process a new event (float seconds)
			# TYPE nginx_ingress_controller_admission_roundtrip_duration gauge
		`
		metadataSecond = `
			# HELP nginx_ingress_controller_admission_render_ingresses The length of ingresses rendered by the admission controller
			# TYPE nginx_ingress_controller_admission_render_ingresses gauge
			# HELP nginx_ingress_controller_admission_tested_duration The processing duration of the admission controller tests (float seconds)
			# TYPE nginx_ingress_controller_admission_tested_duration gauge
		`
		metadataThird = `
			# HELP nginx_ingress_controller_admission_config_size The size of the tested configuration
			# TYPE nginx_ingress_controller_admission_config_size gauge
			# HELP nginx_ingress_controller_admission_render_duration The processing duration of ingresses rendering by the admission controller (float seconds)
			# TYPE nginx_ingress_controller_admission_render_duration gauge
			# HELP nginx_ingress_controller_admission_render_ingresses The length of ingresses rendered by the admission controller
			# TYPE nginx_ingress_controller_admission_render_ingresses gauge
			# HELP nginx_ingress_controller_admission_roundtrip_duration The complete duration of the admission controller at the time to process a new event (float seconds)
			# TYPE nginx_ingress_controller_admission_roundtrip_duration gauge
			# HELP nginx_ingress_controller_admission_tested_ingresses The length of ingresses processed by the admission controller
			# TYPE nginx_ingress_controller_admission_tested_ingresses gauge
			# HELP nginx_ingress_controller_admission_tested_duration The processing duration of the admission controller tests (float seconds)
			# TYPE nginx_ingress_controller_admission_tested_duration gauge
		`
	)
	cases := []struct {
		name    string
		test    func(*AdmissionCollector)
		metrics []string
		want    string
	}{
		{
			name: "should return 0 as values on a fresh initiated collector",
			test: func(am *AdmissionCollector) {
			},
			want: metadataFirst + `
				nginx_ingress_controller_admission_config_size{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 0
				nginx_ingress_controller_admission_roundtrip_duration{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 0
			`,
			metrics: []string{"nginx_ingress_controller_admission_config_size", "nginx_ingress_controller_admission_roundtrip_duration"},
		},
		{
			name: "set admission metrics to 1 in all fields and validate next set",
			test: func(am *AdmissionCollector) {
				am.SetAdmissionMetrics(1, 1, 1, 1, 1, 1)
			},
			want: metadataSecond + `
				nginx_ingress_controller_admission_render_ingresses{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 1
				nginx_ingress_controller_admission_tested_duration{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 1
			`,
			metrics: []string{"nginx_ingress_controller_admission_render_ingresses", "nginx_ingress_controller_admission_tested_duration"},
		},
		{
			name: "set admission metrics to 5 in all fields and validate all sets",
			test: func(am *AdmissionCollector) {
				am.SetAdmissionMetrics(5, 5, 5, 5, 5, 5)
			},
			want: metadataThird + `
				nginx_ingress_controller_admission_config_size{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
				nginx_ingress_controller_admission_render_duration{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
				nginx_ingress_controller_admission_render_ingresses{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
				nginx_ingress_controller_admission_roundtrip_duration{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
				nginx_ingress_controller_admission_tested_ingresses{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
				nginx_ingress_controller_admission_tested_duration{controller_class="nginx",controller_namespace="default",controller_pod="pod"} 5
			`,
			metrics: []string{
				"nginx_ingress_controller_admission_config_size",
				"nginx_ingress_controller_admission_render_duration",
				"nginx_ingress_controller_admission_render_ingresses",
				"nginx_ingress_controller_admission_roundtrip_duration",
				"nginx_ingress_controller_admission_tested_ingresses",
				"nginx_ingress_controller_admission_tested_duration",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			am := NewAdmissionCollector("pod", "default", "nginx")
			reg := prometheus.NewPedanticRegistry()
			if err := reg.Register(am); err != nil {
				t.Errorf("registering collector failed: %s", err)
			}

			c.test(am)

			if err := GatherAndCompare(am, c.want, c.metrics, reg); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}

			reg.Unregister(am)
		})
	}
}
