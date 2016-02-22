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

package instances

import (
	"fmt"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/util/sets"
)

// NewFakeInstanceGroups creates a new FakeInstanceGroups.
func NewFakeInstanceGroups(nodes sets.String) *FakeInstanceGroups {
	return &FakeInstanceGroups{
		instances:  nodes,
		listResult: getInstanceList(nodes),
		namer:      utils.Namer{},
	}
}

// InstanceGroup fakes

// FakeInstanceGroups fakes out the instance groups api.
type FakeInstanceGroups struct {
	instances      sets.String
	instanceGroups []*compute.InstanceGroup
	Ports          []int64
	getResult      *compute.InstanceGroup
	listResult     *compute.InstanceGroupsListInstances
	calls          []int
	namer          utils.Namer
}

// GetInstanceGroup fakes getting an instance group from the cloud.
func (f *FakeInstanceGroups) GetInstanceGroup(name, zone string) (*compute.InstanceGroup, error) {
	f.calls = append(f.calls, utils.Get)
	for _, ig := range f.instanceGroups {
		if ig.Name == name {
			return ig, nil
		}
	}
	// TODO: Return googleapi 404 error
	return nil, fmt.Errorf("Instance group %v not found", name)
}

// CreateInstanceGroup fakes instance group creation.
func (f *FakeInstanceGroups) CreateInstanceGroup(name, zone string) (*compute.InstanceGroup, error) {
	newGroup := &compute.InstanceGroup{Name: name, SelfLink: name}
	f.instanceGroups = append(f.instanceGroups, newGroup)
	return newGroup, nil
}

// DeleteInstanceGroup fakes instance group deletion.
func (f *FakeInstanceGroups) DeleteInstanceGroup(name, zone string) error {
	newGroups := []*compute.InstanceGroup{}
	found := false
	for _, ig := range f.instanceGroups {
		if ig.Name == name {
			found = true
			continue
		}
		newGroups = append(newGroups, ig)
	}
	if !found {
		return fmt.Errorf("Instance Group %v not found", name)
	}
	f.instanceGroups = newGroups
	return nil
}

// ListInstancesInInstanceGroup fakes listing instances in an instance group.
func (f *FakeInstanceGroups) ListInstancesInInstanceGroup(name, zone string, state string) (*compute.InstanceGroupsListInstances, error) {
	return f.listResult, nil
}

// AddInstancesToInstanceGroup fakes adding instances to an instance group.
func (f *FakeInstanceGroups) AddInstancesToInstanceGroup(name, zone string, instanceNames []string) error {
	f.calls = append(f.calls, utils.AddInstances)
	f.instances.Insert(instanceNames...)
	return nil
}

// RemoveInstancesFromInstanceGroup fakes removing instances from an instance group.
func (f *FakeInstanceGroups) RemoveInstancesFromInstanceGroup(name, zone string, instanceNames []string) error {
	f.calls = append(f.calls, utils.RemoveInstances)
	f.instances.Delete(instanceNames...)
	return nil
}

// AddPortToInstanceGroup fakes adding ports to an Instance Group.
func (f *FakeInstanceGroups) AddPortToInstanceGroup(ig *compute.InstanceGroup, port int64) (*compute.NamedPort, error) {
	f.Ports = append(f.Ports, port)
	return &compute.NamedPort{Name: f.namer.BeName(port), Port: port}, nil
}

// getInstanceList returns an instance list based on the given names.
// The names cannot contain a '.', the real gce api validates against this.
func getInstanceList(nodeNames sets.String) *compute.InstanceGroupsListInstances {
	instanceNames := nodeNames.List()
	computeInstances := []*compute.InstanceWithNamedPorts{}
	for _, name := range instanceNames {
		instanceLink := fmt.Sprintf(
			"https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s",
			"project", "zone", name)
		computeInstances = append(
			computeInstances, &compute.InstanceWithNamedPorts{
				Instance: instanceLink})
	}
	return &compute.InstanceGroupsListInstances{
		Items: computeInstances,
	}
}
