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
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog"

	pool "gopkg.in/go-playground/pool.v3"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/task"
)

const (
	updateInterval = 60 * time.Second
)

// Sync ...
type Sync interface {
	Run()
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

	ElectionID string

	UpdateStatusOnShutdown bool

	UseNodeInternalIP bool

	IngressLister ingressLister

	DefaultIngressClass string
	IngressClass        string
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
	// pod contains runtime information about this pod
	pod *k8s.PodInfo

	elector *leaderelection.LeaderElector

	// workqueue used to keep in sync the status IP/s
	// in the Ingress rules
	syncQueue *task.Queue
}

// Run starts the loop to keep the status in sync
func (s statusSync) Run() {
	// we need to use the defined ingress class to allow multiple leaders
	// in order to update information about ingress status
	electionID := fmt.Sprintf("%v-%v", s.Config.ElectionID, s.Config.DefaultIngressClass)
	if s.Config.IngressClass != "" {
		electionID = fmt.Sprintf("%v-%v", s.Config.ElectionID, s.Config.IngressClass)
	}

	// start a new context
	ctx := context.Background()

	var cancelContext context.CancelFunc

	var newLeaderCtx = func(ctx context.Context) context.CancelFunc {
		// allow to cancel the context in case we stop being the leader
		leaderCtx, cancel := context.WithCancel(ctx)
		go s.elector.Run(leaderCtx)
		return cancel
	}

	var stopCh chan struct{}
	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			klog.V(2).Infof("I am the new status update leader")
			stopCh = make(chan struct{})
			go s.syncQueue.Run(time.Second, stopCh)
			// trigger initial sync
			s.syncQueue.EnqueueTask(task.GetDummyObject("sync status"))
			// when this instance is the leader we need to enqueue
			// an item to trigger the update of the Ingress status.
			wait.PollUntil(updateInterval, func() (bool, error) {
				s.syncQueue.EnqueueTask(task.GetDummyObject("sync status"))
				return false, nil
			}, stopCh)
		},
		OnStoppedLeading: func() {
			klog.V(2).Info("I am not status update leader anymore")
			close(stopCh)

			// cancel the context
			cancelContext()

			cancelContext = newLeaderCtx(ctx)
		},
		OnNewLeader: func(identity string) {
			klog.Infof("new leader elected: %v", identity)
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, _ := os.Hostname()

	recorder := broadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{
		Component: "ingress-leader-elector",
		Host:      hostname,
	})

	lock := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{Namespace: s.pod.Namespace, Name: electionID},
		Client:        s.Config.Client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      s.pod.Name,
			EventRecorder: recorder,
		},
	}

	ttl := 30 * time.Second
	var err error
	s.elector, err = leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callbacks,
	})
	if err != nil {
		klog.Fatalf("unexpected error starting leader election: %v", err)
	}

	cancelContext = newLeaderCtx(ctx)
}

// Shutdown stop the sync. In case the instance is the leader it will remove the current IP
// if there is no other instances running.
func (s statusSync) Shutdown() {
	go s.syncQueue.Shutdown()

	// remove IP from Ingress
	if s.elector != nil && !s.elector.IsLeader() {
		return
	}

	if !s.UpdateStatusOnShutdown {
		klog.Warningf("skipping update of status of Ingress rules")
		return
	}

	klog.Info("updating status of Ingress rules (remove)")

	addrs, err := s.runningAddresses()
	if err != nil {
		klog.Errorf("error obtaining running IPs: %v", addrs)
		return
	}

	if len(addrs) > 1 {
		// leave the job to the next leader
		klog.Infof("leaving status update for next leader (%v)", len(addrs))
		return
	}

	if s.isRunningMultiplePods() {
		klog.V(2).Infof("skipping Ingress status update (multiple pods running - another one will be elected as master)")
		return
	}

	klog.Infof("removing address from ingress status (%v)", addrs)
	s.updateStatus([]apiv1.LoadBalancerIngress{})
}

