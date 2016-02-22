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

package controller

import (
	"fmt"

	"k8s.io/contrib/ingress/controllers/gce/backends"
	"k8s.io/contrib/ingress/controllers/gce/healthchecks"
	"k8s.io/contrib/ingress/controllers/gce/instances"
	"k8s.io/contrib/ingress/controllers/gce/loadbalancers"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/cloudprovider"
	gce "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
)

const (
	defaultPort            = 80
	defaultHealthCheckPath = "/"

	// A single instance-group is created per cluster manager.
	// Tagged with the name of the controller.
	instanceGroupPrefix = "k8s-ig"

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
	ClusterNamer           utils.Namer
	defaultBackendNodePort int64
	instancePool           instances.NodePool
	backendPool            backends.BackendPool
	l7Pool                 loadbalancers.LoadBalancerPool
}

// IsHealthy returns an error if the cluster manager is unhealthy.
func (c *ClusterManager) IsHealthy() (err error) {
	// TODO: Expand on this, for now we just want to detect when the GCE client
	// is broken.
	_, err = c.backendPool.List()
	return
}

func (c *ClusterManager) shutdown() error {
	if err := c.l7Pool.Shutdown(); err != nil {
		return err
	}
	// The backend pool will also delete instance groups.
	return c.backendPool.Shutdown()
}

// Checkpoint performs a checkpoint with the cloud.
// - lbNames are the names of L7 loadbalancers we wish to exist. If they already
//   exist, they should not have any broken links between say, a UrlMap and
//   TargetHttpProxy.
// - nodeNames are the names of nodes we wish to add to all loadbalancer
//   instance groups.
// - nodePorts are the ports for which we require BackendServices. Each of
//   these ports must also be opened on the corresponding Instance Group.
// If in performing the checkpoint the cluster manager runs out of quota, a
// googleapi 403 is returned.
func (c *ClusterManager) Checkpoint(lbs []*loadbalancers.L7RuntimeInfo, nodeNames []string, nodePorts []int64) error {
	if err := c.backendPool.Sync(nodePorts); err != nil {
		return err
	}
	if err := c.instancePool.Sync(nodeNames); err != nil {
		return err
	}
	if err := c.l7Pool.Sync(lbs); err != nil {
		return err
	}
	return nil
}

// GC garbage collects unused resources.
// - lbNames are the names of L7 loadbalancers we wish to exist. Those not in
//   this list are removed from the cloud.
// - nodePorts are the ports for which we want BackendServies. BackendServices
//   for ports not in this list are deleted.
// This method ignores googleapi 404 errors (StatusNotFound).
func (c *ClusterManager) GC(lbNames []string, nodePorts []int64) error {

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
	return nil
}

func defaultInstanceGroupName(clusterName string) string {
	return fmt.Sprintf("%v-%v", instanceGroupPrefix, clusterName)
}

// NewClusterManager creates a cluster manager for shared resources.
// - name: is the name used to tag cluster wide shared resources. This is the
//   string passed to glbc via --gce-cluster-name.
// - defaultBackendNodePort: is the node port of glbc's default backend. This is
//	 the kubernetes Service that serves the 404 page if no urls match.
// - defaultHealthCheckPath: is the default path used for L7 health checks, eg: "/healthz"
func NewClusterManager(
	name string,
	defaultBackendNodePort int64,
	defaultHealthCheckPath string) (*ClusterManager, error) {

	cloudInterface, err := cloudprovider.GetCloudProvider("gce", nil)
	if err != nil {
		return nil, err
	}
	cloud := cloudInterface.(*gce.GCECloud)
	cluster := ClusterManager{ClusterNamer: utils.Namer{name}}
	zone, err := cloud.GetZone()
	if err != nil {
		return nil, err
	}
	cluster.instancePool = instances.NewNodePool(cloud, zone.FailureDomain)
	healthChecker := healthchecks.NewHealthChecker(cloud, defaultHealthCheckPath, cluster.ClusterNamer)
	cluster.backendPool = backends.NewBackendPool(
		cloud, healthChecker, cluster.instancePool, cluster.ClusterNamer)
	defaultBackendHealthChecker := healthchecks.NewHealthChecker(cloud, "/healthz", cluster.ClusterNamer)
	defaultBackendPool := backends.NewBackendPool(
		cloud, defaultBackendHealthChecker, cluster.instancePool, cluster.ClusterNamer)
	cluster.defaultBackendNodePort = defaultBackendNodePort
	cluster.l7Pool = loadbalancers.NewLoadBalancerPool(
		cloud, defaultBackendPool, defaultBackendNodePort, cluster.ClusterNamer)
	return &cluster, nil
}
