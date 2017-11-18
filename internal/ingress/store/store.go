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

package store

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	cache_client "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

// Storer is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Storer interface {
	GetConfigMap(key string) (*apiv1.ConfigMap, error)

	// GetSecret returns a Secret using the namespace and name as key
	GetSecret(key string) (*apiv1.Secret, error)

	// GetService returns a Service using the namespace and name as key
	GetService(key string) (*apiv1.Service, error)

	GetServiceEndpoints(svc *apiv1.Service) (*apiv1.Endpoints, error)

	// GetSecret returns an Ingress using the namespace and name as key
	GetIngress(key string) (*extensions.Ingress, error)

	// ListIngresses returns the list of Ingresses
	ListIngresses() []*extensions.Ingress

	// GetIngressAnnotations returns the annotations associated to an Ingress
	GetIngressAnnotations(ing *extensions.Ingress) (*annotations.Ingress, error)

	// GetLocalSecret returns the local copy of a Secret
	GetLocalSecret(name string) (*ingress.SSLCert, error)

	// ListLocalSecrets returns the list of local Secrets
	ListLocalSecrets() []*ingress.SSLCert

	// StartSync initiates the synchronization of the controllers
	StartSync(stopCh chan struct{})
}

// lister returns the stores for ingresses, services, endpoints, secrets and configmaps.
type lister struct {
	Ingress           IngressLister
	Service           ServiceLister
	Endpoint          EndpointLister
	Secret            SecretLister
	ConfigMap         ConfigMapLister
	IngressAnnotation IngressAnnotationsLister
}

// controller defines the required controllers that interact agains the api server
type controller struct {
	Ingress   cache.Controller
	Endpoint  cache.Controller
	Service   cache.Controller
	Secret    cache.Controller
	Configmap cache.Controller
}

