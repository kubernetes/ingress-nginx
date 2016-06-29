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
	cloud InstanceGroups
	// zones is a list of zones seeded by Kubernetes node zones.
	// TODO: we can figure this out.
	snapshotter storage.Snapshotter
	zoneLister
}

// NewNodePool creates a new node pool.
// - cloud: implements InstanceGroups, used to sync Kubernetes nodes with
//   members of the cloud InstanceGroup.
func NewNodePool(cloud InstanceGroups) NodePool {
	return &Instances{cloud, storage.NewInMemoryPool(), nil}
}

// Init initializes the instance pool. The given zoneLister is used to list
// all zones that require an instance group, and to lookup which zone a
// given Kubernetes node is in so we can add it to the right instance group.
func (i *Instances) Init(zl zoneLister) {
	i.zoneLister = zl
}

// AddInstanceGroup creates or gets an instance group if it doesn't exist
// and adds the given port to it. Returns a list of one instance group per zone,
// all of which have the exact same named port.
func (i *Instances) AddInstanceGroup(name string, port int64) ([]*compute.InstanceGroup, *compute.NamedPort, error) {
	igs := []*compute.InstanceGroup{}
	namedPort := &compute.NamedPort{}

	zones, err := i.ListZones()
	if err != nil {
		return igs, namedPort, err
	}

	for _, zone := range zones {
		ig, _ := i.Get(name, zone)
		var err error
		if ig == nil {
			glog.Infof("Creating instance group %v in zone %v", name, zone)
			ig, err = i.cloud.CreateInstanceGroup(name, zone)
			if err != nil {
				return nil, nil, err
			}
		} else {
			glog.V(3).Infof("Instance group %v already exists in zone %v, adding port %d to it", name, zone, port)
		}
		defer i.snapshotter.Add(name, struct{}{})
		namedPort, err = i.cloud.AddPortToInstanceGroup(ig, port)
		if err != nil {
			return nil, nil, err
		}
		igs = append(igs, ig)
	}
	return igs, namedPort, nil
}

// DeleteInstanceGroup deletes the given IG by name, from all zones.
func (i *Instances) DeleteInstanceGroup(name string) error {
	defer i.snapshotter.Delete(name)
	errs := []error{}

	zones, err := i.ListZones()
	if err != nil {
		return err
	}
	for _, zone := range zones {
		glog.Infof("Deleting instance group %v in zone %v", name, zone)
		if err := i.cloud.DeleteInstanceGroup(name, zone); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%v", errs)
}

// list lists all instances in all zones.
func (i *Instances) list(name string) (sets.String, error) {
	nodeNames := sets.NewString()
	zones, err := i.ListZones()
	if err != nil {
		return nodeNames, err
	}

	for _, zone := range zones {
		instances, err := i.cloud.ListInstancesInInstanceGroup(
			name, zone, allInstances)
		if err != nil {
			return nodeNames, err
		}
		for _, ins := range instances.Items {
			// TODO: If round trips weren't so slow one would be inclided
			// to GetInstance using this url and get the name.
			parts := strings.Split(ins.Instance, "/")
			nodeNames.Insert(parts[len(parts)-1])
		}
	}
	return nodeNames, nil
}

// Get returns the Instance Group by name.
func (i *Instances) Get(name, zone string) (*compute.InstanceGroup, error) {
	ig, err := i.cloud.GetInstanceGroup(name, zone)
	if err != nil {
		return nil, err
	}
	i.snapshotter.Add(name, struct{}{})
	return ig, nil
}

// splitNodesByZones takes a list of node names and returns a map of zone:node names.
// It figures out the zones by asking the zoneLister.
func (i *Instances) splitNodesByZone(names []string) map[string][]string {
	nodesByZone := map[string][]string{}
	for _, name := range names {
		zone, err := i.GetZoneForNode(name)
		if err != nil {
			glog.Errorf("Failed to get zones for %v: %v, skipping", name, err)
			continue
		}
		if _, ok := nodesByZone[zone]; !ok {
			nodesByZone[zone] = []string{}
		}
		nodesByZone[zone] = append(nodesByZone[zone], name)
	}
	return nodesByZone
}

// Add adds the given instances to the appropriately zoned Instance Group.
func (i *Instances) Add(groupName string, names []string) error {
	errs := []error{}
	for zone, nodeNames := range i.splitNodesByZone(names) {
		glog.V(1).Infof("Adding nodes %v to %v in zone %v", nodeNames, groupName, zone)
		if err := i.cloud.AddInstancesToInstanceGroup(groupName, zone, nodeNames); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%v", errs)
}

// Remove removes the given instances from the appropriately zoned Instance Group.
func (i *Instances) Remove(groupName string, names []string) error {
	errs := []error{}
	for zone, nodeNames := range i.splitNodesByZone(names) {
		glog.V(1).Infof("Adding nodes %v to %v in zone %v", nodeNames, groupName, zone)
		if err := i.cloud.RemoveInstancesFromInstanceGroup(groupName, zone, nodeNames); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%v", errs)
}

// Sync syncs kubernetes instances with the instances in the instance group.
func (i *Instances) Sync(nodes []string) (err error) {
	glog.V(4).Infof("Syncing nodes %v", nodes)

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
	for igName := range pool {
		gceNodes := sets.NewString()
		gceNodes, err = i.list(igName)
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
				igName, gceNodes.Difference(kubeNodes).List()); err != nil {
				return err
			}
		}

		if len(addNodes) != 0 {
			if err = i.Add(
				igName, kubeNodes.Difference(gceNodes).List()); err != nil {
				return err
			}
		}
	}
	return nil
}
