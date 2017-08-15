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
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/cloudprovider"
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

	// Sleep interval to retry cloud client creation.
	cloudClientRetryInterval = 10 * time.Second
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
// - lbNames are the names of L7 loadbalancers we wish to exist. If they already
//   exist, they should not have any broken links between say, a UrlMap and
//   TargetHttpProxy.
// - nodeNames are the names of nodes we wish to add to all loadbalancer
//   instance groups.
// - nodePorts are the ports for which we require BackendServices. Each of
//   these ports must also be opened on the corresponding Instance Group.
// If in performing the checkpoint the cluster manager runs out of quota, a
// googleapi 403 is returned.
func (c *ClusterManager) Checkpoint(lbs []*loadbalancers.L7RuntimeInfo, nodeNames []string, nodePorts []backends.ServicePort) error {
	// Multiple ingress paths can point to the same service (and hence nodePort)
	// but each nodePort can only have one set of cloud resources behind it. So
	// don't waste time double validating GCE BackendServices.
	portMap := map[int64]backends.ServicePort{}
	for _, p := range nodePorts {
		portMap[p.Port] = p
	}
	nodePorts = []backends.ServicePort{}
	for _, sp := range portMap {
		nodePorts = append(nodePorts, sp)
	}
	if err := c.backendPool.Sync(nodePorts); err != nil {
		return err
	}
	if err := c.instancePool.Sync(nodeNames); err != nil {
		return err
	}
	if err := c.l7Pool.Sync(lbs); err != nil {
		return err
	}

	// TODO: Manage default backend and its firewall rule in a centralized way.
	// DefaultBackend is managed in l7 pool, which doesn't understand instances,
	// which the firewall rule requires.
	fwNodePorts := nodePorts
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

func getGCEClient(config io.Reader) *gce.GCECloud {
	getConfigReader := func() io.Reader { return nil }

	if config != nil {
		allConfig, err := ioutil.ReadAll(config)
		if err != nil {
			glog.Fatalf("Error while reading entire config: %v", err)
		}
		glog.V(2).Infof("Using cloudprovider config file:\n%v ", string(allConfig))

		getConfigReader = func() io.Reader {
			return bytes.NewReader(allConfig)
		}
	} else {
		glog.V(2).Infoln("No cloudprovider config file provided. Continuing with default values.")
	}

	// Creating the cloud interface involves resolving the metadata server to get
	// an oauth token. If this fails, the token provider assumes it's not on GCE.
	// No errors are thrown. So we need to keep retrying till it works because
	// we know we're on GCE.
	for {
		cloudInterface, err := cloudprovider.GetCloudProvider("gce", getConfigReader())
		if err == nil {
			cloud := cloudInterface.(*gce.GCECloud)

			// If this controller is scheduled on a node without compute/rw
			// it won't be allowed to list backends. We can assume that the
			// user has no need for Ingress in this case. If they grant
			// permissions to the node they will have to restart the controller
			// manually to re-create the client.
			if _, err = cloud.ListGlobalBackendServices(); err == nil || utils.IsHTTPErrorCode(err, http.StatusForbidden) {
				return cloud
			}
			glog.Warningf("Failed to list backend services, retrying: %v", err)
		} else {
			glog.Warningf("Failed to retrieve cloud interface, retrying: %v", err)
		}
		time.Sleep(cloudClientRetryInterval)
	}
}

// NewClusterManager creates a cluster manager for shared resources.
// - namer: is the namer used to tag cluster wide shared resources.
// - defaultBackendNodePort: is the node port of glbc's default backend. This is
//	 the kubernetes Service that serves the 404 page if no urls match.
// - defaultHealthCheckPath: is the default path used for L7 health checks, eg: "/healthz".
func NewClusterManager(
	configFilePath string,
	namer *utils.Namer,
	defaultBackendNodePort backends.ServicePort,
	defaultHealthCheckPath string) (*ClusterManager, error) {

	// TODO: Make this more resilient. Currently we create the cloud client
	// and pass it through to all the pools. This makes unit testing easier.
	// However if the cloud client suddenly fails, we should try to re-create it
	// and continue.
	var cloud *gce.GCECloud
	if configFilePath != "" {
		glog.Infof("Reading config from path %v", configFilePath)
		config, err := os.Open(configFilePath)
		if err != nil {
			return nil, err
		}
		defer config.Close()
		cloud = getGCEClient(config)
		glog.Infof("Successfully loaded cloudprovider using config %q", configFilePath)
	} else {
		// While you might be tempted to refactor so we simply assing nil to the
		// config and only invoke getGCEClient once, that will not do the right
		// thing because a nil check against an interface isn't true in golang.
		cloud = getGCEClient(nil)
		glog.Infof("Created GCE client without a config file")
	}

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
