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
	"net/http"
	"reflect"
	"sync"
	"time"

	"k8s.io/contrib/ingress/controllers/gce/loadbalancers"
	"k8s.io/contrib/ingress/controllers/gce/utils"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/record"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/watch"

	"github.com/golang/glog"
)

var (
	keyFunc = framework.DeletionHandlingMetaNamespaceKeyFunc

	// DefaultClusterUID is the uid to use for clusters resources created by an
	// L7 controller created without specifying the --cluster-uid flag.
	DefaultClusterUID = ""
)

// LoadBalancerController watches the kubernetes api and adds/removes services
// from the loadbalancer, via loadBalancerConfig.
type LoadBalancerController struct {
	client              *client.Client
	ingController       *framework.Controller
	nodeController      *framework.Controller
	svcController       *framework.Controller
	ingLister           StoreToIngressLister
	nodeLister          cache.StoreToNodeLister
	svcLister           cache.StoreToServiceLister
	CloudClusterManager *ClusterManager
	recorder            record.EventRecorder
	nodeQueue           *taskQueue
	ingQueue            *taskQueue
	tr                  *GCETranslator
	stopCh              chan struct{}
	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
	// tlsLoader loads secrets from the Kubernetes apiserver for Ingresses.
	tlsLoader tlsLoader
}

// NewLoadBalancerController creates a controller for gce loadbalancers.
// - kubeClient: A kubernetes REST client.
// - clusterManager: A ClusterManager capable of creating all cloud resources
//	 required for L7 loadbalancing.
// - resyncPeriod: Watchers relist from the Kubernetes API server this often.
func NewLoadBalancerController(kubeClient *client.Client, clusterManager *ClusterManager, resyncPeriod time.Duration, namespace string) (*LoadBalancerController, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(""))

	lbc := LoadBalancerController{
		client:              kubeClient,
		CloudClusterManager: clusterManager,
		stopCh:              make(chan struct{}),
		recorder: eventBroadcaster.NewRecorder(
			api.EventSource{Component: "loadbalancer-controller"}),
	}
	lbc.nodeQueue = NewTaskQueue(lbc.syncNodes)
	lbc.ingQueue = NewTaskQueue(lbc.sync)

	// Ingress watch handlers
	pathHandlers := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			lbc.recorder.Eventf(addIng, api.EventTypeNormal, "ADD", fmt.Sprintf("%s/%s", addIng.Namespace, addIng.Name))
			lbc.ingQueue.enqueue(obj)
		},
		DeleteFunc: lbc.ingQueue.enqueue,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				glog.V(3).Infof("Ingress %v changed, syncing",
					cur.(*extensions.Ingress).Name)
			}
			lbc.ingQueue.enqueue(cur)
		},
	}
	lbc.ingLister.Store, lbc.ingController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  ingressListFunc(lbc.client, namespace),
			WatchFunc: ingressWatchFunc(lbc.client, namespace),
		},
		&extensions.Ingress{}, resyncPeriod, pathHandlers)

	// Service watch handlers
	svcHandlers := framework.ResourceEventHandlerFuncs{
		AddFunc: lbc.enqueueIngressForService,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				lbc.enqueueIngressForService(cur)
			}
		},
		// Ingress deletes matter, service deletes don't.
	}

	lbc.svcLister.Store, lbc.svcController = framework.NewInformer(
		cache.NewListWatchFromClient(
			lbc.client, "services", namespace, fields.Everything()),
		&api.Service{}, resyncPeriod, svcHandlers)

	nodeHandlers := framework.ResourceEventHandlerFuncs{
		AddFunc:    lbc.nodeQueue.enqueue,
		DeleteFunc: lbc.nodeQueue.enqueue,
		// Nodes are updated every 10s and we don't care, so no update handler.
	}

	// Node watch handlers
	lbc.nodeLister.Store, lbc.nodeController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(opts api.ListOptions) (runtime.Object, error) {
				return lbc.client.Get().
					Resource("nodes").
					FieldsSelectorParam(fields.Everything()).
					Do().
					Get()
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				return lbc.client.Get().
					Prefix("watch").
					Resource("nodes").
					FieldsSelectorParam(fields.Everything()).
					Param("resourceVersion", options.ResourceVersion).Watch()
			},
		},
		&api.Node{}, 0, nodeHandlers)

	lbc.tr = &GCETranslator{&lbc}
	lbc.tlsLoader = &apiServerTLSLoader{client: lbc.client}
	glog.V(3).Infof("Created new loadbalancer controller")

	return &lbc, nil
}

func ingressListFunc(c *client.Client, ns string) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.Extensions().Ingress(ns).List(opts)
	}
}

func ingressWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return c.Extensions().Ingress(ns).Watch(options)
	}
}

