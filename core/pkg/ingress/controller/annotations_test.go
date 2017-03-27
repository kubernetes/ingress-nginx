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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"

	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	annotationSecureUpstream     = "ingress.kubernetes.io/secure-backends"
	annotationUpsMaxFails        = "ingress.kubernetes.io/upstream-max-fails"
	annotationUpsFailTimeout     = "ingress.kubernetes.io/upstream-fail-timeout"
	annotationPassthrough        = "ingress.kubernetes.io/ssl-passthrough"
	annotationAffinityType       = "ingress.kubernetes.io/affinity"
	annotationAffinityCookieName = "ingress.kubernetes.io/session-cookie-name"
	annotationAffinityCookieHash = "ingress.kubernetes.io/session-cookie-hash"
	annotationAuthTlsSecret      = "ingress.kubernetes.io/auth-tls-secret"
)

type mockCfg struct {
	MockSecrets map[string]*api.Secret
}

func (m mockCfg) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{}
}

func (m mockCfg) GetSecret(name string) (*api.Secret, error) {
	return m.MockSecrets[name], nil
}

func (m mockCfg) GetAuthCertificate(string) (*resolver.AuthSSLCert, error) {
	return nil, nil
}

func TestAnnotationExtractor(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})
	ing := buildIngress()

	m := ec.Extract(ing)
	// the map at least should contains HealthCheck and Proxy information (defaults)
	if _, ok := m["HealthCheck"]; !ok {
		t.Error("expected HealthCheck annotation")
	}
	if _, ok := m["Proxy"]; !ok {
		t.Error("expected Proxy annotation")
	}
}

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

func TestSecureUpstream(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})
	ing := buildIngress()

	fooAnns := []struct {
		annotations map[string]string
		er          bool
	}{
		{map[string]string{annotationSecureUpstream: "true"}, true},
		{map[string]string{annotationSecureUpstream: "false"}, false},
		{map[string]string{annotationSecureUpstream + "_no": "true"}, false},
		{map[string]string{}, false},
		{nil, false},
	}

	for _, foo := range fooAnns {
		ing.SetAnnotations(foo.annotations)
		r := ec.SecureUpstream(ing)
		if r != foo.er {
			t.Errorf("Returned %v but expected %v", r, foo.er)
		}
	}
}

func TestHealthCheck(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})
	ing := buildIngress()

	fooAnns := []struct {
		annotations map[string]string
		eumf        int
		euft        int
	}{
		{map[string]string{annotationUpsMaxFails: "3", annotationUpsFailTimeout: "10"}, 3, 10},
		{map[string]string{annotationUpsMaxFails: "3"}, 3, 0},
		{map[string]string{annotationUpsFailTimeout: "10"}, 0, 10},
		{map[string]string{}, 0, 0},
		{nil, 0, 0},
	}

	for _, foo := range fooAnns {
		ing.SetAnnotations(foo.annotations)
		r := ec.HealthCheck(ing)
		if r == nil {
			t.Errorf("Returned nil but expected a healthcheck.Upstream")
			continue
		}

		if r.FailTimeout != foo.euft {
			t.Errorf("Returned %d but expected %d for FailTimeout", r.FailTimeout, foo.euft)
		}

		if r.MaxFails != foo.eumf {
			t.Errorf("Returned %d but expected %d for MaxFails", r.MaxFails, foo.eumf)
		}
	}
}

func TestSSLPassthrough(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})
	ing := buildIngress()

	fooAnns := []struct {
		annotations map[string]string
		er          bool
	}{
		{map[string]string{annotationPassthrough: "true"}, true},
		{map[string]string{annotationPassthrough: "false"}, false},
		{map[string]string{annotationPassthrough + "_no": "true"}, false},
		{map[string]string{}, false},
		{nil, false},
	}

	for _, foo := range fooAnns {
		ing.SetAnnotations(foo.annotations)
		r := ec.SSLPassthrough(ing)
		if r != foo.er {
			t.Errorf("Returned %v but expected %v", r, foo.er)
		}
	}
}

