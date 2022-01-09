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

package status

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"k8s.io/klog/v2"

	pool "gopkg.in/go-playground/pool.v3"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/task"
)

// UpdateInterval defines the time interval, in seconds, in
// which the status should check if an update is required.
var UpdateInterval = 60

// Syncer ...
type Syncer interface {
	Run(chan struct{})

	Shutdown()
}

type ingressLister interface {
	// ListIngresses returns the list of Ingresses
	ListIngresses() []*ingress.Ingress
}

// Config ...
type Config struct {
	Client clientset.Interface

	PublishService string

	PublishStatusAddress string

	UpdateStatusOnShutdown bool

	UseNodeInternalIP bool

	IngressLister ingressLister
}

// statusSync keeps the status IP in each Ingress rule updated executing a periodic check
// in all the defined rules. To simplify the process leader election is used so the update
// is executed only in one node (Ingress controllers can be scaled to more than one)
// If the controller is running with the flag --publish-service (with a valid service)
// the IP address behind the service is used, if it is running with the flag
// --publish-status-address, the address specified in the flag is used, if neither of the
// two flags are set, the source is the IP/s of the node/s
type statusSync struct {
	Config

	// workqueue used to keep in sync the status IP/s
	// in the Ingress rules
	syncQueue *task.Queue
}

// Start starts the loop to keep the status in sync
func (s statusSync) Run(stopCh chan struct{}) {
	go s.syncQueue.Run(time.Second, stopCh)

	// trigger initial sync
	s.syncQueue.EnqueueTask(task.GetDummyObject("sync status"))

	// when this instance is the leader we need to enqueue
	// an item to trigger the update of the Ingress status.
	wait.PollUntil(time.Duration(UpdateInterval)*time.Second, func() (bool, error) {
		s.syncQueue.EnqueueTask(task.GetDummyObject("sync status"))
		return false, nil
	}, stopCh)
}

// Shutdown stops the sync. In case the instance is the leader it will remove the current IP
// if there is no other instances running.
func (s statusSync) Shutdown() {
	go s.syncQueue.Shutdown()

	if !s.UpdateStatusOnShutdown {
		klog.Warningf("skipping update of status of Ingress rules")
		return
	}

	addrs, err := s.runningAddresses()
	if err != nil {
		klog.ErrorS(err, "error obtaining running IP address")
		return
	}

	if len(addrs) > 1 {
		// leave the job to the next leader
		klog.InfoS("leaving status update for next leader")
		return
	}

	if s.isRunningMultiplePods() {
		klog.V(2).InfoS("skipping Ingress status update (multiple pods running - another one will be elected as master)")
		return
	}

	klog.InfoS("removing value from ingress status", "address", addrs)
	s.updateStatus([]apiv1.LoadBalancerIngress{})
}

func (s *statusSync) sync(key interface{}) error {
	if s.syncQueue.IsShuttingDown() {
		klog.V(2).InfoS("skipping Ingress status update (shutting down in progress)")
		return nil
	}

	addrs, err := s.runningAddresses()
	if err != nil {
		return err
	}
	s.updateStatus(standardizeLoadBalancerIngresses(addrs))

	return nil
}

func (s statusSync) keyfunc(input interface{}) (interface{}, error) {
	return input, nil
}

// NewStatusSyncer returns a new Syncer instance
func NewStatusSyncer(config Config) Syncer {
	st := statusSync{
		Config: config,
	}
	st.syncQueue = task.NewCustomTaskQueue(st.sync, st.keyfunc)

	return st
}

func nameOrIPToLoadBalancerIngress(nameOrIP string) apiv1.LoadBalancerIngress {
	if net.ParseIP(nameOrIP) != nil {
		return apiv1.LoadBalancerIngress{IP: nameOrIP}
	}

	return apiv1.LoadBalancerIngress{Hostname: nameOrIP}
}

// runningAddresses returns a list of IP addresses and/or FQDN where the
// ingress controller is currently running
func (s *statusSync) runningAddresses() ([]apiv1.LoadBalancerIngress, error) {
	if s.PublishStatusAddress != "" {
		re := regexp.MustCompile(`,\s*`)
		multipleAddrs := re.Split(s.PublishStatusAddress, -1)
		addrs := make([]apiv1.LoadBalancerIngress, len(multipleAddrs))
		for i, addr := range multipleAddrs {
			addrs[i] = nameOrIPToLoadBalancerIngress(addr)
		}
		return addrs, nil
	}

	if s.PublishService != "" {
		return statusAddressFromService(s.PublishService, s.Client)
	}

	// get information about all the pods running the ingress controller
	pods, err := s.Client.CoreV1().Pods(k8s.IngressPodDetails.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(k8s.IngressPodDetails.Labels).String(),
	})
	if err != nil {
		return nil, err
	}

	addrs := make([]apiv1.LoadBalancerIngress, 0)
	for i := range pods.Items {
		pod := pods.Items[i]
		// only Running pods are valid
		if pod.Status.Phase != apiv1.PodRunning {
			continue
		}

		// only Ready pods are valid
		isPodReady := false
		for _, cond := range pod.Status.Conditions {
			if cond.Type == apiv1.PodReady && cond.Status == apiv1.ConditionTrue {
				isPodReady = true
				break
			}
		}

		if !isPodReady {
			klog.InfoS("POD is not ready", "pod", klog.KObj(&pod), "node", pod.Spec.NodeName)
			continue
		}

		name := k8s.GetNodeIPOrName(s.Client, pod.Spec.NodeName, s.UseNodeInternalIP)
		if !stringInIngresses(name, addrs) {
			addrs = append(addrs, nameOrIPToLoadBalancerIngress(name))
		}
	}

	return addrs, nil
}

