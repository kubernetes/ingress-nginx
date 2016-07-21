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

package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	podutil "k8s.io/kubernetes/pkg/api/pod"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/record"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/watch"

	"k8s.io/contrib/ingress/controllers/nginx/nginx"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/auth"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/healthcheck"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/ipwhitelist"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/ratelimit"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/rewrite"
	"k8s.io/contrib/ingress/controllers/nginx/nginx/secureupstream"
)

const (
	defUpstreamName          = "upstream-default-backend"
	defServerName            = "_"
	namedPortAnnotation      = "ingress.kubernetes.io/named-ports"
	podStoreSyncedPollPeriod = 1 * time.Second
	rootLocation             = "/"
)

var (
	keyFunc = framework.DeletionHandlingMetaNamespaceKeyFunc
)

type namedPortMapping map[string]string

// getPort returns the port defined in a named port
func (npm namedPortMapping) getPort(name string) (string, bool) {
	val, ok := npm.getPortMappings()[name]
	return val, ok
}

// getPortMappings returns the map containing the
// mapping of named port names and the port number
func (npm namedPortMapping) getPortMappings() map[string]string {
	data := npm[namedPortAnnotation]
	var mapping map[string]string
	if data == "" {
		return mapping
	}
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		glog.Errorf("unexpected error reading annotations: %v", err)
	}

	return mapping
}

// loadBalancerController watches the kubernetes api and adds/removes services
// from the loadbalancer
type loadBalancerController struct {
	client            *client.Client
	ingController     *framework.Controller
	endpController    *framework.Controller
	svcController     *framework.Controller
	secrController    *framework.Controller
	mapController     *framework.Controller
	ingLister         StoreToIngressLister
	svcLister         cache.StoreToServiceLister
	endpLister        cache.StoreToEndpointsLister
	secrLister        StoreToSecretsLister
	mapLister         StoreToConfigmapLister
	nginx             *nginx.Manager
	podInfo           *podInfo
	defaultSvc        string
	nxgConfigMap      string
	tcpConfigMap      string
	udpConfigMap      string
	defSSLCertificate string

	recorder record.EventRecorder

	syncQueue *taskQueue

	// taskQueue used to update the status of the Ingress rules.
	// this avoids a sync execution in the ResourceEventHandlerFuncs
	ingQueue *taskQueue

	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
	stopCh   chan struct{}
}

