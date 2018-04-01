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

	"k8s.io/ingress-nginx/internal/ingress"
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
	}}

	commonConfig := &ingress.Configuration{
		Backends: backends,
		Servers:  servers,
	}

	n := &NGINXController{
		runningConfig: &ingress.Configuration{
			Backends: backends,
			Servers:  servers,
		},
	}

	newConfig := commonConfig
	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("When new config is same as the running config it should be deemed as dynamically configurable")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "another-backend-8081"}},
		Servers:  []*ingress.Server{{Hostname: "myapp1.fake"}},
	}
	if n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to not be dynamically configurable when there's more than just backends change")
	}

	newConfig = &ingress.Configuration{
		Backends: []*ingress.Backend{{Name: "a-backend-8080"}},
		Servers:  servers,
	}
	if !n.IsDynamicConfigurationEnough(newConfig) {
		t.Errorf("Expected to be dynamically configurable when only backends change")
	}

	if !n.runningConfig.Equal(commonConfig) {
		t.Errorf("Expected running config to not change")
	}

	if !newConfig.Equal(&ingress.Configuration{Backends: []*ingress.Backend{{Name: "a-backend-8080"}}, Servers: servers}) {
		t.Errorf("Expected new config to not change")
	}
}

func TestNginxHashBucketSize(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 32},
		{1, 32},
		{2, 32},
		{3, 32},
		// ...
		{13, 32},
		{14, 32},
		{15, 64},
		{16, 64},
		// ...
		{45, 64},
		{46, 64},
		{47, 128},
		{48, 128},
		// ...
		// ...
		{109, 128},
		{110, 128},
		{111, 256},
		{112, 256},
		// ...
		{237, 256},
		{238, 256},
		{239, 512},
		{240, 512},
	}

	for _, test := range tests {
		actual := nginxHashBucketSize(test.n)
		if actual != test.expected {
			t.Errorf("Test nginxHashBucketSize(%d): expected %d but returned %d", test.n, test.expected, actual)
		}
	}
}

func TestNextPowerOf2(t *testing.T) {
	// Powers of 2
	actual := nextPowerOf2(2)
	if actual != 2 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 2, actual)
	}
	actual = nextPowerOf2(4)
	if actual != 4 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 4, actual)
	}
	actual = nextPowerOf2(32)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(256)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Not Powers of 2
	actual = nextPowerOf2(7)
	if actual != 8 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 8, actual)
	}
	actual = nextPowerOf2(9)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(15)
	if actual != 16 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 16, actual)
	}
	actual = nextPowerOf2(17)
	if actual != 32 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 32, actual)
	}
	actual = nextPowerOf2(250)
	if actual != 256 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 256, actual)
	}

	// Other
	actual = nextPowerOf2(0)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-1)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
	actual = nextPowerOf2(-2)
	if actual != 0 {
		t.Errorf("TestNextPowerOf2: expected %d but returned %d.", 0, actual)
	}
}
