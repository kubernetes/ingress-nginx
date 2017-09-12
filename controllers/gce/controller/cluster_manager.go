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

package controller

import (
	"net/http"

	"github.com/golang/glog"

	compute "google.golang.org/api/compute/v1"
	gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"

	"k8s.io/ingress/controllers/gce/backends"
	"k8s.io/ingress/controllers/gce/firewalls"
	"k8s.io/ingress/controllers/gce/healthchecks"
	"k8s.io/ingress/controllers/gce/instances"
	"k8s.io/ingress/controllers/gce/loadbalancers"
	"k8s.io/ingress/controllers/gce/utils"
)

const (
	defaultPort            = 80
	defaultHealthCheckPath = "/"

	// A backend is created per nodePort, tagged with the nodeport.
	// This allows sharing of backends across loadbalancers.
	backendPrefix = "k8s-be"

	// A single target proxy/urlmap/forwarding rule is created per loadbalancer.
	// Tagged with the namespace/name of the Ingress.
	targetProxyPrefix    = "k8s-tp"
	forwardingRulePrefix = "k8s-fw"
	urlMapPrefix         = "k8s-um"

	// Used in the test RunServer method to denote a delete request.
	deleteType = "del"

	// port 0 is used as a signal for port not found/no such port etc.
	invalidPort = 0

	// Names longer than this are truncated, because of GCE restrictions.
	nameLenLimit = 62
)

// ClusterManager manages cluster resource pools.
type ClusterManager struct {
	ClusterNamer           *utils.Namer
	defaultBackendNodePort backends.ServicePort
	instancePool           instances.NodePool
	backendPool            backends.BackendPool
	l7Pool                 loadbalancers.LoadBalancerPool
	firewallPool           firewalls.SingleFirewallPool

	// TODO: Refactor so we simply init a health check pool.
	// Currently health checks are tied to backends because each backend needs
	// the link of the associated health, but both the backend pool and
	// loadbalancer pool manage backends, because the lifetime of the default
	// backend is tied to the last/first loadbalancer not the life of the
	// nodeport service or Ingress.
	healthCheckers []healthchecks.HealthChecker
}

// Init initializes the cluster manager.
func (c *ClusterManager) Init(tr *GCETranslator) {
	c.instancePool.Init(tr)
	c.backendPool.Init(tr)
	// TODO: Initialize other members as needed.
}

// IsHealthy returns an error if the cluster manager is unhealthy.
func (c *ClusterManager) IsHealthy() (err error) {
	// TODO: Expand on this, for now we just want to detect when the GCE client
	// is broken.
	_, err = c.backendPool.List()

	// If this container is scheduled on a node without compute/rw it is
	// effectively useless, but it is healthy. Reporting it as unhealthy
	// will lead to container crashlooping.
	if utils.IsHTTPErrorCode(err, http.StatusForbidden) {
		glog.Infof("Reporting cluster as healthy, but unable to list backends: %v", err)
		return nil
	}
	return
}

func (c *ClusterManager) shutdown() error {
	if err := c.l7Pool.Shutdown(); err != nil {
		return err
	}
	if err := c.firewallPool.Shutdown(); err != nil {
		return err
	}
	// The backend pool will also delete instance groups.
	return c.backendPool.Shutdown()
}

// Checkpoint performs a checkpoint with the cloud.
// - lbs are the single cluster L7 loadbalancers we wish to exist. If they already
//   exist, they should not have any broken links between say, a UrlMap and
//   TargetHttpProxy.
// - nodeNames are the names of nodes we wish to add to all loadbalancer
//   instance groups.
// - backendServicePorts are the ports for which we require BackendServices.
// - namedPorts are the ports which must be opened on instance groups.
// Returns the list of all instance groups corresponding to the given loadbalancers.
// If in performing the checkpoint the cluster manager runs out of quota, a
// googleapi 403 is returned.
func (c *ClusterManager) Checkpoint(lbs []*loadbalancers.L7RuntimeInfo, nodeNames []string, backendServicePorts []backends.ServicePort, namedPorts []backends.ServicePort) ([]*compute.InstanceGroup, error) {
	if len(namedPorts) != 0 {
		// Add the default backend node port to the list of named ports for instance groups.
		namedPorts = append(namedPorts, c.defaultBackendNodePort)
	}
	// Multiple ingress paths can point to the same service (and hence nodePort)
	// but each nodePort can only have one set of cloud resources behind it. So
	// don't waste time double validating GCE BackendServices.
	namedPorts = uniq(namedPorts)
	backendServicePorts = uniq(backendServicePorts)
	// Create Instance Groups.
	igs, err := c.EnsureInstanceGroupsAndPorts(namedPorts)
	if err != nil {
		return igs, err
	}
	if err := c.backendPool.Sync(backendServicePorts, igs); err != nil {
		return igs, err
	}
	if err := c.instancePool.Sync(nodeNames); err != nil {
		return igs, err
	}
	if err := c.l7Pool.Sync(lbs); err != nil {
		return igs, err
	}

	// TODO: Manage default backend and its firewall rule in a centralized way.
	// DefaultBackend is managed in l7 pool, which doesn't understand instances,
	// which the firewall rule requires.
	fwNodePorts := backendServicePorts
	if len(lbs) != 0 {
		// If there are no Ingresses, we shouldn't be allowing traffic to the
		// default backend. Equally importantly if the cluster gets torn down
		// we shouldn't leak the firewall rule.
		fwNodePorts = append(fwNodePorts, c.defaultBackendNodePort)
	}

	var np []int64
	for _, p := range fwNodePorts {
		np = append(np, p.Port)
	}
	if err := c.firewallPool.Sync(np, nodeNames); err != nil {
		return igs, err
	}

	return igs, nil
}