// newLoadBalancerController creates a controller for nginx loadbalancer
func newLoadBalancerController(kubeClient *client.Client, resyncPeriod time.Duration,
	defaultSvc, namespace, nxgConfigMapName, tcpConfigMapName, udpConfigMapName,
	defSSLCertificate string, runtimeInfo *podInfo) (*loadBalancerController, error) {

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(namespace))

	lbc := loadBalancerController{
		client:            kubeClient,
		stopCh:            make(chan struct{}),
		podInfo:           runtimeInfo,
		nginx:             nginx.NewManager(kubeClient),
		nxgConfigMap:      nxgConfigMapName,
		tcpConfigMap:      tcpConfigMapName,
		udpConfigMap:      udpConfigMapName,
		defSSLCertificate: defSSLCertificate,
		defaultSvc:        defaultSvc,
		recorder: eventBroadcaster.NewRecorder(api.EventSource{
			Component: "nginx-ingress-controller",
		}),
	}

	lbc.syncQueue = NewTaskQueue(lbc.sync)
	lbc.ingQueue = NewTaskQueue(lbc.updateIngressStatus)

	ingEventHandler := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			if !isNGINXIngress(addIng) {
				glog.Infof("Ignoring add for ingress %v based on annotation %v", addIng.Name, ingressClassKey)
				return
			}
			lbc.recorder.Eventf(addIng, api.EventTypeNormal, "CREATE", fmt.Sprintf("%s/%s", addIng.Namespace, addIng.Name))
			lbc.ingQueue.enqueue(obj)
			lbc.syncQueue.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			delIng := obj.(*extensions.Ingress)
			if !isNGINXIngress(delIng) {
				glog.Infof("Ignoring add for ingress %v based on annotation %v", delIng.Name, ingressClassKey)
				return
			}
			lbc.recorder.Eventf(delIng, api.EventTypeNormal, "DELETE", fmt.Sprintf("%s/%s", delIng.Namespace, delIng.Name))
			lbc.syncQueue.enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			curIng := cur.(*extensions.Ingress)
			if !isNGINXIngress(curIng) {
				return
			}
			if !reflect.DeepEqual(old, cur) {
				upIng := cur.(*extensions.Ingress)
				lbc.recorder.Eventf(upIng, api.EventTypeNormal, "UPDATE", fmt.Sprintf("%s/%s", upIng.Namespace, upIng.Name))
				lbc.ingQueue.enqueue(cur)
				lbc.syncQueue.enqueue(cur)
			}
		},
	}

	secrEventHandler := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addSecr := obj.(*api.Secret)
			if lbc.secrReferenced(addSecr.Namespace, addSecr.Name) {
				lbc.recorder.Eventf(addSecr, api.EventTypeNormal, "CREATE", fmt.Sprintf("%s/%s", addSecr.Namespace, addSecr.Name))
				lbc.syncQueue.enqueue(obj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			delSecr := obj.(*api.Secret)
			if lbc.secrReferenced(delSecr.Namespace, delSecr.Name) {
				lbc.recorder.Eventf(delSecr, api.EventTypeNormal, "DELETE", fmt.Sprintf("%s/%s", delSecr.Namespace, delSecr.Name))
				lbc.syncQueue.enqueue(obj)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				upSecr := cur.(*api.Secret)
				if lbc.secrReferenced(upSecr.Namespace, upSecr.Name) {
					lbc.recorder.Eventf(upSecr, api.EventTypeNormal, "UPDATE", fmt.Sprintf("%s/%s", upSecr.Namespace, upSecr.Name))
					lbc.syncQueue.enqueue(cur)
				}
			}
		},
	}

	eventHandler := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			lbc.syncQueue.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			lbc.syncQueue.enqueue(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				lbc.syncQueue.enqueue(cur)
			}
		},
	}

	mapEventHandler := framework.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				upCmap := cur.(*api.ConfigMap)
				mapKey := fmt.Sprintf("%s/%s", upCmap.Namespace, upCmap.Name)
				// updates to configuration configmaps can trigger an update
				if mapKey == lbc.nxgConfigMap || mapKey == lbc.tcpConfigMap || mapKey == lbc.udpConfigMap {
					lbc.recorder.Eventf(upCmap, api.EventTypeNormal, "UPDATE", mapKey)
					lbc.syncQueue.enqueue(cur)
				}
			}
		},
	}

	lbc.ingLister.Store, lbc.ingController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  ingressListFunc(lbc.client, namespace),
			WatchFunc: ingressWatchFunc(lbc.client, namespace),
		},
		&extensions.Ingress{}, resyncPeriod, ingEventHandler)

	lbc.endpLister.Store, lbc.endpController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  endpointsListFunc(lbc.client, namespace),
			WatchFunc: endpointsWatchFunc(lbc.client, namespace),
		},
		&api.Endpoints{}, resyncPeriod, eventHandler)

	lbc.svcLister.Store, lbc.svcController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  serviceListFunc(lbc.client, namespace),
			WatchFunc: serviceWatchFunc(lbc.client, namespace),
		},
		&api.Service{}, resyncPeriod, framework.ResourceEventHandlerFuncs{})

	lbc.secrLister.Store, lbc.secrController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  secretsListFunc(lbc.client, namespace),
			WatchFunc: secretsWatchFunc(lbc.client, namespace),
		},
		&api.Secret{}, resyncPeriod, secrEventHandler)

	lbc.mapLister.Store, lbc.mapController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  mapListFunc(lbc.client, namespace),
			WatchFunc: mapWatchFunc(lbc.client, namespace),
		},
		&api.ConfigMap{}, resyncPeriod, mapEventHandler)

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

func serviceListFunc(c *client.Client, ns string) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.Services(ns).List(opts)
	}
}

func serviceWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return c.Services(ns).Watch(options)
	}
}

func endpointsListFunc(c *client.Client, ns string) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.Endpoints(ns).List(opts)
	}
}

func endpointsWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return c.Endpoints(ns).Watch(options)
	}
}

func secretsListFunc(c *client.Client, ns string) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.Secrets(ns).List(opts)
	}
}

func secretsWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return c.Secrets(ns).Watch(options)
	}
}

func mapListFunc(c *client.Client, ns string) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return c.ConfigMaps(ns).List(opts)
	}
}

func mapWatchFunc(c *client.Client, ns string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return c.ConfigMaps(ns).Watch(options)
	}
}

func (lbc *loadBalancerController) controllersInSync() bool {
	return lbc.ingController.HasSynced() &&
		lbc.svcController.HasSynced() &&
		lbc.endpController.HasSynced() &&
		lbc.secrController.HasSynced() &&
		lbc.mapController.HasSynced()
}

