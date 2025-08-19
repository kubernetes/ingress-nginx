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
	networking "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	defaultDemoSecret = "default/demo-secret"
	off               = "off"
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

// mocks the resolver for authTLS
type mockSecret struct {
	resolver.Mock
}

// GetAuthCertificate from mockSecret mocks the GetAuthCertificate for authTLS
func (m mockSecret) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if name != defaultDemoSecret {
		return nil, errors.Errorf("there is no secret with name %v", name)
	}

	return &resolver.AuthSSLCert{
		Secret:     defaultDemoSecret,
		CAFileName: "/ssl/ca.crt",
		CASHA:      "abc",
	}, nil
}

func TestAnnotations(t *testing.T) {
	ing := buildIngress()
	data := map[string]string{}

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSSecret)] = defaultDemoSecret

	ing.SetAnnotations(data)

	fakeSecret := &mockSecret{}
	i, err := NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}

	u, ok := i.(*Config)
	if !ok {
		t.Errorf("expected *Config but got %v", u)
	}

	secret, err := fakeSecret.GetAuthCertificate(defaultDemoSecret)
	if err != nil {
		t.Errorf("unexpected error getting secret %v", err)
	}

	if u.AuthSSLCert.Secret != secret.Secret {
		t.Errorf("expected %v but got %v", secret.Secret, u.AuthSSLCert.Secret)
	}
	if u.VerifyClient != "on" {
		t.Errorf("expected %v but got %v", "on", u.VerifyClient)
	}
	if u.ValidationDepth != 1 {
		t.Errorf("expected %v but got %v", 1, u.ValidationDepth)
	}
	if u.ErrorPage != "" {
		t.Errorf("expected %v but got %v", "", u.ErrorPage)
	}
	if u.PassCertToUpstream != false {
		t.Errorf("expected %v but got %v", false, u.PassCertToUpstream)
	}
	if u.MatchCN != "" {
		t.Errorf("expected empty string, but got %v", u.MatchCN)
	}

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSVerifyClient)] = off
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSVerifyDepth)] = "2"
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSErrorPage)] = "ok.com/error"
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSPassCertToUpstream)] = "true"
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSMatchCN)] = "CN=(hello-app|ok|goodbye)"

	ing.SetAnnotations(data)

	i, err = NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}

	u, ok = i.(*Config)
	if !ok {
		t.Errorf("expected *Config but got %v", u)
	}

	if u.AuthSSLCert.Secret != secret.Secret {
		t.Errorf("expected %v but got %v", secret.Secret, u.AuthSSLCert.Secret)
	}
	if u.VerifyClient != off {
		t.Errorf("expected %v but got %v", off, u.VerifyClient)
	}
	if u.ValidationDepth != 2 {
		t.Errorf("expected %v but got %v", 2, u.ValidationDepth)
	}
	if u.ErrorPage != "ok.com/error" {
		t.Errorf("expected %v but got %v", "ok.com/error", u.ErrorPage)
	}
	if u.PassCertToUpstream != true {
		t.Errorf("expected %v but got %v", true, u.PassCertToUpstream)
	}
	if u.MatchCN != "CN=(hello-app|ok|goodbye)" {
		t.Errorf("expected %v but got %v", "CN=(hello-app|ok|goodbye)", u.MatchCN)
	}
}

