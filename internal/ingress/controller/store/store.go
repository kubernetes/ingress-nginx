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
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eapache/channels"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/ingress-nginx/internal/ingress/inspector"

	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/util/file"
	klog "k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ngx_config "k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/internal/ingress/controller/ingressclass"
	ngx_template "k8s.io/ingress-nginx/internal/ingress/controller/template"
	"k8s.io/ingress-nginx/internal/ingress/defaults"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

// IngressFilterFunc decides if an Ingress should be omitted or not
type IngressFilterFunc func(*ingress.Ingress) bool

// Storer is the interface that wraps the required methods to gather information
// about ingresses, services, secrets and ingress annotations.
type Storer interface {
	// GetBackendConfiguration returns the nginx configuration stored in a configmap
	GetBackendConfiguration() ngx_config.Configuration

	// GetSecurityConfiguration returns the configuration options from Ingress
	GetSecurityConfiguration() defaults.SecurityConfiguration

	// GetConfigMap returns the ConfigMap matching key.
	GetConfigMap(key string) (*corev1.ConfigMap, error)

	// GetSecret returns the Secret matching key.
	GetSecret(key string) (*corev1.Secret, error)

	// GetService returns the Service matching key.
	GetService(key string) (*corev1.Service, error)

	// GetServiceEndpointsSlices returns the EndpointSlices of a Service matching key.
	GetServiceEndpointsSlices(key string) ([]*discoveryv1.EndpointSlice, error)

	// ListIngresses returns a list of all Ingresses in the store.
	ListIngresses() []*ingress.Ingress

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

	// GetIngressClass validates given ingress against ingress class configuration and returns the ingress class.
	GetIngressClass(ing *networkingv1.Ingress, icConfig *ingressclass.Configuration) (string, error)
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

// Event holds the context of an event.
type Event struct {
	Type EventType
	Obj  interface{}
}

// Informer defines the required SharedIndexInformers that interact with the API server.
type Informer struct {
	Ingress       cache.SharedIndexInformer
	IngressClass  cache.SharedIndexInformer
	EndpointSlice cache.SharedIndexInformer
	Service       cache.SharedIndexInformer
	Secret        cache.SharedIndexInformer
	ConfigMap     cache.SharedIndexInformer
	Namespace     cache.SharedIndexInformer
}

// Lister contains object listers (stores).
type Lister struct {
	Ingress               IngressLister
	IngressClass          IngressClassLister
	Service               ServiceLister
	EndpointSlice         EndpointSliceLister
	Secret                SecretLister
	ConfigMap             ConfigMapLister
	Namespace             NamespaceLister
	IngressWithAnnotation IngressWithAnnotationsLister
}

// NotExistsError is returned when an object does not exist in a local store.
type NotExistsError string

// Error implements the error interface.
func (e NotExistsError) Error() string {
	return fmt.Sprintf("no object matching key %q in local store", string(e))
}

// Run initiates the synchronization of the informers against the API server.
func (i *Informer) Run(stopCh chan struct{}) {
	go i.Secret.Run(stopCh)
	go i.EndpointSlice.Run(stopCh)
	if i.IngressClass != nil {
		go i.IngressClass.Run(stopCh)
	}
	go i.Service.Run(stopCh)
	go i.ConfigMap.Run(stopCh)

	// wait for all involved caches to be synced before processing items
	// from the queue
	if !cache.WaitForCacheSync(stopCh,
		i.Service.HasSynced,
		i.Secret.HasSynced,
		i.ConfigMap.HasSynced,
	) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}
	if i.IngressClass != nil && !cache.WaitForCacheSync(stopCh, i.IngressClass.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for ingress classcaches to sync"))
	}

	// when limit controller scope to one namespace, skip sync namespaces at cluster scope
	if i.Namespace != nil {
		go i.Namespace.Run(stopCh)

		if !cache.WaitForCacheSync(stopCh, i.Namespace.HasSynced) {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		}
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
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	}
}

// k8sStore internal Storer implementation using informers and thread safe stores
type k8sStore struct {
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

	// updateCh
	updateCh *channels.RingChannel

	// syncSecretMu protects against simultaneous invocations of syncSecret
	syncSecretMu *sync.Mutex

	// backendConfigMu protects against simultaneous read/write of backendConfig
	backendConfigMu *sync.RWMutex

	defaultSSLCertificate string

	recorder record.EventRecorder
}