func (lbc *loadBalancerController) getConfigMap(ns, name string) (*api.ConfigMap, error) {
	// TODO: check why lbc.mapLister.Store.GetByKey(mapKey) is not stable (random content)
	return lbc.client.ConfigMaps(ns).Get(name)
}

func (lbc *loadBalancerController) getTCPConfigMap(ns, name string) (*api.ConfigMap, error) {
	return lbc.getConfigMap(ns, name)
}

func (lbc *loadBalancerController) getUDPConfigMap(ns, name string) (*api.ConfigMap, error) {
	return lbc.getConfigMap(ns, name)
}

// checkSvcForUpdate verifies if one of the running pods for a service contains
// named port. If the annotation in the service does not exists or is not equals
// to the port mapping obtained from the pod the service must be updated to reflect
// the current state
func (lbc *loadBalancerController) checkSvcForUpdate(svc *api.Service) (map[string]string, error) {
	// get the pods associated with the service
	// TODO: switch this to a watch
	pods, err := lbc.client.Pods(svc.Namespace).List(api.ListOptions{
		LabelSelector: labels.Set(svc.Spec.Selector).AsSelector(),
	})

	namedPorts := map[string]string{}
	if err != nil {
		return namedPorts, fmt.Errorf("error searching service pods %v/%v: %v", svc.Namespace, svc.Name, err)
	}

	if len(pods.Items) == 0 {
		return namedPorts, nil
	}

	// we need to check only one pod searching for named ports
	pod := &pods.Items[0]
	glog.V(4).Infof("checking pod %v/%v for named port information", pod.Namespace, pod.Name)
	for i := range svc.Spec.Ports {
		servicePort := &svc.Spec.Ports[i]

		_, err := strconv.Atoi(servicePort.TargetPort.StrVal)
		if err != nil {
			portNum, err := podutil.FindPort(pod, servicePort)
			if err != nil {
				glog.V(4).Infof("failed to find port for service %s/%s: %v", svc.Namespace, svc.Name, err)
				continue
			}

			if servicePort.TargetPort.StrVal == "" {
				continue
			}

			namedPorts[servicePort.TargetPort.StrVal] = fmt.Sprintf("%v", portNum)
		}
	}

	if svc.ObjectMeta.Annotations == nil {
		svc.ObjectMeta.Annotations = map[string]string{}
	}

	curNamedPort := svc.ObjectMeta.Annotations[namedPortAnnotation]
	if len(namedPorts) > 0 && !reflect.DeepEqual(curNamedPort, namedPorts) {
		data, _ := json.Marshal(namedPorts)

		newSvc, err := lbc.client.Services(svc.Namespace).Get(svc.Name)
		if err != nil {
			return namedPorts, fmt.Errorf("error getting service %v/%v: %v", svc.Namespace, svc.Name, err)
		}

		if newSvc.ObjectMeta.Annotations == nil {
			newSvc.ObjectMeta.Annotations = map[string]string{}
		}

		newSvc.ObjectMeta.Annotations[namedPortAnnotation] = string(data)
		glog.Infof("updating service %v with new named port mappings", svc.Name)
		_, err = lbc.client.Services(svc.Namespace).Update(newSvc)
		if err != nil {
			return namedPorts, fmt.Errorf("error syncing service %v/%v: %v", svc.Namespace, svc.Name, err)
		}

		return newSvc.ObjectMeta.Annotations, nil
	}

	return namedPorts, nil
}

func (lbc *loadBalancerController) sync(key string) error {
	if !lbc.controllersInSync() {
		time.Sleep(podStoreSyncedPollPeriod)
		return fmt.Errorf("deferring sync till endpoints controller has synced")
	}

	// by default no custom configuration configmap
	cfg := &api.ConfigMap{}

	if lbc.nxgConfigMap != "" {
		// Search for custom configmap (defined in main args)
		var err error
		ns, name, _ := parseNsName(lbc.nxgConfigMap)
		cfg, err = lbc.getConfigMap(ns, name)
		if err != nil {
			return fmt.Errorf("unexpected error searching configmap %v: %v", lbc.nxgConfigMap, err)
		}
	}

	ngxConfig := lbc.nginx.ReadConfig(cfg)

	ings := lbc.ingLister.Store.List()
	upstreams, servers := lbc.getUpstreamServers(ngxConfig, ings)

	return lbc.nginx.CheckAndReload(ngxConfig, nginx.IngressConfig{
		Upstreams:    upstreams,
		Servers:      servers,
		TCPUpstreams: lbc.getTCPServices(),
		UDPUpstreams: lbc.getUDPServices(),
	})
}

