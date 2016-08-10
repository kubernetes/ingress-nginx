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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"k8s.io/kubernetes/pkg/util/sets"

	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/healthchecks"
	"k8s.io/contrib/ingress/controllers/gce/instances"
	"k8s.io/contrib/ingress/controllers/gce/storage"
	"k8s.io/contrib/ingress/controllers/gce/utils"
)

// Backends implements BackendPool.
type Backends struct {
	cloud         BackendServices
	nodePool      instances.NodePool
	healthChecker healthchecks.HealthChecker
	snapshotter   storage.Snapshotter
	// ignoredPorts are a set of ports excluded from GC, even
	// after the Ingress has been deleted. Note that invoking
	// a Delete() on these ports will still delete the backend.
	ignoredPorts sets.String
	namer        *utils.Namer
}

func portKey(port int64) string {
	return fmt.Sprintf("%d", port)
}

// NewBackendPool returns a new backend pool.
// - cloud: implements BackendServices and syncs backends with a cloud provider
// - healthChecker: is capable of producing health checks for backends.
// - nodePool: implements NodePool, used to create/delete new instance groups.
// - namer: procudes names for backends.
// - ignorePorts: is a set of ports to avoid syncing/GCing.
// - resyncWithCloud: if true, periodically syncs with cloud resources.
func NewBackendPool(
	cloud BackendServices,
	healthChecker healthchecks.HealthChecker,
	nodePool instances.NodePool,
	namer *utils.Namer,
	ignorePorts []int64,
	resyncWithCloud bool) *Backends {

	ignored := []string{}
	for _, p := range ignorePorts {
		ignored = append(ignored, portKey(p))
	}
	backendPool := &Backends{
		cloud:         cloud,
		nodePool:      nodePool,
		healthChecker: healthChecker,
		namer:         namer,
		ignoredPorts:  sets.NewString(ignored...),
	}
	if !resyncWithCloud {
		backendPool.snapshotter = storage.NewInMemoryPool()
		return backendPool
	}
	backendPool.snapshotter = storage.NewCloudListingPool(
		func(i interface{}) (string, error) {
			bs := i.(*compute.BackendService)
			if !namer.NameBelongsToCluster(bs.Name) {
				return "", fmt.Errorf("Unrecognized name %v", bs.Name)
			}
			port, err := namer.BePort(bs.Name)
			if err != nil {
				return "", err
			}
			return port, nil
		},
		backendPool,
		30*time.Second,
	)
	return backendPool
}

// Get returns a single backend.
func (b *Backends) Get(port int64) (*compute.BackendService, error) {
	be, err := b.cloud.GetBackendService(b.namer.BeName(port))
	if err != nil {
		return nil, err
	}
	b.snapshotter.Add(portKey(port), be)
	return be, nil
}

func (b *Backends) create(igs []*compute.InstanceGroup, namedPort *compute.NamedPort, name string) (*compute.BackendService, error) {
	// Create a new health check
	if err := b.healthChecker.Add(namedPort.Port); err != nil {
		return nil, err
	}
	hc, err := b.healthChecker.Get(namedPort.Port)
	if err != nil {
		return nil, err
	}
	// Create a new backend
	backend := &compute.BackendService{
		Name:     name,
		Protocol: "HTTP",
		Backends: getBackendsForIGs(igs),
		// Api expects one, means little to kubernetes.
		HealthChecks: []string{hc.SelfLink},
		Port:         namedPort.Port,
		PortName:     namedPort.Name,
	}
	if err := b.cloud.CreateBackendService(backend); err != nil {
		return nil, err
	}
	return b.Get(namedPort.Port)
}

// Add will get or create a Backend for the given port.
func (b *Backends) Add(port int64) error {
	// We must track the port even if creating the backend failed, because
	// we might've created a health-check for it.
	be := &compute.BackendService{}
	defer func() { b.snapshotter.Add(portKey(port), be) }()

	igs, namedPort, err := b.nodePool.AddInstanceGroup(b.namer.IGName(), port)
	if err != nil {
		return err
	}
	be, _ = b.Get(port)
	if be == nil {
		glog.Infof("Creating backend for %d instance groups, port %v named port %v",
			len(igs), port, namedPort)
		be, err = b.create(igs, namedPort, b.namer.BeName(port))
		if err != nil {
			return err
		}
	}
	// we won't find any igs till the node pool syncs nodes.
	if len(igs) == 0 {
		return nil
	}
	if err := b.edgeHop(be, igs); err != nil {
		return err
	}
	return err
}