// New creates a new object store to be used in the ingress controller.
//
//nolint:gocyclo // Ignore function complexity error.
func New(
	namespace string,
	namespaceSelector labels.Selector,
	configmap, tcp, udp, defaultSSLCertificate string,
	resyncPeriod time.Duration,
	client clientset.Interface,
	updateCh *channels.RingChannel,
	disableCatchAll bool,
	deepInspector bool,
	icConfig *ingressclass.Configuration,
	disableSyncEvents bool,
) Storer {
	store := &k8sStore{
		informers:             &Informer{},
		listers:               &Lister{},
		sslStore:              NewSSLCertTracker(),
		updateCh:              updateCh,
		backendConfig:         ngx_config.NewDefault(),
		syncSecretMu:          &sync.Mutex{},
		backendConfigMu:       &sync.RWMutex{},
		secretIngressMap:      NewObjectRefMap(),
		defaultSSLCertificate: defaultSSLCertificate,
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	if !disableSyncEvents {
		eventBroadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{
			Interface: client.CoreV1().Events(namespace),
		})
	}
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "nginx-ingress-controller",
	})
	store.recorder = recorder

	// k8sStore fulfills resolver.Resolver interface
	store.annotations = annotations.NewAnnotationExtractor(store)

	store.listers.IngressWithAnnotation.Store = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	// As we currently do not filter out kubernetes objects we list, we can
	// retrieve a huge amount of data from the API server.
	// In a cluster using HELM < v3 configmaps are used to store binary data.
	// If you happen to have a lot of HELM releases in the cluster it will make
	// the memory consumption of nginx-ingress-controller explode.
	// In order to avoid that we filter out labels OWNER=TILLER.
	labelsTweakListOptionsFunc := func(options *metav1.ListOptions) {
		if options.LabelSelector != "" {
			options.LabelSelector += ",OWNER!=TILLER"
		} else {
			options.LabelSelector = "OWNER!=TILLER"
		}
	}

	// As of HELM >= v3 helm releases are stored using Secrets instead of ConfigMaps.
	// In order to avoid listing those secrets we discard type "helm.sh/release.v1"
	secretsTweakListOptionsFunc := func(options *metav1.ListOptions) {
		helmAntiSelector := fields.OneTermNotEqualSelector("type", "helm.sh/release.v1")
		baseSelector, err := fields.ParseSelector(options.FieldSelector)

		if err != nil {
			options.FieldSelector = helmAntiSelector.String()
		} else {
			options.FieldSelector = fields.AndSelectors(baseSelector, helmAntiSelector).String()
		}
	}

	// create informers factory, enable and assign required informers
	infFactory := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
		informers.WithNamespace(namespace),
	)

	// create informers factory for configmaps
	infFactoryConfigmaps := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(labelsTweakListOptionsFunc),
	)

	// create informers factory for secrets
	infFactorySecrets := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(secretsTweakListOptionsFunc),
	)

	store.informers.Ingress = infFactory.Networking().V1().Ingresses().Informer()
	store.listers.Ingress.Store = store.informers.Ingress.GetStore()

	if !icConfig.IgnoreIngressClass {
		store.informers.IngressClass = infFactory.Networking().V1().IngressClasses().Informer()
		store.listers.IngressClass.Store = cache.NewStore(cache.MetaNamespaceKeyFunc)
	}

	store.informers.EndpointSlice = infFactory.Discovery().V1().EndpointSlices().Informer()
	store.listers.EndpointSlice.Store = store.informers.EndpointSlice.GetStore()

	store.informers.Secret = infFactorySecrets.Core().V1().Secrets().Informer()
	store.listers.Secret.Store = store.informers.Secret.GetStore()

	store.informers.ConfigMap = infFactoryConfigmaps.Core().V1().ConfigMaps().Informer()
	store.listers.ConfigMap.Store = store.informers.ConfigMap.GetStore()

	store.informers.Service = infFactory.Core().V1().Services().Informer()
	store.listers.Service.Store = store.informers.Service.GetStore()

	// avoid caching namespaces at cluster scope when watching single namespace
	if namespaceSelector != nil && !namespaceSelector.Empty() {
		// cache informers factory for namespaces
		infFactoryNamespaces := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
			informers.WithTweakListOptions(labelsTweakListOptionsFunc),
		)

		store.informers.Namespace = infFactoryNamespaces.Core().V1().Namespaces().Informer()
		store.listers.Namespace.Store = store.informers.Namespace.GetStore()
	}

	watchedNamespace := func(namespace string) bool {
		if namespaceSelector == nil || namespaceSelector.Empty() {
			return true
		}

		item, ok, err := store.listers.Namespace.GetByKey(namespace)
		if !ok {
			klog.Errorf("Namespace %s not existed: %v.", namespace, err)
			return false
		}
		ns, ok := item.(*corev1.Namespace)
		if !ok {
			return false
		}

		return namespaceSelector.Matches(labels.Set(ns.Labels))
	}

	ingDeleteHandler := func(obj interface{}) {
		ing, ok := toIngress(obj)
		if !ok {
			// If we reached here it means the ingress was deleted but its final state is unrecorded.
			tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
			if !ok {
				klog.ErrorS(nil, "Error obtaining object from tombstone", "key", obj)
				return
			}
			ing, ok = tombstone.Obj.(*networkingv1.Ingress)
			if !ok {
				klog.Errorf("Tombstone contained object that is not an Ingress: %#v", obj)
				return
			}
		}

		if !watchedNamespace(ing.Namespace) {
			return
		}

		_, err := store.GetIngressClass(ing, icConfig)
		if err != nil {
			klog.InfoS("Ignoring ingress because of error while validating ingress class", "ingress", klog.KObj(ing), "error", err)
			return
		}

		if hasCatchAllIngressRule(ing.Spec) && disableCatchAll {
			klog.InfoS("Ignoring delete for catch-all because of --disable-catch-all", "ingress", klog.KObj(ing))
			return
		}

		if err := store.listers.IngressWithAnnotation.Delete(ing); err != nil {
			klog.ErrorS(err, "Error while deleting ingress from store", "ingress", klog.KObj(ing))
			return
		}

		key := k8s.MetaNamespaceKey(ing)
		store.secretIngressMap.Delete(key)

		updateCh.In() <- Event{
			Type: DeleteEvent,
			Obj:  obj,
		}
	}

	ingEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ing, _ := toIngress(obj)

			if !watchedNamespace(ing.Namespace) {
				return
			}

			ic, err := store.GetIngressClass(ing, icConfig)
			if err != nil {
				klog.InfoS("Ignoring ingress because of error while validating ingress class", "ingress", klog.KObj(ing), "error", err)
				return
			}

			klog.InfoS("Found valid IngressClass", "ingress", klog.KObj(ing), "ingressclass", ic)

			if deepInspector {
				if err := inspector.DeepInspect(ing); err != nil {
					klog.ErrorS(err, "received invalid ingress", "ingress", klog.KObj(ing))
					return
				}
			}
			if hasCatchAllIngressRule(ing.Spec) && disableCatchAll {
				klog.InfoS("Ignoring add for catch-all ingress because of --disable-catch-all", "ingress", klog.KObj(ing))
				return
			}

			recorder.Eventf(ing, corev1.EventTypeNormal, "Sync", "Scheduled for sync")

			store.syncIngress(ing)
			store.updateSecretIngressMap(ing)
			store.syncSecrets(ing)

			updateCh.In() <- Event{
				Type: CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: ingDeleteHandler,
		UpdateFunc: func(old, cur interface{}) {
			oldIng, _ := toIngress(old)
			curIng, _ := toIngress(cur)

			if !watchedNamespace(oldIng.Namespace) {
				return
			}

			var errOld, errCur error
			var classCur string
			if !icConfig.IgnoreIngressClass {
				_, errOld = store.GetIngressClass(oldIng, icConfig)
				classCur, errCur = store.GetIngressClass(curIng, icConfig)
			}
			switch {
			case errOld != nil && errCur == nil:
				if hasCatchAllIngressRule(curIng.Spec) && disableCatchAll {
					klog.InfoS("ignoring update for catch-all ingress because of --disable-catch-all", "ingress", klog.KObj(curIng))
					return
				}

				klog.InfoS("creating ingress", "ingress", klog.KObj(curIng), "ingressclass", classCur)
				recorder.Eventf(curIng, corev1.EventTypeNormal, "Sync", "Scheduled for sync")
			case errOld == nil && errCur != nil:
				klog.InfoS("removing ingress because of unknown ingressclass", "ingress", klog.KObj(curIng))
				ingDeleteHandler(old)
				return
			case errCur == nil && !reflect.DeepEqual(old, cur):
				if hasCatchAllIngressRule(curIng.Spec) && disableCatchAll {
					klog.InfoS("ignoring update for catch-all ingress and delete old one because of --disable-catch-all", "ingress", klog.KObj(curIng))
					ingDeleteHandler(old)
					return
				}

				recorder.Eventf(curIng, corev1.EventTypeNormal, "Sync", "Scheduled for sync")
			default:
				klog.V(3).InfoS("No changes on ingress. Skipping update", "ingress", klog.KObj(curIng))
				return
			}

			if deepInspector {
				if err := inspector.DeepInspect(curIng); err != nil {
					klog.ErrorS(err, "received invalid ingress", "ingress", klog.KObj(curIng))
					return
				}
			}

			store.syncIngress(curIng)
			store.updateSecretIngressMap(curIng)
			store.syncSecrets(curIng)

			updateCh.In() <- Event{
				Type: UpdateEvent,
				Obj:  cur,
			}
		},
	}

	ingressClassEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingressclass, ok := obj.(*networkingv1.IngressClass)
			if !ok {
				klog.Errorf("unexpected type: %T", obj)
			}
			foundClassByName := false
			if icConfig.IngressClassByName && ingressclass.Name == icConfig.AnnotationValue {
				klog.InfoS("adding ingressclass as ingress-class-by-name is configured", "ingressclass", klog.KObj(ingressclass))
				foundClassByName = true
			}
			if !foundClassByName && ingressclass.Spec.Controller != icConfig.Controller {
				klog.InfoS("ignoring ingressclass as the spec.controller is not the same of this ingress", "ingressclass", klog.KObj(ingressclass))
				return
			}
			err := store.listers.IngressClass.Add(ingressclass)
			if err != nil {
				klog.InfoS("error adding ingressclass to store", "ingressclass", klog.KObj(ingressclass), "error", err)
				return
			}

			updateCh.In() <- Event{
				Type: CreateEvent,
				Obj:  obj,
			}
		},
		DeleteFunc: func(obj interface{}) {
			ingressclass, ok := obj.(*networkingv1.IngressClass)
			if !ok {
				klog.Errorf("unexpected type: %T", obj)
			}
			if ingressclass.Spec.Controller != icConfig.Controller {
				klog.InfoS("ignoring ingressclass as the spec.controller is not the same of this ingress", "ingressclass", klog.KObj(ingressclass))
				return
			}
			err := store.listers.IngressClass.Delete(ingressclass)
			if err != nil {
				klog.InfoS("error removing ingressclass from store", "ingressclass", klog.KObj(ingressclass), "error", err)
				return
			}
			updateCh.In() <- Event{
				Type: DeleteEvent,
				Obj:  obj,
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oic, ok := old.(*networkingv1.IngressClass)
			if !ok {
				klog.Errorf("unexpected type: %T", old)
			}
			cic, ok := cur.(*networkingv1.IngressClass)
			if !ok {
				klog.Errorf("unexpected type: %T", cur)
			}
			if cic.Spec.Controller != icConfig.Controller {
				klog.InfoS("ignoring ingressclass as the spec.controller is not the same of this ingress", "ingressclass", klog.KObj(cic))
				return
			}
			// TODO: In a future we might be interested in parse parameters and use as
			// current IngressClass for this case, crossing with configmap
			if !reflect.DeepEqual(cic.Spec.Parameters, oic.Spec.Parameters) {
				err := store.listers.IngressClass.Update(cic)
				if err != nil {
					klog.InfoS("error updating ingressclass in store", "ingressclass", klog.KObj(cic), "error", err)
					return
				}
				updateCh.In() <- Event{
					Type: UpdateEvent,
					Obj:  cur,
				}
			}
		},
	}

	secrEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sec, ok := obj.(*corev1.Secret)
			if !ok {
				klog.Errorf("unexpected type: %T", obj)
			}
			key := k8s.MetaNamespaceKey(sec)

			if store.defaultSSLCertificate == key {
				store.syncSecret(store.defaultSSLCertificate)
			}

			// find references in ingresses and update local ssl certs
			if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
				klog.InfoS("Secret was added and it is used in ingress annotations. Parsing", "secret", key)
				for _, ingKey := range ings {
					ing, err := store.getIngress(ingKey)
					if err != nil {
						klog.Errorf("could not find Ingress %v in local store", ingKey)
						continue
					}
					store.syncIngress(ing)
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
				sec, ok := cur.(*corev1.Secret)
				if !ok {
					klog.Errorf("unexpected type: %T", cur)
				}
				key := k8s.MetaNamespaceKey(sec)

				if !watchedNamespace(sec.Namespace) {
					return
				}

				if store.defaultSSLCertificate == key {
					store.syncSecret(store.defaultSSLCertificate)
				}

				// find references in ingresses and update local ssl certs
				if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
					klog.InfoS("secret was updated and it is used in ingress annotations. Parsing", "secret", key)
					for _, ingKey := range ings {
						ing, err := store.getIngress(ingKey)
						if err != nil {
							klog.ErrorS(err, "could not find Ingress in local store", "ingress", ingKey)
							continue
						}
						store.syncSecrets(ing)
						store.syncIngress(ing)
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
					return
				}

				sec, ok = tombstone.Obj.(*corev1.Secret)
				if !ok {
					return
				}
			}

			if !watchedNamespace(sec.Namespace) {
				return
			}

			store.sslStore.Delete(k8s.MetaNamespaceKey(sec))

			key := k8s.MetaNamespaceKey(sec)

			// find references in ingresses
			if ings := store.secretIngressMap.Reference(key); len(ings) > 0 {
				klog.InfoS("secret was deleted and it is used in ingress annotations. Parsing", "secret", key)
				for _, ingKey := range ings {
					ing, err := store.getIngress(ingKey)
					if err != nil {
						klog.Errorf("could not find Ingress %v in local store", ingKey)
						continue
					}
					store.syncIngress(ing)
				}

				updateCh.In() <- Event{
					Type: DeleteEvent,
					Obj:  obj,
				}
			}
		},
	}

	epsEventHandler := cache.ResourceEventHandlerFuncs{
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
			oeps, ok := old.(*discoveryv1.EndpointSlice)
			if !ok {
				klog.Errorf("unexpected type: %T", old)
			}
			ceps, ok := cur.(*discoveryv1.EndpointSlice)
			if !ok {
				klog.Errorf("unexpected type: %T", cur)
			}
			if !reflect.DeepEqual(ceps.Endpoints, oeps.Endpoints) {
				updateCh.In() <- Event{
					Type: UpdateEvent,
					Obj:  cur,
				}
			}
		},
	}

	changeTriggerUpdate := func(name string) bool {
		return name == configmap || name == tcp || name == udp
	}

	handleCfgMapEvent := func(key string, cfgMap *corev1.ConfigMap, eventName string) {
		// updates to configuration configmaps can trigger an update
		triggerUpdate := false
		if changeTriggerUpdate(key) {
			triggerUpdate = true
			recorder.Eventf(cfgMap, corev1.EventTypeNormal, eventName, fmt.Sprintf("ConfigMap %v", key))
			if key == configmap {
				store.setConfig(cfgMap)
			}
		}

		ings := store.listers.IngressWithAnnotation.List()
		for _, ingKey := range ings {
			key := k8s.MetaNamespaceKey(ingKey)
			ing, err := store.getIngress(key)
			if err != nil {
				klog.Errorf("could not find Ingress %v in local store: %v", key, err)
				continue
			}

			if parser.AnnotationsReferencesConfigmap(ing) {
				store.syncIngress(ing)
				continue
			}

			if triggerUpdate {
				store.syncIngress(ing)
			}
		}

		if triggerUpdate {
			updateCh.In() <- Event{
				Type: ConfigurationEvent,
				Obj:  cfgMap,
			}
		}
	}

	cmEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cfgMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				klog.Errorf("unexpected type: %T", obj)
			}
			key := k8s.MetaNamespaceKey(cfgMap)
			handleCfgMapEvent(key, cfgMap, "CREATE")
		},
		UpdateFunc: func(old, cur interface{}) {
			if reflect.DeepEqual(old, cur) {
				return
			}

			cfgMap, ok := cur.(*corev1.ConfigMap)
			if !ok {
				klog.Errorf("unexpected type: %T", cur)
			}
			key := k8s.MetaNamespaceKey(cfgMap)
			handleCfgMapEvent(key, cfgMap, "UPDATE")
		},
	}

	serviceHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				klog.Errorf("unexpected type: %T", obj)
			}
			if svc.Spec.Type == corev1.ServiceTypeExternalName {
				updateCh.In() <- Event{
					Type: CreateEvent,
					Obj:  obj,
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				// If we reached here it means the service was deleted but its final state is unrecorded.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.ErrorS(nil, "Error obtaining object from tombstone", "key", obj)
					return
				}
				svc, ok = tombstone.Obj.(*corev1.Service)
				if !ok {
					klog.Errorf("Tombstone contained object that is not a Service: %#v", obj)
					return
				}
			}
			if svc.Spec.Type == corev1.ServiceTypeExternalName {
				updateCh.In() <- Event{
					Type: DeleteEvent,
					Obj:  obj,
				}
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			oldSvc, ok := old.(*corev1.Service)
			if !ok {
				klog.Errorf("unexpected type: %T", old)
			}
			curSvc, ok := cur.(*corev1.Service)
			if !ok {
				klog.Errorf("unexpected type: %T", cur)
			}

			if reflect.DeepEqual(oldSvc, curSvc) {
				return
			}

			updateCh.In() <- Event{
				Type: UpdateEvent,
				Obj:  cur,
			}
		},
	}

	if _, err := store.informers.Ingress.AddEventHandler(ingEventHandler); err != nil {
		klog.Errorf("Error adding ingress event handler: %v", err)
	}
	if !icConfig.IgnoreIngressClass {
		if _, err := store.informers.IngressClass.AddEventHandler(ingressClassEventHandler); err != nil {
			klog.Errorf("Error adding ingress class event handler: %v", err)
		}
	}
	if _, err := store.informers.EndpointSlice.AddEventHandler(epsEventHandler); err != nil {
		klog.Errorf("Error adding endpoint slice event handler: %v", err)
	}
	if _, err := store.informers.Secret.AddEventHandler(secrEventHandler); err != nil {
		klog.Errorf("Error adding secret event handler: %v", err)
	}
	if _, err := store.informers.ConfigMap.AddEventHandler(cmEventHandler); err != nil {
		klog.Errorf("Error adding configmap event handler: %v", err)
	}
	if _, err := store.informers.Service.AddEventHandler(serviceHandler); err != nil {
		klog.Errorf("Error adding service event handler: %v", err)
	}

	// do not wait for informers to read the configmap configuration
	ns, name, err := k8s.ParseNameNS(configmap)
	if err != nil {
		klog.Errorf("unexpected error parsing name and ns: %v", err)
	}
	cm, err := client.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.Warningf("Unexpected error reading configuration configmap: %v", err)
	}

	store.setConfig(cm)
	return store
}

