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
	"strconv"
	"time"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/loadbalancers"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/util/workqueue"

	"github.com/golang/glog"
)

const (
	allowHTTPKey    = "kubernetes.io/ingress.allow-http"
	staticIPNameKey = "kubernetes.io/ingress.global-static-ip-name"
)

// ingAnnotations represents Ingress annotations.
type ingAnnotations map[string]string

// allowHTTP returns the allowHTTP flag. True by default.
func (ing ingAnnotations) allowHTTP() bool {
	val, ok := ing[allowHTTPKey]
	if !ok {
		return true
	}
	v, err := strconv.ParseBool(val)
	if err != nil {
		return true
	}
	return v
}

func (ing ingAnnotations) staticIPName() string {
	val, ok := ing[staticIPNameKey]
	if !ok {
		return ""
	}
	return val
}

// errorNodePortNotFound is an implementation of error.
type errorNodePortNotFound struct {
	backend extensions.IngressBackend
	origErr error
}

func (e errorNodePortNotFound) Error() string {
	return fmt.Sprintf("Could not find nodeport for backend %+v: %v",
		e.backend, e.origErr)
}

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
		glog.V(3).Infof("Syncing %v", key)
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

// compareLinks returns true if the 2 self links are equal.
func compareLinks(l1, l2 string) bool {
	// TODO: These can be partial links
	return l1 == l2 && l1 != ""
}

// StoreToIngressLister makes a Store that lists Ingress.
// TODO: Move this to cache/listers post 1.1.
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

// GetServiceIngress gets all the Ingress' that have rules pointing to a service.
// Note that this ignores services without the right nodePorts.
func (s *StoreToIngressLister) GetServiceIngress(svc *api.Service) (ings []extensions.Ingress, err error) {
	for _, m := range s.Store.List() {
		ing := *m.(*extensions.Ingress)
		if ing.Namespace != svc.Namespace {
			continue
		}
		for _, rules := range ing.Spec.Rules {
			if rules.IngressRuleValue.HTTP == nil {
				continue
			}
			for _, p := range rules.IngressRuleValue.HTTP.Paths {
				if p.Backend.ServiceName == svc.Name {
					ings = append(ings, ing)
				}
			}
		}
	}
	if len(ings) == 0 {
		err = fmt.Errorf("No ingress for service %v", svc.Name)
	}
	return
}

// GCETranslator helps with kubernetes -> gce api conversion.
type GCETranslator struct {
	*LoadBalancerController
}

// toUrlMap converts an ingress to a map of subdomain: url-regex: gce backend.
func (t *GCETranslator) toUrlMap(ing *extensions.Ingress) (utils.GCEURLMap, error) {
	hostPathBackend := utils.GCEURLMap{}
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			glog.Errorf("Ignoring non http Ingress rule")
			continue
		}
		pathToBackend := map[string]*compute.BackendService{}
		for _, p := range rule.HTTP.Paths {
			backend, err := t.toGCEBackend(&p.Backend, ing.Namespace)
			if err != nil {
				// If a service doesn't have a nodeport we can still forward traffic
				// to all other services under the assumption that the user will
				// modify nodeport.
				if _, ok := err.(errorNodePortNotFound); ok {
					glog.Infof("%v", err)
					continue
				}

				// If a service doesn't have a backend, there's nothing the user
				// can do to correct this (the admin might've limited quota).
				// So keep requeuing the l7 till all backends exist.
				return utils.GCEURLMap{}, err
			}
			// The Ingress spec defines empty path as catch-all, so if a user
			// asks for a single host and multiple empty paths, all traffic is
			// sent to one of the last backend in the rules list.
			path := p.Path
			if path == "" {
				path = loadbalancers.DefaultPath
			}
			pathToBackend[path] = backend
		}
		// If multiple hostless rule sets are specified, last one wins
		host := rule.Host
		if host == "" {
			host = loadbalancers.DefaultHost
		}
		hostPathBackend[host] = pathToBackend
	}
	defaultBackend, _ := t.toGCEBackend(ing.Spec.Backend, ing.Namespace)
	hostPathBackend.PutDefaultBackend(defaultBackend)
	return hostPathBackend, nil
}

func (t *GCETranslator) toGCEBackend(be *extensions.IngressBackend, ns string) (*compute.BackendService, error) {
	if be == nil {
		return nil, nil
	}
	port, err := t.getServiceNodePort(*be, ns)
	if err != nil {
		return nil, err
	}
	backend, err := t.CloudClusterManager.backendPool.Get(int64(port))
	if err != nil {
		return nil, fmt.Errorf(
			"No GCE backend exists for port %v, kube backend %+v", port, be)
	}
	return backend, nil
}

// getServiceNodePort looks in the svc store for a matching service:port,
// and returns the nodeport.
func (t *GCETranslator) getServiceNodePort(be extensions.IngressBackend, namespace string) (int, error) {
	obj, exists, err := t.svcLister.Store.Get(
		&api.Service{
			ObjectMeta: api.ObjectMeta{
				Name:      be.ServiceName,
				Namespace: namespace,
			},
		})
	if !exists {
		return invalidPort, errorNodePortNotFound{be, fmt.Errorf(
			"Service %v/%v not found in store", namespace, be.ServiceName)}
	}
	if err != nil {
		return invalidPort, errorNodePortNotFound{be, err}
	}
	var nodePort int
	for _, p := range obj.(*api.Service).Spec.Ports {
		switch be.ServicePort.Type {
		case intstr.Int:
			if p.Port == int(be.ServicePort.IntVal) {
				nodePort = p.NodePort
				break
			}
		default:
			if p.Name == be.ServicePort.StrVal {
				nodePort = p.NodePort
				break
			}
		}
	}
	if nodePort != invalidPort {
		return nodePort, nil
	}
	return invalidPort, errorNodePortNotFound{be, fmt.Errorf(
		"Could not find matching nodeport from service.")}
}

// toNodePorts converts a pathlist to a flat list of nodeports.
func (t *GCETranslator) toNodePorts(ings *extensions.IngressList) []int64 {
	knownPorts := []int64{}
	for _, ing := range ings.Items {
		defaultBackend := ing.Spec.Backend
		if defaultBackend != nil {
			port, err := t.getServiceNodePort(*defaultBackend, ing.Namespace)
			if err != nil {
				glog.Infof("%v", err)
			} else {
				knownPorts = append(knownPorts, int64(port))
			}
		}
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				glog.Errorf("Ignoring non http Ingress rule.")
				continue
			}
			for _, path := range rule.HTTP.Paths {
				port, err := t.getServiceNodePort(path.Backend, ing.Namespace)
				if err != nil {
					glog.Infof("%v", err)
					continue
				}
				knownPorts = append(knownPorts, int64(port))
			}
		}
	}
	return knownPorts
}
