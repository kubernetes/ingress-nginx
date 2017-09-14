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
	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/ingress/controllers/gce/backends"
	"k8s.io/ingress/controllers/gce/firewalls"
	"k8s.io/ingress/controllers/gce/healthchecks"
	"k8s.io/ingress/controllers/gce/instances"
	"k8s.io/ingress/controllers/gce/loadbalancers"
	"k8s.io/ingress/controllers/gce/utils"
)

var (
	testDefaultBeNodePort = backends.ServicePort{Port: 3000, Protocol: utils.ProtocolHTTP}
	testBackendPort       = intstr.IntOrString{Type: intstr.Int, IntVal: 80}
)

// ClusterManager fake
type fakeClusterManager struct {
	*ClusterManager
	fakeLbs      *loadbalancers.FakeLoadBalancers
	fakeBackends *backends.FakeBackendServices
	fakeIGs      *instances.FakeInstanceGroups
}

// NewFakeClusterManager creates a new fake ClusterManager.
func NewFakeClusterManager(clusterName, firewallName string) *fakeClusterManager {
	fakeLbs := loadbalancers.NewFakeLoadBalancers(clusterName)
	fakeBackends := backends.NewFakeBackendServices(func(op int, be *compute.BackendService) error { return nil })
	fakeIGs := instances.NewFakeInstanceGroups(sets.NewString())
	fakeHCP := healthchecks.NewFakeHealthCheckProvider()
	namer := utils.NewNamer(clusterName, firewallName)

	nodePool := instances.NewNodePool(fakeIGs)
	nodePool.Init(&instances.FakeZoneLister{Zones: []string{"zone-a"}})

	healthChecker := healthchecks.NewHealthChecker(fakeHCP, "/", namer)

	backendPool := backends.NewBackendPool(
		fakeBackends,
		healthChecker, nodePool, namer, []int64{}, false)
	l7Pool := loadbalancers.NewLoadBalancerPool(
		fakeLbs,
		// TODO: change this
		backendPool,
		testDefaultBeNodePort,
		namer,
	)
	frPool := firewalls.NewFirewallPool(firewalls.NewFakeFirewallsProvider(), namer)
	cm := &ClusterManager{
		ClusterNamer: namer,
		instancePool: nodePool,
		backendPool:  backendPool,
		l7Pool:       l7Pool,
		firewallPool: frPool,
	}
	return &fakeClusterManager{cm, fakeLbs, fakeBackends, fakeIGs}
}
