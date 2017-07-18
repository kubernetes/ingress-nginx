/*
Copyright 2016 The Kubernetes Authors.

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
	"testing"
	"time"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress/controllers/gce/backends"
	"k8s.io/ingress/controllers/gce/utils"
)

// Pods created in loops start from this time, for routines that
// sort on timestamp.
var firstPodCreationTime = time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC)

func TestZoneListing(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	zoneToNode := map[string][]string{
		"zone-1": {"n1"},
		"zone-2": {"n2"},
	}
	addNodes(lbc, zoneToNode)
	zones, err := lbc.tr.ListZones()
	if err != nil {
		t.Errorf("Failed to list zones: %v", err)
	}
	for expectedZone := range zoneToNode {
		found := false
		for _, gotZone := range zones {
			if gotZone == expectedZone {
				found = true
			}
		}
		if !found {
			t.Fatalf("Expected zones %v; Got zones %v", zoneToNode, zones)
		}
	}
}

func TestInstancesAddedToZones(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	zoneToNode := map[string][]string{
		"zone-1": {"n1", "n2"},
		"zone-2": {"n3"},
	}
	addNodes(lbc, zoneToNode)

	// Create 2 igs, one per zone.
	testIG := "test-ig"
	testPort := int64(3001)
	lbc.CloudClusterManager.instancePool.AddInstanceGroup(testIG, testPort)

	// node pool syncs kube-nodes, this will add them to both igs.
	lbc.CloudClusterManager.instancePool.Sync([]string{"n1", "n2", "n3"})
	gotZonesToNode := cm.fakeIGs.GetInstancesByZone()

	if cm.fakeIGs.Ports[0] != testPort {
		t.Errorf("Expected the same node port on all igs, got ports %+v", cm.fakeIGs.Ports)
	}

	for z, nodeNames := range zoneToNode {
		if ig, err := cm.fakeIGs.GetInstanceGroup(testIG, z); err != nil {
			t.Errorf("Failed to find ig %v in zone %v, found %+v: %v", testIG, z, ig, err)
		}
		expNodes := sets.NewString(nodeNames...)
		gotNodes := sets.NewString(gotZonesToNode[z]...)
		if !gotNodes.Equal(expNodes) {
			t.Errorf("Nodes not added to zones, expected %+v got %+v", expNodes, gotNodes)
		}
	}
}

func TestProbeGetter(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)

	nodePortToHealthCheck := map[backends.ServicePort]string{
		{Port: 3001, Protocol: utils.ProtocolHTTP}:  "/healthz",
		{Port: 3002, Protocol: utils.ProtocolHTTPS}: "/foo",
	}
	addPods(lbc, nodePortToHealthCheck, api_v1.NamespaceDefault)
	for p, exp := range nodePortToHealthCheck {
		got, err := lbc.tr.GetProbe(p)
		if err != nil || got == nil {
			t.Errorf("Failed to get probe for node port %v: %v", p, err)
		} else if getProbePath(got) != exp {
			t.Errorf("Wrong path for node port %v, got %v expected %v", p, getProbePath(got), exp)
		}
	}
}

func TestProbeGetterNamedPort(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)
	nodePortToHealthCheck := map[backends.ServicePort]string{
		{Port: 3001, Protocol: utils.ProtocolHTTP}: "/healthz",
	}
	addPods(lbc, nodePortToHealthCheck, api_v1.NamespaceDefault)
	for _, p := range lbc.podLister.Indexer.List() {
		pod := p.(*api_v1.Pod)
		pod.Spec.Containers[0].Ports[0].Name = "test"
		pod.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet.Port = intstr.IntOrString{Type: intstr.String, StrVal: "test"}
	}
	for p, exp := range nodePortToHealthCheck {
		got, err := lbc.tr.GetProbe(p)
		if err != nil || got == nil {
			t.Errorf("Failed to get probe for node port %v: %v", p, err)
		} else if getProbePath(got) != exp {
			t.Errorf("Wrong path for node port %v, got %v expected %v", p, getProbePath(got), exp)
		}
	}

}

func TestProbeGetterCrossNamespace(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID, DefaultFirewallName)
	lbc := newLoadBalancerController(t, cm)

	firstPod := &api_v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			// labels match those added by "addPods", but ns and health check
			// path is different. If this pod was created in the same ns, it
			// would become the health check.
			Labels:            map[string]string{"app-3001": "test"},
			Name:              fmt.Sprintf("test-pod-new-ns"),
			Namespace:         "new-ns",
			CreationTimestamp: meta_v1.NewTime(firstPodCreationTime.Add(-time.Duration(time.Hour))),
		},
		Spec: api_v1.PodSpec{
			Containers: []api_v1.Container{
				{
					Ports: []api_v1.ContainerPort{{ContainerPort: 80}},
					ReadinessProbe: &api_v1.Probe{
						Handler: api_v1.Handler{
							HTTPGet: &api_v1.HTTPGetAction{
								Scheme: api_v1.URISchemeHTTP,
								Path:   "/badpath",
								Port: intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	lbc.podLister.Indexer.Add(firstPod)
	nodePortToHealthCheck := map[backends.ServicePort]string{
		{Port: 3001, Protocol: utils.ProtocolHTTP}: "/healthz",
	}
	addPods(lbc, nodePortToHealthCheck, api_v1.NamespaceDefault)

	for p, exp := range nodePortToHealthCheck {
		got, err := lbc.tr.GetProbe(p)
		if err != nil || got == nil {
			t.Errorf("Failed to get probe for node port %v: %v", p, err)
		} else if getProbePath(got) != exp {
			t.Errorf("Wrong path for node port %v, got %v expected %v", p, getProbePath(got), exp)
		}
	}
}

func addPods(lbc *LoadBalancerController, nodePortToHealthCheck map[backends.ServicePort]string, ns string) {
	delay := time.Minute
	for np, u := range nodePortToHealthCheck {
		l := map[string]string{fmt.Sprintf("app-%d", np.Port): "test"}
		svc := &api_v1.Service{
			Spec: api_v1.ServiceSpec{
				Selector: l,
				Ports: []api_v1.ServicePort{
					{
						NodePort: int32(np.Port),
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
			},
		}
		svc.Name = fmt.Sprintf("%d", np.Port)
		svc.Namespace = ns
		lbc.svcLister.Indexer.Add(svc)

		pod := &api_v1.Pod{
			ObjectMeta: meta_v1.ObjectMeta{
				Labels:            l,
				Name:              fmt.Sprintf("%d", np.Port),
				Namespace:         ns,
				CreationTimestamp: meta_v1.NewTime(firstPodCreationTime.Add(delay)),
			},
			Spec: api_v1.PodSpec{
				Containers: []api_v1.Container{
					{
						Ports: []api_v1.ContainerPort{{Name: "test", ContainerPort: 80}},
						ReadinessProbe: &api_v1.Probe{
							Handler: api_v1.Handler{
								HTTPGet: &api_v1.HTTPGetAction{
									Scheme: api_v1.URIScheme(string(np.Protocol)),
									Path:   u,
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 80,
									},
								},
							},
						},
					},
				},
			},
		}
		lbc.podLister.Indexer.Add(pod)
		delay = 2 * delay
	}
}

func addNodes(lbc *LoadBalancerController, zoneToNode map[string][]string) {
	for zone, nodes := range zoneToNode {
		for _, node := range nodes {
			n := &api_v1.Node{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: node,
					Labels: map[string]string{
						zoneKey: zone,
					},
				},
				Status: api_v1.NodeStatus{
					Conditions: []api_v1.NodeCondition{
						{Type: api_v1.NodeReady, Status: api_v1.ConditionTrue},
					},
				},
			}
			lbc.nodeLister.Indexer.Add(n)
		}
	}
	lbc.CloudClusterManager.instancePool.Init(lbc.tr)
}

func getProbePath(p *api_v1.Probe) string {
	return p.Handler.HTTPGet.Path
}
