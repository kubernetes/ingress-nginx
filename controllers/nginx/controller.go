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
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/watch"

	"k8s.io/contrib/ingress/controllers/nginx/nginx"
)

const (
	defUpstreamName = "upstream-default-backend"
)

var (
	keyFunc = framework.DeletionHandlingMetaNamespaceKeyFunc
)

// loadBalancerController watches the kubernetes api and adds/removes services
// from the loadbalancer
type loadBalancerController struct {
	client         *client.Client
	ingController  *framework.Controller
	endpController *framework.Controller
	svcController  *framework.Controller
	ingLister      StoreToIngressLister
	svcLister      cache.StoreToServiceLister
	endpLister     cache.StoreToEndpointsLister
	nginx          *nginx.Manager
	lbInfo         *lbInfo
	defaultSvc     string
	nxgConfigMap   string
	tcpConfigMap   string
	udpConfigMap   string

	syncQueue *taskQueue

	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
	stopCh   chan struct{}
}

// newLoadBalancerController creates a controller for nginx loadbalancer
func newLoadBalancerController(kubeClient *client.Client, resyncPeriod time.Duration, defaultSvc,
	namespace, nxgConfigMapName, tcpConfigMapName, udpConfigMapName string, lbRuntimeInfo *lbInfo) (*loadBalancerController, error) {
	lbc := loadBalancerController{
		client:       kubeClient,
		stopCh:       make(chan struct{}),
		lbInfo:       lbRuntimeInfo,
		nginx:        nginx.NewManager(kubeClient),
		nxgConfigMap: nxgConfigMapName,
		tcpConfigMap: tcpConfigMapName,
		udpConfigMap: udpConfigMapName,
		defaultSvc:   defaultSvc,
	}

	lbc.syncQueue = NewTaskQueue(lbc.sync)

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

	lbc.ingLister.Store, lbc.ingController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  ingressListFunc(lbc.client, namespace),
			WatchFunc: ingressWatchFunc(lbc.client, namespace),
		},
		&extensions.Ingress{}, resyncPeriod, eventHandler)

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

func (lbc *loadBalancerController) controllersInSync() bool {
	return lbc.ingController.HasSynced() && lbc.svcController.HasSynced() && lbc.endpController.HasSynced()
}

func (lbc *loadBalancerController) getConfigMap(ns, name string) (*api.ConfigMap, error) {
	return lbc.client.ConfigMaps(ns).Get(name)
}

func (lbc *loadBalancerController) getTCPConfigMap(ns, name string) (*api.ConfigMap, error) {
	return lbc.client.ConfigMaps(ns).Get(name)
}

func (lbc *loadBalancerController) getUDPConfigMap(ns, name string) (*api.ConfigMap, error) {
	return lbc.client.ConfigMaps(ns).Get(name)
}

func (lbc *loadBalancerController) sync(key string) {
	if !lbc.controllersInSync() {
		lbc.syncQueue.requeue(key, fmt.Errorf("deferring sync till endpoints controller has synced"))
		return
	}

	ings := lbc.ingLister.Store.List()
	upstreams, servers := lbc.getUpstreamServers(ings)

	var cfg *api.ConfigMap

	ns, name, _ := parseNsName(lbc.nxgConfigMap)
	cfg, err := lbc.getConfigMap(ns, name)
	if err != nil {
		cfg = &api.ConfigMap{}
	}

	ngxConfig := lbc.nginx.ReadConfig(cfg)
	lbc.nginx.CheckAndReload(ngxConfig, nginx.IngressConfig{
		Upstreams:    upstreams,
		Servers:      servers,
		TCPUpstreams: lbc.getTCPServices(),
		UDPUpstreams: lbc.getUDPServices(),
	})
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

	return lbc.getServices(tcpMap.Data, api.ProtocolTCP)
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

	return lbc.getServices(tcpMap.Data, api.ProtocolUDP)
}

