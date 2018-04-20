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

	"github.com/eapache/channels"
	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/class"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	ngx_template "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
)

// Storer is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Storer interface {
	// GetBackendConfiguration returns the nginx configuration stored in a configmap
	GetBackendConfiguration() ngx_config.Configuration

	// GetConfigMap returns a ConfigmMap using the namespace and name as key
	GetConfigMap(key string) (*corev1.ConfigMap, error)

	// GetSecret returns a Secret using the namespace and name as key
	GetSecret(key string) (*corev1.Secret, error)

	// GetService returns a Service using the namespace and name as key
	GetService(key string) (*corev1.Service, error)

	GetServiceEndpoints(svc *corev1.Service) (*corev1.Endpoints, error)

	// GetSecret returns an Ingress using the namespace and name as key
	GetIngress(key string) (*extensions.Ingress, error)

	// ListIngresses returns the list of Ingresses
	ListIngresses() []*extensions.Ingress

	// GetIngressAnnotations returns the annotations associated to an Ingress
	GetIngressAnnotations(ing *extensions.Ingress) (*annotations.Ingress, error)

	// GetLocalSSLCert returns the local copy of a SSLCert
	GetLocalSSLCert(name string) (*ingress.SSLCert, error)

	// ListLocalSSLCerts returns the list of local SSLCerts
	ListLocalSSLCerts() []*ingress.SSLCert

	// GetAuthCertificate resolves a given secret name into an SSL certificate.
	// The secret must contain 3 keys named:
	//   ca.crt: contains the certificate chain used for authentication
	GetAuthCertificate(string) (*resolver.AuthSSLCert, error)

	// GetDefaultBackend returns the default backend configuration
	GetDefaultBackend() defaults.Backend

	// Run initiates the synchronization of the controllers
	Run(stopCh chan struct{})
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
	// ConfigurationEvent event associated when a controller configuration object is created or updated
	ConfigurationEvent EventType = "CONFIGURATION"
)

// Event holds the context of an event
type Event struct {
	Type EventType
	Obj  interface{}
}

