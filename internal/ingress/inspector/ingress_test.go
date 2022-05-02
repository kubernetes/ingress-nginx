/*
Copyright 2022 The Kubernetes Authors.

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
	"testing"

	networking "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeSimpleIngress(hostname string, paths ...string) *networking.Ingress {

	newIngress := networking.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{},
						},
					},
				},
			},
		},
	}

	prefix := networking.PathTypePrefix
	for _, path := range paths {
		newPath := networking.HTTPIngressPath{
			Path:     path,
			PathType: &prefix,
		}
		newIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(newIngress.Spec.Rules[0].IngressRuleValue.HTTP.Paths, newPath)
	}
	return &newIngress
}

func TestInspectIngress(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		path     []string
		wantErr  bool
	}{
		{
			name:     "invalid-path-etc",
			hostname: "invalid.etc.com",
			path: []string{
				"/var/run/secrets",
				"/mypage",
			},
			wantErr: true,
		},
		{
			name:     "invalid-path-etc",
			hostname: "invalid.etc.com",
			path: []string{
				"/etc/nginx",
			},
			wantErr: true,
		},
		{
			name:     "invalid-path-etc",
			hostname: "invalid.etc.com",
			path: []string{
				"/mypage",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ingress := makeSimpleIngress(tt.hostname, tt.path...)
			if err := InspectIngress(ingress); (err != nil) != tt.wantErr {
				t.Errorf("InspectIngress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
