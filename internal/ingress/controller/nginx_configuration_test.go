/*
Copyright 2020 The Kubernetes Authors.

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

	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
)

func TestExtractServers(t *testing.T) {
	tests := []struct {
		name      string
		ingresses []*ingress.Ingress
		expected  []string
	}{
		{
			"no ingresses should only return catch all empty server",
			nil,
			[]string{"_"},
		},
		{
			"ingress with no host",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]string{"_"},
		},
		{
			"multiple ingresses with multiple hosts without duplication",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-1",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "foo.bar",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]string{"_", "example.com", "foo.bar"},
		},
		{
			"single ingress with multiple hosts without duplication",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "foo.bar",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]string{"_", "example.com", "foo.bar"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			servers := extractServers(testCase.ingresses)
			if len(testCase.expected) != len(servers) {
				t.Errorf("Expected %d Servers but got %d", len(testCase.expected), len(servers))
			}
		})
	}
}

func TestServerLocations(t *testing.T) {
	tests := []struct {
		name      string
		hostname  string
		ingresses []*ingress.Ingress
		expected  []*ingress.Location
	}{
		{
			"no ingresses should only return catch all empty server",
			"",
			[]*ingress.Ingress{},
			[]*ingress.Location{},
		},
		{
			"ingress with no matching host",
			"_",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathPrefix,
				},
			},
		},
		{
			"multiple ingresses with multiple hosts without duplication",
			"example.com",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-1",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "foo.bar",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathPrefix,
				},
			},
		},
		{
			"single ingress with multiple hosts without duplication",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "foo.bar",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathPrefix,
				},
			},
		},
		{
			"single ingress with multiple hosts same path and different types",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
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
													PathType: &pathPrefix,
													Backend:  sampleBackend(),
												},
												{
													Path:     "/",
													PathType: &pathExact,
													Backend:  sampleBackend(),
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathPrefix,
				},
				{
					Path:     "/",
					PathType: &pathExact,
				},
			},
		},
		{
			"single ingress with multiple hosts same duplicated path with same type",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
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
													PathType: &pathExact,
													Backend:  sampleBackend(),
												},
												{
													Path:     "/",
													PathType: &pathExact,
													Backend:  sampleBackend(),
												},
											},
										},
									},
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathExact,
				},
			},
		},
		{
			"multiple ingresses with multiple hosts with same host and path different types",
			"example.com",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathPrefix),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example-1",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host:             "example.com",
									IngressRuleValue: ingressRule("/", &pathExact),
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "/",
					PathType: &pathPrefix,
				},
				{
					Path:     "/",
					PathType: &pathExact,
				},
			},
		},
		{
			"single ingress with host and no rules",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Rules: []networking.IngressRule{
								{
									Host: "foo.bar",
								},
							},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{},
		},
		{
			"single ingress without host and no rules",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{},
		},
		{
			"single ingress without rules only backend",
			"foo.bar",
			[]*ingress.Ingress{
				{
					Ingress: networking.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "example",
						},
						Spec: networking.IngressSpec{
							Backend: defaultBackend(),
							Rules:   []networking.IngressRule{},
						},
					},
					ParsedAnnotations: &annotations.Ingress{},
				},
			},
			[]*ingress.Location{
				{
					Path:     "_defaultBackend",
					PathType: &pathPrefix,
				},
			},
		}}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			locations := serverLocations(testCase.hostname, testCase.ingresses)
			if len(testCase.expected) != len(locations) {
				t.Errorf("Expected %d Locations but got %d", len(testCase.expected), len(locations))
			}
		})
	}
}

func sampleBackend() networking.IngressBackend {
	return networking.IngressBackend{
		ServiceName: "http-svc",
		ServicePort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 80,
		},
	}
}

func defaultBackend() *networking.IngressBackend {
	db := sampleBackend()
	return &db
}

func ingressRule(path string, pathType *networking.PathType) networking.IngressRuleValue {
	return networking.IngressRuleValue{
		HTTP: &networking.HTTPIngressRuleValue{
			Paths: []networking.HTTPIngressPath{
				{
					Path:     path,
					PathType: pathType,
					Backend:  sampleBackend(),
				},
			},
		},
	}
}

func TestSamePathType(t *testing.T) {
	tests := []struct {
		name     string
		a        *networking.PathType
		b        *networking.PathType
		expected bool
	}{
		{
			"nil PathType are equal",
			nil,
			nil,
			true,
		},
		{
			"Exact and nil are not equal",
			&pathExact,
			nil,
			false,
		},
		{
			"nil and Prefix are not equal",
			nil,
			&pathPrefix,
			false,
		},
		{
			"Exact and Prefix are not equal",
			&pathExact,
			&pathPrefix,
			false,
		},
		{
			"Prefix and Prefix are equal",
			&pathPrefix,
			&pathPrefix,
			true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			equal := samePathType(testCase.a, testCase.b)
			if testCase.expected != equal {
				t.Errorf("expected %v but got %v", testCase.expected, equal)
			}
		})
	}
}
