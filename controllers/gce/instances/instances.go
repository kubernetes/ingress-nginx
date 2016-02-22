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
	"net/http"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/storage"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/golang/glog"
)

const (
	// State string required by gce library to list all instances.
	allInstances = "ALL"
)

// Instances implements NodePool.
type Instances struct {
	cloud       InstanceGroups
	zone        string
	snapshotter storage.Snapshotter
}

// NewNodePool creates a new node pool.
// - cloud: implements InstanceGroups, used to sync Kubernetes nodes with
//   members of the cloud InstanceGroup.
func NewNodePool(cloud InstanceGroups, zone string) NodePool {
	glog.V(3).Infof("NodePool is only aware of instances in zone %v", zone)
	return &Instances{cloud, zone, storage.NewInMemoryPool()}
}

// AddInstanceGroup creates or gets an instance group if it doesn't exist
// and adds the given port to it.
func (i *Instances) AddInstanceGroup(name string, port int64) (*compute.InstanceGroup, *compute.NamedPort, error) {
	ig, _ := i.Get(name)
	if ig == nil {
		glog.Infof("Creating instance group %v", name)
		var err error
		ig, err = i.cloud.CreateInstanceGroup(name, i.zone)
		if err != nil {
			return nil, nil, err
		}
	} else {
		glog.V(3).Infof("Instance group already exists %v", name)
	}
	defer i.snapshotter.Add(name, ig)
	namedPort, err := i.cloud.AddPortToInstanceGroup(ig, port)
	if err != nil {
		return nil, nil, err
	}

	return ig, namedPort, nil
}

// DeleteInstanceGroup deletes the given IG by name.
func (i *Instances) DeleteInstanceGroup(name string) error {
	defer i.snapshotter.Delete(name)
	return i.cloud.DeleteInstanceGroup(name, i.zone)
}

func (i *Instances) list(name string) (sets.String, error) {
	nodeNames := sets.NewString()
	instances, err := i.cloud.ListInstancesInInstanceGroup(
		name, i.zone, allInstances)
	if err != nil {
		return nodeNames, err
	}
	for _, ins := range instances.Items {
		// TODO: If round trips weren't so slow one would be inclided
		// to GetInstance using this url and get the name.
		parts := strings.Split(ins.Instance, "/")
		nodeNames.Insert(parts[len(parts)-1])
	}
	return nodeNames, nil
}

// Get returns the Instance Group by name.
func (i *Instances) Get(name string) (*compute.InstanceGroup, error) {
	ig, err := i.cloud.GetInstanceGroup(name, i.zone)
	if err != nil {
		return nil, err
	}
	i.snapshotter.Add(name, ig)
	return ig, nil
}

// Add adds the given instances to the Instance Group.
func (i *Instances) Add(groupName string, names []string) error {
	glog.V(3).Infof("Adding nodes %v to %v", names, groupName)
	return i.cloud.AddInstancesToInstanceGroup(groupName, i.zone, names)
}

// Remove removes the given instances from the Instance Group.
func (i *Instances) Remove(groupName string, names []string) error {
	glog.V(3).Infof("Removing nodes %v from %v", names, groupName)
	return i.cloud.RemoveInstancesFromInstanceGroup(groupName, i.zone, names)
}

// Sync syncs kubernetes instances with the instances in the instance group.
func (i *Instances) Sync(nodes []string) (err error) {
	glog.V(3).Infof("Syncing nodes %v", nodes)

	defer func() {
		// The node pool is only responsible for syncing nodes to instance
		// groups. It never creates/deletes, so if an instance groups is
		// not found there's nothing it can do about it anyway. Most cases
		// this will happen because the backend pool has deleted the instance
		// group, however if it happens because a user deletes the IG by mistake
		// we should just wait till the backend pool fixes it.
		if utils.IsHTTPErrorCode(err, http.StatusNotFound) {
			glog.Infof("Node pool encountered a 404, ignoring: %v", err)
			err = nil
		}
	}()

	pool := i.snapshotter.Snapshot()
	for name := range pool {
		gceNodes := sets.NewString()
		gceNodes, err = i.list(name)
		if err != nil {
			return err
		}
		kubeNodes := sets.NewString(nodes...)

		// A node deleted via kubernetes could still exist as a gce vm. We don't
		// want to route requests to it. Similarly, a node added to kubernetes
		// needs to get added to the instance group so we do route requests to it.

		removeNodes := gceNodes.Difference(kubeNodes).List()
		addNodes := kubeNodes.Difference(gceNodes).List()
		if len(removeNodes) != 0 {
			if err = i.Remove(
				name, gceNodes.Difference(kubeNodes).List()); err != nil {
				return err
			}
		}

		if len(addNodes) != 0 {
			if err = i.Add(
				name, kubeNodes.Difference(gceNodes).List()); err != nil {
				return err
			}
		}
	}
	return nil
}