func (lbc *loadBalancerController) updateIngressStatus(key string) error {
	if !lbc.controllersInSync() {
		time.Sleep(podStoreSyncedPollPeriod)
		return fmt.Errorf("deferring sync till endpoints controller has synced")
	}

	obj, ingExists, err := lbc.ingLister.Store.GetByKey(key)
	if err != nil {
		return err
	}

	if !ingExists {
		// TODO: what's the correct behavior here?
		return nil
	}

	ing := obj.(*extensions.Ingress)

	ingClient := lbc.client.Extensions().Ingress(ing.Namespace)

	currIng, err := ingClient.Get(ing.Name)
	if err != nil {
		return fmt.Errorf("unexpected error searching Ingress %v/%v: %v", ing.Namespace, ing.Name, err)
	}

	lbIPs := ing.Status.LoadBalancer.Ingress
	if !lbc.isStatusIPDefined(lbIPs) {
		glog.Infof("Updating loadbalancer %v/%v with IP %v", ing.Namespace, ing.Name, lbc.podInfo.NodeIP)
		currIng.Status.LoadBalancer.Ingress = append(currIng.Status.LoadBalancer.Ingress, api.LoadBalancerIngress{
			IP: lbc.podInfo.NodeIP,
		})
		if _, err := ingClient.UpdateStatus(currIng); err != nil {
			lbc.recorder.Eventf(currIng, api.EventTypeWarning, "UPDATE", "error: %v", err)
			return err
		}

		lbc.recorder.Eventf(currIng, api.EventTypeNormal, "CREATE", "ip: %v", lbc.podInfo.NodeIP)
	}

	return nil
}

func (lbc *loadBalancerController) isStatusIPDefined(lbings []api.LoadBalancerIngress) bool {
	for _, lbing := range lbings {
		if lbing.IP == lbc.podInfo.NodeIP {
			return true
		}
	}

	return false
}

func (lbc *loadBalancerController) getTCPServices() []*nginx.Location {
	if lbc.tcpConfigMap == "" {
		// no configmap for TCP services
		return []*nginx.Location{}
	}

	ns, name, err := parseNsName(lbc.tcpConfigMap)
	if err != nil {
		glog.Warningf("%v", err)
		return []*nginx.Location{}
	}
	tcpMap, err := lbc.getTCPConfigMap(ns, name)
	if err != nil {
		glog.V(3).Infof("no configured tcp services found: %v", err)
		return []*nginx.Location{}
	}

	return lbc.getStreamServices(tcpMap.Data, api.ProtocolTCP)
}

func (lbc *loadBalancerController) getUDPServices() []*nginx.Location {
	if lbc.udpConfigMap == "" {
		// no configmap for TCP services
		return []*nginx.Location{}
	}

	ns, name, err := parseNsName(lbc.udpConfigMap)
	if err != nil {
		glog.Warningf("%v", err)
		return []*nginx.Location{}
	}
	tcpMap, err := lbc.getUDPConfigMap(ns, name)
	if err != nil {
		glog.V(3).Infof("no configured tcp services found: %v", err)
		return []*nginx.Location{}
	}

	return lbc.getStreamServices(tcpMap.Data, api.ProtocolUDP)
}

func (lbc *loadBalancerController) getStreamServices(data map[string]string, proto api.Protocol) []*nginx.Location {
	var svcs []*nginx.Location
	// k -> port to expose in nginx
	// v -> <namespace>/<service name>:<port from service to be used>
	for k, v := range data {
		port, err := strconv.Atoi(k)
		if err != nil {
			glog.Warningf("%v is not valid as a TCP port", k)
			continue
		}

		// this ports are required for NGINX
		if k == "80" || k == "443" || k == "8181" {
			glog.Warningf("port %v cannot be used for TCP or UDP services. Is reserved for NGINX", k)
			continue
		}

		nsSvcPort := strings.Split(v, ":")
		if len(nsSvcPort) != 2 {
			glog.Warningf("invalid format (namespace/name:port) '%v'", k)
			continue
		}

		nsName := nsSvcPort[0]
		svcPort := nsSvcPort[1]

		svcNs, svcName, err := parseNsName(nsName)
		if err != nil {
			glog.Warningf("%v", err)
			continue
		}

		svcObj, svcExists, err := lbc.svcLister.Store.GetByKey(nsName)
		if err != nil {
			glog.Warningf("error getting service %v: %v", nsName, err)
			continue
		}

		if !svcExists {
			glog.Warningf("service %v was not found", nsName)
			continue
		}

		svc := svcObj.(*api.Service)

		var endps []nginx.UpstreamServer
		targetPort, err := strconv.Atoi(svcPort)
		if err != nil {
			for _, sp := range svc.Spec.Ports {
				if sp.Name == svcPort {
					endps = lbc.getEndpoints(svc, sp.TargetPort, proto, &healthcheck.Upstream{})
					break
				}
			}
		} else {
			// we need to use the TargetPort (where the endpoints are running)
			for _, sp := range svc.Spec.Ports {
				if sp.Port == int32(targetPort) {
					endps = lbc.getEndpoints(svc, sp.TargetPort, proto, &healthcheck.Upstream{})
					break
				}
			}
		}

		// tcp upstreams cannot contain empty upstreams and there is no
		// default backend equivalent for TCP
		if len(endps) == 0 {
			glog.Warningf("service %v/%v does not have any active endpoints", svcNs, svcName)
			continue
		}

		svcs = append(svcs, &nginx.Location{
			Path: k,
			Upstream: nginx.Upstream{
				Name:     fmt.Sprintf("%v-%v-%v", svcNs, svcName, port),
				Backends: endps,
			},
		})
	}

	return svcs
}

