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

package ingress

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

func TestIsDynamicConfigurationEnough(t *testing.T) {
	backends := []*ingress.Backend{{
		Name: "fakenamespace-myapp-80",
		Endpoints: []ingress.Endpoint{
			{
				Address: "10.0.0.1",
				Port:    "8080",
			},
			{
				Address: "10.0.0.2",
				Port:    "8080",
			},
		},
	}}

	servers := []*ingress.Server{{
		Hostname: "myapp.fake",
		Locations: []*ingress.Location{
			{
				Path:    "/",
				Backend: "fakenamespace-myapp-80",
			},
		},
		SSLCert: &ingress.SSLCert{
			PemCertKey: "fake-certificate",
		},
	}}

	commonConfig := &ingress.Configuration{
		Backends: backends,
		Servers:  servers,
	}

	runningConfig := &ingress.Configuration{
		Backends: backends,
		Servers:  servers,
	}

	newConfig := commonConfig
	if !IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("When new config is same as the running config it should be deemed as dynamically configurable")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "another-backend-8081"}},
		Servers:  []*ingress.Server{{Hostname: "myapp1.fake"}},
	}
	if IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("Expected to not be dynamically configurable when there's more than just backends change")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "a-backend-8080"}},
		Servers:  servers,
	}

	if !IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("Expected to be dynamically configurable when only backends change")
	}

	newServers := []*ingress.Server{{
		Hostname: "myapp1.fake",
		Locations: []*ingress.Location{
			{
				Path:    "/",
				Backend: "fakenamespace-myapp-80",
			},
		},
		SSLCert: &ingress.SSLCert{
			PemCertKey: "fake-certificate",
		},
	}}

	newConfig = &ingress.Configuration{
		Backends: backends,
		Servers:  newServers,
	}
	if IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("Expected to not be dynamically configurable when dynamic certificates is enabled and a non-certificate field in servers is updated")
	}

	newServers[0].Hostname = "myapp.fake"
	newServers[0].SSLCert.PemCertKey = "new-fake-certificate"

	newConfig = &ingress.Configuration{
		Backends: backends,
		Servers:  newServers,
	}
	if !IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("Expected to be dynamically configurable when only SSLCert changes")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "a-backend-8080"}},
		Servers:  newServers,
	}
	if !IsDynamicConfigurationEnough(newConfig, runningConfig) {
		t.Errorf("Expected to be dynamically configurable when backend and SSLCert changes")
	}

	if !runningConfig.Equal(commonConfig) {
		t.Errorf("Expected running config to not change")
	}

	if !newConfig.Equal(&ingress.Configuration{Backends: []*ingress.Backend{{Name: "a-backend-8080"}}, Servers: newServers}) {
		t.Errorf("Expected new config to not change")
	}
}

func generateDumbIngressforPathTest(pathType *networkingv1.PathType, path string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dumb",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "test.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									PathType: pathType,
									Path:     path,
								},
							},
						},
					},
				},
			},
		},
	}
}

func generateComplexIngress(ing *networkingv1.Ingress) *networkingv1.Ingress {

	oldRules := ing.Spec.DeepCopy().Rules
	ing.Spec.Rules = []networkingv1.IngressRule{
		{
			Host: "test1.com",
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							PathType: &pathTypeExact,
							Path:     "/xpto",
						},
					},
				},
			},
		},
		{
			Host: "test2.com",
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							PathType: &pathTypeExact,
							Path:     "/someotherpath",
						},
						{
							PathType: &pathTypePrefix,
							Path:     "/someprefix/~xpto/lala123",
						},
					},
				},
			},
		},
	}
	// we want to invert the order to test better :)
	ing.Spec.Rules = append(ing.Spec.Rules, oldRules...)

	return ing
}

var (
	pathTypeExact        = networkingv1.PathTypeExact
	pathTypePrefix       = networkingv1.PathTypePrefix
	pathTypeImplSpecific = networkingv1.PathTypeImplementationSpecific
)

const (
	defaultAdditionalChars = "^%$[](){}*+?"
)

func TestValidateIngressPath(t *testing.T) {
	tests := []struct {
		name                     string
		copyIng                  *networkingv1.Ingress
		EnablePathTypeValidation bool
		additionalChars          string
		wantErr                  bool
	}{
		{
			name:    "should return nil when ingress = nil",
			wantErr: false,
			copyIng: nil,
		},
		{
			name:    "should accept valid path on pathType Exact",
			wantErr: false,
			copyIng: generateDumbIngressforPathTest(&pathTypeExact, "/xpto/~user9/t-e_st.exe"),
		},
		{
			name:    "should accept valid path on pathType Prefix",
			wantErr: false,
			copyIng: generateDumbIngressforPathTest(&pathTypePrefix, "/xpto/~user9/t-e_st.exe"),
		},
		{
			name:    "should accept valid simple path on pathType Impl Specific",
			wantErr: false,
			copyIng: generateDumbIngressforPathTest(&pathTypeImplSpecific, "/xpto/~user9/t-e_st.exe"),
		},
		{
			name:    "should accept valid path on pathType nil",
			wantErr: false,
			copyIng: generateDumbIngressforPathTest(nil, "/xpto/~user/t-e_st.exe"),
		},
		{
			name:    "should accept empty path",
			wantErr: false,
			copyIng: generateDumbIngressforPathTest(&pathTypePrefix, ""),
		},
		{
			name:            "should deny path with bad characters and pathType not implementationSpecific",
			wantErr:         true,
			additionalChars: "()",
			copyIng:         generateDumbIngressforPathTest(&pathTypeExact, "/foo/bar/(.+)"),
		},
		{
			name:                     "should accept path with regex characters and pathType implementationSpecific",
			wantErr:                  false,
			additionalChars:          defaultAdditionalChars,
			EnablePathTypeValidation: false,
			copyIng:                  generateDumbIngressforPathTest(&pathTypeImplSpecific, "/foo/bar/(.+)"),
		},
		{
			name:                     "should accept path with regex characters and pathType exact, but pathType validation disabled",
			wantErr:                  false,
			additionalChars:          defaultAdditionalChars,
			EnablePathTypeValidation: false,
			copyIng:                  generateDumbIngressforPathTest(&pathTypeExact, "/foo/bar/(.+)"),
		},
		{
			name:                     "should reject path when the allowed additional set does not match",
			wantErr:                  true,
			additionalChars:          "().?",
			EnablePathTypeValidation: true,
			copyIng:                  generateDumbIngressforPathTest(&pathTypeImplSpecific, "/foo/bar/(.+)"),
		},
		{
			name:                     "should accept path when the allowed additional set does match",
			wantErr:                  false,
			additionalChars:          "().?",
			EnablePathTypeValidation: false,
			copyIng:                  generateDumbIngressforPathTest(&pathTypeImplSpecific, "/foo/bar/(.?)"),
		},
		{
			name:                     "should block if at least one path is bad",
			wantErr:                  true,
			EnablePathTypeValidation: false,
			copyIng:                  generateComplexIngress(generateDumbIngressforPathTest(&pathTypeExact, "/foo/bar/(.?)")),
		},
		{
			name:                     "should block if at least one path is bad",
			wantErr:                  true,
			EnablePathTypeValidation: true,
			copyIng:                  generateComplexIngress(generateDumbIngressforPathTest(&pathTypeImplSpecific, "/foo/bar/(.?)")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateIngressPath(tt.copyIng, tt.EnablePathTypeValidation, tt.additionalChars); (err != nil) != tt.wantErr {
				t.Errorf("ValidateIngressPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
