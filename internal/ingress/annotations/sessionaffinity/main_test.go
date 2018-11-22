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
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
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

func TestIngressAffinityCookieConfig(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(annotationAffinityType)] = "cookie"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieHash)] = "sha123"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieName)] = "INGRESSCOOKIE"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieExpires)] = "4500"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookieMaxAge)] = "3000"
	data[parser.GetAnnotationWithPrefix(annotationAffinityCookiePath)] = "/foo"
	ing.SetAnnotations(data)

	affin, _ := NewParser(&resolver.Mock{}).Parse(ing)
	nginxAffinity, ok := affin.(*Config)
	if !ok {
		t.Errorf("expected a Config type")
	}

	if nginxAffinity.Type != "cookie" {
		t.Errorf("expected cookie as affinity but returned %v", nginxAffinity.Type)
	}

	if nginxAffinity.Cookie.Hash != "md5" {
		t.Errorf("expected md5 as session-cookie-hash but returned %v", nginxAffinity.Cookie.Hash)
	}

	if nginxAffinity.Cookie.Name != "INGRESSCOOKIE" {
		t.Errorf("expected route as session-cookie-name but returned %v", nginxAffinity.Cookie.Name)
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
}