func (lbc *loadBalancerController) getDefaultUpstream() *nginx.Upstream {
	upstream := &nginx.Upstream{
		Name: defUpstreamName,
	}
	svcKey := lbc.defaultSvc
	svcObj, svcExists, err := lbc.svcLister.Store.GetByKey(svcKey)
	if err != nil {
		glog.Warningf("unexpected error searching the default backend %v: %v", lbc.defaultSvc, err)
		upstream.Backends = append(upstream.Backends, nginx.NewDefaultServer())
		return upstream
	}

	if !svcExists {
		glog.Warningf("service %v does not exists", svcKey)
		upstream.Backends = append(upstream.Backends, nginx.NewDefaultServer())
		return upstream
	}

	svc := svcObj.(*api.Service)

	endps := lbc.getEndpoints(svc, svc.Spec.Ports[0].TargetPort, api.ProtocolTCP, &healthcheck.Upstream{})
	if len(endps) == 0 {
		glog.Warningf("service %v does not have any active endpoints", svcKey)
		upstream.Backends = append(upstream.Backends, nginx.NewDefaultServer())
	} else {
		upstream.Backends = append(upstream.Backends, endps...)
	}

	return upstream
}

func (lbc *loadBalancerController) getUpstreamServers(ngxCfg config.Configuration, data []interface{}) ([]*nginx.Upstream, []*nginx.Server) {
	upstreams := lbc.createUpstreams(ngxCfg, data)
	upstreams[defUpstreamName] = lbc.getDefaultUpstream()

	servers := lbc.createServers(data)
	if _, ok := servers[defServerName]; !ok {
		// default server - no servername.
		// there is no rule with default backend
		servers[defServerName] = &nginx.Server{
			Name: defServerName,
			Locations: []*nginx.Location{{
				Path:         rootLocation,
				IsDefBackend: true,
				Upstream:     *lbc.getDefaultUpstream(),
			},
			},
		}
	}

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			if rule.IngressRuleValue.HTTP == nil {
				continue
			}

			nginxAuth, err := auth.ParseAnnotations(lbc.client, ing, auth.DefAuthDirectory)
			glog.V(3).Infof("nginx auth %v", nginxAuth)
			if err != nil {
				glog.V(3).Infof("error reading authentication in Ingress %v/%v: %v", ing.GetNamespace(), ing.GetName(), err)
			}

			rl, err := ratelimit.ParseAnnotations(ing)
			glog.V(3).Infof("nginx rate limit %v", rl)
			if err != nil {
				glog.V(3).Infof("error reading rate limit annotation in Ingress %v/%v: %v", ing.GetNamespace(), ing.GetName(), err)
			}

			secUpstream, err := secureupstream.ParseAnnotations(ing)
			if err != nil {
				glog.V(3).Infof("error reading secure upstream in Ingress %v/%v: %v", ing.GetNamespace(), ing.GetName(), err)
			}

			locRew, err := rewrite.ParseAnnotations(ngxCfg, ing)
			if err != nil {
				glog.V(3).Infof("error parsing rewrite annotations for Ingress rule %v/%v: %v", ing.GetNamespace(), ing.GetName(), err)
			}

			wl, err := ipwhitelist.ParseAnnotations(ngxCfg.WhitelistSourceRange, ing)
			glog.V(3).Infof("nginx white list %v", wl)
			if err != nil {
				glog.V(3).Infof("error reading white list annotation in Ingress %v/%v: %v", ing.GetNamespace(), ing.GetName(), err)
			}

			host := rule.Host
			if host == "" {
				host = defServerName
			}
			server := servers[host]
			if server == nil {
				server = servers["_"]
			}

			for _, path := range rule.HTTP.Paths {
				upsName := fmt.Sprintf("%v-%v-%v", ing.GetNamespace(), path.Backend.ServiceName, path.Backend.ServicePort.String())
				ups := upstreams[upsName]

				nginxPath := path.Path
				// if there's no path defined we assume /
				if nginxPath == "" {
					lbc.recorder.Eventf(ing, api.EventTypeWarning, "MAPPING",
						"Ingress rule '%v/%v' contains no path definition. Assuming /", ing.GetNamespace(), ing.GetName())
					nginxPath = rootLocation
				}

				// Validate that there is no another previuous
				// rule for the same host and path.
				addLoc := true
				for _, loc := range server.Locations {
					if loc.Path == rootLocation && nginxPath == rootLocation && loc.IsDefBackend {
						loc.Upstream = *ups
						loc.Auth = *nginxAuth
						loc.RateLimit = *rl
						loc.Redirect = *locRew
						loc.SecureUpstream = secUpstream
						loc.Whitelist = *wl

						addLoc = false
						continue
					}

					if loc.Path == nginxPath {
						lbc.recorder.Eventf(ing, api.EventTypeWarning, "MAPPING",
							"Path '%v' already defined in another Ingress rule", nginxPath)
						addLoc = false
						break
					}
				}

				if addLoc {

					server.Locations = append(server.Locations, &nginx.Location{
						Path:           nginxPath,
						Upstream:       *ups,
						Auth:           *nginxAuth,
						RateLimit:      *rl,
						Redirect:       *locRew,
						SecureUpstream: secUpstream,
						Whitelist:      *wl,
					})
				}
			}
		}
	}

	// TODO: find a way to make this more readable
	// The structs must be ordered to always generate the same file
	// if the content does not change.
	aUpstreams := make([]*nginx.Upstream, 0, len(upstreams))
	for _, value := range upstreams {
		if len(value.Backends) == 0 {
			glog.Warningf("upstream %v does not have any active endpoints. Using default backend", value.Name)
			value.Backends = append(value.Backends, nginx.NewDefaultServer())
		}
		sort.Sort(nginx.UpstreamServerByAddrPort(value.Backends))
		aUpstreams = append(aUpstreams, value)
	}
	sort.Sort(nginx.UpstreamByNameServers(aUpstreams))

	aServers := make([]*nginx.Server, 0, len(servers))
	for _, value := range servers {
		sort.Sort(nginx.LocationByPath(value.Locations))
		aServers = append(aServers, value)
	}
	sort.Sort(nginx.ServerByName(aServers))

	return aUpstreams, aServers
}