func (s *statusSync) isRunningMultiplePods() bool {

	// As a standard, app.kubernetes.io are "reserved well-known" labels.
	// In our case, we add those labels as identifiers of the Ingress
	// deployment in this namespace, so we can select it as a set of Ingress instances.
	// As those labels are also generated as part of a HELM deployment, we can be "safe" they
	// cover 95% of the cases
	podLabel := make(map[string]string)
	for k, v := range k8s.IngressPodDetails.Labels {
		if k != "pod-template-hash" && k != "controller-revision-hash" && k != "pod-template-generation" {
			podLabel[k] = v
		}
	}

	pods, err := s.Client.CoreV1().Pods(k8s.IngressPodDetails.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(podLabel).String(),
	})
	if err != nil {
		return false
	}

	return len(pods.Items) > 1
}

// standardizeLoadBalancerIngresses sorts the list of loadbalancer by
// IP
func standardizeLoadBalancerIngresses(lbi []apiv1.LoadBalancerIngress) []apiv1.LoadBalancerIngress {
	sort.SliceStable(lbi, func(a, b int) bool {
		return lbi[a].IP < lbi[b].IP
	})

	return lbi
}

// updateStatus changes the status information of Ingress rules
func (s *statusSync) updateStatus(newIngressPoint []apiv1.LoadBalancerIngress) {
	ings := s.IngressLister.ListIngresses()

	p := pool.NewLimited(10)
	defer p.Close()

	batch := p.Batch()
	sort.SliceStable(newIngressPoint, lessLoadBalancerIngress(newIngressPoint))

	for _, ing := range ings {
		curIPs := ing.Status.LoadBalancer.Ingress
		sort.SliceStable(curIPs, lessLoadBalancerIngress(curIPs))
		if ingressSliceEqual(curIPs, newIngressPoint) {
			klog.V(3).InfoS("skipping update of Ingress (no change)", "namespace", ing.Namespace, "ingress", ing.Name)
			continue
		}

		batch.Queue(runUpdate(ing, newIngressPoint, s.Client))
	}

	batch.QueueComplete()
	batch.WaitAll()
}

func runUpdate(ing *ingress.Ingress, status []apiv1.LoadBalancerIngress,
	client clientset.Interface) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			return nil, nil
		}

		ingClient := client.NetworkingV1().Ingresses(ing.Namespace)
		currIng, err := ingClient.Get(context.TODO(), ing.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("unexpected error searching Ingress %s/%s: %w", ing.Namespace, ing.Name, err)
		}

		klog.InfoS("updating Ingress status", "namespace", currIng.Namespace, "ingress", currIng.Name, "currentValue", currIng.Status.LoadBalancer.Ingress, "newValue", status)
		currIng.Status.LoadBalancer.Ingress = status
		_, err = ingClient.UpdateStatus(context.TODO(), currIng, metav1.UpdateOptions{})
		if err != nil {
			klog.Warningf("error updating ingress rule: %v", err)
		}

		return true, nil
	}
}

func lessLoadBalancerIngress(addrs []apiv1.LoadBalancerIngress) func(int, int) bool {
	return func(a, b int) bool {
		switch strings.Compare(addrs[a].Hostname, addrs[b].Hostname) {
		case -1:
			return true
		case 1:
			return false
		}
		return addrs[a].IP < addrs[b].IP
	}
}

func ingressSliceEqual(lhs, rhs []apiv1.LoadBalancerIngress) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := range lhs {
		if lhs[i].IP != rhs[i].IP {
			return false
		}
		if lhs[i].Hostname != rhs[i].Hostname {
			return false
		}
	}

	return true
}

func statusAddressFromService(service string, kubeClient clientset.Interface) ([]apiv1.LoadBalancerIngress, error) {
	ns, name, _ := k8s.ParseNameNS(service)
	svc, err := kubeClient.CoreV1().Services(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	switch svc.Spec.Type {
	case apiv1.ServiceTypeExternalName:
		return []apiv1.LoadBalancerIngress{{
			Hostname: svc.Spec.ExternalName,
		}}, nil
	case apiv1.ServiceTypeClusterIP:
		return []apiv1.LoadBalancerIngress{{
			IP: svc.Spec.ClusterIP,
		}}, nil
	case apiv1.ServiceTypeNodePort:
		if svc.Spec.ExternalIPs == nil {
			return []apiv1.LoadBalancerIngress{{
				IP: svc.Spec.ClusterIP,
			}}, nil
		}
		addrs := make([]apiv1.LoadBalancerIngress, len(svc.Spec.ExternalIPs))
		for i, ip := range svc.Spec.ExternalIPs {
			addrs[i] = apiv1.LoadBalancerIngress{IP: ip}
		}
		return addrs, nil
	case apiv1.ServiceTypeLoadBalancer:
		addrs := make([]apiv1.LoadBalancerIngress, len(svc.Status.LoadBalancer.Ingress))
		for i, ingress := range svc.Status.LoadBalancer.Ingress {
			addrs[i] = apiv1.LoadBalancerIngress{}
			if ingress.Hostname != "" {
				addrs[i].Hostname = ingress.Hostname
			}
			if ingress.IP != "" {
				addrs[i].IP = ingress.IP
			}
		}
		for _, ip := range svc.Spec.ExternalIPs {
			if !stringInIngresses(ip, addrs) {
				addrs = append(addrs, apiv1.LoadBalancerIngress{IP: ip})
			}
		}
		return addrs, nil
	}

	return nil, fmt.Errorf("unable to extract IP address/es from service %v", service)
}

// stringInSlice returns true if s is in list
func stringInIngresses(s string, list []apiv1.LoadBalancerIngress) bool {
	for _, v := range list {
		if v.IP == s || v.Hostname == s {
			return true
		}
	}

	return false
}
