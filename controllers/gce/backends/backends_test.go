/*
Copyright 2015 The Kubernetes Authors.

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
	"net/http"
	"testing"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/ingress/controllers/gce/healthchecks"
	"k8s.io/ingress/controllers/gce/instances"
	"k8s.io/ingress/controllers/gce/storage"
	"k8s.io/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/util/sets"

	"google.golang.org/api/googleapi"
)

const defaultZone = "zone-a"

var noOpErrFunc = func(op int, be *compute.BackendService) error { return nil }

func newBackendPool(f BackendServices, fakeIGs instances.InstanceGroups, syncWithCloud bool) BackendPool {
	namer := &utils.Namer{}
	nodePool := instances.NewNodePool(fakeIGs)
	nodePool.Init(&instances.FakeZoneLister{Zones: []string{defaultZone}})
	healthChecks := healthchecks.NewHealthChecker(healthchecks.NewFakeHealthChecks(), "/", namer)
	healthChecks.Init(&healthchecks.FakeHealthCheckGetter{DefaultHealthCheck: nil})
	return NewBackendPool(
		f, healthChecks, nodePool, namer, []int64{}, syncWithCloud)
}

func TestBackendPoolAdd(t *testing.T) {
	f := NewFakeBackendServices(noOpErrFunc)
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs, false)
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
	be.Backends = []*compute.Backend{
		{Group: "test edge hop"},
	}
	f.UpdateBackendService(be)

	pool.Add(nodePort)
	for _, call := range f.calls {
		if call == utils.Create {
			t.Fatalf("Unexpected create for existing backend service")
		}
	}
	gotBackend, err := f.GetBackendService(beName)
	if err != nil {
		t.Fatalf("Failed to find a backend with name %v: %v", beName, err)
	}
	gotGroup, err := fakeIGs.GetInstanceGroup(namer.IGName(), defaultZone)
	if err != nil {
		t.Fatalf("Failed to find instance group %v", namer.IGName())
	}
	backendLinks := sets.NewString()
	for _, be := range gotBackend.Backends {
		backendLinks.Insert(be.Group)
	}
	if !backendLinks.Has(gotGroup.SelfLink) {
		t.Fatalf(
			"Broken instance group link, got: %+v expected: %v",
			backendLinks.List(),
			gotGroup.SelfLink)
	}
}

func TestBackendPoolSync(t *testing.T) {
	// Call sync on a backend pool with a list of ports, make sure the pool
	// creates/deletes required ports.
	svcNodePorts := []int64{81, 82, 83}
	f := NewFakeBackendServices(noOpErrFunc)
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs, true)
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

	svcNodePorts = []int64{81}
	deletedPorts := []int64{82, 83}
	pool.GC(svcNodePorts)
	for _, port := range deletedPorts {
		if _, err := pool.Get(port); err == nil {
			t.Fatalf("Pool contains %v after deletion", port)
		}
	}

	// All these backends should be ignored because they don't belong to the cluster.
	// foo - non k8s managed backend
	// k8s-be-foo - foo is not a nodeport
	// k8s--bar--foo - too many cluster delimiters
	// k8s-be-3001--uid - another cluster tagged with uid
	unrelatedBackends := sets.NewString([]string{"foo", "k8s-be-foo", "k8s--bar--foo", "k8s-be-30001--uid"}...)
	for _, name := range unrelatedBackends.List() {
		f.CreateBackendService(&compute.BackendService{Name: name})
	}

	namer := &utils.Namer{}
	// This backend should get deleted again since it is managed by this cluster.
	f.CreateBackendService(&compute.BackendService{Name: namer.BeName(deletedPorts[0])})

	// TODO: Avoid casting.
	// Repopulate the pool with a cloud list, which now includes the 82 port
	// backend. This would happen if, say, an ingress backend is removed
	// while the controller is restarting.
	pool.(*Backends).snapshotter.(*storage.CloudListingPool).ReplenishPool()

	pool.GC(svcNodePorts)

	currBackends, _ := f.ListBackendServices()
	currSet := sets.NewString()
	for _, b := range currBackends.Items {
		currSet.Insert(b.Name)
	}
	// Port 81 still exists because it's an in-use service NodePort.
	knownBe := namer.BeName(81)
	if !currSet.Has(knownBe) {
		t.Fatalf("Expected %v to exist in backend pool", knownBe)
	}
	currSet.Delete(knownBe)
	if !currSet.Equal(unrelatedBackends) {
		t.Fatalf("Some unrelated backends were deleted. Expected %+v, got %+v", unrelatedBackends, currSet)
	}
}

func TestBackendPoolShutdown(t *testing.T) {
	f := NewFakeBackendServices(noOpErrFunc)
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs, false)
	namer := utils.Namer{}

	pool.Add(80)
	pool.Shutdown()
	if _, err := f.GetBackendService(namer.BeName(80)); err == nil {
		t.Fatalf("%v", err)
	}
}

func TestBackendInstanceGroupClobbering(t *testing.T) {
	f := NewFakeBackendServices(noOpErrFunc)
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs, false)
	namer := utils.Namer{}

	// This will add the instance group k8s-ig to the instance pool
	pool.Add(80)

	be, err := f.GetBackendService(namer.BeName(80))
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Simulate another controller updating the same backend service with
	// a different instance group
	newGroups := []*compute.Backend{
		{Group: "k8s-ig-bar"},
		{Group: "k8s-ig-foo"},
	}
	be.Backends = append(be.Backends, newGroups...)
	if err := f.UpdateBackendService(be); err != nil {
		t.Fatalf("Failed to update backend service %v", be.Name)
	}

	// Make sure repeated adds don't clobber the inserted instance group
	pool.Add(80)
	be, err = f.GetBackendService(namer.BeName(80))
	if err != nil {
		t.Fatalf("%v", err)
	}
	gotGroups := sets.NewString()
	for _, g := range be.Backends {
		gotGroups.Insert(g.Group)
	}

	// seed expectedGroups with the first group native to this controller
	expectedGroups := sets.NewString("k8s-ig")
	for _, newGroup := range newGroups {
		expectedGroups.Insert(newGroup.Group)
	}
	if !expectedGroups.Equal(gotGroups) {
		t.Fatalf("Expected %v Got %v", expectedGroups, gotGroups)
	}
}

func TestBackendCreateBalancingMode(t *testing.T) {
	f := NewFakeBackendServices(noOpErrFunc)

	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	pool := newBackendPool(f, fakeIGs, false)
	namer := utils.Namer{}
	nodePort := int64(8080)
	modes := []BalancingMode{Rate, Utilization}

	// block the creation of Backends with the given balancingMode
	// and verify that a backend with the other balancingMode is
	// created
	for i, bm := range modes {
		f.errFunc = func(op int, be *compute.BackendService) error {
			for _, b := range be.Backends {
				if b.BalancingMode == string(bm) {
					return &googleapi.Error{Code: http.StatusBadRequest}
				}
			}
			return nil
		}

		pool.Add(nodePort)
		be, err := f.GetBackendService(namer.BeName(nodePort))
		if err != nil {
			t.Fatalf("%v", err)
		}

		for _, b := range be.Backends {
			if b.BalancingMode != string(modes[(i+1)%len(modes)]) {
				t.Fatalf("Wrong balancing mode, expected %v got %v", modes[(i+1)%len(modes)], b.BalancingMode)
			}
		}
		pool.GC([]int64{})
	}
}
