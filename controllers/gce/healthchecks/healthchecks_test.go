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

package healthchecks

import (
	"testing"

	"k8s.io/ingress/controllers/gce/utils"
)

func TestFakeHealthCheckActions(t *testing.T) {
	namer := &utils.Namer{}
	healthChecks := NewHealthChecker(NewFakeHealthChecks(), "/", namer)
	healthChecks.Init(&FakeHealthCheckGetter{DefaultHealthCheck: nil})

	err := healthChecks.Add(80)
	if err != nil {
		t.Fatalf("unexpected error")
	}

	_, err1 := healthChecks.Get(8080)
	if err1 == nil {
		t.Errorf("expected error")
	}

	hc, err2 := healthChecks.Get(80)
	if err2 != nil {
		t.Errorf("unexpected error")
	} else {
		if hc == nil {
			t.Errorf("expected a *compute.HttpHealthCheck")
		}
	}

	err = healthChecks.Delete(8080)
	if err == nil {
		t.Errorf("expected error")
	}

	err = healthChecks.Delete(80)
	if err != nil {
		t.Errorf("unexpected error")
	}

	_, err3 := healthChecks.Get(80)
	if err3 == nil {
		t.Errorf("expected error")
	}
}
