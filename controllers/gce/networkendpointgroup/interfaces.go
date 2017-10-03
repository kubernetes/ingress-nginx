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
	computealpha "google.golang.org/api/compute/v0.alpha"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NetworkEndpointGroupCloud is an interface for managing gce network endpoint group.
type NetworkEndpointGroupCloud interface {
	GetNetworkEndpointGroup(name string, zone string) (*computealpha.NetworkEndpointGroup, error)
	ListNetworkEndpointGroup(zone string) ([]*computealpha.NetworkEndpointGroup, error)
	AggregatedListNetworkEndpointGroup() (map[string][]*computealpha.NetworkEndpointGroup, error)
	CreateNetworkEndpointGroup(neg *computealpha.NetworkEndpointGroup, zone string) error
	DeleteNetworkEndpointGroup(name string, zone string) error
	AttachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error
	DetachNetworkEndpoints(name, zone string, endpoints []*computealpha.NetworkEndpoint) error
	ListNetworkEndpoints(name, zone string, showHealthStatus bool) ([]*computealpha.NetworkEndpointWithHealthStatus, error)
	NetworkURL() string
	SubnetworkURL() string
}

// NetworkEndpointGroupNamer is an interface for generating network endpoint group name.
type NetworkEndpointGroupNamer interface {
	NEGName(namespace, name, port string) string
	NEGPrefix() string
}

// ZoneGetter is an interface for retrieve zone related information
type ZoneGetter interface {
	ListZones() ([]string, error)
	GetZoneForNode(name string) (string, error)
}

// Syncer is an interface to interact with syncer
type Syncer interface {
	Start() error
	Stop()
	Sync() bool
	IsStopped() bool
	IsShuttingDown() bool
}

// SyncerManager is an interface for controllers to manage Syncers
type SyncerManager interface {
	EnsureSyncer(namespace, name string, targetPorts sets.String) error
	StopSyncer(namespace, name string)
	Sync(namespace, name string)
	GC() error
	ShutDown()
}
