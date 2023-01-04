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

package sessionaffinity

import (
	"testing"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	defaultBackend := networking.IngressBackend{
		Service: &networking.IngressServiceBackend{
			Name: "default-backend",
			Port: networking.ServiceBackendPort{
				Number: 80,
			},
		},
	}

	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "default-backend",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
			},
			Rules: []networking.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
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

func TestIngressAffinityCookieConfig(t *testing.T) {
	ing := buildIngress()
	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(annotationAffinityType)] = "cookie"
	data[parser.GetAnnotationWithPrefix(annotationAffinityMode)] = "balanced"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieName)] = "INGRESSCOOKIE"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieExpires)] = "4500"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieMaxAge)] = "3000"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookiePath)] = "/foo"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieDomain)] = "foo.bar"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieChangeOnFailure)] = "true"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieSecure)] = "true"
	ing.SetAnnotations(data)

	affin, _ := NewParser(&resolver.Mock{}).Parse(ing)
	nginxAffinity, ok := affin.(*Config)
	if !ok {
		t.Errorf("expected a Config type")
	}

	if nginxAffinity.Type != "cookie" {
		t.Errorf("expected cookie as affinity but returned %v", nginxAffinity.Type)
	}

	if nginxAffinity.Mode != "balanced" {
		t.Errorf("expected balanced as affinity mode but returned %v", nginxAffinity.Mode)
	}

	if nginxAffinity.Cookie.Name != "INGRESSCOOKIE" {
		t.Errorf("expected INGRESSCOOKIE as session-cookie-name but returned %v", nginxAffinity.Cookie.Name)
	}

	if nginxAffinity.Cookie.Expires != "4500" {
		t.Errorf("expected 1h as session-cookie-expires but returned %v", nginxAffinity.Cookie.Expires)
	}

	if nginxAffinity.Cookie.MaxAge != "3000" {
		t.Errorf("expected 3000 as session-cookie-max-age but returned %v", nginxAffinity.Cookie.MaxAge)
	}

	if nginxAffinity.Cookie.Path != "/foo" {
		t.Errorf("expected /foo as session-cookie-path but returned %v", nginxAffinity.Cookie.Path)
	}

	if nginxAffinity.Cookie.Domain != "foo.bar" {
		t.Errorf("expected foo.bar as session-cookie-domain but returned %v", nginxAffinity.Cookie.Domain)
	}

	if !nginxAffinity.Cookie.ChangeOnFailure {
		t.Errorf("expected change of failure parameter set to true but returned %v", nginxAffinity.Cookie.ChangeOnFailure)
	}

	if !nginxAffinity.Cookie.Secure {
		t.Errorf("expected secure parameter set to true but returned %v", nginxAffinity.Cookie.Secure)
	}
}