// hasCatchAllIngressRule returns whether or not an ingress produces a
// catch-all server, and so should be ignored when --disable-catch-all is set
func hasCatchAllIngressRule(spec networkingv1.IngressSpec) bool {
	return spec.DefaultBackend != nil
}

func checkBadAnnotationValue(annotationMap map[string]string, badwords string) error {
	arraybadWords := strings.Split(strings.TrimSpace(badwords), ",")

	for annotation, value := range annotationMap {
		if strings.HasPrefix(annotation, fmt.Sprintf("%s/", parser.AnnotationsPrefix)) {
			for _, forbiddenvalue := range arraybadWords {
				if strings.Contains(value, forbiddenvalue) {
					return fmt.Errorf("%s annotation contains invalid word %s", annotation, forbiddenvalue)
				}
			}
		}
	}
	return nil
}

// setCompressionPriority sets the compression priority from the backend config
func (s *k8sStore) setCompressionPriority() {
	if strings.TrimSpace(s.backendConfig.CompressionPriority) == "" {
		return
	}

	compPri := strings.Split(s.backendConfig.CompressionPriority, ",")

	// we expect user to use lower cases but be safe and normalize the input
	pri := strings.ToLower(strings.TrimSpace(compPri[0]))

	switch pri {
	case "brotli":
		// prefer brotli: enable brotli and (optionally) disable gzip so
		// the template only renders brotli directives. This enforces the
		// priority at config-render time.
		s.backendConfig.EnableBrotli = true
		s.backendConfig.UseGzip = false
	case "gzip":
		// prefer gzip: enable gzip and disable brotli
		s.backendConfig.UseGzip = true
		s.backendConfig.EnableBrotli = false
	default:
		klog.Warningf("unknown compression priority %q, defaulting to gzip", s.backendConfig.CompressionPriority)
		s.backendConfig.UseGzip = true
		s.backendConfig.EnableBrotli = false
	}
}

