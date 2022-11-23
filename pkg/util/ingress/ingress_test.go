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
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
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

func generateDumbIngressforPathTest(regexEnabled bool) *networkingv1.Ingress {
	var annotations = make(map[string]string)
	regexAnnotation := fmt.Sprintf("%s/use-regex", parser.AnnotationsPrefix)
	if regexEnabled {
		annotations[regexAnnotation] = "true"
	}
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "dumb",
			Namespace:   "default",
			Annotations: annotations,
		},
	}
}

func TestIsSafePath(t *testing.T) {
	tests := []struct {
		name    string
		copyIng *networkingv1.Ingress
		path    string
		want    bool
	}{
		{
			name:    "should accept valid path with regex disabled",
			want:    true,
			copyIng: generateDumbIngressforPathTest(false),
			path:    "/xpto/~user/t-e_st.exe",
		},
		{
			name:    "should accept valid path / with regex disabled",
			want:    true,
			copyIng: generateDumbIngressforPathTest(false),
			path:    "/",
		},
		{
			name:    "should reject invalid path with invalid chars",
			want:    false,
			copyIng: generateDumbIngressforPathTest(false),
			path:    "/foo/bar/;xpto",
		},
		{
			name:    "should reject regex path when regex is disabled",
			want:    false,
			copyIng: generateDumbIngressforPathTest(false),
			path:    "/foo/bar/(.+)",
		},
		{
			name:    "should accept valid path / with regex enabled",
			want:    true,
			copyIng: generateDumbIngressforPathTest(true),
			path:    "/",
		},
		{
			name:    "should accept regex path when regex is enabled",
			want:    true,
			copyIng: generateDumbIngressforPathTest(true),
			path:    "/foo/bar/(.+)",
		},
		{
			name:    "should reject regex path when regex is enabled but the path is invalid",
			want:    false,
			copyIng: generateDumbIngressforPathTest(true),
			path:    "/foo/bar/;xpto",
		},
		{
			name:    "should reject regex path when regex is enabled but the path is invalid",
			want:    false,
			copyIng: generateDumbIngressforPathTest(true),
			path:    ";xpto",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafePath(tt.copyIng, tt.path); got != tt.want {
				t.Errorf("IsSafePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