func (c *ClusterManager) EnsureInstanceGroupsAndPorts(servicePorts []backends.ServicePort) ([]*compute.InstanceGroup, error) {
	var igs []*compute.InstanceGroup
	var err error
	for _, p := range servicePorts {
		// EnsureInstanceGroupsAndPorts always returns all the instance groups, so we can return
		// the output of any call, no need to append the return from all calls.
		// TODO: Ideally, we want to call CreateInstaceGroups only the first time and
		// then call AddNamedPort multiple times. Need to update the interface to
		// achieve this.
		igs, _, err = instances.EnsureInstanceGroupsAndPorts(c.instancePool, c.ClusterNamer, p.Port)
		if err != nil {
			return nil, err
		}
	}
	return igs, nil
}

// GC garbage collects unused resources.
// - lbNames are the names of L7 loadbalancers we wish to exist. Those not in
//   this list are removed from the cloud.
// - nodePorts are the ports for which we want BackendServies. BackendServices
//   for ports not in this list are deleted.
// This method ignores googleapi 404 errors (StatusNotFound).
func (c *ClusterManager) GC(lbNames []string, nodePorts []backends.ServicePort) error {

	// On GC:
	// * Loadbalancers need to get deleted before backends.
	// * Backends are refcounted in a shared pool.
	// * We always want to GC backends even if there was an error in GCing
	//   loadbalancers, because the next Sync could rely on the GC for quota.
	// * There are at least 2 cases for backend GC:
	//   1. The loadbalancer has been deleted.
	//   2. An update to the url map drops the refcount of a backend. This can
	//      happen when an Ingress is updated, if we don't GC after the update
	//      we'll leak the backend.

	lbErr := c.l7Pool.GC(lbNames)
	beErr := c.backendPool.GC(nodePorts)
	if lbErr != nil {
		return lbErr
	}
	if beErr != nil {
		return beErr
	}

	// TODO(ingress#120): Move this to the backend pool so it mirrors creation
	var igErr error
	if len(lbNames) == 0 {
		igName := c.ClusterNamer.IGName()
		glog.Infof("Deleting instance group %v", igName)
		igErr = c.instancePool.DeleteInstanceGroup(igName)
	}
	if igErr != nil {
		return igErr
	}
	return nil
}

// NewClusterManager creates a cluster manager for shared resources.
// - namer: is the namer used to tag cluster wide shared resources.
// - defaultBackendNodePort: is the node port of glbc's default backend. This is
//	 the kubernetes Service that serves the 404 page if no urls match.
// - defaultHealthCheckPath: is the default path used for L7 health checks, eg: "/healthz".
func NewClusterManager(
	cloud *gce.GCECloud,
	namer *utils.Namer,
	defaultBackendNodePort backends.ServicePort,
	defaultHealthCheckPath string) (*ClusterManager, error) {

	// Names are fundamental to the cluster, the uid allocator makes sure names don't collide.
	cluster := ClusterManager{ClusterNamer: namer}

	// NodePool stores GCE vms that are in this Kubernetes cluster.
	cluster.instancePool = instances.NewNodePool(cloud)

	// BackendPool creates GCE BackendServices and associated health checks.
	healthChecker := healthchecks.NewHealthChecker(cloud, defaultHealthCheckPath, cluster.ClusterNamer)
	// Loadbalancer pool manages the default backend and its health check.
	defaultBackendHealthChecker := healthchecks.NewHealthChecker(cloud, "/healthz", cluster.ClusterNamer)

	cluster.healthCheckers = []healthchecks.HealthChecker{healthChecker, defaultBackendHealthChecker}

	// TODO: This needs to change to a consolidated management of the default backend.
	cluster.backendPool = backends.NewBackendPool(cloud, healthChecker, cluster.instancePool, cluster.ClusterNamer, []int64{defaultBackendNodePort.Port}, true)
	defaultBackendPool := backends.NewBackendPool(cloud, defaultBackendHealthChecker, cluster.instancePool, cluster.ClusterNamer, []int64{}, false)
	cluster.defaultBackendNodePort = defaultBackendNodePort

	// L7 pool creates targetHTTPProxy, ForwardingRules, UrlMaps, StaticIPs.
	cluster.l7Pool = loadbalancers.NewLoadBalancerPool(cloud, defaultBackendPool, defaultBackendNodePort, cluster.ClusterNamer)
	cluster.firewallPool = firewalls.NewFirewallPool(cloud, cluster.ClusterNamer)
	return &cluster, nil
}