// Informer defines the required SharedIndexInformers that interact with the API server.
type Informer struct {
	Ingress   cache.SharedIndexInformer
	Endpoint  cache.SharedIndexInformer
	Service   cache.SharedIndexInformer
	Secret    cache.SharedIndexInformer
	ConfigMap cache.SharedIndexInformer
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

// Run initiates the synchronization of the informers against the API server.
func (i *Informer) Run(stopCh chan struct{}) {
	go i.Endpoint.Run(stopCh)
	go i.Service.Run(stopCh)
	go i.Secret.Run(stopCh)
	go i.ConfigMap.Run(stopCh)

	// wait for all involved caches to be synced before processing items
	// from the queue
	if !cache.WaitForCacheSync(stopCh,
		i.Endpoint.HasSynced,
		i.Service.HasSynced,
		i.Secret.HasSynced,
		i.ConfigMap.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}

	// in big clusters, deltas can keep arriving even after HasSynced
	// functions have returned 'true'
	time.Sleep(1 * time.Second)

	// we can start syncing ingress objects only after other caches are
	// ready, because ingress rules require content from other listers, and
	// 'add' events get triggered in the handlers during caches population.
	go i.Ingress.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh,
		i.Ingress.HasSynced,
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

	// informer contains the cache Informers
	informers *Informer

	// listers contains the cache.Store interfaces used in the ingress controller
	listers *Lister

	// sslStore local store of SSL certificates (certificates used in ingress)
	// this is required because the certificates must be present in the
	// container filesystem
	sslStore *SSLCertTracker

	annotations annotations.Extractor

	// secretIngressMap contains information about which ingress references a
	// secret in the annotations.
	secretIngressMap ObjectRefMap

	filesystem file.Filesystem

	// updateCh
	updateCh *channels.RingChannel

	// mu protects against simultaneous invocations of syncSecret
	mu *sync.Mutex

	defaultSSLCertificate string
}

// New creates a new object store to be used in the ingress controller
func New(checkOCSP bool,
	namespace, configmap, tcp, udp, defaultSSLCertificate string,
	resyncPeriod time.Duration,
	client clientset.Interface,
	fs file.Filesystem,
	updateCh *channels.RingChannel) Storer {

	store := &k8sStore{
		isOCSPCheckEnabled:    checkOCSP,
		informers:             &Informer{},
		listers:               &Lister{},
		sslStore:              NewSSLCertTracker(),
		filesystem:            fs,
		updateCh:              updateCh,
		backendConfig:         ngx_config.NewDefault(),
		mu:                    &sync.Mutex{},
		secretIngressMap:      NewObjectRefMap(),
		defaultSSLCertificate: defaultSSLCertificate,
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{
		Interface: client.CoreV1().Events(namespace),
	})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "nginx-ingress-controller",
	})

	// k8sStore fulfils resolver.Resolver interface
	store.annotations = annotations.NewAnnotationExtractor(store)

	store.listers.IngressAnnotation.Store = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	// create informers factory, enable and assign required informers
	infFactory := informers.NewFilteredSharedInformerFactory(client, resyncPeriod, namespace, func(*metav1.ListOptions) {})

	store.informers.Ingress = infFactory.Extensions().V1beta1().Ingresses().Informer()
	store.listers.Ingress.Store = store.informers.Ingress.GetStore()

	store.informers.Endpoint = infFactory.Core().V1().Endpoints().Informer()
	store.listers.Endpoint.Store = store.informers.Endpoint.GetStore()

	store.informers.Secret = infFactory.Core().V1().Secrets().Informer()
	store.listers.Secret.Store = store.informers.Secret.GetStore()

	store.informers.ConfigMap = infFactory.Core().V1().ConfigMaps().Informer()
	store.listers.ConfigMap.Store = store.informers.ConfigMap.GetStore()

	store.informers.Service = infFactory.Core().V1().Services().Informer()
	store.listers.Service.Store = store.informers.Service.GetStore()

	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ing := obj.(*extensions.Ingress)
			if !class.IsValid(ing) {
				a, _ := parser.GetStringAnnotation(class.IngressKey, ing)
				glog.Infof("ignoring add for ingress %v based on annotation %v with value %v", ing.Name, class.IngressKey, a)
				return
			}
			recorder.Eventf(ing, corev1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", ing.Namespace, ing.Name))

			store.extractAnnotations(ing)
			store.updateSecretIngressMap(ing)
			store.syncSecrets(ing)

			updateCh.In() <- Event{
				Type: CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: func(obj interface{}) {
			ing, ok := obj.(*extensions.Ingress)
			if !ok {
				// If we reached here it means the ingress was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				ing, ok = tombstone.Obj.(*extensions.Ingress)
				if !ok {
					glog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
					return
				}
			}
			if !class.IsValid(ing) {
				glog.Infof("ignoring delete for ingress %v based on annotation %v", ing.Name, class.IngressKey)
				return
			}
			recorder.Eventf(ing, corev1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", ing.Namespace, ing.Name))

			store.listers.IngressAnnotation.Delete(ing)

			key := k8s.MetaNamespaceKey(ing)
			store.secretIngressMap.Delete(key)

			updateCh.In() <- Event{
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
				recorder.Eventf(curIng, corev1.EventTypeNormal, "CREATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validOld && !validCur {
				glog.Infof("removing ingress %v based on annotation %v", curIng.Name, class.IngressKey)
				recorder.Eventf(curIng, corev1.EventTypeNormal, "DELETE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			} else if validCur && !reflect.DeepEqual(old, cur) {
				recorder.Eventf(curIng, corev1.EventTypeNormal, "UPDATE", fmt.Sprintf("Ingress %s/%s", curIng.Namespace, curIng.Name))
			}

			store.extractAnnotations(curIng)
			store.updateSecretIngressMap(curIng)
			store.syncSecrets(curIng)

			updateCh.In() <- Event{
				Type: UpdateEvent,
				Obj:  cur,
			}
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sec := obj.(*corev1.Secret)
			key := k8s.MetaNamespaceKey(sec)

			if store.defaultSSLCertificate == key {
				store.syncSecret(store.defaultSSLCertificate)
			}

			// find references in ingresses and update local ssl certs
			if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
				glog.Infof("secret %v was added and it is used in ingress annotations. Parsing...", key)
				for _, ingKey := range ings {
					ing, err := store.GetIngress(ingKey)
					if err != nil {
						glog.Errorf("could not find Ingress %v in local store", ingKey)
						continue
					}
					store.extractAnnotations(ing)
					store.syncSecrets(ing)
				}
				updateCh.In() <- Event{
					Type: CreateEvent,
					Obj:  obj,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				sec := cur.(*corev1.Secret)
				key := k8s.MetaNamespaceKey(sec)

				if store.defaultSSLCertificate == key {
					store.syncSecret(store.defaultSSLCertificate)
				}

				// find references in ingresses and update local ssl certs
				if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
					glog.Infof("secret %v was updated and it is used in ingress annotations. Parsing...", key)
					for _, ingKey := range ings {
						ing, err := store.GetIngress(ingKey)
						if err != nil {
							glog.Errorf("could not find Ingress %v in local store", ingKey)
							continue
						}
						store.extractAnnotations(ing)
						store.syncSecrets(ing)
					}
					updateCh.In() <- Event{
						Type: UpdateEvent,
						Obj:  cur,
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			sec, ok := obj.(*corev1.Secret)
			if !ok {
				// If we reached here it means the secret was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					glog.Errorf("couldn't get object from tombstone %#v", obj)
					return
				}
				sec, ok = tombstone.Obj.(*corev1.Secret)
				if !ok {
					glog.Errorf("Tombstone contained object that is not a Secret: %#v", obj)
					return
				}
			}

			store.sslStore.Delete(k8s.MetaNamespaceKey(sec))

			key := k8s.MetaNamespaceKey(sec)

			// find references in ingresses
			if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
				glog.Infof("secret %v was deleted and it is used in ingress annotations. Parsing...", key)
				for _, ingKey := range ings {
					ing, err := store.GetIngress(ingKey)
					if err != nil {
						glog.Errorf("could not find Ingress %v in local store", ingKey)
						continue
					}
					store.extractAnnotations(ing)
				}
				updateCh.In() <- Event{
					Type: DeleteEvent,
					Obj:  obj,
				}
			}
		},
	}

	epEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateCh.In() <- Event{
				Type: CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: func(obj interface{}) {
			updateCh.In() <- Event{
				Type: DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oep := old.(*corev1.Endpoints)
			cep := cur.(*corev1.Endpoints)
			if !reflect.DeepEqual(cep.Subsets, oep.Subsets) {
				updateCh.In() <- Event{
					Type: UpdateEvent,
					Obj:  cur,
				}
			}
		},
	}

	cmEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			key := k8s.MetaNamespaceKey(cm)
			// updates to configuration configmaps can trigger an update
			if key == configmap || key == tcp || key == udp {
				recorder.Eventf(cm, corev1.EventTypeNormal, "CREATE", fmt.Sprintf("ConfigMap %v", key))
				if key == configmap {
					store.setConfig(cm)
				}
				updateCh.In() <- Event{
					Type: ConfigurationEvent,
					Obj:  obj,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				cm := cur.(*corev1.ConfigMap)
				key := k8s.MetaNamespaceKey(cm)
				// updates to configuration configmaps can trigger an update
				if key == configmap || key == tcp || key == udp {
					recorder.Eventf(cm, corev1.EventTypeNormal, "UPDATE", fmt.Sprintf("ConfigMap %v", key))
					if key == configmap {
						store.setConfig(cm)
					}
					updateCh.In() <- Event{
						Type: ConfigurationEvent,
						Obj:  cur,
					}
				}
			}
		},
	}

	store.informers.Ingress.AddEventHandler(ingEventHandler)
	store.informers.Endpoint.AddEventHandler(epEventHandler)
	store.informers.Secret.AddEventHandler(secrEventHandler)
	store.informers.ConfigMap.AddEventHandler(cmEventHandler)
	store.informers.Service.AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return store
}

// extractAnnotations parses ingress annotations converting the value of the
// annotation to a go struct and also information about the referenced secrets
func (s *k8sStore) extractAnnotations(ing *extensions.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	glog.V(3).Infof("updating annotations information for ingress %v", key)

	anns := s.annotations.Extract(ing)

	err := s.listers.IngressAnnotation.Update(anns)
	if err != nil {
		glog.Error(err)
	}
}

// updateSecretIngressMap takes an Ingress and updates all Secret objects it
// references in secretIngressMap.
func (s *k8sStore) updateSecretIngressMap(ing *extensions.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	glog.V(3).Infof("updating references to secrets for ingress %v", key)

	// delete all existing references first
	s.secretIngressMap.Delete(key)

	var refSecrets []string

	for _, tls := range ing.Spec.TLS {
		secrName := tls.SecretName
		if secrName != "" {
			secrKey := fmt.Sprintf("%v/%v", ing.Namespace, secrName)
			refSecrets = append(refSecrets, secrKey)
		}
	}

	// We can not rely on cached ingress annotations because these are
	// discarded when the referenced secret does not exist in the local
	// store. As a result, adding a secret *after* the ingress(es) which
	// references it would not trigger a resync of that secret.
	secretAnnotations := []string{
		"auth-secret",
		"auth-tls-secret",
	}
	for _, ann := range secretAnnotations {
		secrKey, err := objectRefAnnotationNsKey(ann, ing)
		if err != nil && !errors.IsMissingAnnotations(err) {
			glog.Errorf("error reading secret reference in annotation %q: %s", ann, err)
			continue
		}
		if secrKey != "" {
			refSecrets = append(refSecrets, secrKey)
		}
	}

	// populate map with all secret references
	s.secretIngressMap.Insert(key, refSecrets...)
}

// objectRefAnnotationNsKey returns an object reference formatted as a
// 'namespace/name' key from the given annotation name.
func objectRefAnnotationNsKey(ann string, ing *extensions.Ingress) (string, error) {
	annValue, err := parser.GetStringAnnotation(ann, ing)
	if annValue == "" {
		return "", err
	}

	secrNs, secrName, err := cache.SplitMetaNamespaceKey(annValue)
	if secrName == "" {
		return "", err
	}

	if secrNs == "" {
		return fmt.Sprintf("%v/%v", ing.Namespace, secrName), nil
	}
	return annValue, nil
}

// syncSecrets synchronizes data from all Secrets referenced by the given
// Ingress with the local store and file system.
func (s k8sStore) syncSecrets(ing *extensions.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	for _, secrKey := range s.secretIngressMap.ReferencedBy(key) {
		s.syncSecret(secrKey)
	}
}

// GetSecret returns a Secret using the namespace and name as key
func (s k8sStore) GetSecret(key string) (*corev1.Secret, error) {
	return s.listers.Secret.ByKey(key)
}

// ListLocalSSLCerts returns the list of local SSLCerts
func (s k8sStore) ListLocalSSLCerts() []*ingress.SSLCert {
	var certs []*ingress.SSLCert
	for _, item := range s.sslStore.List() {
		if s, ok := item.(*ingress.SSLCert); ok {
			certs = append(certs, s)
		}
	}

	return certs
}

// GetService returns a Service using the namespace and name as key
func (s k8sStore) GetService(key string) (*corev1.Service, error) {
	return s.listers.Service.ByKey(key)
}

// GetIngress returns an Ingress using the namespace and name as key
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
		for ri, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for pi, path := range rule.HTTP.Paths {
				if path.Path == "" {
					ing.Spec.Rules[ri].HTTP.Paths[pi].Path = "/"
				}
			}
		}
		ingresses = append(ingresses, ing)
	}

	return ingresses
}

