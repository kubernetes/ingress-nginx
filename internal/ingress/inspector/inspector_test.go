/*
Copyright 2023 The Kubernetes Authors.

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

package inspector

import (
	"errors"
	"fmt"
	"testing"

	networking "k8s.io/api/networking/v1"
)

var (
	exact  = networking.PathTypeExact
	prefix = networking.PathTypePrefix
)

var (
	validIngress = &networking.Ingress{
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path: "/test",
								},
								{
									PathType: &prefix,
									Path:     "/xpto/ab0/x_ss-9",
								},
								{
									PathType: &exact,
									Path:     "/bla/",
								},
							},
						},
					},
				},
			},
		},
	}

	emptyIngress = &networking.Ingress{
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									PathType: &exact,
								},
							},
						},
					},
				},
			},
		},
	}

	invalidIngress = &networking.Ingress{
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									PathType: &exact,
									Path:     "/foo.+",
								},
								{
									PathType: &exact,
									Path:     "xpto/lala",
								},
								{
									PathType: &exact,
									Path:     "/xpto/lala",
								},
								{
									PathType: &prefix,
									Path:     "/foo/bar/[a-z]{3}",
								},
								{
									PathType: &prefix,
									Path:     "/lala/xp\ntest",
								},
							},
						},
					},
				},
			},
		},
	}

	validImplSpecific = &networking.Ingress{
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									PathType: &implSpecific,
									Path:     "/foo.+",
								},
								{
									PathType: &implSpecific,
									Path:     "xpto/lala",
								},
							},
						},
					},
				},
			},
		},
	}
)

var aErr = func(s, pathType string) error {
	return fmt.Errorf("path %s cannot be used with pathType %s", s, pathType)
}

func TestValidatePathType(t *testing.T) {
	tests := []struct {
		name    string
		ing     *networking.Ingress
		wantErr bool
		err     error
	}{
		{
			name:    "nil should return an error",
			ing:     nil,
			wantErr: true,
			err:     fmt.Errorf("received null ingress"),
		},
		{
			name:    "valid should not return an error",
			ing:     validIngress,
			wantErr: false,
		},
		{
			name:    "empty should not return an error",
			ing:     emptyIngress,
			wantErr: false,
		},
		{
			name:    "empty should not return an error",
			ing:     validImplSpecific,
			wantErr: false,
		},
		{
			name:    "invalid should return multiple errors",
			ing:     invalidIngress,
			wantErr: true,
			err: errors.Join(
				aErr("/foo.+", "Exact"),
				aErr("xpto/lala", "Exact"),
				aErr("/foo/bar/[a-z]{3}", "Prefix"),
				aErr("/lala/xp\ntest", "Prefix"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathType(tt.ing)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil && tt.err != nil) && tt.err.Error() != err.Error() {
				t.Errorf("received invalid error: want = %v, expected %v", tt.err, err)
			}
		})
	}
}
