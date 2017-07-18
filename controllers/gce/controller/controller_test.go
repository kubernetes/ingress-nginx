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
	"fmt"
	"math/rand"
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"

	api_v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/api"

	"k8s.io/ingress/controllers/gce/firewalls"
	"k8s.io/ingress/controllers/gce/loadbalancers"
	"k8s.io/ingress/controllers/gce/utils"
)

const testClusterName = "testcluster"

var (
	testPathMap   = map[string]string{"/foo": defaultBackendName(testClusterName)}
	testIPManager = testIP{}
)

// TODO: Use utils.Namer instead of this function.
func defaultBackendName(clusterName string) string {
	return fmt.Sprintf("%v-%v", backendPrefix, clusterName)
}

// newLoadBalancerController create a loadbalancer controller.
func newLoadBalancerController(t *testing.T, cm *fakeClusterManager) *LoadBalancerController {
	kubeClient := fake.NewSimpleClientset()
	lb, err := NewLoadBalancerController(kubeClient, cm.ClusterManager, 1*time.Second, api_v1.NamespaceAll)
	if err != nil {
		t.Fatalf("%v", err)
	}
	lb.hasSynced = func() bool { return true }
	return lb
}

// toHTTPIngressPaths converts the given pathMap to a list of HTTPIngressPaths.
func toHTTPIngressPaths(pathMap map[string]string) []extensions.HTTPIngressPath {
	httpPaths := []extensions.HTTPIngressPath{}
	for path, backend := range pathMap {
		httpPaths = append(httpPaths, extensions.HTTPIngressPath{
			Path: path,
			Backend: extensions.IngressBackend{
				ServiceName: backend,
				ServicePort: testBackendPort,
			},
		})
	}
	return httpPaths
}

// toIngressRules converts the given ingressRule map to a list of IngressRules.
func toIngressRules(hostRules map[string]utils.FakeIngressRuleValueMap) []extensions.IngressRule {
	rules := []extensions.IngressRule{}
	for host, pathMap := range hostRules {
		rules = append(rules, extensions.IngressRule{
			Host: host,
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: toHTTPIngressPaths(pathMap),
				},
			},
		})
	}
	return rules
}

// newIngress returns a new Ingress with the given path map.
func newIngress(hostRules map[string]utils.FakeIngressRuleValueMap) *extensions.Ingress {
	return &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      fmt.Sprintf("%v", uuid.NewUUID()),
			Namespace: api.NamespaceNone,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: defaultBackendName(testClusterName),
				ServicePort: testBackendPort,
			},
			Rules: toIngressRules(hostRules),
		},
		Status: extensions.IngressStatus{
			LoadBalancer: api_v1.LoadBalancerStatus{
				Ingress: []api_v1.LoadBalancerIngress{
					{IP: testIPManager.ip()},
				},
			},
		},
	}
}

// validIngress returns a valid Ingress.
func validIngress() *extensions.Ingress {
	return newIngress(map[string]utils.FakeIngressRuleValueMap{
		"foo.bar.com": testPathMap,
	})
}

// getKey returns the key for an ingress.
func getKey(ing *extensions.Ingress, t *testing.T) string {
	key, err := keyFunc(ing)
	if err != nil {
		t.Fatalf("Unexpected error getting key for Ingress %v: %v", ing.Name, err)
	}
	return key
}

// nodePortManager is a helper to allocate ports to services and
// remember the allocations.
type nodePortManager struct {
	portMap map[string]int
	start   int
	end     int
	namer   utils.Namer
}

// randPort generated pseudo random port numbers.
func (p *nodePortManager) getNodePort(svcName string) int {
	if port, ok := p.portMap[svcName]; ok {
		return port
	}
	p.portMap[svcName] = rand.Intn(p.end-p.start) + p.start
	return p.portMap[svcName]
}

// toNodePortSvcNames converts all service names in the given map to gce node
// port names, eg foo -> k8-be-<foo nodeport>
func (p *nodePortManager) toNodePortSvcNames(inputMap map[string]utils.FakeIngressRuleValueMap) map[string]utils.FakeIngressRuleValueMap {
	expectedMap := map[string]utils.FakeIngressRuleValueMap{}
	for host, rules := range inputMap {
		ruleMap := utils.FakeIngressRuleValueMap{}
		for path, svc := range rules {
			ruleMap[path] = p.namer.BeName(int64(p.portMap[svc]))
		}
		expectedMap[host] = ruleMap
	}
	return expectedMap
}

func newPortManager(st, end int) *nodePortManager {
	return &nodePortManager{map[string]int{}, st, end, utils.Namer{}}
}

