/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"reflect"
	"testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/ingress/core/pkg/ingress/annotations/service"
)

func buildSimpleClientSet() *fake.Clientset {
	return fake.NewSimpleClientset(
		&api_v1.PodList{
			Items: []api_v1.Pod{
				{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "foo1",
						Namespace: api.NamespaceDefault,
						Labels: map[string]string{
							"lable_sig": "foo_pod",
						},
					},
					Spec: api_v1.PodSpec{
						NodeName: "foo_node_1",
						Containers: []api_v1.Container{
							{
								Ports: []api_v1.ContainerPort{
									{
										Name:          "foo1_named_port_c1",
										Protocol:      api_v1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      "foo1",
						Namespace: api.NamespaceSystem,
						Labels: map[string]string{
							"lable_sig": "foo_pod",
						},
					},
				},
			},
		},
		&api_v1.ServiceList{Items: []api_v1.Service{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: api.NamespaceDefault,
					Name:      "named_port_test_service",
				},
			},
		}},
	)
}

func buildGenericController() *GenericController {
	return &GenericController{
		cfg: &Configuration{
			Client: buildSimpleClientSet(),
		},
	}
}

func buildService() *api_v1.Service {
	return &api_v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: api.NamespaceSystem,
			Name:      "named_port_test_service",
		},
		Spec: api_v1.ServiceSpec{
			ClusterIP: "10.10.10.10",
		},
	}
}

func TestCheckSvcForUpdate(t *testing.T) {
	foos := []struct {
		n   string
		ns  string
		sps []api_v1.ServicePort
		sl  map[string]string
		er  string
	}{
		{
			"pods_have_not_been_found_in_this_namespace",
			api.NamespaceSystem,
			[]api_v1.ServicePort{
				{Name: "foo_port_1", Port: 8080, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("foo1_named_port_c1")},
				{Name: "foo_port_2", Port: 8181, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(81)},
				{Name: "foo_port_3", Port: 8282, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("")},
			},
			map[string]string{
				"lable_sig": "foo_pod",
			},
			"",
		},
		{
			"ports_have_not_been_found_in_this_pod",
			api.NamespaceDefault,
			[]api_v1.ServicePort{
				{Name: "foo_port_1", Port: 8080, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("foo1_named_port_cXX")},
				{Name: "foo_port_2", Port: 8181, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(81)},
				{Name: "foo_port_3", Port: 8282, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("")},
			},
			map[string]string{
				"lable_sig": "foo_pod",
			},
			"",
		},

		{
			"ports_fixed",
			api.NamespaceDefault,
			[]api_v1.ServicePort{
				{Name: "foo_port_1", Port: 8080, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(80)},
				{Name: "foo_port_2", Port: 8181, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(81)},
				{Name: "foo_port_3", Port: 8282, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("")},
			},
			map[string]string{
				"lable_sig": "foo_pod",
			},
			"",
		},
		{
			"nil_selector",
			api.NamespaceDefault,
			[]api_v1.ServicePort{
				{Name: "foo_port_1", Port: 8080, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("foo1_named_port_c1")},
				{Name: "foo_port_2", Port: 8181, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(81)},
				{Name: "foo_port_3", Port: 8282, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("")},
			},
			nil,
			"{\"foo1_named_port_c1\":\"80\"}",
		},
		{
			"normal_update",
			api.NamespaceDefault,
			[]api_v1.ServicePort{
				{Name: "foo_port_1", Port: 8080, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("foo1_named_port_c1")},
				{Name: "foo_port_2", Port: 8181, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromInt(81)},
				{Name: "foo_port_3", Port: 8282, Protocol: api_v1.ProtocolTCP, TargetPort: intstr.FromString("")},
			},
			map[string]string{
				"lable_sig": "foo_pod",
			},
			"{\"foo1_named_port_c1\":\"80\"}",
		},
	}

	for _, foo := range foos {
		t.Run(foo.n, func(t *testing.T) {
			gc := buildGenericController()
			s := buildService()
			s.SetNamespace(foo.ns)
			s.Spec.Ports = foo.sps
			s.Spec.Selector = foo.sl

			err := gc.checkSvcForUpdate(s)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			rs, _ := gc.cfg.Client.Core().Services(api.NamespaceDefault).Get("named_port_test_service", meta_v1.GetOptions{})
			rr := rs.ObjectMeta.Annotations[service.NamedPortAnnotation]
			if !reflect.DeepEqual(rr, foo.er) {
				t.Errorf("Returned %s, but expected %s for %s", rr, foo.er, foo.n)
			}
		})
	}
}