func TestInvalidAnnotations(t *testing.T) {
	ing := buildIngress()
	fakeSecret := &mockSecret{}
	data := map[string]string{}

	// No annotation
	_, err := NewParser(fakeSecret).Parse(ing)
	if err == nil {
		t.Errorf("Expected error with ingress but got nil")
	}

	// Invalid NameSpace
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSSecret)] = "demo-secret"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err == nil {
		t.Errorf("Expected error with ingress but got nil")
	}

	// Invalid Cross NameSpace
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSSecret)] = "nondefault/demo-secret"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	expErr := errors.NewLocationDenied("cross namespace secrets are not supported")
	if err.Error() != expErr.Error() {
		t.Errorf("received error is different from cross namespace error: %s Expected %s", err, expErr)
	}

	// Invalid Auth Certificate
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSSecret)] = "default/invalid-demo-secret"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err == nil {
		t.Errorf("Expected error with ingress but got nil")
	}

	// Invalid optional Annotations
	data[parser.GetAnnotationWithPrefix(annotationAuthTLSSecret)] = "default/demo-secret"

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSVerifyClient)] = "w00t"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Error should be nil and verify client should be defaulted")
	}

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSVerifyDepth)] = "abcd"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Error should be nil and verify depth should be defaulted")
	}

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSPassCertToUpstream)] = "nahh"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err == nil {
		t.Errorf("Expected error with ingress but got nil")
	}
	delete(data, parser.GetAnnotationWithPrefix(annotationAuthTLSPassCertToUpstream))

	data[parser.GetAnnotationWithPrefix(annotationAuthTLSMatchCN)] = "<script>nope</script>"
	ing.SetAnnotations(data)
	_, err = NewParser(fakeSecret).Parse(ing)
	if err == nil {
		t.Errorf("Expected error with ingress CN but got nil")
	}
	delete(data, parser.GetAnnotationWithPrefix(annotationAuthTLSMatchCN))

	ing.SetAnnotations(data)

	i, err := NewParser(fakeSecret).Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	u, ok := i.(*Config)
	if !ok {
		t.Errorf("expected *Config but got %v", u)
	}

	if u.VerifyClient != "on" {
		t.Errorf("expected %v but got %v", "on", u.VerifyClient)
	}
	if u.ValidationDepth != 1 {
		t.Errorf("expected %v but got %v", 1, u.ValidationDepth)
	}
	if u.PassCertToUpstream != false {
		t.Errorf("expected %v but got %v", false, u.PassCertToUpstream)
	}
	if u.MatchCN != "" {
		t.Errorf("expected empty string but got %v", u.MatchCN)
	}
}

func TestEquals(t *testing.T) {
	cfg1 := &Config{}
	cfg2 := &Config{}

	// compare nil
	result := cfg1.Equal(nil)
	if result != false {
		t.Errorf("Expected false")
	}

	// Different Certs
	sslCert1 := resolver.AuthSSLCert{
		Secret:     defaultDemoSecret,
		CAFileName: "/ssl/ca.crt",
		CASHA:      "abc",
	}
	sslCert2 := resolver.AuthSSLCert{
		Secret:     "default/other-demo-secret",
		CAFileName: "/ssl/ca.crt",
		CASHA:      "abc",
	}
	cfg1.AuthSSLCert = sslCert1
	cfg2.AuthSSLCert = sslCert2
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.AuthSSLCert = sslCert1

	// Different Verify Client
	cfg1.VerifyClient = "on"
	cfg2.VerifyClient = off
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.VerifyClient = "on"

	// Different Validation Depth
	cfg1.ValidationDepth = 1
	cfg2.ValidationDepth = 2
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.ValidationDepth = 1

	// Different Error Page
	cfg1.ErrorPage = "error-1"
	cfg2.ErrorPage = "error-2"
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.ErrorPage = "error-1"

	// Different Pass to Upstream
	cfg1.PassCertToUpstream = true
	cfg2.PassCertToUpstream = false
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.PassCertToUpstream = true

	// Different MatchCN
	cfg1.MatchCN = "CN=(hello-app|goodbye)"
	cfg2.MatchCN = "CN=(hello-app)"
	result = cfg1.Equal(cfg2)
	if result != false {
		t.Errorf("Expected false")
	}
	cfg2.MatchCN = "CN=(hello-app|goodbye)"

	// Equal Configs
	result = cfg1.Equal(cfg2)
	if result != true {
		t.Errorf("Expected true")
	}
}
