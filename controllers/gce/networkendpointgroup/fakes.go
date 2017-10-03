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

package networkendpointgroup

import (
	"fmt"
	computealpha "google.golang.org/api/compute/v0.alpha"
	"k8s.io/apimachinery/pkg/util/sets"
	"reflect"
	"sync"
)

const (
	TestZone1     = "zone1"
	TestZone2     = "zone2"
	TestInstance1 = "instance1"
	TestInstance2 = "instance2"
	TestInstance3 = "instance3"
	TestInstance4 = "instance4"
)

type fakeZoneGetter struct {
	zoneInstanceMap map[string]sets.String
}

func NewFakeZoneGetter() *fakeZoneGetter {
	return &fakeZoneGetter{
		zoneInstanceMap: map[string]sets.String{
			TestZone1: sets.NewString(TestInstance1, TestInstance2),
			TestZone2: sets.NewString(TestInstance3, TestInstance4),
		},
	}
}

func (f *fakeZoneGetter) ListZones() ([]string, error) {
	ret := []string{}
	for key := range f.zoneInstanceMap {
		ret = append(ret, key)
	}
	return ret, nil
}
func (f *fakeZoneGetter) GetZoneForNode(name string) (string, error) {
	for zone, instances := range f.zoneInstanceMap {
		if instances.Has(name) {
			return zone, nil
		}
	}
	return "", NotFoundError
}

type FakeNetworkEndpointGroupCloud struct {
	NetworkEndpointGroups map[string][]*computealpha.NetworkEndpointGroup
	NetworkEndpoints      map[string][]*computealpha.NetworkEndpoint
	Subnetwork            string
	Network               string
	mu                    sync.Mutex
}

func NewFakeNetworkEndpointGroupCloud(subnetwork, network string) NetworkEndpointGroupCloud {
	return &FakeNetworkEndpointGroupCloud{
		Subnetwork:            subnetwork,
		Network:               network,
		NetworkEndpointGroups: map[string][]*computealpha.NetworkEndpointGroup{},
		NetworkEndpoints:      map[string][]*computealpha.NetworkEndpoint{},
	}
}

var NotFoundError = fmt.Errorf("Not Found")

func (cloud *FakeNetworkEndpointGroupCloud) GetNetworkEndpointGroup(name string, zone string) (*computealpha.NetworkEndpointGroup, error) {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	negs, ok := cloud.NetworkEndpointGroups[zone]
	if ok {
		for _, neg := range negs {
			if neg.Name == name {
				return neg, nil
			}
		}
	}
	return nil, NotFoundError
}

func networkEndpointKey(name, zone string) string {
	return fmt.Sprintf("%s-%s", zone, name)
}

func (cloud *FakeNetworkEndpointGroupCloud) ListNetworkEndpointGroup(zone string) ([]*computealpha.NetworkEndpointGroup, error) {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	return cloud.NetworkEndpointGroups[zone], nil
}

func (cloud *FakeNetworkEndpointGroupCloud) AggregatedListNetworkEndpointGroup() (map[string][]*computealpha.NetworkEndpointGroup, error) {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	return cloud.NetworkEndpointGroups, nil
}

func (cloud *FakeNetworkEndpointGroupCloud) CreateNetworkEndpointGroup(neg *computealpha.NetworkEndpointGroup, zone string) error {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	if _, ok := cloud.NetworkEndpointGroups[zone]; !ok {
		cloud.NetworkEndpointGroups[zone] = []*computealpha.NetworkEndpointGroup{}
	}
	cloud.NetworkEndpointGroups[zone] = append(cloud.NetworkEndpointGroups[zone], neg)
	cloud.NetworkEndpoints[networkEndpointKey(neg.Name, zone)] = []*computealpha.NetworkEndpoint{}
	return nil
}

func (cloud *FakeNetworkEndpointGroupCloud) DeleteNetworkEndpointGroup(name string, zone string) error {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	delete(cloud.NetworkEndpoints, networkEndpointKey(name, zone))
	negs := cloud.NetworkEndpointGroups[zone]
	newList := []*computealpha.NetworkEndpointGroup{}
	found := false
	for _, neg := range negs {
		if neg.Name == name {
			found = true
			continue
		}
		newList = append(newList, neg)
	}
	if !found {
		return NotFoundError
	}
	cloud.NetworkEndpointGroups[zone] = newList
	return nil
}

func (cloud *FakeNetworkEndpointGroupCloud) AttachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	cloud.NetworkEndpoints[networkEndpointKey(name, zone)] = append(cloud.NetworkEndpoints[networkEndpointKey(name, zone)], endpoints...)
	return nil
}

func (cloud *FakeNetworkEndpointGroupCloud) DetachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	newList := []*computealpha.NetworkEndpoint{}
	for _, ne := range cloud.NetworkEndpoints[networkEndpointKey(name, zone)] {
		found := false
		for _, remove := range endpoints {
			if reflect.DeepEqual(*ne, *remove) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		newList = append(newList, ne)
	}
	cloud.NetworkEndpoints[networkEndpointKey(name, zone)] = newList
	return nil
}

func (cloud *FakeNetworkEndpointGroupCloud) ListNetworkEndpoints(name, zone string, showHealthStatus bool) ([]*computealpha.NetworkEndpointWithHealthStatus, error) {
	cloud.mu.Lock()
	defer cloud.mu.Unlock()
	ret := []*computealpha.NetworkEndpointWithHealthStatus{}
	nes, ok := cloud.NetworkEndpoints[networkEndpointKey(name, zone)]
	if !ok {
		return nil, NotFoundError
	}
	for _, ne := range nes {
		ret = append(ret, &computealpha.NetworkEndpointWithHealthStatus{NetworkEndpoint: ne})
	}
	return ret, nil
}

func (cloud *FakeNetworkEndpointGroupCloud) NetworkURL() string {
	return cloud.Network
}

func (cloud *FakeNetworkEndpointGroupCloud) SubnetworkURL() string {
	return cloud.Subnetwork
}