// Delete deletes the Backend for the given port.
func (b *Backends) Delete(port int64) (err error) {
	name := b.namer.BeName(port)
	glog.Infof("Deleting backend %v", name)
	defer func() {
		if utils.IsHTTPErrorCode(err, http.StatusNotFound) {
			err = nil
		}
		if err == nil {
			b.snapshotter.Delete(portKey(port))
		}
	}()
	// Try deleting health checks even if a backend is not found.
	if err = b.cloud.DeleteBackendService(name); err != nil &&
		!utils.IsHTTPErrorCode(err, http.StatusNotFound) {
		return err
	}
	if err = b.healthChecker.Delete(port); err != nil &&
		!utils.IsHTTPErrorCode(err, http.StatusNotFound) {
		return err
	}
	return nil
}

// List lists all backends.
func (b *Backends) List() ([]interface{}, error) {
	// TODO: for consistency with the rest of this sub-package this method
	// should return a list of backend ports.
	interList := []interface{}{}
	be, err := b.cloud.ListBackendServices()
	if err != nil {
		return interList, err
	}
	for i := range be.Items {
		interList = append(interList, be.Items[i])
	}
	return interList, nil
}

func getBackendsForIGs(igs []*compute.InstanceGroup) []*compute.Backend {
	backends := []*compute.Backend{}
	for _, ig := range igs {
		backends = append(backends, &compute.Backend{Group: ig.SelfLink})
	}
	return backends
}

// edgeHop checks the links of the given backend by executing an edge hop.
// It fixes broken links.
func (b *Backends) edgeHop(be *compute.BackendService, igs []*compute.InstanceGroup) error {
	beIGs := sets.String{}
	for _, beToIG := range be.Backends {
		beIGs.Insert(beToIG.Group)
	}
	igLinks := sets.String{}
	for _, igToBE := range igs {
		igLinks.Insert(igToBE.SelfLink)
	}
	if beIGs.IsSuperset(igLinks) {
		return nil
	}
	glog.Infof("Backend %v has a broken edge, expected igs %+v, current igs %+v",
		be.Name, igLinks.List(), beIGs.List())

	newBackends := []*compute.Backend{}
	for _, b := range getBackendsForIGs(igs) {
		if !beIGs.Has(b.Group) {
			newBackends = append(newBackends, b)
		}
	}
	be.Backends = append(be.Backends, newBackends...)
	if err := b.cloud.UpdateBackendService(be); err != nil {
		return err
	}
	return nil
}

// Sync syncs backend services corresponding to ports in the given list.
func (b *Backends) Sync(svcNodePorts []int64) error {
	glog.V(3).Infof("Sync: backends %v", svcNodePorts)

	// create backends for new ports, perform an edge hop for existing ports
	for _, port := range svcNodePorts {
		if err := b.Add(port); err != nil {
			return err
		}
	}
	return nil
}

// GC garbage collects services corresponding to ports in the given list.
func (b *Backends) GC(svcNodePorts []int64) error {
	knownPorts := sets.NewString()
	for _, port := range svcNodePorts {
		knownPorts.Insert(portKey(port))
	}
	pool := b.snapshotter.Snapshot()
	for port := range pool {
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		nodePort := int64(p)
		if knownPorts.Has(portKey(nodePort)) || b.ignoredPorts.Has(portKey(nodePort)) {
			continue
		}
		glog.V(3).Infof("GCing backend for port %v", p)
		if err := b.Delete(nodePort); err != nil {
			return err
		}
	}
	if len(svcNodePorts) == 0 {
		glog.Infof("Deleting instance group %v", b.namer.IGName())
		if err := b.nodePool.DeleteInstanceGroup(b.namer.IGName()); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown deletes all backends and the default backend.
// This will fail if one of the backends is being used by another resource.
func (b *Backends) Shutdown() error {
	if err := b.GC([]int64{}); err != nil {
		return err
	}
	return nil
}

// Status returns the status of the given backend by name.
func (b *Backends) Status(name string) string {
	backend, err := b.cloud.GetBackendService(name)
	if err != nil {
		return "Unknown"
	}
	// TODO: Include port, ip in the status, since it's in the health info.
	hs, err := b.cloud.GetHealth(name, backend.Backends[0].Group)
	if err != nil || len(hs.HealthStatus) == 0 || hs.HealthStatus[0] == nil {
		return "Unknown"
	}
	// TODO: State transition are important, not just the latest.
	return hs.HealthStatus[0].HealthState
}
