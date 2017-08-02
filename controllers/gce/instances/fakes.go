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
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress/controllers/gce/utils"
)

// NewFakeInstanceGroups creates a new FakeInstanceGroups.
func NewFakeInstanceGroups(nodes sets.String) *FakeInstanceGroups {
	return &FakeInstanceGroups{
		instances:        nodes,
		listResult:       getInstanceList(nodes),
		namer:            utils.Namer{},
		zonesToInstances: map[string][]string{},
	}
}

// InstanceGroup fakes

// FakeZoneLister records zones for nodes.
type FakeZoneLister struct {
	Zones []string
}

// ListZones returns the list of zones.
func (z *FakeZoneLister) ListZones() ([]string, error) {
	return z.Zones, nil
}

// GetZoneForNode returns the only zone stored in the fake zone lister.
func (z *FakeZoneLister) GetZoneForNode(name string) (string, error) {
	// TODO: evolve as required, it's currently needed just to satisfy the
	// interface in unittests that don't care about zones. See unittests in
	// controller/util_test for actual zoneLister testing.
	return z.Zones[0], nil
}

// FakeInstanceGroups fakes out the instance groups api.
type FakeInstanceGroups struct {
	instances        sets.String
	instanceGroups   []*compute.InstanceGroup
	Ports            []int64
	getResult        *compute.InstanceGroup
	listResult       *compute.InstanceGroupsListInstances
	calls            []int
	namer            utils.Namer
	zonesToInstances map[string][]string
}

// GetInstanceGroup fakes getting an instance group from the cloud.
func (f *FakeInstanceGroups) GetInstanceGroup(name, zone string) (*compute.InstanceGroup, error) {
	f.calls = append(f.calls, utils.Get)
	for _, ig := range f.instanceGroups {
		if ig.Name == name && ig.Zone == zone {
			return ig, nil
		}
	}

	return nil, utils.FakeGoogleAPINotFoundErr()
}

// CreateInstanceGroup fakes instance group creation.
func (f *FakeInstanceGroups) CreateInstanceGroup(ig *compute.InstanceGroup, zone string) error {
	ig.SelfLink = ig.Name
	ig.Zone = zone
	f.instanceGroups = append(f.instanceGroups, ig)
	return nil
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
		return fmt.Errorf("instance group %v not found", name)
	}
	f.instanceGroups = newGroups
	return nil
}

// ListInstancesInInstanceGroup fakes listing instances in an instance group.
func (f *FakeInstanceGroups) ListInstancesInInstanceGroup(name, zone string, state string) (*compute.InstanceGroupsListInstances, error) {
	return f.listResult, nil
}

// AddInstancesToInstanceGroup fakes adding instances to an instance group.
func (f *FakeInstanceGroups) AddInstancesToInstanceGroup(name, zone string, instanceRefs []*compute.InstanceReference) error {
	instanceNames := toInstanceNames(instanceRefs)
	f.calls = append(f.calls, utils.AddInstances)
	f.instances.Insert(instanceNames...)
	if _, ok := f.zonesToInstances[zone]; !ok {
		f.zonesToInstances[zone] = []string{}
	}
	f.zonesToInstances[zone] = append(f.zonesToInstances[zone], instanceNames...)
	return nil
}

// GetInstancesByZone returns the zone to instances map.
func (f *FakeInstanceGroups) GetInstancesByZone() map[string][]string {
	return f.zonesToInstances
}

// RemoveInstancesFromInstanceGroup fakes removing instances from an instance group.
func (f *FakeInstanceGroups) RemoveInstancesFromInstanceGroup(name, zone string, instanceRefs []*compute.InstanceReference) error {
	instanceNames := toInstanceNames(instanceRefs)
	f.calls = append(f.calls, utils.RemoveInstances)
	f.instances.Delete(instanceNames...)
	l, ok := f.zonesToInstances[zone]
	if !ok {
		return nil
	}
	newIns := []string{}
	delIns := sets.NewString(instanceNames...)
	for _, oldIns := range l {
		if delIns.Has(oldIns) {
			continue
		}
		newIns = append(newIns, oldIns)
	}
	f.zonesToInstances[zone] = newIns
	return nil
}

func (f *FakeInstanceGroups) SetNamedPortsOfInstanceGroup(igName, zone string, namedPorts []*compute.NamedPort) error {
	found := false
	for _, ig := range f.instanceGroups {
		if ig.Name == igName && ig.Zone == zone {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Failed to find instance group %q in zone %q", igName, zone)
	}

	f.Ports = f.Ports[:0]
	for _, port := range namedPorts {
		f.Ports = append(f.Ports, port.Port)
	}
	return nil
}

// getInstanceList returns an instance list based on the given names.
// The names cannot contain a '.', the real gce api validates against this.
func getInstanceList(nodeNames sets.String) *compute.InstanceGroupsListInstances {
	instanceNames := nodeNames.List()
	computeInstances := []*compute.InstanceWithNamedPorts{}
	for _, name := range instanceNames {
		instanceLink := getInstanceUrl(name)
		computeInstances = append(
			computeInstances, &compute.InstanceWithNamedPorts{
				Instance: instanceLink})
	}
	return &compute.InstanceGroupsListInstances{
		Items: computeInstances,
	}
}

func (f *FakeInstanceGroups) ToInstanceReferences(zone string, instanceNames []string) (refs []*compute.InstanceReference) {
	for _, ins := range instanceNames {
		instanceLink := getInstanceUrl(ins)
		refs = append(refs, &compute.InstanceReference{Instance: instanceLink})
	}
	return refs
}

func getInstanceUrl(instanceName string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s",
		"project", "zone", instanceName)
}

func toInstanceNames(instanceRefs []*compute.InstanceReference) []string {
	instanceNames := make([]string, len(instanceRefs))
	for ix := range instanceRefs {
		url := instanceRefs[ix].Instance
		parts := strings.Split(url, "/")
		instanceNames[ix] = parts[len(parts)-1]
	}
	return instanceNames
}
