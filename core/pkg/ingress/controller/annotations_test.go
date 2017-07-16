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

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	annotationSecureUpstream     = "ingress.kubernetes.io/secure-backends"
	annotationSecureVerifyCACert = "ingress.kubernetes.io/secure-verify-ca-secret"
	annotationUpsMaxFails        = "ingress.kubernetes.io/upstream-max-fails"
	annotationUpsFailTimeout     = "ingress.kubernetes.io/upstream-fail-timeout"
	annotationPassthrough        = "ingress.kubernetes.io/ssl-passthrough"
	annotationAffinityType       = "ingress.kubernetes.io/affinity"
	annotationAffinityCookieName = "ingress.kubernetes.io/session-cookie-name"
	annotationAffinityCookieHash = "ingress.kubernetes.io/session-cookie-hash"
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

func (m mockCfg) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if secret, _ := m.GetSecret(name); secret != nil {
		return &resolver.AuthSSLCert{
			Secret:     name,
			CAFileName: "/opt/ca.pem",
			PemSHA:     "123",
		}, nil
	}
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
		if r.Secure != foo.er {
			t.Errorf("Returned %v but expected %v", r, foo.er)
		}
	}
}

func TestSecureVerifyCACert(t *testing.T) {
	ec := newAnnotationExtractor(mockCfg{
		MockSecrets: map[string]*api.Secret{
			"default/secure-verify-ca": {
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "secure-verify-ca",
				},
			},
		},
	})

	anns := []struct {
		it          int
		annotations map[string]string
		exists      bool
	}{
		{1, map[string]string{annotationSecureUpstream: "true", annotationSecureVerifyCACert: "not"}, false},
		{2, map[string]string{annotationSecureUpstream: "false", annotationSecureVerifyCACert: "secure-verify-ca"}, false},
		{3, map[string]string{annotationSecureUpstream: "true", annotationSecureVerifyCACert: "secure-verify-ca"}, true},
		{4, map[string]string{annotationSecureUpstream: "true", annotationSecureVerifyCACert + "_not": "secure-verify-ca"}, false},
		{5, map[string]string{annotationSecureUpstream: "true"}, false},
		{6, map[string]string{}, false},
		{7, nil, false},
	}

	for _, ann := range anns {
		ing := buildIngress()
		ing.SetAnnotations(ann.annotations)
		res := ec.SecureUpstream(ing)
		if (res.CACert.CAFileName != "") != ann.exists {
			t.Errorf("Expected exists was %v on iteration %v", ann.exists, ann.it)
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