func (lbc *loadBalancerController) getServices(data map[string]string, proto api.Protocol) []*nginx.Location {
	var svcs []*nginx.Location
	// k -> port to expose in nginx
	// v -> <namespace>/<service name>:<port from service to be used>
	for k, v := range data {
		port, err := strconv.Atoi(k)
		if err != nil {
			glog.Warningf("%v is not valid as a TCP port", k)
			continue
		}

		svcPort := strings.Split(v, ":")
		if len(svcPort) != 2 {
			glog.Warningf("invalid format (namespace/name:port) '%v'", k)
			continue
		}

		svcNs, svcName, err := parseNsName(svcPort[0])
		if err != nil {
			glog.Warningf("%v", err)
			continue
		}

		svcObj, svcExists, err := lbc.svcLister.Store.GetByKey(svcPort[0])
		if err != nil {
			glog.Warningf("error getting service %v: %v", svcPort[0], err)
			continue
		}

		if !svcExists {
			glog.Warningf("service %v was not found", svcPort[0])
			continue
		}

		svc := svcObj.(*api.Service)

		var endps []nginx.UpstreamServer
		targetPort, err := strconv.Atoi(svcPort[1])
		if err != nil {
			endps = lbc.getEndpoints(svc, intstr.FromString(svcPort[1]), proto)
		} else {
			// we need to use the TargetPort (where the endpoints are running)
			for _, sp := range svc.Spec.Ports {
				if sp.Port == targetPort {
					endps = lbc.getEndpoints(svc, sp.TargetPort, proto)
					break
				}
			}
		}

		// tcp upstreams cannot contain empty upstreams and there is no
		// default backend equivalent for TCP
		if len(endps) == 0 {
			glog.Warningf("service %v/%v does no have any active endpoints", svcNs, svcName)
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
		glog.Warningf("service %v does no exists", svcKey)
		upstream.Backends = append(upstream.Backends, nginx.NewDefaultServer())
		return upstream
	}

	svc := svcObj.(*api.Service)

	endps := lbc.getEndpoints(svc, svc.Spec.Ports[0].TargetPort, api.ProtocolTCP)
	if len(endps) == 0 {
		glog.Warningf("service %v does no have any active endpoints", svcKey)
		upstream.Backends = append(upstream.Backends, nginx.NewDefaultServer())
	} else {
		upstream.Backends = append(upstream.Backends, endps...)
	}

	return upstream
}

func (lbc *loadBalancerController) getUpstreamServers(data []interface{}) ([]*nginx.Upstream, []*nginx.Server) {
	upstreams := lbc.createUpstreams(data)
	servers := lbc.createServers(data)

	upstreams[defUpstreamName] = lbc.getDefaultUpstream()

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			if rule.IngressRuleValue.HTTP == nil {
				continue
			}

			server := servers[rule.Host]
			locations := []*nginx.Location{}

			for _, path := range rule.HTTP.Paths {
				upsName := fmt.Sprintf("%v-%v-%v", ing.GetNamespace(), path.Backend.ServiceName, path.Backend.ServicePort.IntValue())
				ups := upstreams[upsName]

				svcKey := fmt.Sprintf("%v/%v", ing.GetNamespace(), path.Backend.ServiceName)
				svcObj, svcExists, err := lbc.svcLister.Store.GetByKey(svcKey)
				if err != nil {
					glog.Infof("error getting service %v from the cache: %v", svcKey, err)
					continue
				}

				if !svcExists {
					glog.Warningf("service %v does no exists", svcKey)
					continue
				}

				svc := svcObj.(*api.Service)

				for _, servicePort := range svc.Spec.Ports {
					if servicePort.Port == path.Backend.ServicePort.IntValue() {
						endps := lbc.getEndpoints(svc, servicePort.TargetPort, api.ProtocolTCP)
						if len(endps) == 0 {
							glog.Warningf("service %v does no have any active endpoints", svcKey)
						}

						ups.Backends = append(ups.Backends, endps...)
						break
					}
				}

				for _, ups := range upstreams {
					if upsName == ups.Name {
						loc := &nginx.Location{Path: path.Path}
						loc.Upstream = *ups
						locations = append(locations, loc)
						break
					}
				}
			}

			for _, loc := range locations {
				server.Locations = append(server.Locations, loc)
			}
		}
	}

	// TODO: find a way to make this more readable
	// The structs must be ordered to always generate the same file
	// if the content does not change.
	aUpstreams := make([]*nginx.Upstream, 0, len(upstreams))
	for _, value := range upstreams {
		if len(value.Backends) == 0 {
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

func (lbc *loadBalancerController) createUpstreams(data []interface{}) map[string]*nginx.Upstream {
	upstreams := make(map[string]*nginx.Upstream)
	upstreams[defUpstreamName] = nginx.NewUpstream(defUpstreamName)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			if rule.IngressRuleValue.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				name := fmt.Sprintf("%v-%v-%v", ing.GetNamespace(), path.Backend.ServiceName, path.Backend.ServicePort.IntValue())
				if _, ok := upstreams[name]; !ok {
					upstreams[name] = nginx.NewUpstream(name)
				}
			}
		}
	}

	return upstreams
}