// createUpstreams creates the NGINX upstreams for each service referenced in
// Ingress rules. The servers inside the upstream are endpoints.
func (lbc *loadBalancerController) createUpstreams(ngxCfg config.Configuration, data []interface{}) map[string]*nginx.Upstream {
	upstreams := make(map[string]*nginx.Upstream)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		hz := healthcheck.ParseAnnotations(ngxCfg, ing)

		for _, rule := range ing.Spec.Rules {
			if rule.IngressRuleValue.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				name := fmt.Sprintf("%v-%v-%v", ing.GetNamespace(), path.Backend.ServiceName, path.Backend.ServicePort.String())
				if _, ok := upstreams[name]; ok {
					continue
				}

				glog.V(3).Infof("creating upstream %v", name)
				upstreams[name] = nginx.NewUpstream(name)

				svcKey := fmt.Sprintf("%v/%v", ing.GetNamespace(), path.Backend.ServiceName)
				svcObj, svcExists, err := lbc.svcLister.Store.GetByKey(svcKey)

				if err != nil {
					glog.Infof("error getting service %v from the cache: %v", svcKey, err)
					continue
				}

				if !svcExists {
					glog.Warningf("service %v does not exists", svcKey)
					continue
				}

				svc := svcObj.(*api.Service)
				glog.V(3).Infof("obtaining port information for service %v", svcKey)
				bp := path.Backend.ServicePort.String()
				for _, servicePort := range svc.Spec.Ports {
					// targetPort could be a string, use the name or the port (int)
					if strconv.Itoa(int(servicePort.Port)) == bp || servicePort.TargetPort.String() == bp || servicePort.Name == bp {
						endps := lbc.getEndpoints(svc, servicePort.TargetPort, api.ProtocolTCP, hz)
						if len(endps) == 0 {
							glog.Warningf("service %v does not have any active endpoints", svcKey)
						}

						upstreams[name].Backends = append(upstreams[name].Backends, endps...)
						break
					}
				}
			}
		}
	}

	return upstreams
}

