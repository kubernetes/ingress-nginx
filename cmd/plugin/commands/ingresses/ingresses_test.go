/*
Copyright 2021 The Kubernetes Authors.

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

package ingresses

import (
	"testing"

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGetIngressInformation(t *testing.T) {

	testcases := map[string]struct {
		ServiceBackend *networking.IngressServiceBackend
		wantName       string
		wantPort       intstr.IntOrString
	}{
		"empty ingressServiceBackend": {
			ServiceBackend: &networking.IngressServiceBackend{},
			wantName:       "",
			wantPort:       intstr.IntOrString{},
		},
		"ingressServiceBackend with port 8080": {
			ServiceBackend: &networking.IngressServiceBackend{
				Name: "test",
				Port: networking.ServiceBackendPort{
					Number: 8080,
				},
			},
			wantName: "test",
			wantPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 8080,
			},
		},
		"ingressServiceBackend with port name a-svc": {
			ServiceBackend: &networking.IngressServiceBackend{
				Name: "test",
				Port: networking.ServiceBackendPort{
					Name: "a-svc",
				},
			},
			wantName: "test",
			wantPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "a-svc",
			},
		},
	}

	for title, testCase := range testcases {
		gotName, gotPort := serviceToNameAndPort(testCase.ServiceBackend)
		if gotName != testCase.wantName {
			t.Fatalf("%s: expected '%v' but returned %v", title, testCase.wantName, gotName)
		}
		if gotPort != testCase.wantPort {
			t.Fatalf("%s: expected '%v' but returned %v", title, testCase.wantPort, gotPort)
		}
	}
}
