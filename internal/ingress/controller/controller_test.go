/*
Copyright 2018 The Kubernetes Authors.

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
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eapache/channels"
	corev1 "k8s.io/api/core/v1"

	discoveryv1 "k8s.io/api/discovery/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/ingress-nginx/pkg/apis/ingress"

	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/canary"
	"k8s.io/ingress-nginx/internal/ingress/annotations/ipallowlist"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxyssl"
	"k8s.io/ingress-nginx/internal/ingress/annotations/sessionaffinity"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/ingressclass"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/metric"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/net/ssl"

	"k8s.io/ingress-nginx/pkg/util/file"
)

const (
	exampleBackend = "example-http-svc-1-80"
	TRUE           = "true"
)

type fakeIngressStore struct {
	ingresses     []*ingress.Ingress
	configuration ngx_config.Configuration
}

func (fakeIngressStore) GetIngressClass(_ *networking.Ingress, _ *ingressclass.Configuration) (string, error) {
	return "nginx", nil
}

func (fis *fakeIngressStore) GetBackendConfiguration() ngx_config.Configuration {
	return fis.configuration
}

func (fis *fakeIngressStore) GetSecurityConfiguration() defaults.SecurityConfiguration {
	return defaults.SecurityConfiguration{
		AnnotationsRiskLevel:         fis.configuration.AnnotationsRiskLevel,
		AllowCrossNamespaceResources: fis.configuration.AllowCrossNamespaceResources,
	}
}

func (fakeIngressStore) GetConfigMap(_ string) (*corev1.ConfigMap, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetSecret(_ string) (*corev1.Secret, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetService(_ string) (*corev1.Service, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetServiceEndpointsSlices(_ string) ([]*discoveryv1.EndpointSlice, error) {
	return nil, fmt.Errorf("test error")
}

func (fis *fakeIngressStore) ListIngresses() []*ingress.Ingress {
	return fis.ingresses
}

func (fis *fakeIngressStore) FilterIngresses(ingresses []*ingress.Ingress, _ store.IngressFilterFunc) []*ingress.Ingress {
	return ingresses
}

func (fakeIngressStore) GetLocalSSLCert(_ string) (*ingress.SSLCert, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) ListLocalSSLCerts() []*ingress.SSLCert {
	return nil
}

func (fakeIngressStore) GetAuthCertificate(string) (*resolver.AuthSSLCert, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetDefaultBackend() defaults.Backend {
	return defaults.Backend{}
}

func (fakeIngressStore) Run(_ chan struct{}) {}

type testNginxTestCommand struct {
	t        *testing.T
	expected string
	out      []byte
	err      error
}

func (ntc testNginxTestCommand) ExecCommand(_ ...string) *exec.Cmd {
	return nil
}

func (ntc testNginxTestCommand) Test(cfg string) ([]byte, error) {
	fd, err := os.Open(cfg)
	if err != nil {
		ntc.t.Errorf("could not read generated nginx configuration: %v", err.Error())
		return nil, err
	}
	defer fd.Close()
	bytes, err := io.ReadAll(fd)
	if err != nil {
		ntc.t.Errorf("could not read generated nginx configuration: %v", err.Error())
	}
	if string(bytes) != ntc.expected {
		ntc.t.Errorf("unexpected generated configuration %v. Expecting %v", string(bytes), ntc.expected)
	}
	return ntc.out, ntc.err
}

type fakeTemplate struct{}

func (fakeTemplate) Write(conf *ngx_config.TemplateConfig) ([]byte, error) {
	r := []byte{}
	for _, s := range conf.Servers {
		if len(r) > 0 {
			r = append(r, ',')
		}
		r = append(r, []byte(s.Hostname)...)
	}
	return r, nil
}

func TestCheckIngress(t *testing.T) {
	defer func() {
		err := filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, _ error) error {
			if info.IsDir() && os.TempDir() != path {
				return filepath.SkipDir
			}
			if strings.HasPrefix(info.Name(), tempNginxPattern) {
				os.Remove(path)
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	err := file.CreateRequiredDirectories()
	if err != nil {
		t.Fatal(err)
	}

	// Ensure no panic with wrong arguments
	var nginx *NGINXController
	if err := nginx.CheckIngress(nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	nginx = newNGINXController(t)
	if err := nginx.CheckIngress(nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	nginx.metricCollector = metric.DummyCollector{}

	nginx.t = fakeTemplate{}
	nginx.store = &fakeIngressStore{
		ingresses: []*ingress.Ingress{},
	}

	ing := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-ingress",
			Namespace:   "user-namespace",
			Annotations: map[string]string{},
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: "example.com",
				},
			},
		},
	}
	t.Run("when the class is the nginx one", func(t *testing.T) {
		ing.ObjectMeta.Annotations["kubernetes.io/ingress.class"] = "nginx"
		nginx.command = testNginxTestCommand{
			t:        t,
			err:      nil,
			expected: "_,example.com",
		}
		if nginx.CheckIngress(ing) != nil {
			t.Errorf("with a new ingress without error, no error should be returned")
		}

		t.Run("When the hostname is updated", func(t *testing.T) {
			nginx.store = &fakeIngressStore{
				ingresses: []*ingress.Ingress{
					{
						Ingress:           *ing,
						ParsedAnnotations: &annotations.Ingress{},
					},
				},
			}
			ing.Spec.Rules[0].Host = "test.example.com"
			nginx.command = testNginxTestCommand{
				t:        t,
				err:      nil,
				expected: "_,test.example.com",
			}
			if nginx.CheckIngress(ing) != nil {
				t.Errorf("with a new ingress without error, no error should be returned")
			}
		})

		t.Run("When nginx test returns an error", func(t *testing.T) {
			nginx.command = testNginxTestCommand{
				t:        t,
				err:      fmt.Errorf("test error"),
				out:      []byte("this is the test command output"),
				expected: "_,test.example.com",
			}
			if nginx.CheckIngress(ing) == nil {
				t.Errorf("with a new ingress with an error, an error should be returned")
			}
		})

		t.Run("When the default annotation prefix is used despite an override", func(t *testing.T) {
			defer func() {
				parser.AnnotationsPrefix = "nginx.ingress.kubernetes.io"
			}()
			parser.AnnotationsPrefix = "ingress.kubernetes.io"
			ing.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "GRPC"
			nginx.command = testNginxTestCommand{
				t:   t,
				err: nil,
			}
			if nginx.CheckIngress(ing) == nil {
				t.Errorf("with a custom annotation prefix, ingresses using the default should be rejected")
			}
		})

		t.Run("When snippets are disabled and user tries to use snippet annotation", func(t *testing.T) {
			nginx.store = &fakeIngressStore{
				ingresses: []*ingress.Ingress{},
				configuration: ngx_config.Configuration{
					AllowSnippetAnnotations: false,
				},
			}
			nginx.command = testNginxTestCommand{
				t:   t,
				err: nil,
			}
			ing.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/server-snippet"] = "bla"
			if err := nginx.CheckIngress(ing); err == nil {
				t.Errorf("with a snippet annotation, ingresses using the default should be rejected")
			}
		})

		t.Run("When invalid directives are used in annotation values", func(t *testing.T) {
			nginx.store = &fakeIngressStore{
				ingresses: []*ingress.Ingress{},
				configuration: ngx_config.Configuration{
					AnnotationValueWordBlocklist: "invalid_directive, another_directive",
				},
			}
			nginx.command = testNginxTestCommand{
				t:   t,
				err: nil,
			}
			ing.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/custom-headers"] = "invalid_directive"
			if err := nginx.CheckIngress(ing); err == nil {
				t.Errorf("with an invalid value in annotation the ingress should be rejected")
			}
			ing.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/custom-headers"] = "another_directive"
			if err := nginx.CheckIngress(ing); err == nil {
				t.Errorf("with an invalid value in annotation the ingress should be rejected")
			}
		})

		t.Run("When a new catch-all ingress is being created despite catch-alls being disabled ", func(t *testing.T) {
			backendBefore := ing.Spec.DefaultBackend
			disableCatchAllBefore := nginx.cfg.DisableCatchAll

			nginx.command = testNginxTestCommand{
				t:   t,
				err: nil,
			}
			nginx.cfg.DisableCatchAll = true

			ing.Spec.DefaultBackend = &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: "http-svc",
					Port: networking.ServiceBackendPort{
						Number: 80,
					},
				},
			}

			if nginx.CheckIngress(ing) == nil {
				t.Errorf("with a new catch-all ingress and catch-alls disable, should return error")
			}

			// reset backend and catch-all flag
			ing.Spec.DefaultBackend = backendBefore
			nginx.cfg.DisableCatchAll = disableCatchAllBefore
		})

		t.Run("When the ingress is in a different namespace than the watched one", func(t *testing.T) {
			defer func() {
				nginx.cfg.Namespace = "test-namespace"
			}()
			nginx.command = testNginxTestCommand{
				t:   t,
				err: fmt.Errorf("test error"),
			}
			nginx.cfg.Namespace = "other-namespace"
			ing.ObjectMeta.Namespace = "test-namespace"
			if nginx.CheckIngress(ing) != nil {
				t.Errorf("with a new ingress without error, no error should be returned")
			}
		})

		t.Run("When there is a duplicated ingress with same host and path it should error", func(t *testing.T) {
			ing := &networking.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress",
					Namespace: "test-namespace",
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
				Spec: networking.IngressSpec{
					Rules: []networking.IngressRule{
						{
							Host: "example.com",
							IngressRuleValue: networking.IngressRuleValue{
								HTTP: &networking.HTTPIngressRuleValue{
									Paths: []networking.HTTPIngressPath{
										{
											Path:     "/",
											PathType: &pathTypePrefix,
											Backend: networking.IngressBackend{
												Service: &networking.IngressServiceBackend{
													Name: "http-svc",
													Port: networking.ServiceBackendPort{
														Number: 80,
														Name:   "http",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			nginx.store = &fakeIngressStore{
				ingresses: []*ingress.Ingress{
					{
						Ingress:           *ing,
						ParsedAnnotations: &annotations.Ingress{},
					},
				},
			}
			duplicatedIngress := ing.DeepCopy()
			duplicatedIngress.ObjectMeta.Name = "duplicated-ingress"

			nginx.cfg.DisablePathOverlapValidation = false
			nginx.command = testNginxTestCommand{
				t:        t,
				err:      nil,
				expected: "_,example.com",
			}

			err = nginx.CheckIngress(duplicatedIngress)
			if err == nil {
				t.Errorf("expected errors but noone occurred")
			}
			t.Run("if disablePathOverlap is enabled should not throw any error", func(t *testing.T) {
				duplicatedIngress := ing.DeepCopy()
				duplicatedIngress.ObjectMeta.Name = "duplicated-ingress"

				nginx.cfg.DisablePathOverlapValidation = true
				nginx.command = testNginxTestCommand{
					t:        t,
					err:      nil,
					expected: "_,example.com",
				}

				err = nginx.CheckIngress(duplicatedIngress)
				if err != nil {
					t.Errorf("expected no errors but one %+v occurred", err)
				}
			})
		})
	})

	t.Run("When the ingress is marked as deleted", func(t *testing.T) {
		ing.DeletionTimestamp = &metav1.Time{
			Time: time.Now(),
		}

		if nginx.CheckIngress(ing) != nil {
			t.Errorf("when the ingress is marked as deleted, no error should be returned")
		}
	})
}

func TestCheckWarning(t *testing.T) {
	// Ensure no panic with wrong arguments
	nginx := &NGINXController{}

	nginx.t = fakeTemplate{}
	nginx.store = &fakeIngressStore{
		ingresses: []*ingress.Ingress{},
	}

	ing := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-ingress-warning",
			Namespace:   "user-namespace",
			Annotations: map[string]string{},
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: "example.com",
				},
			},
		},
	}
	t.Run("when a deprecated annotation is used a warning should be returned", func(t *testing.T) {
		ing.ObjectMeta.Annotations[parser.GetAnnotationWithPrefix("enable-influxdb")] = TRUE
		defer func() {
			ing.ObjectMeta.Annotations = map[string]string{}
		}()

		warnings, err := nginx.CheckWarning(ing)
		if err != nil {
			t.Errorf("no error should be returned, but %s was returned", err)
		}
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning to occur but %d occurred", len(warnings))
		} else {
			t.Logf("got warning %s correctly", warnings[0])
		}
	})

	t.Run("When an invalid pathType is used, a warning should be returned", func(t *testing.T) {
		rules := ing.Spec.DeepCopy().Rules
		ing.Spec.Rules = []networking.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								Path:     "/xpto{$2}",
								PathType: &pathTypePrefix,
							},
							{
								Path:     "/ok",
								PathType: &pathTypeExact,
							},
						},
					},
				},
			},
		}
		defer func() {
			ing.Spec.Rules = rules
		}()

		warnings, err := nginx.CheckWarning(ing)
		if err != nil {
			t.Errorf("no error should be returned, but %s was returned", err)
		}
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning to occur but %d occurred", len(warnings))
		} else {
			t.Logf("got warnings %v correctly", warnings)
		}

		t.Run("adding invalid annotations increases the warning count", func(t *testing.T) {
			ing.ObjectMeta.Annotations[parser.GetAnnotationWithPrefix("enable-influxdb")] = TRUE
			ing.ObjectMeta.Annotations[parser.GetAnnotationWithPrefix("secure-verify-ca-secret")] = TRUE
			ing.ObjectMeta.Annotations[parser.GetAnnotationWithPrefix("influxdb-host")] = "blabla"
			defer func() {
				ing.ObjectMeta.Annotations = map[string]string{}
			}()
			warnings, err := nginx.CheckWarning(ing)
			if err != nil {
				t.Errorf("no error should be returned, but %s was returned", err)
			}
			if len(warnings) != 4 {
				t.Errorf("expected 4 warning to occur but %d occurred", len(warnings))
			} else {
				t.Logf("got warnings %v correctly", warnings)
			}
		})
	})

	t.Run("When the ingress is marked as deleted", func(t *testing.T) {
		ing.DeletionTimestamp = &metav1.Time{
			Time: time.Now(),
		}

		if warnings, err := nginx.CheckWarning(ing); err != nil || len(warnings) != 0 {
			t.Errorf("when the ingress is marked as deleted, no warning should be returned")
		}
	})
}

//nolint:dupl // Ignore dupl errors for similar test case
func TestMergeAlternativeBackends(t *testing.T) {
	testCases := map[string]struct {
		ingress      *ingress.Ingress
		upstreams    map[string]*ingress.Backend
		servers      map[string]*ingress.Server
		expUpstreams map[string]*ingress.Backend
		expServers   map[string]*ingress.Server
	}{
		"alternative backend has no server and embeds into matching real backend": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend merges with the correct real backend when multiple are present": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "foo-http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-foo-http-svc-80": {
					Name:     "example-foo-http-svc-80",
					NoServer: false,
				},
				"example-foo-http-svc-canary-80": {
					Name:     "example-foo-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"foo.bar": {
					Hostname: "foo.bar",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-foo-http-svc-80",
						},
					},
				},
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-foo-http-svc-80": {
					Name:                "example-foo-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-foo-http-svc-canary-80"},
				},
				"example-foo-http-svc-canary-80": {
					Name:     "example-foo-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend does not merge into itself": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{},
			map[string]*ingress.Backend{},
			map[string]*ingress.Server{},
		},
		"catch-all alternative backend has no server and embeds into matching real backend": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						DefaultBackend: &networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "http-svc-canary",
								Port: networking.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"catch-all alternative backend does not merge into itself": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						DefaultBackend: &networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "http-svc-canary",
								Port: networking.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"upstream-default-backend": {
					Name:     "upstream-default-backend",
					NoServer: false,
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "upstream-default-backend",
						},
					},
				},
			},
			map[string]*ingress.Backend{},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "upstream-default-backend",
						},
					},
				},
			},
		},
		"non-host canary ingress use default server name as host to merge": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"_": {
					Hostname: "_",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend gets SessionAffinitySettings configured when CanaryBehavior is 'sticky'": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ParsedAnnotations: &annotations.Ingress{
					SessionAffinity: sessionaffinity.Config{
						CanaryBehavior: "sticky",
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend gets SessionAffinitySettings configured when CanaryBehavior is not 'legacy'": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ParsedAnnotations: &annotations.Ingress{
					SessionAffinity: sessionaffinity.Config{
						CanaryBehavior: "", // In fact any value but 'legacy' would do the trick.
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend doesn't get SessionAffinitySettings configured when CanaryBehavior is 'legacy'": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ParsedAnnotations: &annotations.Ingress{
					SessionAffinity: sessionaffinity.Config{
						CanaryBehavior: "legacy",
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:     "example-http-svc-80",
					NoServer: false,
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-80": {
					Name:                "example-http-svc-80",
					NoServer:            false,
					AlternativeBackends: []string{"example-http-svc-canary-80"},
					SessionAffinity: ingress.SessionAffinityConfig{
						AffinityType: "cookie",
						AffinityMode: "balanced",
						CookieSessionAffinity: ingress.CookieSessionAffinity{
							Name: "test",
						},
					},
				},
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
		"alternative backend does not merge for missing upstream": {
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "example",
					},
					Spec: networking.IngressSpec{
						Rules: []networking.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networking.IngressRuleValue{
									HTTP: &networking.HTTPIngressRuleValue{
										Paths: []networking.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: networking.IngressBackend{
													Service: &networking.IngressServiceBackend{
														Name: "http-svc-canary",
														Port: networking.ServiceBackendPort{
															Number: 80,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]*ingress.Backend{
				"example-http-svc-canary-80": {
					Name:     "example-http-svc-canary-80",
					NoServer: true,
					TrafficShapingPolicy: ingress.TrafficShapingPolicy{
						Weight: 20,
					},
				},
			},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
			map[string]*ingress.Backend{},
			map[string]*ingress.Server{
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:     "/",
							PathType: &pathTypePrefix,
							Backend:  "example-http-svc-80",
						},
					},
				},
			},
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			mergeAlternativeBackends(tc.ingress, tc.upstreams, tc.servers)

			for upsName, expUpstream := range tc.expUpstreams {
				actualUpstream, ok := tc.upstreams[upsName]
				if !ok {
					t.Errorf("expected upstream %s to exist but it did not", upsName)
				}

				if !actualUpstream.Equal(expUpstream) {
					t.Logf("actual upstream %s alternative backends: %s", actualUpstream.Name, actualUpstream.AlternativeBackends)
					t.Logf("expected upstream %s alternative backends: %s", expUpstream.Name, expUpstream.AlternativeBackends)
					t.Errorf("upstream %s was not equal to what was expected", actualUpstream.Name)
				}
			}

			for serverName, expServer := range tc.expServers {
				actualServer, ok := tc.servers[serverName]
				if !ok {
					t.Errorf("expected server %s to exist but it did not", serverName)
				}

				if !actualServer.Equal(expServer) {
					t.Errorf("server %s was not equal to what was expected", serverName)
				}
			}
		})
	}
}

func TestExtractTLSSecretName(t *testing.T) {
	testCases := map[string]struct {
		host    string
		ingress *ingress.Ingress
		fn      func(string) (*ingress.SSLCert, error)
		expName string
	}{
		"nil ingress": {
			"foo.bar",
			nil,
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		"empty ingress": {
			"foo.bar",
			&ingress.Ingress{},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		"ingress tls, nil secret": {
			"foo.bar",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{SecretName: "demo"},
						},
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"",
		},
		"ingress tls, no host, matching cert cn": {
			"foo.bar",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{SecretName: "demo"},
						},
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return &ingress.SSLCert{
					Certificate: fakeX509Cert([]string{"foo.bar", "example.com"}),
				}, nil
			},
			"demo",
		},
		"ingress tls, no host, wildcard cert with matching cn": {
			"foo.bar",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{
								SecretName: "demo",
							},
						},
						Rules: []networking.IngressRule{
							{
								Host: "test.foo.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return &ingress.SSLCert{
					Certificate: fakeX509Cert([]string{"*.foo.bar", "foo.bar"}),
				}, nil
			},
			"demo",
		},
		"ingress tls, hosts, matching cert cn": {
			"foo.bar",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{
								Hosts:      []string{"foo.bar", "example.com"},
								SecretName: "demo",
							},
						},
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"demo",
		},
		"ingress tls, hosts, matching cert cn, uppercase host": {
			"FOO.BAR",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{
								Hosts:      []string{"foo.bar", "example.com"},
								SecretName: "demo",
							},
						},
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				return nil, nil
			},
			"demo",
		},
		"ingress tls, hosts, bad format cert, host not in tls Hosts": {
			"foo1.bar",
			&ingress.Ingress{
				Ingress: networking.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: networking.IngressSpec{
						TLS: []networking.IngressTLS{
							{
								Hosts:      []string{"foo.bar"},
								SecretName: "demo",
							},
						},
						Rules: []networking.IngressRule{
							{
								Host: "foo.bar",
							},
							{
								Host: "foo1.bar",
							},
						},
					},
				},
			},
			func(string) (*ingress.SSLCert, error) {
				secretData := map[string]string{
					"ca.crt":    "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EVXhOVEEzTXpJMU5sb1hEVE13TURVeE16QTNNekkxTmxvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTUpGClVDcXBxY09mb2pPeUduTGRsRDlZSG5VVVFiTEQzZkxyN0FzM0RNTk9FNVkyTDJ2TUFhRXQxYkRvYWNqUDFMRlkKcy9ieFBVRmFmVEZ5cmY0SU1iN1FHUlZMYW84aVhmU3p4TlcwanE2dTc0OGZHN3E3c3QvNWR2c28yekpxcHNrcQo2QzNIS3liajMxMTBPNlh2N1I2VDlqdkxzT0M2Vm5VK3BtVHo4RzZ0YVdUeTFBdktQdU5Cc1puUWVRWis3Nk5hClRqSVpsaGlIMnZTUStSMFFzenBoU2tDQmVYbmdkaFloSDYyRVJvZDVZaDNiV1E3T2U0UjFKbHpNUjN3VzFSVGcKZE83aXoxSXdJT3drdlN6T1RlaElqU0pISHVWV3pJOXVqOTBkRTJUQmladzNheUVxdnhrbUxFT3Y1SFRONzRvbwpJQ216WEZzK1kxOFRCb2ZNRXNNQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFKajArK3lCVXBIekYxdWMxSmcyczhGRGNRRHIKdFBuanFROWE0QkwwN2JZdWZyeUJqdUlpaDJReWVKcFB3cTh0RklFa3ZuWjI4VDY4bDFLajRmMnRIU0Y0MG14WAoxVGlkNWxMc1ZjZytja1B2YWVRL1MrUmxnZDNCT1hjY3FFUWR0dithanhqOTdCZXZSelQ1SWd6UFVna3VtSU5wCkxrNS9kSWdxYTIrbmorVUpxdm9TWEhtZG4rMEdvNFJRMXMyZlBJUDhhRFIyL1paQThSTE1rSTJ2R1FVVUJ3RHMKeVkzVy9oWmRWeUhpWEcvRkJKRHNZU1cyZjFrZ1AzRzlyNjdnZG1WT05JckNCdHBSWkU3NllTRjFqNUtocFlUNgp3UDFpSVNDOUc0ZmRCeGxkaXJqRWM3VU9PQTlNQ0JzYXZ4R1IreCtBTEVnTnhOUlVZdnEvZWl0OGtRVT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
					"namespace": "ZGVtbw==",
					"token":     "ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklqaExObWcyVnpWM01ERm9Ua1ZpVFVwTlYwbDRPV3RMVkhaTE1XRnpOa010VjI5WE55MTRaRzR6VFVVaWZRLmV5SnBjM01pT2lKcmRXSmxjbTVsZEdWekwzTmxjblpwWTJWaFkyTnZkVzUwSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXVZVzFsYzNCaFkyVWlPaUprWlcxdklpd2lhM1ZpWlhKdVpYUmxjeTVwYnk5elpYSjJhV05sWVdOamIzVnVkQzl6WldOeVpYUXVibUZ0WlNJNkltUmxabUYxYkhRdGRHOXJaVzR0Wkc0MmVHSWlMQ0pyZFdKbGNtNWxkR1Z6TG1sdkwzTmxjblpwWTJWaFkyTnZkVzUwTDNObGNuWnBZMlV0WVdOamIzVnVkQzV1WVcxbElqb2laR1ZtWVhWc2RDSXNJbXQxWW1WeWJtVjBaWE11YVc4dmMyVnlkbWxqWldGalkyOTFiblF2YzJWeWRtbGpaUzFoWTJOdmRXNTBMblZwWkNJNkltVm1OR0kxWW1NMExUTXdPREV0TkRFNU15MWlZakl6TFRoaE5qRmhNV0ptTWpRNFlTSXNJbk4xWWlJNkluTjVjM1JsYlRwelpYSjJhV05sWVdOamIzVnVkRHBrWlcxdk9tUmxabUYxYkhRaWZRLnEzaGFxVVFDN2Z6a1V3UldKazM0RjRsamktbWs5cWdPcDJHSFlSZ1JrWUk0WW8xclhoSURCSnUzWkFPdjhMN3doZkgzcmo4ZjFnNFpMSFBkd3JKT2lZdWlvXzVXdDZPSXZtbXFaU2VncnRmV1MwUFZXYzJ1d0xweDJpSElTbUlHd21uQ1hYQzNRX05RNFRlQnZxWEMyUHR4REFwM19QM3QyZnRKN0w2Z1kzTkcyZUsyQTVFZG82azQtR2wzN0Zaam51NmRzc0FocVZaeld0NE9ZS3hTWWtpN003dnh5ZWtJQ091UmJ6SW5DNmhldEhtbHhyaF9ObWplMHhfY2M4V3ZkUnJYbFlpRWxnYXZCY1FtMTJ2YkxBQWlzWkFrT2Y1T3VvaEhLUmpEOGlMS1pRMXdKRHNnRmYzd1BFWGxTWkg2QkVZdS1TU0laSDNKYWVWU3llWjExdw==",
				}
				ca, err := base64.StdEncoding.DecodeString(secretData["ca.crt"])
				if err != nil {
					t.Fatalf("unexpected error decoding ca.crt: %v", err)
				}
				cert, err := ssl.CreateCACert(ca)
				if err != nil {
					t.Fatalf("unexpected error creating SSL Cert: %v", err)
				}
				err = ssl.ConfigureCACert("demo", ca, cert)
				if err != nil {
					t.Fatalf("error configuring CA certificate: %v", err)
				}
				cert.Name = "default-token-dn6xb"
				cert.Namespace = "demo"
				return cert, nil
			},
			"",
		},
	}

	for title, tc := range testCases {
		t.Run(title, func(t *testing.T) {
			name := extractTLSSecretName(tc.host, tc.ingress, tc.fn)
			if name != tc.expName {
				t.Errorf("Expected Secret name %q (got %q)", tc.expName, name)
			}
		})
	}
}

//nolint:gocyclo // Ignore function complexity error
func TestGetBackendServers(t *testing.T) {
	testCases := []struct {
		Ingresses    []*ingress.Ingress
		Validate     func(ingresses []*ingress.Ingress, upstreams []*ingress.Backend, servers []*ingress.Server)
		SetConfigMap func(namespace string) *corev1.ConfigMap
	}{
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							DefaultBackend: &networking.IngressBackend{
								Service: &networking.IngressServiceBackend{
									Name: "http-svc-canary",
									Port: networking.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 1 {
					t.Errorf("servers count should be 1, got %d", len(servers))
					return
				}

				s := servers[0]
				if s.Hostname != "_" {
					t.Errorf("server hostname should be '_', got '%s'", s.Hostname)
				}
				if !s.Locations[0].IsDefBackend {
					t.Errorf("server location 0 should be default backend")
				}

				if s.Locations[0].Backend != defUpstreamName {
					t.Errorf("location backend should be '%s', got '%s'", defUpstreamName, s.Locations[0].Backend)
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							DefaultBackend: &networking.IngressBackend{
								Service: &networking.IngressServiceBackend{
									Name: "http-svc-canary",
									Port: networking.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							DefaultBackend: &networking.IngressBackend{
								Service: &networking.IngressServiceBackend{
									Name: "http-svc",
									Port: networking.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: false,
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 1 {
					t.Errorf("servers count should be 1, got %d", len(servers))
					return
				}

				s := servers[0]
				if s.Hostname != "_" {
					t.Errorf("server hostname should be '_', got '%s'", s.Hostname)
				}
				if s.Locations[0].IsDefBackend {
					t.Errorf("server location 0 should not be default backend")
				}

				if s.Locations[0].Backend != "example-http-svc-80" {
					t.Errorf("location backend should be 'example-http-svc-80', got '%s'", s.Locations[0].Backend)
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-canary",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 1 {
					t.Errorf("servers count should be 1, got %d", len(servers))
					return
				}

				s := servers[0]
				if s.Hostname != "_" {
					t.Errorf("server hostname should be '_', got '%s'", s.Hostname)
				}
				if !s.Locations[0].IsDefBackend {
					t.Errorf("server location 0 should be default backend")
				}

				if s.Locations[0].Backend != defUpstreamName {
					t.Errorf("location backend should be '%s', got '%s'", defUpstreamName, s.Locations[0].Backend)
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: false,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-canary",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-canary",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 2, got %d", len(servers))
					return
				}

				s := servers[0]
				if s.Hostname != "_" {
					t.Errorf("server hostname should be '_', got '%s'", s.Hostname)
				}
				if !s.Locations[0].IsDefBackend {
					t.Errorf("server location 0 should be default backend")
				}

				if s.Locations[0].Backend != defUpstreamName {
					t.Errorf("location backend should be '%s', got '%s'", defUpstreamName, s.Locations[0].Backend)
				}

				s = servers[1]
				if s.Hostname != "example.com" {
					t.Errorf("server hostname should be 'example.com', got '%s'", s.Hostname)
				}

				if s.Locations[0].Backend != "example-http-svc-80" {
					t.Errorf("location backend should be 'example-http-svc-80', got '%s'", s.Locations[0].Backend)
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-a",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/a",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-1",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: false,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-a-canary",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/a",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-2",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-b",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/b",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-2",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: false,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-b-canary",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/b",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-1",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-c",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/c",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-1",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: false,
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-c-canary",
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/c",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "http-svc-2",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Canary: canary.Config{
							Enabled: true,
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, upstreams []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 2, got %d", len(servers))
					return
				}

				s := servers[0]
				if s.Hostname != "_" {
					t.Errorf("server hostname should be '_', got '%s'", s.Hostname)
				}
				if !s.Locations[0].IsDefBackend {
					t.Errorf("server location 0 should be default backend")
				}

				if s.Locations[0].Backend != defUpstreamName {
					t.Errorf("location backend should be '%s', got '%s'", defUpstreamName, s.Locations[0].Backend)
				}

				s = servers[1]
				if s.Hostname != "example.com" {
					t.Errorf("server hostname should be 'example.com', got '%s'", s.Hostname)
				}

				if s.Locations[0].Backend != exampleBackend || s.Locations[1].Backend != exampleBackend || s.Locations[2].Backend != exampleBackend {
					t.Errorf("all location backend should be 'example-http-svc-1-80'")
				}

				if len(upstreams) != 3 {
					t.Errorf("upstreams count should be 3, got %d", len(upstreams))
					return
				}

				if upstreams[0].Name != exampleBackend {
					t.Errorf("example-http-svc-1-80 should be first upstream, got %s", upstreams[0].Name)
					return
				}
				if upstreams[0].NoServer {
					t.Errorf("'example-http-svc-1-80' should be primary upstream, got as alternative upstream")
				}
				if len(upstreams[0].AlternativeBackends) != 1 || upstreams[0].AlternativeBackends[0] != "example-http-svc-2-80" {
					t.Errorf("example-http-svc-2-80 should be alternative upstream for 'example-http-svc-1-80'")
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			//nolint:dupl // Ignore dupl errors for similar test case
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "proxy-ssl-1",
							Namespace: "proxyssl",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path1",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "path1-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						ProxySSL: proxyssl.Config{
							AuthSSLCert: resolver.AuthSSLCert{
								CAFileName: "cafile1.crt",
								Secret:     "secret1",
							},
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "proxy-ssl-2",
							Namespace: "proxyssl",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path2",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "path2-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						ProxySSL: proxyssl.Config{
							AuthSSLCert: resolver.AuthSSLCert{
								CAFileName: "cafile1.crt",
								Secret:     "secret1",
							},
						},
					},
				},
			},
			Validate: func(ingresses []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 2, got %d", len(servers))
					return
				}

				s := servers[1]

				if s.ProxySSL.CAFileName != ingresses[0].ParsedAnnotations.ProxySSL.CAFileName {
					t.Errorf("server cafilename should be '%s', got '%s'", ingresses[0].ParsedAnnotations.ProxySSL.CAFileName, s.ProxySSL.CAFileName)
				}

				if s.Locations[0].ProxySSL.CAFileName != ingresses[0].ParsedAnnotations.ProxySSL.CAFileName {
					t.Errorf("location cafilename should be '%s', got '%s'", ingresses[0].ParsedAnnotations.ProxySSL.CAFileName, s.Locations[0].ProxySSL.CAFileName)
				}

				if s.Locations[1].ProxySSL.CAFileName != ingresses[1].ParsedAnnotations.ProxySSL.CAFileName {
					t.Errorf("location cafilename should be '%s', got '%s'", ingresses[1].ParsedAnnotations.ProxySSL.CAFileName, s.Locations[0].ProxySSL.CAFileName)
				}
			},
			SetConfigMap: testConfigMap,
		},
		{
			//nolint:dupl // Ignore dupl errors for similar test case
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "proxy-ssl-1",
							Namespace: "proxyssl",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path1",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "path1-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						ProxySSL: proxyssl.Config{
							AuthSSLCert: resolver.AuthSSLCert{
								CAFileName: "cafile1.crt",
								Secret:     "secret1",
							},
						},
					},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "proxy-ssl-2",
							Namespace: "proxyssl",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path2",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "path2-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						ProxySSL: proxyssl.Config{
							AuthSSLCert: resolver.AuthSSLCert{
								CAFileName: "cafile1.crt",
								Secret:     "secret1",
							},
						},
					},
				},
			},
			Validate: func(ingresses []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 2, got %d", len(servers))
					return
				}

				s := servers[1]

				if s.ProxySSL.CAFileName != "" {
					t.Errorf("server cafilename should be empty, got '%s'", s.ProxySSL.CAFileName)
				}

				if s.Locations[0].ProxySSL.CAFileName != ingresses[0].ParsedAnnotations.ProxySSL.CAFileName {
					t.Errorf("location cafilename should be '%s', got '%s'", ingresses[0].ParsedAnnotations.ProxySSL.CAFileName, s.Locations[0].ProxySSL.CAFileName)
				}

				if s.Locations[1].ProxySSL.CAFileName != ingresses[1].ParsedAnnotations.ProxySSL.CAFileName {
					t.Errorf("location cafilename should be '%s', got '%s'", ingresses[1].ParsedAnnotations.ProxySSL.CAFileName, s.Locations[0].ProxySSL.CAFileName)
				}
			},
			SetConfigMap: func(ns string) *corev1.ConfigMap {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:     "config",
						SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
					},
					Data: map[string]string{
						"proxy-ssl-location-only": TRUE,
					},
				}
			},
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "proxy-ssl-1",
							Namespace: "proxyssl",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path1",
													PathType: &pathTypePrefix,
													Backend:  networking.IngressBackend{},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						ProxySSL: proxyssl.Config{
							AuthSSLCert: resolver.AuthSSLCert{
								CAFileName: "cafile1.crt",
								Secret:     "secret1",
							},
						},
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 1, got %d", len(servers))
					return
				}

				s := servers[1]

				if s.Locations[0].Backend != "upstream-default-backend" {
					t.Errorf("backend should be upstream-default-backend, got '%s'", s.Locations[0].Backend)
				}
			},
			SetConfigMap: func(ns string) *corev1.ConfigMap {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:     "config",
						SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
					},
					Data: map[string]string{
						"proxy-ssl-location-only": TRUE,
					},
				}
			},
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "not-allowed-snippet",
							Namespace: "default",
							Annotations: map[string]string{
								"nginx.ingress.kubernetes.io/server-snippet":         "bla",
								"nginx.ingress.kubernetes.io/configuration-snippet":  "blo",
								"nginx.ingress.kubernetes.io/whitelist-source-range": "10.0.0.0/24",
							},
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "example.com",
									IngressRuleValue: networking.IngressRuleValue{
										HTTP: &networking.HTTPIngressRuleValue{
											Paths: []networking.HTTPIngressPath{
												{
													Path:     "/path1",
													PathType: &pathTypePrefix,
													Backend: networking.IngressBackend{
														Service: &networking.IngressServiceBackend{
															Name: "path1-svc",
															Port: networking.ServiceBackendPort{
																Number: 80,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{
						Allowlist:            ipallowlist.SourceRange{CIDR: []string{"10.0.0.0/24"}},
						ServerSnippet:        "bla",
						ConfigurationSnippet: "blo",
					},
				},
			},
			Validate: func(_ []*ingress.Ingress, _ []*ingress.Backend, servers []*ingress.Server) {
				if len(servers) != 2 {
					t.Errorf("servers count should be 2, got %d", len(servers))
					return
				}
				s := servers[1]

				if s.ServerSnippet != "" {
					t.Errorf("server snippet should be empty, got '%s'", s.ServerSnippet)
				}

				if s.Locations[0].ConfigurationSnippet != "" {
					t.Errorf("config snippet should be empty, got '%s'", s.Locations[0].ConfigurationSnippet)
				}

				if len(s.Locations[0].Allowlist.CIDR) != 1 || s.Locations[0].Allowlist.CIDR[0] != "10.0.0.0/24" {
					t.Errorf("allow list was incorrectly dropped, len should be 1 and contain 10.0.0.0/24")
				}
			},
			SetConfigMap: func(ns string) *corev1.ConfigMap {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:     "config",
						SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
					},
					Data: map[string]string{
						"allow-snippet-annotations": "false",
					},
				}
			},
		},
	}

	for _, testCase := range testCases {
		nginxController := newDynamicNginxController(t, testCase.SetConfigMap)
		upstreams, servers := nginxController.getBackendServers(testCase.Ingresses)
		testCase.Validate(testCase.Ingresses, upstreams, servers)
	}
}

func testConfigMap(ns string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "config",
			SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
		},
	}
}

func newNGINXController(t *testing.T) *NGINXController {
	ns := corev1.NamespaceDefault

	clientSet := fake.NewSimpleClientset()

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "config",
			SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
		},
	}

	_, err := clientSet.CoreV1().ConfigMaps(ns).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating the configuration map: %v", err)
	}

	k8s.IngressPodDetails = &k8s.PodInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpod",
			Namespace: ns,
			Labels: map[string]string{
				"pod-template-hash": "1234",
			},
		},
	}

	storer := store.New(
		ns,
		labels.Nothing(),
		fmt.Sprintf("%v/config", ns),
		fmt.Sprintf("%v/tcp", ns),
		fmt.Sprintf("%v/udp", ns),
		"",
		10*time.Minute,
		clientSet,
		channels.NewRingChannel(10),
		false,
		true,
		&ingressclass.Configuration{
			Controller:      "k8s.io/ingress-nginx",
			AnnotationValue: "nginx",
		},
		false,
	)

	sslCert := ssl.GetFakeSSLCert()
	config := &Configuration{
		FakeCertificate: sslCert,
		ListenPorts: &ngx_config.ListenPorts{
			Default: 80,
		},
	}

	return &NGINXController{
		store:   storer,
		cfg:     config,
		command: NewNginxCommand(),
	}
}

var oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}

func fakeX509Cert(dnsNames []string) *x509.Certificate {
	return &x509.Certificate{
		DNSNames: dnsNames,
		Extensions: []pkix.Extension{
			{Id: oidExtensionSubjectAltName},
		},
	}
}

func newDynamicNginxController(t *testing.T, setConfigMap func(string) *corev1.ConfigMap) *NGINXController {
	ns := corev1.NamespaceDefault

	clientSet := fake.NewSimpleClientset()
	configMap := setConfigMap(ns)

	_, err := clientSet.CoreV1().ConfigMaps(ns).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating the configuration map: %v", err)
	}

	k8s.IngressPodDetails = &k8s.PodInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpod",
			Namespace: ns,
			Labels: map[string]string{
				"pod-template-hash": "1234",
			},
		},
	}

	storer := store.New(
		ns,
		labels.Nothing(),
		fmt.Sprintf("%v/config", ns),
		fmt.Sprintf("%v/tcp", ns),
		fmt.Sprintf("%v/udp", ns),
		"",
		10*time.Minute,
		clientSet,
		channels.NewRingChannel(10),
		false,
		true,
		&ingressclass.Configuration{
			Controller:      "k8s.io/ingress-nginx",
			AnnotationValue: "nginx",
		},
		false)

	sslCert := ssl.GetFakeSSLCert()
	config := &Configuration{
		FakeCertificate: sslCert,
		ListenPorts: &ngx_config.ListenPorts{
			Default: 80,
		},
	}

	return &NGINXController{
		store:           storer,
		cfg:             config,
		command:         NewNginxCommand(),
		metricCollector: metric.DummyCollector{},
	}
}
