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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/sets"
)

// Pods created in loops start from this time, for routines that
// sort on timestamp.
var firstPodCreationTime = time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC)

func TestZoneListing(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID)
	lbc := newLoadBalancerController(t, cm, "")
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
	cm := NewFakeClusterManager(DefaultClusterUID)
	lbc := newLoadBalancerController(t, cm, "")
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

	i := 0
	for z, nodeNames := range zoneToNode {
		if ig, err := cm.fakeIGs.GetInstanceGroup(testIG, z); err != nil {
			t.Errorf("Failed to find ig %v in zone %v, found %+v: %v", testIG, z, ig, err)
		}
		if cm.fakeIGs.Ports[i] != testPort {
			t.Errorf("Expected the same node port on all igs, got ports %+v", cm.fakeIGs.Ports)
		}
		expNodes := sets.NewString(nodeNames...)
		gotNodes := sets.NewString(gotZonesToNode[z]...)
		if !gotNodes.Equal(expNodes) {
			t.Errorf("Nodes not added to zones, expected %+v got %+v", expNodes, gotNodes)
		}
		i++
	}
}

func TestProbeGetter(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID)
	lbc := newLoadBalancerController(t, cm, "")
	nodePortToHealthCheck := map[int64]string{
		3001: "/healthz",
		3002: "/foo",
	}
	addPods(lbc, nodePortToHealthCheck, api.NamespaceDefault)
	for p, exp := range nodePortToHealthCheck {
		got, err := lbc.tr.HealthCheck(p)
		if err != nil {
			t.Errorf("Failed to get health check for node port %v: %v", p, err)
		} else if got.RequestPath != exp {
			t.Errorf("Wrong health check for node port %v, got %v expected %v", p, got.RequestPath, exp)
		}
	}
}

func TestProbeGetterCrossNamespace(t *testing.T) {
	cm := NewFakeClusterManager(DefaultClusterUID)
	lbc := newLoadBalancerController(t, cm, "")

	firstPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			// labels match those added by "addPods", but ns and health check
			// path is different. If this pod was created in the same ns, it
			// would become the health check.
			Labels:            map[string]string{"app-3001": "test"},
			Name:              fmt.Sprintf("test-pod-new-ns"),
			Namespace:         "new-ns",
			CreationTimestamp: unversioned.NewTime(firstPodCreationTime.Add(-time.Duration(time.Hour))),
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Ports: []api.ContainerPort{{ContainerPort: 80}},
					ReadinessProbe: &api.Probe{
						Handler: api.Handler{
							HTTPGet: &api.HTTPGetAction{
								Scheme: api.URISchemeHTTP,
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
	nodePortToHealthCheck := map[int64]string{
		3001: "/healthz",
	}
	addPods(lbc, nodePortToHealthCheck, api.NamespaceDefault)

	for p, exp := range nodePortToHealthCheck {
		got, err := lbc.tr.HealthCheck(p)
		if err != nil {
			t.Errorf("Failed to get health check for node port %v: %v", p, err)
		} else if got.RequestPath != exp {
			t.Errorf("Wrong health check for node port %v, got %v expected %v", p, got.RequestPath, exp)
		}
	}
}

func addPods(lbc *LoadBalancerController, nodePortToHealthCheck map[int64]string, ns string) {
	delay := time.Minute
	for np, u := range nodePortToHealthCheck {
		l := map[string]string{fmt.Sprintf("app-%d", np): "test"}
		svc := &api.Service{
			Spec: api.ServiceSpec{
				Selector: l,
				Ports: []api.ServicePort{
					{
						NodePort: int32(np),
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 80,
						},
					},
				},
			},
		}
		svc.Name = fmt.Sprintf("%d", np)
		svc.Namespace = ns
		lbc.svcLister.Indexer.Add(svc)

		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{
				Labels:            l,
				Name:              fmt.Sprintf("%d", np),
				Namespace:         ns,
				CreationTimestamp: unversioned.NewTime(firstPodCreationTime.Add(delay)),
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Ports: []api.ContainerPort{{ContainerPort: 80}},
						ReadinessProbe: &api.Probe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Scheme: api.URISchemeHTTP,
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
			n := &api.Node{
				ObjectMeta: api.ObjectMeta{
					Name: node,
					Labels: map[string]string{
						zoneKey: zone,
					},
				},
				Status: api.NodeStatus{
					Conditions: []api.NodeCondition{
						{Type: api.NodeReady, Status: api.ConditionTrue},
					},
				},
			}
			lbc.nodeLister.Store.Add(n)
		}
	}
	lbc.CloudClusterManager.instancePool.Init(lbc.tr)
}
