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

package authtls

import (
	"testing"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
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

// mocks the resolver for authTLS
type mockSecret struct {
	resolver.Mock
}

// GetAuthCertificate from mockSecret mocks the GetAuthCertificate for authTLS
func (m mockSecret) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if name != "default/demo-secret" {
		return nil, errors.Errorf("there is no secret with name %v", name)
	}

	return &resolver.AuthSSLCert{
		Secret:     "default/demo-secret",
		CAFileName: "/ssl/ca.crt",
		PemSHA:     "abc",
	}, nil

}

func TestAnnotations(t *testing.T) {
	ing := buildIngress()
	data := map[string]string{}

	data[parser.GetAnnotationWithPrefix("auth-tls-secret")] = "default/demo-secret"
	data[parser.GetAnnotationWithPrefix("auth-tls-verify-client")] = "off"
	data[parser.GetAnnotationWithPrefix("auth-tls-verify-depth")] = "1"
	data[parser.GetAnnotationWithPrefix("auth-tls-error-page")] = "ok.com/error"
	data[parser.GetAnnotationWithPrefix("auth-tls-pass-certificate-to-upstream")] = "true"

	ing.SetAnnotations(data)

	fakeSecret := &mockSecret{}
	i, err := NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Uxpected error with ingress: %v", err)
	}

	u, ok := i.(*Config)
	if !ok {
		t.Errorf("expected *Config but got %v", u)
	}

	secret, err := fakeSecret.GetAuthCertificate("default/demo-secret")
	if err != nil {
		t.Errorf("unexpected error getting secret %v", err)
	}

	if u.AuthSSLCert.Secret != secret.Secret {
		t.Errorf("expected %v but got %v", secret.Secret, u.AuthSSLCert.Secret)
	}
	if u.VerifyClient != "off" {
		t.Errorf("expected %v but got %v", "off", u.VerifyClient)
	}
	if u.ValidationDepth != 1 {
		t.Errorf("expected %v but got %v", 1, u.ValidationDepth)
	}
	if u.ErrorPage != "ok.com/error" {
		t.Errorf("expected %v but got %v", "ok.com/error", u.ErrorPage)
	}
	if u.PassCertToUpstream != true {
		t.Errorf("expected %v but got %v", true, u.PassCertToUpstream)
	}
}