// syncIngress parses ingress annotations converting the value of the
// annotation to a go struct
func (s *k8sStore) syncIngress(ing *networkingv1.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	klog.V(3).Infof("updating annotations information for ingress %v", key)

	copyIng := &networkingv1.Ingress{}
	ing.ObjectMeta.DeepCopyInto(&copyIng.ObjectMeta)

	if s.backendConfig.AnnotationValueWordBlocklist != "" {
		if err := checkBadAnnotationValue(copyIng.Annotations, s.backendConfig.AnnotationValueWordBlocklist); err != nil {
			klog.Warningf("skipping ingress %s: %s", key, err)
			return
		}
	}

	ing.Spec.DeepCopyInto(&copyIng.Spec)
	ing.Status.DeepCopyInto(&copyIng.Status)

	for ri, rule := range copyIng.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		for pi, path := range rule.HTTP.Paths {
			if path.Path == "" {
				copyIng.Spec.Rules[ri].HTTP.Paths[pi].Path = "/"
			}
		}
	}

	k8s.SetDefaultNGINXPathType(copyIng)

	parsed, err := s.annotations.Extract(ing)
	if err != nil {
		klog.Error(err)
		return
	}
	if parsed.Denied != nil {
		s.recorder.Eventf(ing, corev1.EventTypeWarning, "AnnotationParsingFailed", fmt.Sprintf("Error parsing annotations: %v", *parsed.Denied))
	}
	err = s.listers.IngressWithAnnotation.Update(&ingress.Ingress{
		Ingress:           *copyIng,
		ParsedAnnotations: parsed,
	})
	if err != nil {
		klog.Error(err)
	}
}