func (lbc *loadBalancerController) createServers(data []interface{}) map[string]*nginx.Server {
	servers := make(map[string]*nginx.Server)

	pems := lbc.getPemsFromIngress(data)

	var ngxCert nginx.SSLCert
	var err error

	if lbc.defSSLCertificate == "" {
		// use system certificated generated at image build time
		cert, key := getFakeSSLCert()
		ngxCert, err = lbc.nginx.AddOrUpdateCertAndKey("system-snake-oil-certificate", cert, key)
	} else {
		ngxCert, err = lbc.getPemCertificate(lbc.defSSLCertificate)
	}

	if err == nil {
		pems["_"] = ngxCert
	} else {
		glog.Warningf("%v", err)
	}

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}

			if _, ok := servers[host]; !ok {
				locs := []*nginx.Location{}
				locs = append(locs, &nginx.Location{
					Path:         rootLocation,
					IsDefBackend: true,
					Upstream:     *lbc.getDefaultUpstream(),
				})
				servers[host] = &nginx.Server{Name: host, Locations: locs}
			}

			if ngxCert, ok := pems[host]; ok {
				server := servers[host]
				server.SSL = true
				server.SSLCertificate = ngxCert.PemFileName
				server.SSLCertificateKey = ngxCert.PemFileName
				server.SSLPemChecksum = ngxCert.PemSHA
			}
		}
	}

	return servers
}

func (lbc *loadBalancerController) getPemsFromIngress(data []interface{}) map[string]nginx.SSLCert {
	pems := make(map[string]nginx.SSLCert)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)
		for _, tls := range ing.Spec.TLS {
			secretName := tls.SecretName
			secretKey := fmt.Sprintf("%s/%s", ing.Namespace, secretName)

			ngxCert, err := lbc.getPemCertificate(secretKey)
			if err != nil {
				glog.Warningf("%v", err)
				continue
			}

			for _, host := range tls.Hosts {
				if isHostValid(host, ngxCert.CN) {
					pems[host] = ngxCert
				} else {
					glog.Warningf("SSL Certificate stored in secret %v is not valid for the host %v defined in the Ingress rule %v", secretName, host, ing.Name)
				}
			}
		}
	}

	return pems
}

func (lbc *loadBalancerController) getPemCertificate(secretName string) (nginx.SSLCert, error) {
	secretInterface, exists, err := lbc.secrLister.Store.GetByKey(secretName)
	if err != nil {
		return nginx.SSLCert{}, fmt.Errorf("Error retriveing secret %v: %v", secretName, err)
	}
	if !exists {
		return nginx.SSLCert{}, fmt.Errorf("Secret %v does not exists", secretName)
	}

	secret := secretInterface.(*api.Secret)
	cert, ok := secret.Data[api.TLSCertKey]
	if !ok {
		return nginx.SSLCert{}, fmt.Errorf("Secret %v has no private key", secretName)
	}
	key, ok := secret.Data[api.TLSPrivateKeyKey]
	if !ok {
		return nginx.SSLCert{}, fmt.Errorf("Secret %v has no cert", secretName)
	}

	nsSecName := strings.Replace(secretName, "/", "-", -1)
	return lbc.nginx.AddOrUpdateCertAndKey(nsSecName, string(cert), string(key))
}

// check if secret is referenced in this controller's config
func (lbc *loadBalancerController) secrReferenced(namespace string, name string) bool {
	for _, ingIf := range lbc.ingLister.Store.List() {
		ing := ingIf.(*extensions.Ingress)
		if ing.Namespace != namespace {
			continue
		}
		for _, tls := range ing.Spec.TLS {
			if tls.SecretName == name {
				return true
			}
		}
	}
	return false
}

