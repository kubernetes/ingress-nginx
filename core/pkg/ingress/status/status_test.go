/*
Copyright 2017 The Kubernetes Authors.

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

package status

import (
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	api_v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/api"

	"k8s.io/ingress/core/pkg/ingress/annotations/class"
	cache_store "k8s.io/ingress/core/pkg/ingress/store"
	"k8s.io/ingress/core/pkg/k8s"
	"k8s.io/ingress/core/pkg/task"
)

func buildLoadBalancerIngressByIP() loadBalancerIngressByIP {
	return []api_v1.LoadBalancerIngress{
		{
			IP:       "10.0.0.1",
			Hostname: "foo1",
		},
		{
			IP:       "10.0.0.2",
			Hostname: "foo2",
		},
		{
			IP:       "10.0.0.3",
			Hostname: "",
		},
		{
			IP:       "",
			Hostname: "foo4",
		},
	}
}

func buildSimpleClientSet() *testclient.Clientset {
	return testclient.NewSimpleClientset(
		&api_v1.PodList{Items: []api_v1.Pod{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "foo1",
					Namespace: api_v1.NamespaceDefault,
					Labels: map[string]string{
						"lable_sig": "foo_pod",
					},
				},
				Spec: api_v1.PodSpec{
					NodeName: "foo_node_2",
				},
			},
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "foo2",
					Namespace: api_v1.NamespaceDefault,
					Labels: map[string]string{
						"lable_sig": "foo_no",
					},
				},
			},
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "foo3",
					Namespace: api.NamespaceSystem,
					Labels: map[string]string{
						"lable_sig": "foo_pod",
					},
				},
				Spec: api_v1.PodSpec{
					NodeName: "foo_node_2",
				},
			},
		}},
		&api_v1.ServiceList{Items: []api_v1.Service{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "foo",
					Namespace: api_v1.NamespaceDefault,
				},
				Status: api_v1.ServiceStatus{
					LoadBalancer: api_v1.LoadBalancerStatus{
						Ingress: buildLoadBalancerIngressByIP(),
					},
				},
			},
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "foo_non_exist",
					Namespace: api_v1.NamespaceDefault,
				},
			},
		}},
		&api_v1.NodeList{Items: []api_v1.Node{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "foo_node_1",
				},
				Status: api_v1.NodeStatus{
					Addresses: []api_v1.NodeAddress{
						{
							Type:    api_v1.NodeInternalIP,
							Address: "10.0.0.1",
						}, {
							Type:    api_v1.NodeExternalIP,
							Address: "10.0.0.2",
						},
					},
				},
			},
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "foo_node_2",
				},
				Status: api_v1.NodeStatus{
					Addresses: []api_v1.NodeAddress{
						{
							Type:    api_v1.NodeInternalIP,
							Address: "11.0.0.1",
						},
						{
							Type:    api_v1.NodeExternalIP,
							Address: "11.0.0.2",
						},
					},
				},
			},
		}},
		&api_v1.EndpointsList{Items: []api_v1.Endpoints{
			{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "ingress-controller-leader",
					Namespace: api_v1.NamespaceDefault,
					SelfLink:  "/api/v1/namespaces/default/endpoints/ingress-controller-leader",
				},
			}}},
		&extensions.IngressList{Items: buildExtensionsIngresses()},
	)
}

func fakeSynFn(interface{}) error {
	return nil
}

func buildExtensionsIngresses() []extensions.Ingress {
	return []extensions.Ingress{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "foo_ingress_1",
				Namespace: api_v1.NamespaceDefault,
			},
			Status: extensions.IngressStatus{
				LoadBalancer: api_v1.LoadBalancerStatus{
					Ingress: []api_v1.LoadBalancerIngress{
						{
							IP:       "10.0.0.1",
							Hostname: "foo1",
						},
					},
				},
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "foo_ingress_different_class",
				Namespace: api.NamespaceDefault,
				Annotations: map[string]string{
					class.IngressKey: "no-nginx",
				},
			},
			Status: extensions.IngressStatus{
				LoadBalancer: api_v1.LoadBalancerStatus{
					Ingress: []api_v1.LoadBalancerIngress{
						{
							IP:       "0.0.0.0",
							Hostname: "foo.bar.com",
						},
					},
				},
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "foo_ingress_2",
				Namespace: api_v1.NamespaceDefault,
			},
			Status: extensions.IngressStatus{
				LoadBalancer: api_v1.LoadBalancerStatus{
					Ingress: []api_v1.LoadBalancerIngress{},
				},
			},
		},
	}
}

func buildIngressListener() cache_store.IngressLister {
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	store.Add(&extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo_ingress_non_01",
			Namespace: api_v1.NamespaceDefault,
		}})
	store.Add(&extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo_ingress_1",
			Namespace: api_v1.NamespaceDefault,
		},
		Status: extensions.IngressStatus{
			LoadBalancer: api_v1.LoadBalancerStatus{
				Ingress: buildLoadBalancerIngressByIP(),
			},
		},
	})
	return cache_store.IngressLister{Store: store}
}

func buildStatusSync() statusSync {
	return statusSync{
		pod: &k8s.PodInfo{
			Name:      "foo_base_pod",
			Namespace: api_v1.NamespaceDefault,
			Labels: map[string]string{
				"lable_sig": "foo_pod",
			},
		},
		runLock:   &sync.Mutex{},
		syncQueue: task.NewTaskQueue(fakeSynFn),
		Config: Config{
			Client:         buildSimpleClientSet(),
			PublishService: api_v1.NamespaceDefault + "/" + "foo",
			IngressLister:  buildIngressListener(),
			CustomIngressStatus: func(*extensions.Ingress) []api_v1.LoadBalancerIngress {
				return nil
			},
		},
	}
}

func TestStatusActions(t *testing.T) {
	// make sure election can be created
	os.Setenv("POD_NAME", "foo1")
	os.Setenv("POD_NAMESPACE", api_v1.NamespaceDefault)
	c := Config{
		Client:                 buildSimpleClientSet(),
		PublishService:         "",
		IngressLister:          buildIngressListener(),
		DefaultIngressClass:    "nginx",
		IngressClass:           "",
		UpdateStatusOnShutdown: true,
		CustomIngressStatus: func(*extensions.Ingress) []api_v1.LoadBalancerIngress {
			return nil
		},
	}
	// create object
	fkSync := NewStatusSyncer(c)
	if fkSync == nil {
		t.Fatalf("expected a valid Sync")
	}

	fk := fkSync.(statusSync)

	ns := make(chan struct{})
	// start it and wait for the election and syn actions
	go fk.Run(ns)
	//  wait for the election
	time.Sleep(100 * time.Millisecond)
	// execute sync
	fk.sync("just-test")
	// PublishService is empty, so the running address is: ["11.0.0.2"]
	// after updated, the ingress's ip should only be "11.0.0.2"
	newIPs := []api_v1.LoadBalancerIngress{{
		IP: "11.0.0.2",
	}}
	fooIngress1, err1 := fk.Client.Extensions().Ingresses(api_v1.NamespaceDefault).Get("foo_ingress_1", meta_v1.GetOptions{})
	if err1 != nil {
		t.Fatalf("unexpected error")
	}
	fooIngress1CurIPs := fooIngress1.Status.LoadBalancer.Ingress
	if !ingressSliceEqual(fooIngress1CurIPs, newIPs) {
		t.Fatalf("returned %v but expected %v", fooIngress1CurIPs, newIPs)
	}

	// execute shutdown
	fk.Shutdown()
	// ingress should be empty
	newIPs2 := []api_v1.LoadBalancerIngress{}
	fooIngress2, err2 := fk.Client.Extensions().Ingresses(api_v1.NamespaceDefault).Get("foo_ingress_1", meta_v1.GetOptions{})
	if err2 != nil {
		t.Fatalf("unexpected error")
	}
	fooIngress2CurIPs := fooIngress2.Status.LoadBalancer.Ingress
	if !ingressSliceEqual(fooIngress2CurIPs, newIPs2) {
		t.Fatalf("returned %v but expected %v", fooIngress2CurIPs, newIPs2)
	}

	oic, err := fk.Client.Extensions().Ingresses(api.NamespaceDefault).Get("foo_ingress_different_class", meta_v1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if oic.Status.LoadBalancer.Ingress[0].IP != "0.0.0.0" && oic.Status.LoadBalancer.Ingress[0].Hostname != "foo.bar.com" {
		t.Fatalf("invalid ingress status for rule with different class")
	}

	// end test
	ns <- struct{}{}
}

func TestCallback(t *testing.T) {
	fk := buildStatusSync()
	//  do nothing
	fk.callback("foo_base_pod")
}

func TestKeyfunc(t *testing.T) {
	fk := buildStatusSync()
	i := "foo_base_pod"
	r, err := fk.keyfunc(i)

	if err != nil {
		t.Fatalf("unexpected error")
	}
	if r != i {
		t.Errorf("returned %v but expected %v", r, i)
	}
}

func TestRunningAddresessWithPublishService(t *testing.T) {
	fk := buildStatusSync()

	r, _ := fk.runningAddresses()
	if r == nil {
		t.Fatalf("returned nil but expected valid []string")
	}
	rl := len(r)
	if len(r) != 4 {
		t.Errorf("returned %v but expected %v", rl, 4)
	}
}

func TestRunningAddresessWithPods(t *testing.T) {
	fk := buildStatusSync()
	fk.PublishService = ""

	r, _ := fk.runningAddresses()
	if r == nil {
		t.Fatalf("returned nil but expected valid []string")
	}
	rl := len(r)
	if len(r) != 1 {
		t.Fatalf("returned %v but expected %v", rl, 1)
	}
	rv := r[0]
	if rv != "11.0.0.2" {
		t.Errorf("returned %v but expected %v", rv, "11.0.0.2")
	}
}

func TestUpdateStatus(t *testing.T) {
	fk := buildStatusSync()
	newIPs := buildLoadBalancerIngressByIP()
	sort.Sort(loadBalancerIngressByIP(newIPs))
	fk.updateStatus(newIPs)

	fooIngress1, err1 := fk.Client.Extensions().Ingresses(api_v1.NamespaceDefault).Get("foo_ingress_1", meta_v1.GetOptions{})
	if err1 != nil {
		t.Fatalf("unexpected error")
	}
	fooIngress1CurIPs := fooIngress1.Status.LoadBalancer.Ingress
	if !ingressSliceEqual(fooIngress1CurIPs, newIPs) {
		t.Fatalf("returned %v but expected %v", fooIngress1CurIPs, newIPs)
	}

	fooIngress2, err2 := fk.Client.Extensions().Ingresses(api_v1.NamespaceDefault).Get("foo_ingress_2", meta_v1.GetOptions{})
	if err2 != nil {
		t.Fatalf("unexpected error")
	}
	fooIngress2CurIPs := fooIngress2.Status.LoadBalancer.Ingress
	if !ingressSliceEqual(fooIngress2CurIPs, []api_v1.LoadBalancerIngress{}) {
		t.Fatalf("returned %v but expected %v", fooIngress2CurIPs, []api_v1.LoadBalancerIngress{})
	}
}

func TestSliceToStatus(t *testing.T) {
	fkEndpoints := []string{
		"10.0.0.1",
		"2001:db8::68",
		"opensource-k8s-ingress",
	}

	r := sliceToStatus(fkEndpoints)

	if r == nil {
		t.Fatalf("returned nil but expected a valid []api_v1.LoadBalancerIngress")
	}
	rl := len(r)
	if rl != 3 {
		t.Fatalf("returned %v but expected %v", rl, 3)
	}
	re1 := r[0]
	if re1.Hostname != "opensource-k8s-ingress" {
		t.Fatalf("returned %v but expected %v", re1, api_v1.LoadBalancerIngress{Hostname: "opensource-k8s-ingress"})
	}
	re2 := r[1]
	if re2.IP != "10.0.0.1" {
		t.Fatalf("returned %v but expected %v", re2, api_v1.LoadBalancerIngress{IP: "10.0.0.1"})
	}
	re3 := r[2]
	if re3.IP != "2001:db8::68" {
		t.Fatalf("returned %v but expected %v", re3, api_v1.LoadBalancerIngress{IP: "2001:db8::68"})
	}
}

func TestIngressSliceEqual(t *testing.T) {
	fk1 := buildLoadBalancerIngressByIP()
	fk2 := append(buildLoadBalancerIngressByIP(), api_v1.LoadBalancerIngress{
		IP:       "10.0.0.5",
		Hostname: "foo5",
	})
	fk3 := buildLoadBalancerIngressByIP()
	fk3[0].Hostname = "foo_no_01"
	fk4 := buildLoadBalancerIngressByIP()
	fk4[2].IP = "11.0.0.3"

	fooTests := []struct {
		lhs []api_v1.LoadBalancerIngress
		rhs []api_v1.LoadBalancerIngress
		er  bool
	}{
		{fk1, fk1, true},
		{fk2, fk1, false},
		{fk3, fk1, false},
		{fk4, fk1, false},
		{fk1, nil, false},
		{nil, nil, true},
		{[]api_v1.LoadBalancerIngress{}, []api_v1.LoadBalancerIngress{}, true},
	}

	for _, fooTest := range fooTests {
		r := ingressSliceEqual(fooTest.lhs, fooTest.rhs)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v", r, fooTest.er)
		}
	}
}

func TestLoadBalancerIngressByIPLen(t *testing.T) {
	fooTests := []struct {
		ips loadBalancerIngressByIP
		el  int
	}{
		{[]api_v1.LoadBalancerIngress{}, 0},
		{buildLoadBalancerIngressByIP(), 4},
		{nil, 0},
	}

	for _, fooTest := range fooTests {
		r := fooTest.ips.Len()
		if r != fooTest.el {
			t.Errorf("returned %v but expected %v ", r, fooTest.el)
		}
	}
}

func TestLoadBalancerIngressByIPSwap(t *testing.T) {
	fooTests := []struct {
		ips loadBalancerIngressByIP
		i   int
		j   int
	}{
		{buildLoadBalancerIngressByIP(), 0, 1},
		{buildLoadBalancerIngressByIP(), 2, 1},
	}

	for _, fooTest := range fooTests {
		fooi := fooTest.ips[fooTest.i]
		fooj := fooTest.ips[fooTest.j]
		fooTest.ips.Swap(fooTest.i, fooTest.j)
		if fooi.IP != fooTest.ips[fooTest.j].IP ||
			fooj.IP != fooTest.ips[fooTest.i].IP {
			t.Errorf("failed to swap for loadBalancerIngressByIP")
		}
	}
}

func TestLoadBalancerIngressByIPLess(t *testing.T) {
	fooTests := []struct {
		ips loadBalancerIngressByIP
		i   int
		j   int
		er  bool
	}{
		{buildLoadBalancerIngressByIP(), 0, 1, true},
		{buildLoadBalancerIngressByIP(), 2, 1, false},
	}

	for _, fooTest := range fooTests {
		r := fooTest.ips.Less(fooTest.i, fooTest.j)
		if r != fooTest.er {
			t.Errorf("returned %v but expected %v ", r, fooTest.er)
		}
	}
}
