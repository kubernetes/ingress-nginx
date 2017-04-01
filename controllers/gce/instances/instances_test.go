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

package instances

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

const defaultZone = "default-zone"

func newNodePool(f *FakeInstanceGroups, zone string) NodePool {
	pool := NewNodePool(f)
	pool.Init(&FakeZoneLister{[]string{zone}})
	return pool
}

func TestNodePoolSync(t *testing.T) {
	f := NewFakeInstanceGroups(sets.NewString(
		[]string{"n1", "n2"}...))
	pool := newNodePool(f, defaultZone)
	pool.AddInstanceGroup("test", 80)

	// KubeNodes: n1
	// GCENodes: n1, n2
	// Remove n2 from the instance group.

	f.calls = []int{}
	kubeNodes := sets.NewString([]string{"n1"}...)
	pool.Sync(kubeNodes.List())
	if f.instances.Len() != kubeNodes.Len() || !kubeNodes.IsSuperset(f.instances) {
		t.Fatalf("%v != %v", kubeNodes, f.instances)
	}

	// KubeNodes: n1, n2
	// GCENodes: n1
	// Try to add n2 to the instance group.

	f = NewFakeInstanceGroups(sets.NewString([]string{"n1"}...))
	pool = newNodePool(f, defaultZone)
	pool.AddInstanceGroup("test", 80)

	f.calls = []int{}
	kubeNodes = sets.NewString([]string{"n1", "n2"}...)
	pool.Sync(kubeNodes.List())
	if f.instances.Len() != kubeNodes.Len() ||
		!kubeNodes.IsSuperset(f.instances) {
		t.Fatalf("%v != %v", kubeNodes, f.instances)
	}

	// KubeNodes: n1, n2
	// GCENodes: n1, n2
	// Do nothing.

	f = NewFakeInstanceGroups(sets.NewString([]string{"n1", "n2"}...))
	pool = newNodePool(f, defaultZone)
	pool.AddInstanceGroup("test", 80)

	f.calls = []int{}
	kubeNodes = sets.NewString([]string{"n1", "n2"}...)
	pool.Sync(kubeNodes.List())
	if len(f.calls) != 0 {
		t.Fatalf(
			"Did not expect any calls, got %+v", f.calls)
	}
}