// GetIngressAnnotations returns the annotations associated to an Ingress
func (s k8sStore) GetIngressAnnotations(ing *extensions.Ingress) (*annotations.Ingress, error) {
	key := k8s.MetaNamespaceKey(ing)
	item, exists, err := s.listers.IngressAnnotation.GetByKey(key)
	if err != nil {
		return &annotations.Ingress{}, fmt.Errorf("unexpected error getting ingress annotation %v: %v", key, err)
	}
	if !exists {
		return &annotations.Ingress{}, fmt.Errorf("ingress annotations %v was not found", key)
	}
	return item.(*annotations.Ingress), nil
}

// GetLocalSSLCert returns the local copy of a SSLCert
func (s k8sStore) GetLocalSSLCert(key string) (*ingress.SSLCert, error) {
	return s.sslStore.ByKey(key)
}

func (s k8sStore) GetConfigMap(key string) (*corev1.ConfigMap, error) {
	return s.listers.ConfigMap.ByKey(key)
}

func (s k8sStore) GetServiceEndpoints(svc *corev1.Service) (*corev1.Endpoints, error) {
	return s.listers.Endpoint.GetServiceEndpoints(svc)
}

// GetAuthCertificate is used by the auth-tls annotations to get a cert from a secret
func (s k8sStore) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if _, err := s.GetLocalSSLCert(name); err != nil {
		s.syncSecret(name)
	}

	cert, err := s.GetLocalSSLCert(name)
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

func (s *k8sStore) setConfig(cmap *corev1.ConfigMap) {
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

// Run initiates the synchronization of the informers and the initial
// synchronization of the secrets.
func (s k8sStore) Run(stopCh chan struct{}) {
	// start informers
	s.informers.Run(stopCh)

	if s.isOCSPCheckEnabled {
		go wait.Until(s.checkSSLChainIssues, 60*time.Second, stopCh)
	}
}