// addIngress adds an ingress to the loadbalancer controllers ingress store. If
// a nodePortManager is supplied, it also adds all backends to the service store
// with a nodePort acquired through it.
func addIngress(lbc *LoadBalancerController, ing *extensions.Ingress, pm *nodePortManager) {
	lbc.ingLister.Store.Add(ing)
	if pm == nil {
		return
	}
	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			svc := &api_v1.Service{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      path.Backend.ServiceName,
					Namespace: ing.Namespace,
				},
			}
			var svcPort api_v1.ServicePort
			switch path.Backend.ServicePort.Type {
			case intstr.Int:
				svcPort = api_v1.ServicePort{Port: path.Backend.ServicePort.IntVal}
			default:
				svcPort = api_v1.ServicePort{Name: path.Backend.ServicePort.StrVal}
			}
			svcPort.NodePort = int32(pm.getNodePort(path.Backend.ServiceName))
			svc.Spec.Ports = []api_v1.ServicePort{svcPort}
			lbc.svcLister.Indexer.Add(svc)
		}
	}
}

func TestLbCreateDelete(t *testing.T) {
	testFirewallName := "quux"
	cm := NewFakeClusterManager(DefaultClusterUID, testFirewallName)
	lbc := newLoadBalancerController(t, cm)
	inputMap1 := map[string]utils.FakeIngressRuleValueMap{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
		"bar.example.com": {
			"/bar1": "bar1svc",
			"/bar2": "bar2svc",
		},
	}
	inputMap2 := map[string]utils.FakeIngressRuleValueMap{
		"baz.foobar.com": {
			"/foo": "foo1svc",
			"/bar": "bar1svc",
		},
	}
	pm := newPortManager(1, 65536)
	ings := []*extensions.Ingress{}
	for _, m := range []map[string]utils.FakeIngressRuleValueMap{inputMap1, inputMap2} {
		newIng := newIngress(m)
		addIngress(lbc, newIng, pm)
		ingStoreKey := getKey(newIng, t)
		lbc.sync(ingStoreKey)
		l7, err := cm.l7Pool.Get(ingStoreKey)
		if err != nil {
			t.Fatalf("%v", err)
		}
		cm.fakeLbs.CheckURLMap(t, l7, pm.toNodePortSvcNames(m))
		ings = append(ings, newIng)
	}
	lbc.ingLister.Store.Delete(ings[0])
	lbc.sync(getKey(ings[0], t))

	// BackendServices associated with ports of deleted Ingress' should get gc'd
	// when the Ingress is deleted, regardless of the service. At the same time
	// we shouldn't pull shared backends out from existing loadbalancers.
	unexpected := []int{pm.portMap["foo2svc"], pm.portMap["bar2svc"]}
	expected := []int{pm.portMap["foo1svc"], pm.portMap["bar1svc"]}
	firewallPorts := sets.NewString()
	pm.namer.SetFirewallName(testFirewallName)
	firewallName := pm.namer.FrName(pm.namer.FrSuffix())

	if firewallRule, err := cm.firewallPool.(*firewalls.FirewallRules).GetFirewall(firewallName); err != nil {
		t.Fatalf("%v", err)
	} else {
		if len(firewallRule.Allowed) != 1 {
			t.Fatalf("Expected a single firewall rule")
		}
		for _, p := range firewallRule.Allowed[0].Ports {
			firewallPorts.Insert(p)
		}
	}

	for _, port := range expected {
		if _, err := cm.backendPool.Get(int64(port)); err != nil {
			t.Fatalf("%v", err)
		}
		if !firewallPorts.Has(fmt.Sprintf("%v", port)) {
			t.Fatalf("Expected a firewall rule for port %v", port)
		}
	}
	for _, port := range unexpected {
		if be, err := cm.backendPool.Get(int64(port)); err == nil {
			t.Fatalf("Found backend %+v for port %v", be, port)
		}
	}
	lbc.ingLister.Store.Delete(ings[1])
	lbc.sync(getKey(ings[1], t))

	// No cluster resources (except the defaults used by the cluster manager)
	// should exist at this point.
	for _, port := range expected {
		if be, err := cm.backendPool.Get(int64(port)); err == nil {
			t.Fatalf("Found backend %+v for port %v", be, port)
		}
	}
	if len(cm.fakeLbs.Fw) != 0 || len(cm.fakeLbs.Um) != 0 || len(cm.fakeLbs.Tp) != 0 {
		t.Fatalf("Loadbalancer leaked resources")
	}
	for _, lbName := range []string{getKey(ings[0], t), getKey(ings[1], t)} {
		if l7, err := cm.l7Pool.Get(lbName); err == nil {
			t.Fatalf("Found unexpected loadbalandcer %+v: %v", l7, err)
		}
	}
	if firewallRule, err := cm.firewallPool.(*firewalls.FirewallRules).GetFirewall(firewallName); err == nil {
		t.Fatalf("Found unexpected firewall rule %v", firewallRule)
	}
}

func TestLbFaultyUpdate(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	inputMap := map[string]utils.FakeIngressRuleValueMap{
		"foo.example.com": {
			"/foo1": "foo1svc",
			"/foo2": "foo2svc",
		},
		"bar.example.com": {
			"/bar1": "bar1svc",
			"/bar2": "bar2svc",
		},
	}
	ing := newIngress(inputMap)
	pm := newPortManager(1, 65536)
	addIngress(lbc, ing, pm)

	ingStoreKey := getKey(ing, t)
	lbc.sync(ingStoreKey)
	l7, err := cm.l7Pool.Get(ingStoreKey)
	if err != nil {
		t.Fatalf("%v", err)
	}
	cm.fakeLbs.CheckURLMap(t, l7, pm.toNodePortSvcNames(inputMap))

	// Change the urlmap directly through the lb pool, resync, and
	// make sure the controller corrects it.
	l7.UpdateUrlMap(utils.GCEURLMap{
		"foo.example.com": {
			"/foo1": &compute.BackendService{SelfLink: "foo2svc"},
		},
	})

	lbc.sync(ingStoreKey)
	cm.fakeLbs.CheckURLMap(t, l7, pm.toNodePortSvcNames(inputMap))
}

