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
	"sort"
	"strconv"
	"time"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/contrib/ingress/controllers/gce/loadbalancers"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/util/workqueue"

	"github.com/golang/glog"
)

const (
	// allowHTTPKey tells the Ingress controller to allow/block HTTP access.
	// If either unset or set to true, the controller will create a
	// forwarding-rule for port 80, and any additional rules based on the TLS
	// section of the Ingress. If set to false, the controller will only create
	// rules for port 443 based on the TLS section.
	allowHTTPKey = "kubernetes.io/ingress.allow-http"

	// staticIPNameKey tells the Ingress controller to use a specific GCE
	// static ip for its forwarding rules. If specified, the Ingress controller
	// assigns the static ip by this name to the forwarding rules of the given
	// Ingress. The controller *does not* manage this ip, it is the users
	// responsibility to create/delete it.
	staticIPNameKey = "kubernetes.io/ingress.global-static-ip-name"

	// ingressClassKey picks a specific "class" for the Ingress. The controller
	// only processes Ingresses with this annotation either unset, or set
	// to either gceIngessClass or the empty string.
	ingressClassKey = "kubernetes.io/ingress.class"
	gceIngressClass = "gce"

	// Label key to denote which GCE zone a Kubernetes node is in.
	zoneKey     = "failure-domain.beta.kubernetes.io/zone"
	defaultZone = ""
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

func (ing ingAnnotations) ingressClass() string {
	val, ok := ing[ingressClassKey]
	if !ok {
		return ""
	}
	return val
}

// isGCEIngress returns true if the given Ingress either doesn't specify the
// ingress.class annotation, or it's set to "gce".
func isGCEIngress(ing *extensions.Ingress) bool {
	class := ingAnnotations(ing.ObjectMeta.Annotations).ingressClass()
	return class == "" || class == gceIngressClass
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
	queue workqueue.RateLimitingInterface
	// sync is called for each item in the queue
	sync func(string) error
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

// worker processes work in the queue through sync.
func (t *taskQueue) worker() {
	for {
		key, quit := t.queue.Get()
		if quit {
			close(t.workerDone)
			return
		}
		glog.V(3).Infof("Syncing %v", key)
		if err := t.sync(key.(string)); err != nil {
			glog.Errorf("Requeuing %v, err %v", key, err)
			t.queue.AddRateLimited(key)
		} else {
			t.queue.Forget(key)
		}
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
func NewTaskQueue(syncFn func(string) error) *taskQueue {
	return &taskQueue{
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
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
		newIng := m.(*extensions.Ingress)
		if isGCEIngress(newIng) {
			ing.Items = append(ing.Items, *newIng)
		}
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
			if p.Port == be.ServicePort.IntVal {
				nodePort = int(p.NodePort)
				break
			}
		default:
			if p.Name == be.ServicePort.StrVal {
				nodePort = int(p.NodePort)
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

func getZone(n api.Node) string {
	zone, ok := n.Labels[zoneKey]
	if !ok {
		return defaultZone
	}
	return zone
}

// GetZoneForNode returns the zone for a given node by looking up its zone label.
func (t *GCETranslator) GetZoneForNode(name string) (string, error) {
	nodes, err := t.nodeLister.NodeCondition(getNodeReadyPredicate()).List()
	if err != nil {
		return "", err
	}
	for _, n := range nodes.Items {
		if n.Name == name {
			// TODO: Make this more resilient to label changes by listing
			// cloud nodes and figuring out zone.
			return getZone(n), nil
		}
	}
	return "", fmt.Errorf("Node not found %v", name)
}

// ListZones returns a list of zones this Kubernetes cluster spans.
func (t *GCETranslator) ListZones() ([]string, error) {
	zones := sets.String{}
	readyNodes, err := t.nodeLister.NodeCondition(getNodeReadyPredicate()).List()
	if err != nil {
		return zones.List(), err
	}
	for _, n := range readyNodes.Items {
		zones.Insert(getZone(n))
	}
	return zones.List(), nil
}

// isPortEqual compares the given IntOrString ports
func isPortEqual(port, targetPort intstr.IntOrString) bool {
	if targetPort.Type == intstr.Int {
		return port.IntVal == targetPort.IntVal
	}
	return port.StrVal == targetPort.StrVal
}

// geHTTPProbe returns the http readiness probe from the first container
// that matches targetPort, from the set of pods matching the given labels.
func (t *GCETranslator) getHTTPProbe(l map[string]string, targetPort intstr.IntOrString) (*api.Probe, error) {
	// Lookup any container with a matching targetPort from the set of pods
	// with a matching label selector.
	pl, err := t.podLister.List(labels.SelectorFromSet(labels.Set(l)))
	if err != nil {
		return nil, err
	}

	// If multiple endpoints have different health checks, take the first
	sort.Sort(PodsByCreationTimestamp(pl))

	for _, pod := range pl {
		logStr := fmt.Sprintf("Pod %v matching service selectors %v (targetport %+v)", pod.Name, l, targetPort)
		for _, c := range pod.Spec.Containers {
			if !isSimpleHTTPProbe(c.ReadinessProbe) {
				continue
			}
			for _, p := range c.Ports {
				cPort := intstr.IntOrString{IntVal: p.ContainerPort, StrVal: p.Name}
				if isPortEqual(cPort, targetPort) {
					if isPortEqual(c.ReadinessProbe.Handler.HTTPGet.Port, targetPort) {
						return c.ReadinessProbe, nil
					}
					glog.Infof("%v: found matching targetPort on container %v, but not on readinessProbe (%+v)",
						logStr, c.Name, c.ReadinessProbe.Handler.HTTPGet.Port)
				}
			}
		}
		glog.V(4).Infof("%v: lacks a matching HTTP probe for use in health checks.", logStr)
	}
	return nil, nil
}

// isSimpleHTTPProbe returns true if the given Probe is:
// - an HTTPGet probe, as opposed to a tcp or exec probe
// - has a scheme of HTTP, as opposed to HTTPS
// - has no special host or headers fields
func isSimpleHTTPProbe(probe *api.Probe) bool {
	return (probe != nil && probe.Handler.HTTPGet != nil && probe.Handler.HTTPGet.Host == "" &&
		probe.Handler.HTTPGet.Scheme == api.URISchemeHTTP && len(probe.Handler.HTTPGet.HTTPHeaders) == 0)
}

// HealthCheck returns the http readiness probe for the endpoint backing the
// given nodePort. If no probe is found it returns a health check with "" as
// the request path, callers are responsible for swapping this out for the
// appropriate default.
func (t *GCETranslator) HealthCheck(port int64) (*compute.HttpHealthCheck, error) {
	sl, err := t.svcLister.List()
	if err != nil {
		return nil, err
	}
	// Find the label and target port of the one service with the given nodePort
	for _, s := range sl.Items {
		for _, p := range s.Spec.Ports {
			if int32(port) == p.NodePort {
				rp, err := t.getHTTPProbe(s.Spec.Selector, p.TargetPort)
				if err != nil {
					return nil, err
				}
				if rp == nil {
					glog.Infof("No pod in service %v with node port %v has declared a matching readiness probe for health checks.", s.Name, port)
					break
				}
				healthPath := rp.Handler.HTTPGet.Path
				// GCE requires a leading "/" for health check urls.
				if string(healthPath[0]) != "/" {
					healthPath = fmt.Sprintf("/%v", healthPath)
				}
				host := rp.Handler.HTTPGet.Host
				glog.Infof("Found custom health check for Service %v nodeport %v: %v%v", s.Name, port, host, healthPath)
				return &compute.HttpHealthCheck{
					Port:               port,
					RequestPath:        healthPath,
					Host:               host,
					Description:        "kubernetes L7 health check from readiness probe.",
					CheckIntervalSec:   int64(rp.PeriodSeconds),
					TimeoutSec:         int64(rp.TimeoutSeconds),
					HealthyThreshold:   int64(rp.SuccessThreshold),
					UnhealthyThreshold: int64(rp.FailureThreshold),
					// TODO: include headers after updating compute godep.
				}, nil
			}
		}
	}
	return utils.DefaultHealthCheckTemplate(port), nil
}

// PodsByCreationTimestamp sorts a list of Pods by creation timestamp, using their names as a tie breaker.
type PodsByCreationTimestamp []*api.Pod

func (o PodsByCreationTimestamp) Len() int      { return len(o) }
func (o PodsByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o PodsByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(o[j].CreationTimestamp)
}