func (s *statusSync) sync(key interface{}) error {
	if s.syncQueue.IsShuttingDown() {
		klog.V(2).Infof("skipping Ingress status update (shutting down in progress)")
		return nil
	}

	if s.elector != nil && !s.elector.IsLeader() {
		return fmt.Errorf("i am not the current leader. Skiping status update")
	}

	addrs, err := s.runningAddresses()
	if err != nil {
		return err
	}
	s.updateStatus(sliceToStatus(addrs))

	return nil
}

func (s statusSync) keyfunc(input interface{}) (interface{}, error) {
	return input, nil
}

// NewStatusSyncer returns a new Sync instance
func NewStatusSyncer(config Config) Sync {
	pod, err := k8s.GetPodDetails(config.Client)
	if err != nil {
		klog.Fatalf("unexpected error obtaining pod information: %v", err)
	}

	st := statusSync{
		pod: pod,

		Config: config,
	}
	st.syncQueue = task.NewCustomTaskQueue(st.sync, st.keyfunc)

	return st
}

// runningAddresses returns a list of IP addresses and/or FQDN where the
// ingress controller is currently running
func (s *statusSync) runningAddresses() ([]string, error) {
	addrs := []string{}

	if s.PublishService != "" {
		ns, name, _ := k8s.ParseNameNS(s.PublishService)
		svc, err := s.Client.CoreV1().Services(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if svc.Spec.Type == apiv1.ServiceTypeExternalName {
			addrs = append(addrs, svc.Spec.ExternalName)
			return addrs, nil
		}

		for _, ip := range svc.Status.LoadBalancer.Ingress {
			if ip.IP == "" {
				addrs = append(addrs, ip.Hostname)
			} else {
				addrs = append(addrs, ip.IP)
			}
		}

		addrs = append(addrs, svc.Spec.ExternalIPs...)
		return addrs, nil
	}

	if s.PublishStatusAddress != "" {
		addrs = append(addrs, s.PublishStatusAddress)
		return addrs, nil
	}

	// get information about all the pods running the ingress controller
	pods, err := s.Client.CoreV1().Pods(s.pod.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(s.pod.Labels).String(),
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		// only Running pods are valid
		if pod.Status.Phase != apiv1.PodRunning {
			continue
		}

		name := k8s.GetNodeIPOrName(s.Client, pod.Spec.NodeName, s.UseNodeInternalIP)
		if !sliceutils.StringInSlice(name, addrs) {
			addrs = append(addrs, name)
		}
	}

	return addrs, nil
}

func (s *statusSync) isRunningMultiplePods() bool {
	pods, err := s.Client.CoreV1().Pods(s.pod.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(s.pod.Labels).String(),
	})
	if err != nil {
		return false
	}

	return len(pods.Items) > 1
}

// sliceToStatus converts a slice of IP and/or hostnames to LoadBalancerIngress
func sliceToStatus(endpoints []string) []apiv1.LoadBalancerIngress {
	lbi := []apiv1.LoadBalancerIngress{}
	for _, ep := range endpoints {
		if net.ParseIP(ep) == nil {
			lbi = append(lbi, apiv1.LoadBalancerIngress{Hostname: ep})
		} else {
			lbi = append(lbi, apiv1.LoadBalancerIngress{IP: ep})
		}
	}

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
			klog.V(3).Infof("skipping update of Ingress %v/%v (no change)", ing.Namespace, ing.Name)
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

		ingClient := client.ExtensionsV1beta1().Ingresses(ing.Namespace)

		currIng, err := ingClient.Get(ing.Name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("unexpected error searching Ingress %v/%v", ing.Namespace, ing.Name))
		}

		klog.Infof("updating Ingress %v/%v status from %v to %v", currIng.Namespace, currIng.Name, currIng.Status.LoadBalancer.Ingress, status)
		currIng.Status.LoadBalancer.Ingress = status
		_, err = ingClient.UpdateStatus(currIng)
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
