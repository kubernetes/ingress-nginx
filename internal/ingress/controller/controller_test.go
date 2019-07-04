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
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eapache/channels"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/canary"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/store"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/metric"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/net/ssl"
)

const fakeCertificateName = "default-fake-certificate"

type fakeIngressStore struct {
	ingresses []*ingress.Ingress
}

func (fakeIngressStore) GetBackendConfiguration() ngx_config.Configuration {
	return ngx_config.Configuration{}
}

func (fakeIngressStore) GetConfigMap(key string) (*corev1.ConfigMap, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetSecret(key string) (*corev1.Secret, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetService(key string) (*corev1.Service, error) {
	return nil, fmt.Errorf("test error")
}

func (fakeIngressStore) GetServiceEndpoints(key string) (*corev1.Endpoints, error) {
	return nil, fmt.Errorf("test error")
}

func (fis fakeIngressStore) ListIngresses(store.IngressFilterFunc) []*ingress.Ingress {
	return fis.ingresses
}

func (fakeIngressStore) GetRunningControllerPodsCount() int {
	return 0
}

func (fakeIngressStore) GetLocalSSLCert(name string) (*ingress.SSLCert, error) {
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

func (fakeIngressStore) Run(stopCh chan struct{}) {}

type testNginxTestCommand struct {
	t        *testing.T
	expected string
	out      []byte
	err      error
}

func (ntc testNginxTestCommand) ExecCommand(args ...string) *exec.Cmd {
	return nil
}

func (ntc testNginxTestCommand) Test(cfg string) ([]byte, error) {
	fd, err := os.Open(cfg)
	if err != nil {
		ntc.t.Errorf("could not read generated nginx configuration: %v", err.Error())
		return nil, err
	}
	defer fd.Close()
	bytes, err := ioutil.ReadAll(fd)
	if err != nil {
		ntc.t.Errorf("could not read generated nginx configuration: %v", err.Error())
	}
	if string(bytes) != ntc.expected {
		ntc.t.Errorf("unexpected generated configuration %v. Expecting %v", string(bytes), ntc.expected)
	}
	return ntc.out, ntc.err
}

type fakeTemplate struct{}

func (fakeTemplate) Write(conf config.TemplateConfig) ([]byte, error) {
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
		filepath.Walk(os.TempDir(), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && os.TempDir() != path {
				return filepath.SkipDir
			}
			if strings.HasPrefix(info.Name(), tempNginxPattern) {
				os.Remove(path)
			}
			return nil
		})
	}()

	// Ensure no panic with wrong arguments
	var nginx *NGINXController
	nginx.CheckIngress(nil)
	nginx = newNGINXController(t)
	nginx.CheckIngress(nil)
	nginx.metricCollector = metric.DummyCollector{}

	nginx.t = fakeTemplate{}
	nginx.store = fakeIngressStore{
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

	t.Run("When the ingress class differs from nginx", func(t *testing.T) {
		ing.ObjectMeta.Annotations["kubernetes.io/ingress.class"] = "different"
		nginx.command = testNginxTestCommand{
			t:   t,
			err: fmt.Errorf("test error"),
		}
		if nginx.CheckIngress(ing) != nil {
			t.Errorf("with a different ingress class, no error should be returned")
		}
	})

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
			nginx.store = fakeIngressStore{
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

		t.Run("When the ingress is in a different namespace than the watched one", func(t *testing.T) {
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
	})
}

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
												Path: "/",
												Backend: networking.IngressBackend{
													ServiceName: "http-svc-canary",
													ServicePort: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 80,
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
							Path:    "/",
							Backend: "example-http-svc-80",
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
							Path:    "/",
							Backend: "example-http-svc-80",
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
												Path: "/",
												Backend: networking.IngressBackend{
													ServiceName: "foo-http-svc-canary",
													ServicePort: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 80,
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
												Path: "/",
												Backend: networking.IngressBackend{
													ServiceName: "http-svc-canary",
													ServicePort: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 80,
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
							Path:    "/",
							Backend: "example-foo-http-svc-80",
						},
					},
				},
				"example.com": {
					Hostname: "example.com",
					Locations: []*ingress.Location{
						{
							Path:    "/",
							Backend: "example-http-svc-80",
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
							Path:    "/",
							Backend: "example-http-svc-80",
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
												Path: "/",
												Backend: networking.IngressBackend{
													ServiceName: "http-svc-canary",
													ServicePort: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 80,
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
						Backend: &networking.IngressBackend{
							ServiceName: "http-svc-canary",
							ServicePort: intstr.IntOrString{
								IntVal: 80,
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
							Path:    "/",
							Backend: "example-http-svc-80",
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
							Path:    "/",
							Backend: "example-http-svc-80",
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
						Backend: &networking.IngressBackend{
							ServiceName: "http-svc-canary",
							ServicePort: intstr.IntOrString{
								IntVal: 80,
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
							Path:    "/",
							Backend: "upstream-default-backend",
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
							Path:    "/",
							Backend: "upstream-default-backend",
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
					t.Errorf("upstream %s was not equal to what was expected: ", upsName)
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

func TestGetBackendServers(t *testing.T) {
	ctl := newNGINXController(t)

	testCases := []struct {
		Ingresses []*ingress.Ingress
		Validate  func(servers []*ingress.Server)
	}{
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Backend: &networking.IngressBackend{
								ServiceName: "http-svc-canary",
								ServicePort: intstr.IntOrString{
									IntVal: 80,
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
			Validate: func(servers []*ingress.Server) {
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
		},
		{
			Ingresses: []*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "example",
						},
						Spec: networking.IngressSpec{
							Backend: &networking.IngressBackend{
								ServiceName: "http-svc-canary",
								ServicePort: intstr.IntOrString{
									IntVal: 80,
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
							Backend: &networking.IngressBackend{
								ServiceName: "http-svc",
								ServicePort: intstr.IntOrString{
									IntVal: 80,
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
			Validate: func(servers []*ingress.Server) {
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
													Path: "/",
													Backend: networking.IngressBackend{
														ServiceName: "http-svc-canary",
														ServicePort: intstr.IntOrString{
															Type:   intstr.Int,
															IntVal: 80,
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
			Validate: func(servers []*ingress.Server) {
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
													Path: "/",
													Backend: networking.IngressBackend{
														ServiceName: "http-svc",
														ServicePort: intstr.IntOrString{
															Type:   intstr.Int,
															IntVal: 80,
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
													Path: "/",
													Backend: networking.IngressBackend{
														ServiceName: "http-svc-canary",
														ServicePort: intstr.IntOrString{
															Type:   intstr.Int,
															IntVal: 80,
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
			Validate: func(servers []*ingress.Server) {
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
		},
	}

	for _, testCase := range testCases {
		_, servers := ctl.getBackendServers(testCase.Ingresses)
		testCase.Validate(servers)
	}
}

func newNGINXController(t *testing.T) *NGINXController {
	ns := v1.NamespaceDefault
	pod := &k8s.PodInfo{
		Name:      "testpod",
		Namespace: ns,
		Labels: map[string]string{
			"pod-template-hash": "1234",
		},
	}

	clientSet := fake.NewSimpleClientset()
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "config",
			SelfLink: fmt.Sprintf("/api/v1/namespaces/%s/configmaps/config", ns),
		},
	}
	_, err := clientSet.CoreV1().ConfigMaps(ns).Create(configMap)
	if err != nil {
		t.Fatalf("error creating the configuration map: %v", err)
	}

	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	storer := store.New(
		ns,
		fmt.Sprintf("%v/config", ns),
		fmt.Sprintf("%v/tcp", ns),
		fmt.Sprintf("%v/udp", ns),
		"",
		10*time.Minute,
		clientSet,
		fs,
		channels.NewRingChannel(10),
		pod,
		false)

	sslCert := ssl.GetFakeSSLCert(fs)
	config := &Configuration{
		FakeCertificate: sslCert,
		ListenPorts: &ngx_config.ListenPorts{
			Default: 80,
		},
	}

	return &NGINXController{
		store:      storer,
		cfg:        config,
		command:    NewNginxCommand(),
		fileSystem: fs,
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
