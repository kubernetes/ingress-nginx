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

package controller

import (
	"fmt"
	"reflect"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	fcache "k8s.io/client-go/tools/cache/testing"

	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/annotations/class"
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
)

type cacheController struct {
	Ingress   cache.Controller
	Endpoint  cache.Controller
	Service   cache.Controller
	Node      cache.Controller
	Secret    cache.Controller
	Configmap cache.Controller
}

func (c *cacheController) Run(stopCh chan struct{}) {
	go c.Ingress.Run(stopCh)
	go c.Endpoint.Run(stopCh)
	go c.Service.Run(stopCh)
	go c.Node.Run(stopCh)
	go c.Secret.Run(stopCh)
	go c.Configmap.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh,
		c.Ingress.HasSynced,
		c.Endpoint.HasSynced,
		c.Service.HasSynced,
		c.Node.HasSynced,
		c.Secret.HasSynced,
		c.Configmap.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}
}

func (ic *GenericController) createListers(disableNodeLister bool) (*ingress.StoreLister, *cacheController) {
	// from here to the end of the method all the code is just boilerplate
	// required to watch Ingress, Secrets, ConfigMaps and Endoints.
	// This is used to detect new content, updates or removals and act accordingly
	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			if !class.IsValid(addIng, ic.cfg.IngressClass, ic.cfg.DefaultIngressClass) {
				a, _ := parser.GetStringAnnotation(class.IngressKey, addIng)
				glog.Infof("ignoring add for ingress %v based on annotation %v with value %v", addIng.Name, class.IngressKey, a)
				return
			}
			ic.recorder.Eventf(addIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", addIng.Namespace, addIng.Name))
			ic.syncQueue.Enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			delIng, ok := obj.(*extensions.Ingress)
			if !ok {
				// If we reached here it means the ingress was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				delIng, ok = tombstone.Obj.(*extensions.Ingress)
				if !ok {
					glog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			if !class.IsValid(delIng, ic.cfg.IngressClass, ic.cfg.DefaultIngressClass) {
				glog.Infof("ignoring delete for ingress %v based on annotation %v", delIng.Name, class.IngressKey)
				return
			}
			ic.recorder.Eventf(delIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", delIng.Namespace, delIng.Name))
			ic.syncQueue.Enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			oldIng := old.(*extensions.Ingress)
			curIng := cur.(*extensions.Ingress)
			validOld := class.IsValid(oldIng, ic.cfg.IngressClass, ic.cfg.DefaultIngressClass)
			validCur := class.IsValid(curIng, ic.cfg.IngressClass, ic.cfg.DefaultIngressClass)
			if !validOld && validCur {
				glog.Infof("creating ingress %v based on annotation %v", curIng.Name, class.IngressKey)
				ic.recorder.Eventf(curIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validOld && !validCur {
				glog.Infof("removing ingress %v based on annotation %v", curIng.Name, class.IngressKey)
				ic.recorder.Eventf(curIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validCur && !reflect.DeepEqual(old, cur) {
				ic.recorder.Eventf(curIng, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			}

			ic.syncQueue.Enqueue(cur)
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				sec := cur.(*apiv1.Secret)
				key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)
				ic.syncSecret(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			sec, ok := obj.(*apiv1.Secret)
			if !ok {
				// If we reached here it means the secret was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				sec, ok = tombstone.Obj.(*apiv1.Secret)
				if !ok {
					glog.Errorf("Tombstone contained object that is not a Secret: %#v", obj)
					return
				}
			}
			key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)
			ic.sslCertTracker.DeleteAll(key)
			ic.syncQueue.Enqueue(key)
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ic.syncQueue.Enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			ic.syncQueue.Enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*apiv1.Endpoints)
			ocur := cur.(*apiv1.Endpoints)
			if !reflect.DeepEqual(ocur.Subsets, oep.Subsets) {
				ic.syncQueue.Enqueue(cur)
			}
		},
	}

	mapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			upCmap := obj.(*apiv1.ConfigMap)
			mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
			if mapKey == ic.cfg.ConfigMapName {
				glog.V(2).Infof("adding configmap %v to backend", mapKey)
				ic.cfg.Backend.SetConfig(upCmap)
				ic.setForceReload(true)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				upCmap := cur.(*apiv1.ConfigMap)
				mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
				if mapKey == ic.cfg.ConfigMapName {
					glog.V(2).Infof("updating configmap backend (%v)", mapKey)
					ic.cfg.Backend.SetConfig(upCmap)
					ic.setForceReload(true)
				}
				// updates to configuration configmaps can trigger an update
				if mapKey == ic.cfg.ConfigMapName || mapKey == ic.cfg.TCPConfigMapName || mapKey == ic.cfg.UDPConfigMapName {
					ic.recorder.Eventf(upCmap, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", mapKey))
					ic.syncQueue.Enqueue(cur)
				}
			}
		},
	}

	watchNs := apiv1.NamespaceAll
	if ic.cfg.ForceNamespaceIsolation && ic.cfg.Namespace != apiv1.NamespaceAll {
		watchNs = ic.cfg.Namespace
	}

	lister := &ingress.StoreLister{}

	controller := &cacheController{}

	lister.Ingress.Store, controller.Ingress = cache.NewInformer(
		cache.NewListWatchFromClient(ic.cfg.Client.ExtensionsV1beta1().RESTClient(), "ingresses", ic.cfg.Namespace, fields.Everything()),
		&extensions.Ingress{}, ic.cfg.ResyncPeriod, ingEventHandler)

	lister.Endpoint.Store, controller.Endpoint = cache.NewInformer(
		cache.NewListWatchFromClient(ic.cfg.Client.CoreV1().RESTClient(), "endpoints", ic.cfg.Namespace, fields.Everything()),
		&apiv1.Endpoints{}, ic.cfg.ResyncPeriod, eventHandler)

	lister.Secret.Store, controller.Secret = cache.NewInformer(
		cache.NewListWatchFromClient(ic.cfg.Client.CoreV1().RESTClient(), "secrets", watchNs, fields.Everything()),
		&apiv1.Secret{}, ic.cfg.ResyncPeriod, secrEventHandler)

	lister.ConfigMap.Store, controller.Configmap = cache.NewInformer(
		cache.NewListWatchFromClient(ic.cfg.Client.CoreV1().RESTClient(), "configmaps", watchNs, fields.Everything()),
		&apiv1.ConfigMap{}, ic.cfg.ResyncPeriod, mapEventHandler)

	lister.Service.Store, controller.Service = cache.NewInformer(
		cache.NewListWatchFromClient(ic.cfg.Client.CoreV1().RESTClient(), "services", ic.cfg.Namespace, fields.Everything()),
		&apiv1.Service{}, ic.cfg.ResyncPeriod, cache.ResourceEventHandlerFuncs{})

	var nodeListerWatcher cache.ListerWatcher
	if disableNodeLister {
		nodeListerWatcher = fcache.NewFakeControllerSource()
	} else {
		nodeListerWatcher = cache.NewListWatchFromClient(ic.cfg.Client.CoreV1().RESTClient(), "nodes", apiv1.NamespaceAll, fields.Everything())
	}
	lister.Node.Store, controller.Node = cache.NewInformer(
		nodeListerWatcher,
		&apiv1.Node{}, ic.cfg.ResyncPeriod, cache.ResourceEventHandlerFuncs{})

	return lister, controller
}
