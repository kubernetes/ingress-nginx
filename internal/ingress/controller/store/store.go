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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	cache_client "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	ngx_template "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
)

// Storer is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Storer interface {
	// GetBackendConfiguration returns the nginx configuration stored in a configmap
	GetBackendConfiguration() ngx_config.Configuration

	// GetConfigMap returns a ConfigmMap using the namespace and name as key
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

	// GetAuthCertificate resolves a given secret name into an SSL certificate.
	// The secret must contain 3 keys named:
	//   ca.crt: contains the certificate chain used for authentication
	GetAuthCertificate(string) (*resolver.AuthSSLCert, error)

	// GetDefaultBackend returns the default backend configuration
	GetDefaultBackend() defaults.Backend

	// Run initiates the synchronization of the controllers
	Run(stopCh chan struct{})

	// ReadSecrets extracts information about secrets from an Ingress rule
	ReadSecrets(*extensions.Ingress)
}

// EventType type of event associated with an informer
type EventType string

const (
	// CreateEvent event associated with new objects in an informer
	CreateEvent EventType = "CREATE"
	// UpdateEvent event associated with an object update in an informer
	UpdateEvent EventType = "UPDATE"
	// DeleteEvent event associated when an object is removed from an informer
	DeleteEvent EventType = "DELETE"
	// ConfigurationEvent event associated when a configuration object is created or updated
	ConfigurationEvent EventType = "CONFIGURATION"
)

// Event holds the context of an event
type Event struct {
	Type EventType
	Obj  interface{}
}

// Lister returns the stores for ingresses, services, endpoints, secrets and configmaps.
type Lister struct {
	Ingress           IngressLister
	Service           ServiceLister
	Endpoint          EndpointLister
	Secret            SecretLister
	ConfigMap         ConfigMapLister
	IngressAnnotation IngressAnnotationsLister
}

// Controller defines the required controllers that interact agains the api server
type Controller struct {
	Ingress   cache.Controller
	Endpoint  cache.Controller
	Service   cache.Controller
	Secret    cache.Controller
	Configmap cache.Controller
}

// Run initiates the synchronization of the controllers against the api server
func (c *Controller) Run(stopCh chan struct{}) {
	go c.Endpoint.Run(stopCh)
	go c.Service.Run(stopCh)
	go c.Secret.Run(stopCh)
	go c.Configmap.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh,
		c.Endpoint.HasSynced,
		c.Service.HasSynced,
		c.Secret.HasSynced,
		c.Configmap.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}

	// We need to wait before start syncing the ingress rules
	// because the rules requires content from other listers
	time.Sleep(1 * time.Second)
	go c.Ingress.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		c.Ingress.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}
}

// k8sStore internal Storer implementation using informers and thread safe stores
type k8sStore struct {
	isOCSPCheckEnabled bool

	// backendConfig contains the running configuration from the configmap
	// this is required because this rarely changes but is a very expensive
	// operation to execute in each OnUpdate invocation
	backendConfig ngx_config.Configuration

	// cache contains the cache Controllers
	cache *Controller

	// listers contains the cache.Store used in the ingress controller
	listers *Lister

	// sslStore local store of SSL certificates (certificates used in ingress)
	// this is required because the certificates must be present in the
	// container filesystem
	sslStore *SSLCertTracker

	annotations annotations.Extractor

	// secretIngressMap contains information about which ingress references a
	// secret in the annotations.
	secretIngressMap map[string]sets.String

	filesystem file.Filesystem

	// updateCh
	updateCh chan Event

	// mu mutex used to avoid simultaneous incovations to syncSecret
	mu *sync.Mutex

	defaultSSLCertificate string
}