// updateSecretIngressMap takes an Ingress and updates all Secret objects it
// references in secretIngressMap.
func (s *k8sStore) updateSecretIngressMap(ing *networkingv1.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	klog.V(3).Infof("updating references to secrets for ingress %v", key)

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
		"proxy-ssl-secret",
		"secure-verify-ca-secret",
	}

	secConfig := s.GetSecurityConfiguration().AllowCrossNamespaceResources
	for _, ann := range secretAnnotations {
		secrKey, err := objectRefAnnotationNsKey(ann, ing, secConfig)
		if err != nil && !errors.IsMissingAnnotations(err) {
			klog.Errorf("error reading secret reference in annotation %q: %s", ann, err)
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
func objectRefAnnotationNsKey(ann string, ing *networkingv1.Ingress, allowCrossNamespace bool) (string, error) {
	// We pass nil fields, as this is an internal process and we don't need to validate it.
	annValue, err := parser.GetStringAnnotation(ann, ing, nil)
	if err != nil {
		return "", err
	}

	secrNs, secrName, err := cache.SplitMetaNamespaceKey(annValue)
	if secrName == "" {
		return "", err
	}

	if secrNs == "" {
		return fmt.Sprintf("%v/%v", ing.Namespace, secrName), nil
	}
	if !allowCrossNamespace && secrNs != ing.Namespace {
		return "", fmt.Errorf("cross namespace secret is not supported")
	}
	return annValue, nil
}

// syncSecrets synchronizes data from all Secrets referenced by the given
// Ingress with the local store and file system.
func (s *k8sStore) syncSecrets(ing *networkingv1.Ingress) {
	key := k8s.MetaNamespaceKey(ing)
	for _, secrKey := range s.secretIngressMap.ReferencedBy(key) {
		s.syncSecret(secrKey)
	}
}

// GetSecret returns the Secret matching key.
func (s *k8sStore) GetSecret(key string) (*corev1.Secret, error) {
	return s.listers.Secret.ByKey(key)
}

// ListLocalSSLCerts returns the list of local SSLCerts
func (s *k8sStore) ListLocalSSLCerts() []*ingress.SSLCert {
	var certs []*ingress.SSLCert
	for _, item := range s.sslStore.List() {
		if s, ok := item.(*ingress.SSLCert); ok {
			certs = append(certs, s)
		}
	}

	return certs
}

// GetService returns the Service matching key.
func (s *k8sStore) GetService(key string) (*corev1.Service, error) {
	return s.listers.Service.ByKey(key)
}

func (s *k8sStore) GetIngressClass(ing *networkingv1.Ingress, icConfig *ingressclass.Configuration) (string, error) {
	// First we try ingressClassName
	if !icConfig.IgnoreIngressClass && ing.Spec.IngressClassName != nil {
		iclass, err := s.listers.IngressClass.ByKey(*ing.Spec.IngressClassName)
		if err != nil {
			return "", err
		}
		return iclass.Name, nil
	}

	// Then we try annotation
	if class, ok := ing.GetAnnotations()[ingressclass.IngressKey]; ok {
		if class != icConfig.AnnotationValue {
			return "", fmt.Errorf("ingress class annotation is not equal to the expected by Ingress Controller")
		}
		return class, nil
	}

	// Then we accept if the WithoutClass is enabled
	if icConfig.WatchWithoutClass {
		// Reserving "_" as a "wildcard" name
		return "_", nil
	}
	return "", fmt.Errorf("ingress does not contain a valid IngressClass")
}

// getIngress returns the Ingress matching key.
func (s *k8sStore) getIngress(key string) (*networkingv1.Ingress, error) {
	ing, err := s.listers.IngressWithAnnotation.ByKey(key)
	if err != nil {
		return nil, err
	}

	return &ing.Ingress, nil
}

func sortIngressSlice(ingresses []*ingress.Ingress) {
	// sort Ingresses using the CreationTimestamp field
	sort.SliceStable(ingresses, func(i, j int) bool {
		ir := ingresses[i].CreationTimestamp
		jr := ingresses[j].CreationTimestamp
		if ir.Equal(&jr) {
			in := fmt.Sprintf("%v/%v", ingresses[i].Namespace, ingresses[i].Name)
			jn := fmt.Sprintf("%v/%v", ingresses[j].Namespace, ingresses[j].Name)
			klog.V(3).Infof("Ingress %v and %v have identical CreationTimestamp", in, jn)
			return in > jn
		}
		return ir.Before(&jr)
	})
}

// ListIngresses returns the list of Ingresses
func (s *k8sStore) ListIngresses() []*ingress.Ingress {
	// filter ingress rules
	ingresses := make([]*ingress.Ingress, 0)
	for _, item := range s.listers.IngressWithAnnotation.List() {
		ing, ok := item.(*ingress.Ingress)
		if !ok {
			klog.Errorf("unexpected type: %T", item)
		}
		ingresses = append(ingresses, ing)
	}

	sortIngressSlice(ingresses)

	return ingresses
}

// GetLocalSSLCert returns the local copy of a SSLCert
func (s *k8sStore) GetLocalSSLCert(key string) (*ingress.SSLCert, error) {
	return s.sslStore.ByKey(key)
}

// GetConfigMap returns the ConfigMap matching key.
func (s *k8sStore) GetConfigMap(key string) (*corev1.ConfigMap, error) {
	return s.listers.ConfigMap.ByKey(key)
}

func (s *k8sStore) GetServiceEndpointsSlices(key string) ([]*discoveryv1.EndpointSlice, error) {
	return s.listers.EndpointSlice.MatchByKey(key)
}

// GetAuthCertificate is used by the auth-tls annotations to get a cert from a secret
func (s *k8sStore) GetAuthCertificate(name string) (*resolver.AuthSSLCert, error) {
	if _, err := s.GetLocalSSLCert(name); err != nil {
		s.syncSecret(name)
	}

	cert, err := s.GetLocalSSLCert(name)
	if err != nil {
		return nil, err
	}

	return &resolver.AuthSSLCert{
		Secret:      name,
		CAFileName:  cert.CAFileName,
		CASHA:       cert.CASHA,
		CRLFileName: cert.CRLFileName,
		CRLSHA:      cert.CRLSHA,
		PemFileName: cert.PemFileName,
	}, nil
}

func (s *k8sStore) writeSSLSessionTicketKey(cmap *corev1.ConfigMap, fileName string) {
	ticketString := ngx_template.ReadConfig(cmap.Data).SSLSessionTicketKey
	s.backendConfig.SSLSessionTicketKey = ""

	if ticketString != "" {
		ticketBytes := base64.StdEncoding.WithPadding(base64.StdPadding).DecodedLen(len(ticketString))

		// 81 used instead of 80 because of padding
		if !(ticketBytes == 48 || ticketBytes == 81) {
			klog.Warningf("ssl-session-ticket-key must contain either 48 or 80 bytes")
		}

		decodedTicket, err := base64.StdEncoding.DecodeString(ticketString)
		if err != nil {
			klog.Errorf("unexpected error decoding ssl-session-ticket-key: %v", err)
			return
		}

		err = os.WriteFile(fileName, decodedTicket, file.ReadWriteByUser)
		if err != nil {
			klog.Errorf("unexpected error writing ssl-session-ticket-key to %s: %v", fileName, err)
			return
		}

		s.backendConfig.SSLSessionTicketKey = ticketString
	}
}

// GetDefaultBackend returns the default backend
func (s *k8sStore) GetDefaultBackend() defaults.Backend {
	return s.GetBackendConfiguration().Backend
}

func (s *k8sStore) GetBackendConfiguration() ngx_config.Configuration {
	s.backendConfigMu.RLock()
	defer s.backendConfigMu.RUnlock()

	return s.backendConfig
}

func (s *k8sStore) GetSecurityConfiguration() defaults.SecurityConfiguration {
	s.backendConfigMu.RLock()
	defer s.backendConfigMu.RUnlock()

	secConfig := defaults.SecurityConfiguration{
		AllowCrossNamespaceResources: s.backendConfig.AllowCrossNamespaceResources,
		AnnotationsRiskLevel:         s.backendConfig.AnnotationsRiskLevel,
	}
	return secConfig
}

func (s *k8sStore) setConfig(cmap *corev1.ConfigMap) {
	s.backendConfigMu.Lock()
	defer s.backendConfigMu.Unlock()

	if cmap == nil {
		return
	}

	s.backendConfig = ngx_template.ReadConfig(cmap.Data)
	if s.backendConfig.UseGeoIP2 && !nginx.GeoLite2DBExists() {
		klog.Warning("The GeoIP2 feature is enabled but the databases are missing. Disabling")
		s.backendConfig.UseGeoIP2 = false
	}

	// Apply compression priority settings (if any) so that the running
	// backendConfig reflects the desired preference between brotli and gzip.
	s.setCompressionPriority()

	s.writeSSLSessionTicketKey(cmap, "/etc/ingress-controller/tickets.key")
}

// Run initiates the synchronization of the informers and the initial
// synchronization of the secrets.
func (s *k8sStore) Run(stopCh chan struct{}) {
	// start informers
	s.informers.Run(stopCh)
}

var runtimeScheme = k8sruntime.NewScheme()

func init() {
	runtime.Must(networkingv1.AddToScheme(runtimeScheme))
}

func toIngress(obj interface{}) (*networkingv1.Ingress, bool) {
	if ing, ok := obj.(*networkingv1.Ingress); ok {
		k8s.SetDefaultNGINXPathType(ing)
		return ing, true
	}

	return nil, false
}
