/*
Copyright 2016 The Kubernetes Authors.

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

package stickysession

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path:    "/foo",
									Backend: defaultBackend,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestIngressHealthCheck(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[stickyEnabled] = "true"
	data[stickyHash] = "md5"
	data[stickyName] = "route1"
	ing.SetAnnotations(data)

	sti, _ := NewParser().Parse(ing)
	nginxSti, ok := sti.(*StickyConfig)
	if !ok {
		t.Errorf("expected a StickyConfig type")
	}

	if nginxSti.Hash != "md5" {
		t.Errorf("expected md5 as sticky-hash but returned %v", nginxSti.Hash)
	}

	if nginxSti.Hash != "md5" {
		t.Errorf("expected md5 as sticky-hash but returned %v", nginxSti.Hash)
	}

	if !nginxSti.Enabled {
		t.Errorf("expected sticky-enabled  but returned %v", nginxSti.Enabled)
	}
}