// New creates a new object store to be used in the ingress controller
func New(checkOCSP bool,
	namespace, configmap, tcp, udp, defaultSSLCertificate string,
	resyncPeriod time.Duration,
	client clientset.Interface,
	fs file.Filesystem,
	updateCh chan Event) Storer {

	store := &k8sStore{
		isOCSPCheckEnabled:    checkOCSP,
		cache:                 &Controller{},
		listers:               &Lister{},
		sslStore:              NewSSLCertTracker(),
		filesystem:            fs,
		updateCh:              updateCh,
		backendConfig:         ngx_config.NewDefault(),
		mu:                    &sync.Mutex{},
		secretIngressMap:      make(map[string]sets.String),
		defaultSSLCertificate: defaultSSLCertificate,
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{
		Interface: client.CoreV1().Events(namespace),
	})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{
		Component: "nginx-ingress-controller",
	})

	// k8sStore fulfils resolver.Resolver interface
	store.annotations = annotations.NewAnnotationExtractor(store)

	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			if !class.IsValid(addIng) {
				a, _ := parser.GetStringAnnotation(class.IngressKey, addIng)
				glog.Infof("ignoring add for ingress %v based on annotation %v with value %v", addIng.Name, class.IngressKey, a)
				return
			}

			store.extractAnnotations(addIng)
			recorder.Eventf(addIng, apiv1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", addIng.Namespace, addIng.Name))
			updateCh <- Event{
				Type: CreateEvent,
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
			if !class.IsValid(delIng) {
				glog.Infof("ignoring delete for ingress %v based on annotation %v", delIng.Name, class.IngressKey)
				return
			}
			recorder.Eventf(delIng, apiv1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", delIng.Namespace, delIng.Name))
			store.listers.IngressAnnotation.Delete(delIng)
			updateCh <- Event{
				Type: DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oldIng := old.(*extensions.Ingress)
			curIng := cur.(*extensions.Ingress)
			validOld := class.IsValid(oldIng)
			validCur := class.IsValid(curIng)
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
			updateCh <- Event{
				Type: UpdateEvent,
				Obj:  cur,
			}
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				sec := cur.(*apiv1.Secret)
				key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)

				// parse the ingress annotations (again)
				if set, ok := store.secretIngressMap[key]; ok {
					glog.Infof("secret %v changed and it is used in ingress annotations. Parsing...", key)
					_, err := store.GetLocalSecret(k8s.MetaNamespaceKey(sec))
					if err == nil {
						store.syncSecret(key)
						updateCh <- Event{
							Type: UpdateEvent,
							Obj:  cur,
						}
					}

					for _, name := range set.List() {
						ing, _ := store.GetIngress(name)
						store.extractAnnotations(ing)
					}

					updateCh <- Event{
						Type: ConfigurationEvent,
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
			store.sslStore.Delete(k8s.MetaNamespaceKey(sec))
			updateCh <- Event{
				Type: DeleteEvent,
				Obj:  obj,
			}

			// parse the ingress annotations (again)c
			key := fmt.Sprintf("%v/%v", sec.Namespace, sec.Name)
			if set, ok := store.secretIngressMap[key]; ok {
				glog.Infof("secret %v was removed and it is used in ingress annotations. Parsing...", key)
				for _, name := range set.List() {
					ing, _ := store.GetIngress(name)
					if ing != nil {
						store.extractAnnotations(ing)
					}
				}

				updateCh <- Event{
					Type: ConfigurationEvent,
					Obj:  sec,
				}
			}
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateCh <- Event{
				Type: CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: func(obj interface{}) {
			updateCh <- Event{
				Type: DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*apiv1.Endpoints)
			ocur := cur.(*apiv1.Endpoints)
			if !reflect.DeepEqual(ocur.Subsets, oep.Subsets) {
				updateCh <- Event{
					Type: UpdateEvent,
					Obj:  cur,
				}
			}
		},
	}

	mapEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m := obj.(*apiv1.ConfigMap)
			mapKey := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
			if mapKey == configmap {
				glog.V(2).Infof("adding configmap %v to backend", mapKey)
				store.setConfig(m)
				updateCh <- Event{
					Type: ConfigurationEvent,
					Obj:  obj,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				m := cur.(*apiv1.ConfigMap)
				mapKey := fmt.Sprintf("%s/%s", m.Namespace, m.Name)
				if mapKey == configmap {
					recorder.Eventf(m, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", mapKey))
					store.setConfig(m)
					updateCh <- Event{
						Type: ConfigurationEvent,
						Obj:  cur,
					}
				}
				// updates to configuration configmaps can trigger an update
				if mapKey == tcp || mapKey == udp {
					recorder.Eventf(m, apiv1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", mapKey))
					updateCh <- Event{
						Type: ConfigurationEvent,
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

// extractAnnotations parses ingress annotations converting the value of the
// annotation to a go struct and also information about the referenced secrets
func (s *k8sStore) extractAnnotations(ing *extensions.Ingress) {
	key := fmt.Sprintf("%v/%v", ing.Namespace, ing.Name)
	glog.V(3).Infof("updating annotations information for ingres %v", key)

	anns := s.annotations.Extract(ing)

	secName := anns.BasicDigestAuth.Secret
	if secName != "" {
		if _, ok := s.secretIngressMap[secName]; !ok {
			s.secretIngressMap[secName] = sets.NewString()
		}
		v := s.secretIngressMap[secName]
		if !v.Has(key) {
			v.Insert(key)
		}
	}

	secName = anns.CertificateAuth.Secret
	if secName != "" {
		if _, ok := s.secretIngressMap[secName]; !ok {
			s.secretIngressMap[secName] = sets.NewString()
		}
		v := s.secretIngressMap[secName]
		if !v.Has(key) {
			v.Insert(key)
		}
	}

	err := s.listers.IngressAnnotation.Update(anns)
	if err != nil {
		glog.Error(err)
	}
}

// GetSecret returns a Secret using the namespace and name as key
func (s k8sStore) GetSecret(key string) (*apiv1.Secret, error) {
	return s.listers.Secret.ByKey(key)
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
	return s.listers.Service.ByKey(key)
}

// GetSecret returns an Ingress using the namespace and name as key
func (s k8sStore) GetIngress(key string) (*extensions.Ingress, error) {
	return s.listers.Ingress.ByKey(key)
}

// ListIngresses returns the list of Ingresses
func (s k8sStore) ListIngresses() []*extensions.Ingress {
	// filter ingress rules
	var ingresses []*extensions.Ingress
	for _, item := range s.listers.Ingress.List() {
		ing := item.(*extensions.Ingress)
		if !class.IsValid(ing) {
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
		return &annotations.Ingress{}, fmt.Errorf("unexpected error getting ingress annotation %v: %v", key, err)
	}
	if !exists {
		return &annotations.Ingress{}, fmt.Errorf("ingress annotations %v was not found", key)
	}
	return item.(*annotations.Ingress), nil
}

// GetLocalSecret returns the local copy of a Secret
func (s k8sStore) GetLocalSecret(key string) (*ingress.SSLCert, error) {
	return s.sslStore.ByKey(key)
}

func (s k8sStore) GetConfigMap(key string) (*apiv1.ConfigMap, error) {
	return s.listers.ConfigMap.ByKey(key)
}

func (s k8sStore) GetServiceEndpoints(svc *apiv1.Service) (*apiv1.Endpoints, error) {
	return s.listers.Endpoint.GetServiceEndpoints(svc)
}

// GetAuthCertificate is used by the auth-tls annotations to get a cert from a secret
func (s k8sStore) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if _, err := s.GetLocalSecret(name); err != nil {
		s.syncSecret(name)
	}

	cert, err := s.GetLocalSecret(name)
	if err != nil {
		return nil, err
	}

	return &resolver.AuthSSLCert{
		Secret:     name,
		CAFileName: cert.CAFileName,
		PemSHA:     cert.PemSHA,
	}, nil
}

// GetDefaultBackend returns the default backend
func (s k8sStore) GetDefaultBackend() defaults.Backend {
	return s.backendConfig.Backend
}

func (s k8sStore) GetBackendConfiguration() ngx_config.Configuration {
	return s.backendConfig
}

func (s *k8sStore) setConfig(cmap *apiv1.ConfigMap) {
	s.backendConfig = ngx_template.ReadConfig(cmap.Data)

	// TODO: this should not be done here
	if s.backendConfig.SSLSessionTicketKey != "" {
		d, err := base64.StdEncoding.DecodeString(s.backendConfig.SSLSessionTicketKey)
		if err != nil {
			glog.Warningf("unexpected error decoding key ssl-session-ticket-key: %v", err)
			s.backendConfig.SSLSessionTicketKey = ""
		}
		ioutil.WriteFile("/etc/nginx/tickets.key", d, 0644)
	}
}

// Run initiates the synchronization of the controllers
// and the initial synchronization of the secrets.
func (s k8sStore) Run(stopCh chan struct{}) {
	// start controllers
	s.cache.Run(stopCh)

	// initial sync of secrets to avoid unnecessary reloads
	glog.Info("running initial sync of secrets")
	for _, ing := range s.ListIngresses() {
		s.ReadSecrets(ing)
	}

	if s.defaultSSLCertificate != "" {
		s.syncSecret(s.defaultSSLCertificate)
	}

	// start goroutine to check for missing local secrets
	go wait.Until(s.checkMissingSecrets, 10*time.Second, stopCh)

	if s.isOCSPCheckEnabled {
		go wait.Until(s.checkSSLChainIssues, 60*time.Second, stopCh)
	}
}
