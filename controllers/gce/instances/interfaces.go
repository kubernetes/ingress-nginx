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
	compute "google.golang.org/api/compute/v1"
)

// NodePool is an interface to manage a pool of kubernetes nodes synced with vm instances in the cloud
// through the InstanceGroups interface.
type NodePool interface {
	AddInstanceGroup(name string, port int64) (*compute.InstanceGroup, *compute.NamedPort, error)
	DeleteInstanceGroup(name string) error

	// TODO: Refactor for modularity
	Add(groupName string, nodeNames []string) error
	Remove(groupName string, nodeNames []string) error
	Sync(nodeNames []string) error
	Get(name string) (*compute.InstanceGroup, error)
}

// InstanceGroups is an interface for managing gce instances groups, and the instances therein.
type InstanceGroups interface {
	GetInstanceGroup(name, zone string) (*compute.InstanceGroup, error)
	CreateInstanceGroup(name, zone string) (*compute.InstanceGroup, error)
	DeleteInstanceGroup(name, zone string) error

	// TODO: Refactor for modulatiry.
	ListInstancesInInstanceGroup(name, zone string, state string) (*compute.InstanceGroupsListInstances, error)
	AddInstancesToInstanceGroup(name, zone string, instanceNames []string) error
	RemoveInstancesFromInstanceGroup(name, zone string, instanceName []string) error
	AddPortToInstanceGroup(ig *compute.InstanceGroup, port int64) (*compute.NamedPort, error)
}