func TestAffinitySession(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})
	ing := buildIngress()

	fooAnns := []struct {
		annotations  map[string]string
		affinitytype string
		hash         string
		name         string
	}{
		{map[string]string{annotationAffinityType: "cookie", annotationAffinityCookieHash: "md5", annotationAffinityCookieName: "route"}, "cookie", "md5", "route"},
		{map[string]string{annotationAffinityType: "cookie", annotationAffinityCookieHash: "xpto", annotationAffinityCookieName: "route1"}, "cookie", "md5", "route1"},
		{map[string]string{annotationAffinityType: "cookie", annotationAffinityCookieHash: "", annotationAffinityCookieName: ""}, "cookie", "md5", "INGRESSCOOKIE"},
		{map[string]string{}, "", "", ""},
		{nil, "", "", ""},
	}

	for _, foo := range fooAnns {
		ing.SetAnnotations(foo.annotations)
		r := ec.SessionAffinity(ing)
		t.Logf("Testing pass %v %v %v", foo.affinitytype, foo.hash, foo.name)
		if r == nil {
			t.Errorf("Returned nil but expected a SessionAffinity.AffinityConfig")
			continue
		}

		if r.CookieConfig.Hash != foo.hash {
			t.Errorf("Returned %v but expected %v for Hash", r.CookieConfig.Hash, foo.hash)
		}

		if r.CookieConfig.Name != foo.name {
			t.Errorf("Returned %v but expected %v for Name", r.CookieConfig.Name, foo.name)
		}
	}
}

func TestContainsCertificateAuth(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{})

	foos := []struct {
		name        string
		annotations map[string]string
		result      bool
	}{
		{"nil_annotations", nil, false},
		{"empty_annotations", map[string]string{}, false},
		{"not_exist_annotations", map[string]string{annotationAffinityType: "cookie"}, false},
		{"exist_annotations", map[string]string{annotationAuthTlsSecret: "default/foo_secret"}, true},
	}

	for _, foo := range foos {
		t.Run(foo.name, func(t *testing.T) {
			ing := buildIngress()
			ing.SetAnnotations(foo.annotations)
			r := ec.ContainsCertificateAuth(ing)
			if r != foo.result {
				t.Errorf("Returned %t but expected %t for %s", r, foo.result, foo.name)
			}
		})
	}
}

func TestCertificateAuthSecret(t *testing.T) {
	resolver := mockCfg{}
	resolver.MockSecrets = map[string]*api.Secret{
		"default/foo_secret": {
			ObjectMeta: api.ObjectMeta{
				Name: "foo_secret_name",
			},
		},
	}
	ec := newAnnotationExtractor(resolver)

	foos := []struct {
		name        string
		annotations map[string]string
		eerr        bool
		ename       string
	}{
		{"nil_annotations", nil, true, ""},
		{"empty_annotations", map[string]string{}, true, ""},
		{"not_exist_annotations", map[string]string{annotationAffinityType: "cookie"}, true, ""},
		{"exist_annotations", map[string]string{annotationAuthTlsSecret: "default/foo_secret"}, false, "foo_secret_name"},
	}

	for _, foo := range foos {
		t.Run(foo.name, func(t *testing.T) {
			ing := buildIngress()
			ing.SetAnnotations(foo.annotations)
			r, err := ec.CertificateAuthSecret(ing)

			if foo.eerr {
				if err == nil {
					t.Fatalf("Exepected error for %s", foo.name)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error %v for %s", err, foo.name)
				}

				rname := ""
				if r != nil {
					rname = r.GetName()
				}
				if rname != foo.ename {
					t.Errorf("Returned %s but expected %s for %s", rname, foo.ename, foo.name)
				}
			}
		})
	}
}