// Run initiates the synchronization of the controllers against the api server
func (c *controller) Run(stopCh chan struct{}) {
	go c.Ingress.Run(stopCh)
	go c.Endpoint.Run(stopCh)
	go c.Service.Run(stopCh)
	go c.Secret.Run(stopCh)
	go c.Configmap.Run(stopCh)

	// wait for all involved caches to be synced, before processing items
	// from the queue is started
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

type k8sStore struct {
	cache *controller
	// listers
	listers *lister

	// sslStore local store of SSL certificates (certificates used in ingress)
	sslStore *SSLCertTracker

	// annotations parser
	annotations annotations.Extractor
}

// New creates a new object store to be used in the ingress controller
func New(namespace, configmap, tcp, udp string,
	resyncPeriod time.Duration,
	recorder record.EventRecorder,
	client clientset.Interface,
	annotations annotations.Extractor,
	r resolver.Resolver,
	updateCh chan ingress.Event) Storer {

	store := &k8sStore{
		cache:       &controller{},
		listers:     &lister{},
		sslStore:    NewSSLCertTracker(),
		annotations: annotations,
	}

	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			if !class.IsValid(addIng, ingress.IngressClass, ingress.DefaultIngressClass) {
				a, _ := parser.GetStringAnnotation(class.IngressKey, addIng, r)
				glog.Infof("ignoring add for ingress %v based on annotation %v with value %v", addIng.Name, class.IngressKey, a)
				return
			}

			store.extractAnnotations(addIng)
			recorder.Eventf(addIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", addIng.Namespace, addIng.Name))
			updateCh <- ingress.Event{
				Type: ingress.CreateEvent,
				Obj:  obj,
			}
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
			if !class.IsValid(delIng, ingress.IngressClass, ingress.DefaultIngressClass) {
				glog.Infof("ignoring delete for ingress %v based on annotation %v", delIng.Name, class.IngressKey)
				return
			}
			recorder.Eventf(delIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", delIng.Namespace, delIng.Name))
			store.listers.IngressAnnotation.Delete(delIng)
			updateCh <- ingress.Event{
				Type: ingress.DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oldIng := old.(*extensions.Ingress)
			curIng := cur.(*extensions.Ingress)
			validOld := class.IsValid(oldIng, ingress.IngressClass, ingress.DefaultIngressClass)
			validCur := class.IsValid(curIng, ingress.IngressClass, ingress.DefaultIngressClass)
			if !validOld && validCur {
				glog.Infof("creating ingress %v based on annotation %v", curIng.Name, class.IngressKey)
				recorder.Eventf(curIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validOld && !validCur {
				glog.Infof("removing ingress %v based on annotation %v", curIng.Name, class.IngressKey)
				recorder.Eventf(curIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validCur && !reflect.DeepEqual(old, cur) {
				recorder.Eventf(curIng, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			}

			store.extractAnnotations(curIng)
			updateCh <- ingress.Event{
				Type: ingress.UpdateEvent,
				Obj:  cur,
			}
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				sec := cur.(*apiv1.Secret)
				key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)
				_, exists := store.sslStore.Get(key)
				if exists {
					updateCh <- ingress.Event{
						Type: ingress.UpdateEvent,
						Obj:  cur,
					}
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
			store.sslStore.Delete(key)
			updateCh <- ingress.Event{
				Type: ingress.DeleteEvent,
				Obj:  obj,
			}
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateCh <- ingress.Event{
				Type: ingress.CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: func(obj interface{}) {
			updateCh <- ingress.Event{
				Type: ingress.DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*apiv1.Endpoints)
			ocur := cur.(*apiv1.Endpoints)
			if !reflect.DeepEqual(ocur.Subsets, oep.Subsets) {
				updateCh <- ingress.Event{
					Type: ingress.UpdateEvent,
					Obj:  cur,
				}
			}
		},
	}

	mapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			upCmap := obj.(*apiv1.ConfigMap)
			mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
			if mapKey == configmap {
				glog.V(2).Infof("adding configmap %v to backend", mapKey)
				updateCh <- ingress.Event{
					Type: ingress.CreateEvent,
					Obj:  obj,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				upCmap := cur.(*apiv1.ConfigMap)
				mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
				if mapKey == configmap {
					glog.V(2).Infof("updating configmap backend (%v)", mapKey)
					updateCh <- ingress.Event{
						Type: ingress.UpdateEvent,
						Obj:  cur,
					}
				}
				// updates to configuration configmaps can trigger an update
				if mapKey == configmap || mapKey == tcp || mapKey == udp {
					recorder.Eventf(upCmap, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", mapKey))
					updateCh <- ingress.Event{
						Type: ingress.UpdateEvent,
						Obj:  cur,
					}
				}
			}
		},
	}

	store.listers.IngressAnnotation.Store = cache_client.NewStore(cache_client.DeletionHandlingMetaNamespaceKeyFunc)

	store.listers.Ingress.Store, store.cache.Ingress = cache.NewInformer(
		cache.NewListWatchFromClient(client.ExtensionsV1beta1().RESTClient(), "ingresses", namespace, fields.Everything()),
		&extensions.Ingress{}, resyncPeriod, ingEventHandler)

	store.listers.Endpoint.Store, store.cache.Endpoint = cache.NewInformer(
		cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "endpoints", namespace, fields.Everything()),
		&apiv1.Endpoints{}, resyncPeriod, eventHandler)

	store.listers.Secret.Store, store.cache.Secret = cache.NewInformer(
		cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "secrets", namespace, fields.Everything()),
		&apiv1.Secret{}, resyncPeriod, secrEventHandler)

	store.listers.ConfigMap.Store, store.cache.Configmap = cache.NewInformer(
		cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "configmaps", namespace, fields.Everything()),
		&apiv1.ConfigMap{}, resyncPeriod, mapEventHandler)

	store.listers.Service.Store, store.cache.Service = cache.NewInformer(
		cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "services", namespace, fields.Everything()),
		&apiv1.Service{}, resyncPeriod, cache.ResourceEventHandlerFuncs{})

	return store
}

func (s k8sStore) extractAnnotations(ing *extensions.Ingress) {
	anns := s.annotations.Extract(ing)
	glog.V(3).Infof("updating annotations information for ingres %v/%v", anns.Namespace, anns.Name)
	s.listers.IngressAnnotation.Update(anns)
}

// GetSecret returns a Secret using the namespace and name as key
func (s k8sStore) GetSecret(key string) (*apiv1.Secret, error) {
	return s.listers.Secret.GetByNamespaceName(key)
}

// ListLocalSecrets returns the list of local Secrets
func (s k8sStore) ListLocalSecrets() []*ingress.SSLCert {
	var certs []*ingress.SSLCert
	for _, item := range s.sslStore.List() {
		if s, ok := item.(*ingress.SSLCert); ok {
			certs = append(certs, s)
		}
	}

	return certs
}

// GetService returns a Service using the namespace and name as key
func (s k8sStore) GetService(key string) (*apiv1.Service, error) {
	return s.listers.Service.GetByNamespaceName(key)
}

// GetSecret returns an Ingress using the namespace and name as key
func (s k8sStore) GetIngress(key string) (*extensions.Ingress, error) {
	return s.listers.Ingress.GetByNamespaceName(key)
}

// ListIngresses returns the list of Ingresses
func (s k8sStore) ListIngresses() []*extensions.Ingress {
	// filter ingress rules
	var ingresses []*extensions.Ingress
	for _, item := range s.listers.Ingress.List() {
		ing := item.(*extensions.Ingress)
		if !class.IsValid(ing, ingress.IngressClass, ingress.DefaultIngressClass) {
			continue
		}

		ingresses = append(ingresses, ing)
	}

	return ingresses
}

// GetIngressAnnotations returns the annotations associated to an Ingress
func (s k8sStore) GetIngressAnnotations(ing *extensions.Ingress) (*annotations.Ingress, error) {
	key := fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
	item, exists, err := s.listers.IngressAnnotation.GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("unexpected error getting ingress annotation %v: %v", key, err)
	}
	if !exists {
		return nil, fmt.Errorf("ingress annotation %v was not found", key)
	}
	return item.(*annotations.Ingress), nil
}

// GetLocalSecret returns the local copy of a Secret
func (s k8sStore) GetLocalSecret(key string) (*ingress.SSLCert, error) {
	return s.sslStore.GetByNamespaceName(key)
}

func (s k8sStore) GetConfigMap(key string) (*apiv1.ConfigMap, error) {
	return s.listers.ConfigMap.GetByNamespaceName(key)
}

func (s k8sStore) GetServiceEndpoints(svc *apiv1.Service) (*apiv1.Endpoints, error) {
	return s.listers.Endpoint.GetServiceEndpoints(svc)
}

// StartSync initiates the synchronization of the controllers
func (s k8sStore) StartSync(stopCh chan struct{}) {
	// start controllers
	s.cache.Run(stopCh)
	// start goroutine to check for missing local secrets
	go wait.Until(s.checkMissingSecrets, 30*time.Second, stopCh)
}