// enqueueIngressForService enqueues all the Ingress' for a Service.
func (lbc *LoadBalancerController) enqueueIngressForService(obj interface{}) {
	svc := obj.(*api.Service)
	ings, err := lbc.ingLister.GetServiceIngress(svc)
	if err != nil {
		glog.V(5).Infof("ignoring service %v: %v", svc.Name, err)
		return
	}
	for _, ing := range ings {
		lbc.ingQueue.enqueue(&ing)
	}
}

// Run starts the loadbalancer controller.
func (lbc *LoadBalancerController) Run() {
	glog.Infof("Starting loadbalancer controller")
	go lbc.ingController.Run(lbc.stopCh)
	go lbc.nodeController.Run(lbc.stopCh)
	go lbc.svcController.Run(lbc.stopCh)
	go lbc.ingQueue.run(time.Second, lbc.stopCh)
	go lbc.nodeQueue.run(time.Second, lbc.stopCh)
	<-lbc.stopCh
	glog.Infof("Shutting down Loadbalancer Controller")
}

// Stop stops the loadbalancer controller. It also deletes cluster resources
// if deleteAll is true.
func (lbc *LoadBalancerController) Stop(deleteAll bool) error {
	// Stop is invoked from the http endpoint.
	lbc.stopLock.Lock()
	defer lbc.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !lbc.shutdown {
		close(lbc.stopCh)
		glog.Infof("Shutting down controller queues.")
		lbc.ingQueue.shutdown()
		lbc.nodeQueue.shutdown()
		lbc.shutdown = true
	}

	// Deleting shared cluster resources is idempotent.
	if deleteAll {
		glog.Infof("Shutting down cluster manager.")
		return lbc.CloudClusterManager.shutdown()
	}
	return nil
}

// sync manages Ingress create/updates/deletes.
func (lbc *LoadBalancerController) sync(key string) {
	glog.V(3).Infof("Syncing %v", key)

	paths, err := lbc.ingLister.List()
	if err != nil {
		lbc.ingQueue.requeue(key, err)
		return
	}
	nodePorts := lbc.tr.toNodePorts(&paths)
	lbNames := lbc.ingLister.Store.ListKeys()
	lbs, _ := lbc.ListRuntimeInfo()
	nodeNames, err := lbc.getReadyNodeNames()
	if err != nil {
		lbc.ingQueue.requeue(key, err)
		return
	}
	obj, ingExists, err := lbc.ingLister.Store.GetByKey(key)
	if err != nil {
		lbc.ingQueue.requeue(key, err)
		return
	}

	// This performs a 2 phase checkpoint with the cloud:
	// * Phase 1 creates/verifies resources are as expected. At the end of a
	//   successful checkpoint we know that existing L7s are WAI, and the L7
	//   for the Ingress associated with "key" is ready for a UrlMap update.
	//   If this encounters an error, eg for quota reasons, we want to invoke
	//   Phase 2 right away and retry checkpointing.
	// * Phase 2 performs GC by refcounting shared resources. This needs to
	//   happen periodically whether or not stage 1 fails. At the end of a
	//   successful GC we know that there are no dangling cloud resources that
	//   don't have an associated Kubernetes Ingress/Service/Endpoint.

	defer func() {
		if err := lbc.CloudClusterManager.GC(lbNames, nodePorts); err != nil {
			lbc.ingQueue.requeue(key, err)
		}
		glog.V(3).Infof("Finished syncing %v", key)
	}()

	// Record any errors during sync and throw a single error at the end. This
	// allows us to free up associated cloud resources ASAP.
	var syncError error
	if err := lbc.CloudClusterManager.Checkpoint(lbs, nodeNames, nodePorts); err != nil {
		// TODO: Implement proper backoff for the queue.
		eventMsg := "GCE"
		if utils.IsHTTPErrorCode(err, http.StatusForbidden) {
			eventMsg += " :Quota"
		}
		if ingExists {
			lbc.recorder.Eventf(obj.(*extensions.Ingress), api.EventTypeWarning, eventMsg, err.Error())
		} else {
			err = fmt.Errorf("%v Error: %v", eventMsg, err)
		}
		syncError = err
	}

	if !ingExists {
		if syncError != nil {
			lbc.ingQueue.requeue(key, err)
		}
		return
	}
	// Update the UrlMap of the single loadbalancer that came through the watch.
	l7, err := lbc.CloudClusterManager.l7Pool.Get(key)
	if err != nil {
		lbc.ingQueue.requeue(key, fmt.Errorf("%v, unable to get loadbalancer: %v", syncError, err))
		return
	}

	ing := *obj.(*extensions.Ingress)
	if urlMap, err := lbc.tr.toUrlMap(&ing); err != nil {
		syncError = fmt.Errorf("%v, convert to url map error %v", syncError, err)
	} else if err := l7.UpdateUrlMap(urlMap); err != nil {
		lbc.recorder.Eventf(&ing, api.EventTypeWarning, "UrlMap", err.Error())
		syncError = fmt.Errorf("%v, update url map error: %v", syncError, err)
	} else if lbc.updateIngressStatus(l7, ing); err != nil {
		lbc.recorder.Eventf(&ing, api.EventTypeWarning, "Status", err.Error())
		syncError = fmt.Errorf("%v, update ingress error: %v", syncError, err)
	}
	if syncError != nil {
		lbc.ingQueue.requeue(key, syncError)
	}
	return
}