// getEndpoints returns a list of <endpoint ip>:<port> for a given service/target port combination.
func (lbc *loadBalancerController) getEndpoints(s *api.Service, servicePort intstr.IntOrString, proto api.Protocol, hz *healthcheck.Upstream) []nginx.UpstreamServer {
	glog.V(3).Infof("getting endpoints for service %v/%v and port %v", s.Namespace, s.Name, servicePort.String())
	ep, err := lbc.endpLister.GetServiceEndpoints(s)
	if err != nil {
		glog.Warningf("unexpected error obtaining service endpoints: %v", err)
		return []nginx.UpstreamServer{}
	}

	upsServers := []nginx.UpstreamServer{}

	for _, ss := range ep.Subsets {
		for _, epPort := range ss.Ports {

			if !reflect.DeepEqual(epPort.Protocol, proto) {
				continue
			}

			var targetPort int32
			switch servicePort.Type {
			case intstr.Int:
				if int(epPort.Port) == servicePort.IntValue() {
					targetPort = epPort.Port
				}
			case intstr.String:
				namedPorts := s.ObjectMeta.Annotations
				val, ok := namedPortMapping(namedPorts).getPort(servicePort.StrVal)
				if ok {
					port, err := strconv.Atoi(val)
					if err != nil {
						glog.Warningf("%v is not valid as a port", val)
						continue
					}

					targetPort = int32(port)
				} else {
					newnp, err := lbc.checkSvcForUpdate(s)
					if err != nil {
						glog.Warningf("error mapping service ports: %v", err)
						continue
					}
					val, ok := namedPortMapping(newnp).getPort(servicePort.StrVal)
					if ok {
						port, err := strconv.Atoi(val)
						if err != nil {
							glog.Warningf("%v is not valid as a port", val)
							continue
						}

						targetPort = int32(port)
					}
				}
			}

			if targetPort == 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ups := nginx.UpstreamServer{
					Address:     epAddress.IP,
					Port:        fmt.Sprintf("%v", targetPort),
					MaxFails:    hz.MaxFails,
					FailTimeout: hz.FailTimeout,
				}
				upsServers = append(upsServers, ups)
			}
		}
	}

	glog.V(3).Infof("endpoints found: %v", upsServers)
	return upsServers
}

// Stop stops the loadbalancer controller.
func (lbc *loadBalancerController) Stop() error {
	// Stop is invoked from the http endpoint.
	lbc.stopLock.Lock()
	defer lbc.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !lbc.shutdown {
		lbc.shutdown = true
		close(lbc.stopCh)

		ings := lbc.ingLister.Store.List()
		glog.Infof("removing IP address %v from ingress rules", lbc.podInfo.NodeIP)
		lbc.removeFromIngress(ings)

		glog.Infof("Shutting down controller queues.")
		lbc.syncQueue.shutdown()
		lbc.ingQueue.shutdown()

		return nil
	}

	return fmt.Errorf("shutdown already in progress")
}

// removeFromIngress removes the IP address of the node where the Ingres
// controller is running before shutdown to avoid incorrect status
// information in Ingress rules
func (lbc *loadBalancerController) removeFromIngress(ings []interface{}) {
	glog.Infof("updating %v Ingress rule/s", len(ings))
	for _, cur := range ings {
		ing := cur.(*extensions.Ingress)

		ingClient := lbc.client.Extensions().Ingress(ing.Namespace)
		currIng, err := ingClient.Get(ing.Name)
		if err != nil {
			glog.Errorf("unexpected error searching Ingress %v/%v: %v", ing.Namespace, ing.Name, err)
			continue
		}

		lbIPs := ing.Status.LoadBalancer.Ingress
		if len(lbIPs) > 0 && lbc.isStatusIPDefined(lbIPs) {
			glog.Infof("Updating loadbalancer %v/%v. Removing IP %v", ing.Namespace, ing.Name, lbc.podInfo.NodeIP)

			for idx, lbStatus := range currIng.Status.LoadBalancer.Ingress {
				if lbStatus.IP == lbc.podInfo.NodeIP {
					currIng.Status.LoadBalancer.Ingress = append(currIng.Status.LoadBalancer.Ingress[:idx],
						currIng.Status.LoadBalancer.Ingress[idx+1:]...)
					break
				}
			}

			if _, err := ingClient.UpdateStatus(currIng); err != nil {
				lbc.recorder.Eventf(currIng, api.EventTypeWarning, "UPDATE", "error: %v", err)
				continue
			}

			lbc.recorder.Eventf(currIng, api.EventTypeNormal, "DELETE", "ip: %v", lbc.podInfo.NodeIP)
		}
	}
}

// Run starts the loadbalancer controller.
func (lbc *loadBalancerController) Run() {
	glog.Infof("starting NGINX loadbalancer controller")
	go lbc.nginx.Start()

	go lbc.ingController.Run(lbc.stopCh)
	go lbc.endpController.Run(lbc.stopCh)
	go lbc.svcController.Run(lbc.stopCh)
	go lbc.secrController.Run(lbc.stopCh)
	go lbc.mapController.Run(lbc.stopCh)

	go lbc.syncQueue.run(time.Second, lbc.stopCh)
	go lbc.ingQueue.run(time.Second, lbc.stopCh)

	<-lbc.stopCh
}
