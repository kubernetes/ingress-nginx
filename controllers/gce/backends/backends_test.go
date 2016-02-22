/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package backends

import (
	"testing"

	"k8s.io/contrib/ingress/controllers/gce/healthchecks"
	"k8s.io/contrib/ingress/controllers/gce/instances"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/util/sets"
)

func newBackendPool(f BackendServices, fakeIGs instances.InstanceGroups) BackendPool {
	namer := utils.Namer{}
	return NewBackendPool(
		f,
		healthchecks.NewHealthChecker(healthchecks.NewFakeHealthChecks(), "/", namer),
		instances.NewNodePool(fakeIGs, "default-zone"), namer)
}

func TestBackendPoolAdd(t *testing.T) {
	f := NewFakeBackendServices()
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs)
	namer := utils.Namer{}

	// Add a backend for a port, then re-add the same port and
	// make sure it corrects a broken link from the backend to
	// the instance group.
	nodePort := int64(8080)
	pool.Add(nodePort)
	beName := namer.BeName(nodePort)

	// Check that the new backend has the right port
	be, err := f.GetBackendService(beName)
	if err != nil {
		t.Fatalf("Did not find expected backend %v", beName)
	}
	if be.Port != nodePort {
		t.Fatalf("Backend %v has wrong port %v, expected %v", be.Name, be.Port, nodePort)
	}
	// Check that the instance group has the new port
	var found bool
	for _, port := range fakeIGs.Ports {
		if port == nodePort {
			found = true
		}
	}
	if !found {
		t.Fatalf("Port %v not added to instance group", nodePort)
	}

	// Mess up the link between backend service and instance group.
	// This simulates a user doing foolish things through the UI.
	f.calls = []int{}
	be, err = f.GetBackendService(beName)
	be.Backends[0].Group = "test edge hop"
	f.UpdateBackendService(be)

	pool.Add(nodePort)
	for _, call := range f.calls {
		if call == utils.Create {
			t.Fatalf("Unexpected create for existing backend service")
		}
	}
	gotBackend, _ := f.GetBackendService(beName)
	gotGroup, _ := fakeIGs.GetInstanceGroup(namer.IGName(), "default-zone")
	if gotBackend.Backends[0].Group != gotGroup.SelfLink {
		t.Fatalf(
			"Broken instance group link: %v %v",
			gotBackend.Backends[0].Group,
			gotGroup.SelfLink)
	}
}

func TestBackendPoolSync(t *testing.T) {

	// Call sync on a backend pool with a list of ports, make sure the pool
	// creates/deletes required ports.
	svcNodePorts := []int64{81, 82, 83}
	f := NewFakeBackendServices()
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs)
	pool.Add(81)
	pool.Add(90)
	pool.Sync(svcNodePorts)
	pool.GC(svcNodePorts)
	if _, err := pool.Get(90); err == nil {
		t.Fatalf("Did not expect to find port 90")
	}
	for _, port := range svcNodePorts {
		if _, err := pool.Get(port); err != nil {
			t.Fatalf("Expected to find port %v", port)
		}
	}

}

func TestBackendPoolShutdown(t *testing.T) {
	f := NewFakeBackendServices()
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs)
	namer := utils.Namer{}

	pool.Add(80)
	pool.Shutdown()
	if _, err := f.GetBackendService(namer.BeName(80)); err == nil {
		t.Fatalf("%v", err)
	}

}
