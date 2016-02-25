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

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/contrib/ingress/controllers/nginx-third-party/nginx"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/util/workqueue"
)

const (
	httpPort  = "80"
	httpsPort = "443"
)

var (
	errMissingPodInfo = fmt.Errorf("Unable to get POD information")

	errInvalidKind = fmt.Errorf("Please check the field Kind, only ReplicationController or DaemonSet are allowed")
)

// taskQueue manages a work queue through an independent worker that
// invokes the given sync function for every work item inserted.
type taskQueue struct {
	// queue is the work queue the worker polls
	queue *workqueue.Type
	// sync is called for each item in the queue
	sync func(string)
	// workerDone is closed when the worker exits
	workerDone chan struct{}
}

func (t *taskQueue) run(period time.Duration, stopCh <-chan struct{}) {
	wait.Until(t.worker, period, stopCh)
}

// enqueue enqueues ns/name of the given api object in the task queue.
func (t *taskQueue) enqueue(obj interface{}) {
	key, err := keyFunc(obj)
	if err != nil {
		glog.Infof("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	t.queue.Add(key)
}

func (t *taskQueue) requeue(key string, err error) {
	glog.Errorf("Requeuing %v, err %v", key, err)
	t.queue.Add(key)
}

// worker processes work in the queue through sync.
func (t *taskQueue) worker() {
	for {
		key, quit := t.queue.Get()
		if quit {
			close(t.workerDone)
			return
		}
		glog.V(2).Infof("Syncing %v", key)
		t.sync(key.(string))
		t.queue.Done(key)
	}
}

// shutdown shuts down the work queue and waits for the worker to ACK
func (t *taskQueue) shutdown() {
	t.queue.ShutDown()
	<-t.workerDone
}

// NewTaskQueue creates a new task queue with the given sync function.
// The sync function is called for every element inserted into the queue.
func NewTaskQueue(syncFn func(string)) *taskQueue {
	return &taskQueue{
		queue:      workqueue.New(),
		sync:       syncFn,
		workerDone: make(chan struct{}),
	}
}

// StoreToIngressLister makes a Store that lists Ingress.
// TODO: use cache/listers post 1.1.
type StoreToIngressLister struct {
	cache.Store
}

// List lists all Ingress' in the store.
func (s *StoreToIngressLister) List() (ing extensions.IngressList, err error) {
	for _, m := range s.Store.List() {
		ing.Items = append(ing.Items, *(m.(*extensions.Ingress)))
	}
	return ing, nil
}

// StoreToConfigMapLister makes a Store that lists existing ConfigMap.
type StoreToConfigMapLister struct {
	cache.Store
}

// getLBDetails returns runtime information about the pod (name, IP) and replication
// controller (namespace and name)
// This is required to watch for changes in annotations or configuration (ConfigMap)
func getLBDetails(kubeClient *unversioned.Client) (rc *lbInfo, err error) {
	podIP := os.Getenv("POD_IP")
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	pod, _ := kubeClient.Pods(podNs).Get(podName)
	if pod == nil {
		return nil, errMissingPodInfo
	}

	annotations := pod.Annotations["kubernetes.io/created-by"]
	var sref api.SerializedReference
	err = json.Unmarshal([]byte(annotations), &sref)
	if err != nil {
		return nil, err
	}

	rc = &lbInfo{
		ObjectName:   sref.Reference.Name,
		PodIP:        podIP,
		Podname:      podName,
		PodNamespace: podNs,
	}

	switch sref.Reference.Kind {
	case "ReplicationController":
		rc.DeployType = &api.ReplicationController{}
		return rc, nil
	case "DaemonSet":
		rc.DeployType = &extensions.DaemonSet{}
		return rc, nil
	default:
		return nil, errInvalidKind
	}
}

func getService(kubeClient *unversioned.Client, name string) (nginx.Service, error) {
	if name == "" {
		return nginx.Service{}, fmt.Errorf("Empty string is not a valid service name")
	}

	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		glog.Fatalf("Please check the service format (namespace/name) in service %v", name)
	}

	defaultPort, err := getServicePorts(kubeClient, parts[0], parts[1])
	if err != nil {
		return nginx.Service{}, fmt.Errorf("Error obtaining service %v: %v", name, err)
	}

	return nginx.Service{
		ServiceName: parts[1],
		ServicePort: defaultPort[0], //TODO: which port?
		Namespace:   parts[0],
	}, nil
}

// getServicePorts returns the ports defined in a service spec
func getServicePorts(kubeClient *unversioned.Client, ns, name string) (ports []string, err error) {
	var svc *api.Service
	glog.Infof("Checking service %v/%v", ns, name)
	svc, err = kubeClient.Services(ns).Get(name)
	if err != nil {
		return
	}

	for _, p := range svc.Spec.Ports {
		if p.Port != 0 {
			ports = append(ports, strconv.Itoa(p.Port))
			break
		}
	}

	glog.Infof("Ports for %v/%v : %v", ns, name, ports)

	return
}

func getTcpServices(kubeClient *unversioned.Client, tcpServices string) []nginx.Service {
	svcs := []nginx.Service{}
	for _, svc := range strings.Split(tcpServices, ",") {
		if svc == "" {
			continue
		}

		namePort := strings.Split(svc, ":")
		if len(namePort) == 2 {
			tcpSvc, err := getService(kubeClient, namePort[0])
			if err != nil {
				glog.Errorf("%s", err)
				continue
			}

			// the exposed TCP service cannot use 80 or 443 as ports
			if namePort[1] == httpPort || namePort[1] == httpsPort {
				glog.Errorf("The TCP service %v cannot use ports 80 or 443 (it creates a conflict with nginx)", svc)
				continue
			}

			tcpSvc.ExposedPort = namePort[1]
			svcs = append(svcs, tcpSvc)
		} else {
			glog.Errorf("TCP services should take the form namespace/name:port not %v from %v", namePort, svc)
		}
	}

	return svcs
}

func checkDNS(name string) (*net.IPAddr, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		glog.Fatalf("Please check the service format (namespace/name) in service %v", name)
	}

	host := fmt.Sprintf("%v.%v", parts[1], parts[0])
	return net.ResolveIPAddr("ip4", host)
}
