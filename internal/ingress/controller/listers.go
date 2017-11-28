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
	cache_client "k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
)

type cacheController struct {
	Ingress   cache.Controller
	Endpoint  cache.Controller
	Service   cache.Controller
	Secret    cache.Controller
	Configmap cache.Controller
}

func (c *cacheController) Run(stopCh chan struct{}) {
	go c.Ingress.Run(stopCh)
	go c.Endpoint.Run(stopCh)
	go c.Service.Run(stopCh)
	go c.Secret.Run(stopCh)
	go c.Configmap.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh,
		c.Ingress.HasSynced,
		c.Endpoint.HasSynced,
		c.Service.HasSynced,
		c.Secret.HasSynced,
		c.Configmap.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}
}

func (n *NGINXController) createListers(stopCh chan struct{}) (*ingress.StoreLister, *cacheController) {
	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			if !class.IsValid(addIng) {
				a := addIng.GetAnnotations()[class.IngressKey]
				glog.Infof("ignoring add for ingress %v based on annotation %v with value %v", addIng.Name, class.IngressKey, a)
				return
			}

			n.extractAnnotations(addIng)
			n.recorder.Eventf(addIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", addIng.Namespace, addIng.Name))
			n.syncQueue.Enqueue(obj)
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
			if !class.IsValid(delIng) {
				glog.Infof("ignoring delete for ingress %v based on annotation %v", delIng.Name, class.IngressKey)
				return
			}
			n.recorder.Eventf(delIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", delIng.Namespace, delIng.Name))
			n.listers.IngressAnnotation.Delete(delIng)
			n.syncQueue.Enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			oldIng := old.(*extensions.Ingress)
			curIng := cur.(*extensions.Ingress)
			validOld := class.IsValid(oldIng)
			validCur := class.IsValid(curIng)

			c := curIng.GetAnnotations()[class.IngressKey]
			if !validOld && validCur {
				glog.Infof("creating ingress %v/%v based on annotation %v with value '%v'", curIng.Namespace, curIng.Name, class.IngressKey, c)
				n.recorder.Eventf(curIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validOld && !validCur {
				glog.Infof("removing ingress %v/%v based on annotation %v with value '%v'", curIng.Namespace, curIng.Name, class.IngressKey, c)
				n.recorder.Eventf(curIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validCur && !reflect.DeepEqual(old, cur) {
				n.recorder.Eventf(curIng, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			}

			n.extractAnnotations(curIng)
			n.syncQueue.Enqueue(cur)
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				sec := cur.(*apiv1.Secret)
				key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)
				_, exists := n.sslCertTracker.Get(key)
				if exists {
					n.syncSecret(key)
				}
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
			n.sslCertTracker.Delete(key)
			n.syncQueue.Enqueue(key)
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n.syncQueue.Enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			n.syncQueue.Enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*apiv1.Endpoints)
			ocur := cur.(*apiv1.Endpoints)
			if !reflect.DeepEqual(ocur.Subsets, oep.Subsets) {
				n.syncQueue.Enqueue(cur)
			}
		},
	}

	mapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			upCmap := obj.(*apiv1.ConfigMap)
			mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
			if mapKey == n.cfg.ConfigMapName {
				glog.V(2).Infof("adding configmap %v to backend", mapKey)
				n.SetConfig(upCmap)
				n.SetForceReload(true)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				upCmap := cur.(*apiv1.ConfigMap)
				mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
				if mapKey == n.cfg.ConfigMapName {
					glog.V(2).Infof("updating configmap backend (%v)", mapKey)
					n.SetConfig(upCmap)
					n.SetForceReload(true)
				}
				// updates to configuration configmaps can trigger an update
				if mapKey == n.cfg.ConfigMapName || mapKey == n.cfg.TCPConfigMapName || mapKey == n.cfg.UDPConfigMapName {
					n.recorder.Eventf(upCmap, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", mapKey))
					n.syncQueue.Enqueue(cur)
				}
			}
		},
	}

	watchNs := apiv1.NamespaceAll
	if n.cfg.ForceNamespaceIsolation && n.cfg.Namespace != apiv1.NamespaceAll {
		watchNs = n.cfg.Namespace
	}

	lister := &ingress.StoreLister{}
	lister.IngressAnnotation.Store = cache_client.NewStore(cache_client.DeletionHandlingMetaNamespaceKeyFunc)

	controller := &cacheController{}

	lister.Ingress.Store, controller.Ingress = cache.NewInformer(
		cache.NewListWatchFromClient(n.cfg.Client.ExtensionsV1beta1().RESTClient(), "ingresses", n.cfg.Namespace, fields.Everything()),
		&extensions.Ingress{}, n.cfg.ResyncPeriod, ingEventHandler)

	lister.Endpoint.Store, controller.Endpoint = cache.NewInformer(
		cache.NewListWatchFromClient(n.cfg.Client.CoreV1().RESTClient(), "endpoints", n.cfg.Namespace, fields.Everything()),
		&apiv1.Endpoints{}, n.cfg.ResyncPeriod, eventHandler)

	lister.Secret.Store, controller.Secret = cache.NewInformer(
		cache.NewListWatchFromClient(n.cfg.Client.CoreV1().RESTClient(), "secrets", watchNs, fields.Everything()),
		&apiv1.Secret{}, n.cfg.ResyncPeriod, secrEventHandler)

	lister.ConfigMap.Store, controller.Configmap = cache.NewInformer(
		cache.NewListWatchFromClient(n.cfg.Client.CoreV1().RESTClient(), "configmaps", watchNs, fields.Everything()),
		&apiv1.ConfigMap{}, n.cfg.ResyncPeriod, mapEventHandler)

	lister.Service.Store, controller.Service = cache.NewInformer(
		cache.NewListWatchFromClient(n.cfg.Client.CoreV1().RESTClient(), "services", n.cfg.Namespace, fields.Everything()),
		&apiv1.Service{}, n.cfg.ResyncPeriod, cache.ResourceEventHandlerFuncs{})

	return lister, controller
}
