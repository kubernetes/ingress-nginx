/*
Copyright 2015 The Kubernetes Authors.

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

package rewrite

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
)

const (
	defRoute = "/demo"
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

	r := ingAnnotations(ing.GetAnnotations()).rewriteTo()
	if r != "" {
		t.Error("Expected no redirect")
	}

	f := ingAnnotations(ing.GetAnnotations()).addBaseURL()
	if f {
		t.Errorf("Expected false in add-base-url but %v was returend", f)
	}

	data := map[string]string{}
	data[rewriteTo] = defRoute
	data[addBaseURL] = "true"
	ing.SetAnnotations(data)

	r = ingAnnotations(ing.GetAnnotations()).rewriteTo()
	if r != defRoute {
		t.Errorf("Expected %v in rewrite but %v was returend", defRoute, r)
	}

	f = ingAnnotations(ing.GetAnnotations()).addBaseURL()
	if !f {
		t.Errorf("Expected true in add-base-url but %v was returend", f)
	}
}

func TestWithoutAnnotations(t *testing.T) {
	ing := buildIngress()
	_, err := ParseAnnotations(config.NewDefault(), ing)
	if err == nil {
		t.Error("Expected error with ingress without annotations")
	}
}

func TestRedirect(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[rewriteTo] = defRoute
	ing.SetAnnotations(data)

	redirect, err := ParseAnnotations(config.NewDefault(), ing)
	if err != nil {
		t.Errorf("Uxpected error with ingress: %v", err)
	}

	if redirect.Target != defRoute {
		t.Errorf("Expected %v as redirect but returned %s", defRoute, redirect.Target)
	}
}

func TestSSLRedirect(t *testing.T) {
	ing := buildIngress()

	cfg := config.Configuration{SSLRedirect: true}

	data := map[string]string{}

	ing.SetAnnotations(data)

	redirect, _ := ParseAnnotations(cfg, ing)

	if !redirect.SSLRedirect {
		t.Errorf("Expected true but returned false")
	}

	data[sslRedirect] = "false"
	ing.SetAnnotations(data)

	redirect, _ = ParseAnnotations(cfg, ing)

	if redirect.SSLRedirect {
		t.Errorf("Expected false but returned true")
	}
}