// updateIngressStatus updates the IP and annotations of a loadbalancer.
// The annotations are parsed by kubectl describe.
func (lbc *LoadBalancerController) updateIngressStatus(l7 *loadbalancers.L7, ing extensions.Ingress) error {
	ingClient := lbc.client.Extensions().Ingress(ing.Namespace)

	// Update IP through update/status endpoint
	ip := l7.GetIP()
	currIng, err := ingClient.Get(ing.Name)
	if err != nil {
		return err
	}
	currIng.Status = extensions.IngressStatus{
		LoadBalancer: api.LoadBalancerStatus{
			Ingress: []api.LoadBalancerIngress{
				{IP: ip},
			},
		},
	}
	lbIPs := ing.Status.LoadBalancer.Ingress
	if len(lbIPs) == 0 && ip != "" || lbIPs[0].IP != ip {
		// TODO: If this update fails it's probably resource version related,
		// which means it's advantageous to retry right away vs requeuing.
		glog.Infof("Updating loadbalancer %v/%v with IP %v", ing.Namespace, ing.Name, ip)
		if _, err := ingClient.UpdateStatus(currIng); err != nil {
			return err
		}
		lbc.recorder.Eventf(currIng, api.EventTypeNormal, "CREATE", "ip: %v", ip)
	}

	// Update annotations through /update endpoint
	currIng, err = ingClient.Get(ing.Name)
	if err != nil {
		return err
	}
	currIng.Annotations = loadbalancers.GetLBAnnotations(l7, currIng.Annotations, lbc.CloudClusterManager.backendPool)
	if !reflect.DeepEqual(ing.Annotations, currIng.Annotations) {
		glog.V(3).Infof("Updating annotations of %v/%v", ing.Namespace, ing.Name)
		if _, err := ingClient.Update(currIng); err != nil {
			return err
		}
	}
	return nil
}

// ListRuntimeInfo lists L7RuntimeInfo as understood by the loadbalancer module.
func (lbc *LoadBalancerController) ListRuntimeInfo() (lbs []*loadbalancers.L7RuntimeInfo, err error) {
	for _, m := range lbc.ingLister.Store.List() {
		ing := m.(*extensions.Ingress)
		k, err := keyFunc(ing)
		if err != nil {
			glog.Warningf("Cannot get key for Ingress %v/%v: %v", ing.Namespace, ing.Name, err)
			continue
		}
		tls, err := lbc.tlsLoader.load(ing)
		if err != nil {
			glog.Warningf("Cannot get certs for Ingress %v/%v: %v", ing.Namespace, ing.Name, err)
		}
		annotations := ingAnnotations(ing.ObjectMeta.Annotations)
		lbs = append(lbs, &loadbalancers.L7RuntimeInfo{
			Name:         k,
			TLS:          tls,
			AllowHTTP:    annotations.allowHTTP(),
			StaticIPName: annotations.staticIPName(),
		})
	}
	return lbs, nil
}

// syncNodes manages the syncing of kubernetes nodes to gce instance groups.
// The instancegroups are referenced by loadbalancer backends.
func (lbc *LoadBalancerController) syncNodes(key string) {
	nodeNames, err := lbc.getReadyNodeNames()
	if err != nil {
		lbc.nodeQueue.requeue(key, err)
		return
	}
	if err := lbc.CloudClusterManager.instancePool.Sync(nodeNames); err != nil {
		lbc.nodeQueue.requeue(key, err)
	}
	return
}

func nodeReady(node api.Node) bool {
	for ix := range node.Status.Conditions {
		condition := &node.Status.Conditions[ix]
		if condition.Type == api.NodeReady {
			return condition.Status == api.ConditionTrue
		}
	}
	return false
}

// getReadyNodeNames returns names of schedulable, ready nodes from the node lister.
func (lbc *LoadBalancerController) getReadyNodeNames() ([]string, error) {
	nodeNames := []string{}
	nodes, err := lbc.nodeLister.NodeCondition(nodeReady).List()
	if err != nil {
		return nodeNames, err
	}
	for _, n := range nodes.Items {
		if n.Spec.Unschedulable {
			continue
		}
		nodeNames = append(nodeNames, n.Name)
	}
	return nodeNames, nil
}
