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

package networkendpointgroup

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	unversionedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	glbc "k8s.io/ingress/controllers/gce/controller"
	"k8s.io/ingress/controllers/gce/utils"
)

const (
	maxRetries = 15
)

// network endpoint group controller
type Controller struct {
	manager      SyncerManager
	resyncPeriod time.Duration

	ingressSynced  cache.InformerSynced
	serviceSynced  cache.InformerSynced
	endpointSynced cache.InformerSynced
	ingressLister  glbc.StoreToIngressLister
	serviceLister  cache.Indexer

	// serviceQueue takes service key as work item. Service key with format "namespace/name".
	serviceQueue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubernetes.Interface,
	cloud NetworkEndpointGroupCloud,
	ctx *utils.ControllerContext,
	zoneGetter ZoneGetter,
	namer NetworkEndpointGroupNamer,
	resyncPeriod time.Duration,
) (*Controller, error) {
	// init event recorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&unversionedcore.EventSinkImpl{
		Interface: kubeClient.Core().Events(""),
	})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme,
		apiv1.EventSource{Component: "networkendpointgroup-controller"})

	manager := newSyncerManager(namer,
		recorder,
		cloud,
		zoneGetter,
		ctx.ServiceInformer.GetIndexer(),
		ctx.EndpointInformer.GetIndexer())

	negController := &Controller{
		manager:        manager,
		resyncPeriod:   resyncPeriod,
		ingressSynced:  ctx.IngressInformer.HasSynced,
		serviceSynced:  ctx.ServiceInformer.HasSynced,
		endpointSynced: ctx.EndpointInformer.HasSynced,
		ingressLister:  glbc.StoreToIngressLister{ctx.IngressInformer.GetIndexer()},
		serviceLister:  ctx.ServiceInformer.GetIndexer(),
		serviceQueue:   workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	ctx.ServiceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    negController.enqueueService,
		DeleteFunc: negController.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			negController.enqueueService(cur)
		},
	})

	ctx.IngressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    negController.enqueueIngressServices,
		DeleteFunc: negController.enqueueIngressServices,
		UpdateFunc: func(old, cur interface{}) {
			keys := gatherIngressServiceKeys(old)
			keys = keys.Union(gatherIngressServiceKeys(cur))
			for _, key := range keys.List() {
				negController.enqueueService(cache.ExplicitKey(key))
			}
		},
	})

	ctx.EndpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    negController.processEndpoint,
		DeleteFunc: negController.processEndpoint,
		UpdateFunc: func(old, cur interface{}) {
			negController.processEndpoint(cur)
		},
	})
	return negController, nil
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	wait.PollUntil(5*time.Second, func() (bool, error) {
		return c.hasSync(), nil
	}, stopCh)

	glog.V(2).Infof("Starting network endpoint group controller")
	defer c.stop()
	defer glog.V(2).Infof("Shutting down network endpoint group controller")

	go wait.Until(c.serviceWorker, time.Second, stopCh)
	go wait.Until(c.garbageCollection, c.resyncPeriod, stopCh)

	<-stopCh
}

func (c *Controller) stop() {
	c.serviceQueue.ShutDown()
	c.manager.ShutDown()
}

// processEndpoint finds the related syncers and signal it to sync
func (c *Controller) processEndpoint(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Failed to generate endpoint key: %v", err)
		return
	}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return
	}
	c.manager.Sync(namespace, name)
}

func (c *Controller) serviceWorker() {
	for {
		func() {
			key, quit := c.serviceQueue.Get()
			if quit {
				return
			}
			defer c.serviceQueue.Done(key)
			err := c.processService(key.(string))
			c.handleErr(err, key)
		}()
	}
}

// processService takes a service and determines whether it needs NEGs or not.
func (c *Controller) processService(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	svc, exists, err := c.serviceLister.GetByKey(key)
	if err != nil {
		return err
	}

	var service *apiv1.Service
	var enabled bool
	if exists {
		service = svc.(*apiv1.Service)
		enabled = utils.NEGEnabled(service.Annotations)
	}

	// if service is deleted or neg is not enabled
	if !exists || !enabled {
		c.manager.StopSyncer(namespace, name)
		return nil
	}

	glog.V(2).Infof("Syncing service %q", key)
	ings, err := c.ingressLister.GetServiceIngress(service)
	if err != nil {
		return err
	}
	c.manager.EnsureSyncer(namespace, name, gatherSerivceTargetPortUsedByIngress(ings, service))
	return nil
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.serviceQueue.Forget(key)
		return
	}

	glog.Errorf("Error processing service %q: %v", key, err)
	if c.serviceQueue.NumRequeues(key) < maxRetries {
		c.serviceQueue.AddRateLimited(key)
		return
	}

	glog.Warningf("Dropping service %q out of the queue: %v", key, err)
	c.serviceQueue.Forget(key)
}

func (c *Controller) enqueueService(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Failed to generate service key: %v", err)
		return
	}
	c.serviceQueue.Add(key)
}

func (c *Controller) enqueueIngressServices(obj interface{}) {
	keys := gatherIngressServiceKeys(obj)
	for key := range keys {
		c.enqueueService(cache.ExplicitKey(key))
	}
}

func (c *Controller) garbageCollection() {
	if err := c.manager.GC(); err != nil {
		glog.Errorf("NEG controller garbage collection failed: %v", err)
	}
}

func (c *Controller) hasSync() bool {
	return c.endpointSynced() &&
		c.serviceSynced() &&
		c.ingressSynced()
}

// gatherSerivceTargetPortUsedByIngress returns all target ports of the service that are referenced by ingresses
func gatherSerivceTargetPortUsedByIngress(ings []extensions.Ingress, svc *apiv1.Service) sets.String {
	servicePorts := sets.NewInt()
	targetPorts := sets.NewString()
	for _, ing := range ings {
		if ing.Spec.Backend.ServiceName == svc.Name {
			servicePorts.Insert(ing.Spec.Backend.ServicePort.IntValue())
		}
	}
	for _, ing := range ings {
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.IngressRuleValue.HTTP.Paths {
				if path.Backend.ServiceName == svc.Name {
					servicePorts.Insert(path.Backend.ServicePort.IntValue())
				}
			}
		}
	}
	for _, svcPort := range svc.Spec.Ports {
		if servicePorts.Has(int(svcPort.Port)) {
			targetPorts.Insert(svcPort.TargetPort.String())
		}
	}

	return targetPorts
}

// gatherIngressServiceKeys returns all service key (formatted as namespace/name) referenced in the ingress
func gatherIngressServiceKeys(obj interface{}) sets.String {
	set := sets.NewString()
	ing, ok := obj.(*extensions.Ingress)
	if !ok {
		glog.Errorf("Expecting ingress type but got: %+v", ing)
		return set
	}
	set.Insert(serviceKeyFunc(ing.Namespace, ing.Spec.Backend.ServiceName))

	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			set.Insert(serviceKeyFunc(ing.Namespace, path.Backend.ServiceName))
		}
	}
	return set
}

func serviceKeyFunc(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
