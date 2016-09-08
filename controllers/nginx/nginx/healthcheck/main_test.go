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

package healthcheck

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
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

func TestAnnotations(t *testing.T) {
	ing := buildIngress()

	_, err := ingAnnotations(ing.GetAnnotations()).maxFails()
	if err == nil {
		t.Error("Expected a validation error")
	}

	_, err = ingAnnotations(ing.GetAnnotations()).failTimeout()
	if err == nil {
		t.Error("Expected a validation error")
	}

	data := map[string]string{}
	data[upsMaxFails] = "1"
	data[upsFailTimeout] = "1"
	ing.SetAnnotations(data)

	mf, err := ingAnnotations(ing.GetAnnotations()).maxFails()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mf != 1 {
		t.Errorf("Expected 1 but returned %v", mf)
	}

	ft, err := ingAnnotations(ing.GetAnnotations()).failTimeout()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ft != 1 {
		t.Errorf("Expected 1 but returned %v", ft)
	}
}

func TestIngressHealthCheck(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[upsMaxFails] = "2"
	ing.SetAnnotations(data)

	cfg := config.Configuration{}
	cfg.UpstreamFailTimeout = 1

	nginxHz := ParseAnnotations(cfg, ing)

	if nginxHz.MaxFails != 2 {
		t.Errorf("Expected 2 as max-fails but returned %v", nginxHz.MaxFails)
	}

	if nginxHz.FailTimeout != 1 {
		t.Errorf("Expected 0 as fail-timeout but returned %v", nginxHz.FailTimeout)
	}
}