func TestLbDefaulting(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	// Make sure the controller plugs in the default values accepted by GCE.
	ing := newIngress(map[string]utils.FakeIngressRuleValueMap{"": {"": "foo1svc"}})
	pm := newPortManager(1, 65536)
	addIngress(lbc, ing, pm)

	ingStoreKey := getKey(ing, t)
	lbc.sync(ingStoreKey)
	l7, err := cm.l7Pool.Get(ingStoreKey)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expectedMap := map[string]utils.FakeIngressRuleValueMap{loadbalancers.DefaultHost: {loadbalancers.DefaultPath: "foo1svc"}}
	cm.fakeLbs.CheckURLMap(t, l7, pm.toNodePortSvcNames(expectedMap))
}

func TestLbNoService(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	inputMap := map[string]utils.FakeIngressRuleValueMap{
		"foo.example.com": {
			"/foo1": "foo1svc",
		},
	}
	ing := newIngress(inputMap)
	ing.Spec.Backend.ServiceName = "foo1svc"
	ingStoreKey := getKey(ing, t)

	// Adds ingress to store, but doesn't create an associated service.
	// This will still create the associated loadbalancer, it will just
	// have empty rules. The rules will get corrected when the service
	// pops up.
	addIngress(lbc, ing, nil)
	lbc.sync(ingStoreKey)

	l7, err := cm.l7Pool.Get(ingStoreKey)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Creates the service, next sync should have complete url map.
	pm := newPortManager(1, 65536)
	addIngress(lbc, ing, pm)
	lbc.enqueueIngressForService(&api_v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo1svc",
			Namespace: ing.Namespace,
		},
	})
	// TODO: This will hang if the previous step failed to insert into queue
	key, _ := lbc.ingQueue.queue.Get()
	lbc.sync(key.(string))

	inputMap[utils.DefaultBackendKey] = map[string]string{
		utils.DefaultBackendKey: "foo1svc",
	}
	expectedMap := pm.toNodePortSvcNames(inputMap)
	cm.fakeLbs.CheckURLMap(t, l7, expectedMap)
}

func TestLbChangeStaticIP(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	inputMap := map[string]utils.FakeIngressRuleValueMap{
		"foo.example.com": {
			"/foo1": "foo1svc",
		},
	}
	ing := newIngress(inputMap)
	ing.Spec.Backend.ServiceName = "foo1svc"
	cert := extensions.IngressTLS{SecretName: "foo"}
	ing.Spec.TLS = []extensions.IngressTLS{cert}

	// Add some certs so we get 2 forwarding rules, the changed static IP
	// should be assigned to both the HTTP and HTTPS forwarding rules.
	lbc.tlsLoader = &fakeTLSSecretLoader{
		fakeCerts: map[string]*loadbalancers.TLSCerts{
			cert.SecretName: {Key: "foo", Cert: "bar"},
		},
	}

	pm := newPortManager(1, 65536)
	addIngress(lbc, ing, pm)
	ingStoreKey := getKey(ing, t)

	// First sync creates forwarding rules and allocates an IP.
	lbc.sync(ingStoreKey)

	// First allocate a static ip, then specify a userip in annotations.
	// The forwarding rules should contain the user ip.
	// The static ip should get cleaned up on lb tear down.
	oldIP := ing.Status.LoadBalancer.Ingress[0].IP
	oldRules := cm.fakeLbs.GetForwardingRulesWithIPs([]string{oldIP})
	if len(oldRules) != 2 || oldRules[0].IPAddress != oldRules[1].IPAddress {
		t.Fatalf("Expected 2 forwarding rules with the same IP.")
	}

	ing.Annotations = map[string]string{staticIPNameKey: "testip"}
	cm.fakeLbs.ReserveGlobalAddress(&compute.Address{Name: "testip", Address: "1.2.3.4"})

	// Second sync reassigns 1.2.3.4 to existing forwarding rule (by recreating it)
	lbc.sync(ingStoreKey)

	newRules := cm.fakeLbs.GetForwardingRulesWithIPs([]string{"1.2.3.4"})
	if len(newRules) != 2 || newRules[0].IPAddress != newRules[1].IPAddress || newRules[1].IPAddress != "1.2.3.4" {
		t.Fatalf("Found unexpected forwaring rules after changing static IP annotation.")
	}
}

type testIP struct {
	start int
}

func (t *testIP) ip() string {
	t.start++
	return fmt.Sprintf("0.0.0.%v", t.start)
}

// TODO: Test lb status update when annotation stabilize