func (lbc *loadBalancerController) createServers(data []interface{}) map[string]*nginx.Server {
	servers := make(map[string]*nginx.Server)

	pems := lbc.getPemsFromIngress(data)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			if _, ok := servers[rule.Host]; !ok {
				servers[rule.Host] = &nginx.Server{Name: rule.Host, Locations: []*nginx.Location{}}
			}

			if pemFile, ok := pems[rule.Host]; ok {
				server := servers[rule.Host]
				server.SSL = true
				server.SSLCertificate = pemFile
				server.SSLCertificateKey = pemFile
			}
		}
	}

	return servers
}

func (lbc *loadBalancerController) getPemsFromIngress(data []interface{}) map[string]string {
	pems := make(map[string]string)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, tls := range ing.Spec.TLS {
			secretName := tls.SecretName
			secret, err := lbc.client.Secrets(ing.Namespace).Get(secretName)
			if err != nil {
				glog.Warningf("Error retriveing secret %v for ing %v: %v", secretName, ing.Name, err)
				continue
			}
			cert, ok := secret.Data[api.TLSCertKey]
			if !ok {
				glog.Warningf("Secret %v has no private key", secretName)
				continue
			}
			key, ok := secret.Data[api.TLSPrivateKeyKey]
			if !ok {
				glog.Warningf("Secret %v has no cert", secretName)
				continue
			}

			pemFileName := lbc.nginx.AddOrUpdateCertAndKey(secretName, string(cert), string(key))
			cn, err := lbc.nginx.CheckSSLCertificate(pemFileName)
			if err != nil {
				glog.Warningf("No valid SSL certificate found in secret %v", secretName)
				continue
			}

			for _, host := range tls.Hosts {
				if isHostValid(host, cn) {
					pems[host] = pemFileName
				} else {
					glog.Warningf("SSL Certificate stored in secret %v is not valid for the host %v defined in the Ingress rule %v", secretName, host, ing.Name)
				}
			}
		}
	}

	return pems
}

// getEndpoints returns a list of <endpoint ip>:<port> for a given service/target port combination.
func (lbc *loadBalancerController) getEndpoints(s *api.Service, servicePort intstr.IntOrString, proto api.Protocol) []nginx.UpstreamServer {
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

			var targetPort int
			switch servicePort.Type {
			case intstr.Int:
				if epPort.Port == servicePort.IntValue() {
					targetPort = epPort.Port
				}
			case intstr.String:
				if epPort.Name == servicePort.StrVal {
					targetPort = epPort.Port
				}
			}

			if targetPort == 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ups := nginx.UpstreamServer{Address: epAddress.IP, Port: fmt.Sprintf("%v", targetPort)}
				upsServers = append(upsServers, ups)
			}
		}
	}

	glog.V(3).Infof("endpoints found: %v", upsServers)
	return upsServers
}

// Stop stops the loadbalancer controller.
func (lbc *loadBalancerController) Stop() {
	// Stop is invoked from the http endpoint.
	lbc.stopLock.Lock()
	defer lbc.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !lbc.shutdown {
		close(lbc.stopCh)
		glog.Infof("shutting down controller queues")
		lbc.shutdown = true
		lbc.syncQueue.shutdown()
	}
}

// Run starts the loadbalancer controller.
func (lbc *loadBalancerController) Run() {
	glog.Infof("starting NGINX loadbalancer controller")
	go lbc.nginx.Start()

	go lbc.ingController.Run(lbc.stopCh)
	go lbc.endpController.Run(lbc.stopCh)
	go lbc.svcController.Run(lbc.stopCh)

	go lbc.syncQueue.run(time.Second, lbc.stopCh)

	<-lbc.stopCh
	glog.Infof("shutting down NGINX loadbalancer controller")
}
