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
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"

	"k8s.io/contrib/ingress/controllers/nginx-third-party/nginx"
)

const (
	// Name of the default config map that contains the configuration for nginx.
	// Takes the form namespace/name.
	// If the annotation does not exists the controller will create a new annotation with the default
	// configuration.
	lbConfigName = "lbconfig"

	// If you have pure tcp services or https services that need L3 routing, you
	// must specify them by name. Note that you are responsible for:
	// 1. Making sure there is no collision between the service ports of these services.
	//  - You can have multiple <mysql svc name>:3306 specifications in this map, and as
	//    long as the service ports of your mysql service don't clash, you'll get
	//    loadbalancing for each one.
	// 2. Exposing the service ports as node ports on a pod.
	// 3. Adding firewall rules so these ports can ingress traffic.

	// Comma separated list of tcp/https
	// namespace/serviceName:portToExport pairings. This assumes you've opened up the right
	// hostPorts for each service that serves ingress traffic. Te value of portToExport indicates the
	// port to listen inside nginx, not the port of the service.
	lbTCPServices = "tcpservices"

	k8sAnnotationPrefix = "nginx-ingress.kubernetes.io"
)

// loadBalancerController watches the kubernetes api and adds/removes services
// from the loadbalancer
type loadBalancerController struct {
	client           *client.Client
	ingController    *framework.Controller
	configController *framework.Controller
	endpController   *framework.Controller
	svcController    *framework.Controller
	ingLister        StoreToIngressLister
	svcLister        cache.StoreToServiceLister
	configLister     StoreToConfigMapLister
	endpLister       cache.StoreToEndpointsLister
	stopCh           chan struct{}
	nginx            *nginx.NginxManager
	lbInfo           *lbInfo
	// stopLock is used to enforce only a single call to Stop is active.
	// Needed because we allow stopping through an http endpoint and
	// allowing concurrent stoppers leads to stack traces.
	stopLock sync.Mutex
	shutdown bool
}

type annotations map[string]string

func (a annotations) getNginxConfig() (string, bool) {
	val, ok := a[fmt.Sprintf("%v/%v", k8sAnnotationPrefix, lbConfigName)]
	return val, ok
}

func (a annotations) getTCPServices() (string, bool) {
	val, ok := a[fmt.Sprintf("%v/%v", k8sAnnotationPrefix, lbTCPServices)]
	return val, ok
}

// newLoadBalancerController creates a controller for nginx loadbalancer
func newLoadBalancerController(kubeClient *client.Client, resyncPeriod time.Duration, defaultSvc, customErrorSvc nginx.Service, namespace string, lbInfo *lbInfo) (*loadBalancerController, error) {
	lbc := loadBalancerController{
		client: kubeClient,
		stopCh: make(chan struct{}),
		lbInfo: lbInfo,
		nginx:  nginx.NewManager(kubeClient, defaultSvc, customErrorSvc),
	}

	lbc.ingLister.Store, lbc.ingController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  ingressListFunc(lbc.client, namespace),
			WatchFunc: ingressWatchFunc(lbc.client, namespace),
		},
		&extensions.Ingress{}, resyncPeriod, framework.ResourceEventHandlerFuncs{})

	lbc.configLister.Store, lbc.configController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  configListFunc(kubeClient, lbc.lbInfo.DeployType, namespace, lbInfo.ObjectName),
			WatchFunc: configWatchFunc(kubeClient, lbc.lbInfo.DeployType, namespace, lbInfo.ObjectName),
		},
		&api.ReplicationController{}, resyncPeriod, framework.ResourceEventHandlerFuncs{})

	lbc.endpLister.Store, lbc.endpController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  endpointsListFunc(kubeClient, namespace),
			WatchFunc: endpointsWatchFunc(kubeClient, namespace),
		},
		&api.Endpoints{}, resyncPeriod, framework.ResourceEventHandlerFuncs{})

	lbc.svcLister.Store, lbc.svcController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  serviceListFunc(kubeClient, namespace),
			WatchFunc: serviceWatchFunc(kubeClient, namespace),
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

func configListFunc(c *client.Client, deployType runtime.Object, ns, name string) func(api.ListOptions) (runtime.Object, error) {
	return func(api.ListOptions) (runtime.Object, error) {
		switch deployType.(type) {
		case *api.ReplicationController:
			rc, err := c.ReplicationControllers(ns).Get(name)
			return &api.ReplicationControllerList{
				Items: []api.ReplicationController{*rc},
			}, err
		case *extensions.DaemonSet:
			ds, err := c.Extensions().DaemonSets(ns).Get(name)
			return &extensions.DaemonSetList{
				Items: []extensions.DaemonSet{*ds},
			}, err
		default:
			return nil, errInvalidKind
		}
	}
}

func configWatchFunc(c *client.Client, deployType runtime.Object, ns, name string) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		switch deployType.(type) {
		case *api.ReplicationController:
			options.LabelSelector = labels.SelectorFromSet(labels.Set{"name": name})
			return c.ReplicationControllers(ns).Watch(options)
		case *extensions.DaemonSet:
			options.LabelSelector = labels.SelectorFromSet(labels.Set{"name": name})
			return c.Extensions().DaemonSets(ns).Watch(options)
		default:
			return nil, errInvalidKind
		}
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

func (lbc *loadBalancerController) registerHandlers() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := lbc.nginx.IsHealthy(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("nginx error"))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		lbc.Stop()
	})

	glog.Fatalf(fmt.Sprintf("%v", http.ListenAndServe(fmt.Sprintf(":%v", *healthzPort), nil)))
}

func (lbc *loadBalancerController) sync() {
	ings := lbc.ingLister.Store.List()
	upstreams, servers := lbc.getUpstreamServers(ings)

	var kindAnnotations map[string]string
	ngxCfgAnn, _ := annotations(kindAnnotations).getNginxConfig()
	tcpSvcAnn, _ := annotations(kindAnnotations).getTCPServices()
	ngxConfig, err := lbc.nginx.ReadConfig(ngxCfgAnn)
	if err != nil {
		glog.Warningf("%v", err)
	}

	tcpServices := getTCPServices(lbc.client, tcpSvcAnn)
	lbc.nginx.CheckAndReload(ngxConfig, upstreams, servers, tcpServices)
}

func (lbc *loadBalancerController) getUpstreamServers(data []interface{}) ([]nginx.Upstream, []nginx.Server) {
	pems := make(map[string]string)

	upstreams := make(map[string]nginx.Upstream)
	servers := make(map[string]nginx.Server)

	for _, ingIf := range data {
		ing := ingIf.(*extensions.Ingress)

		for _, rule := range ing.Spec.Rules {
			if rule.IngressRuleValue.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				name := ing.GetNamespace() + "-" + path.Backend.ServiceName

				var ups nginx.Upstream

				if existent, ok := upstreams[name]; ok {
					ups = existent
				} else {
					ups = nginx.NewUpstream(name)
				}

				svcKey := ing.GetNamespace() + "/" + path.Backend.ServiceName
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
						endps := lbc.getEndpoints(svc, servicePort.TargetPort)
						if len(endps) == 0 {
							glog.Warningf("service %v does no have any active endpoints", svcKey)
						}

						ups.Backends = append(ups.Backends, endps...)
						break
					}
				}

				upstreams[name] = ups
			}
		}

		for _, rule := range ing.Spec.Rules {
			var server nginx.Server
			if existent, ok := servers[rule.Host]; ok {
				server = existent
			} else {
				server = nginx.Server{Name: rule.Host}
			}

			if pemFile, ok := pems[rule.Host]; ok {
				server.SSL = true
				server.SSLCertificate = pemFile
				server.SSLCertificateKey = pemFile
			}

			var locations []nginx.Location

			for _, path := range rule.HTTP.Paths {
				loc := nginx.Location{Path: path.Path}
				upsName := ing.GetNamespace() + "-" + path.Backend.ServiceName

				for _, ups := range upstreams {
					if upsName == ups.Name {
						loc.Upstream = ups
					}
				}
				locations = append(locations, loc)
			}

			server.Locations = append(server.Locations, locations...)
			servers[rule.Host] = server
		}
	}

	aUpstreams := make([]nginx.Upstream, 0, len(upstreams))
	for _, value := range upstreams {
		if len(value.Backends) == 0 {
			value.Backends = append(value.Backends, nginx.NewDefaultServer())
		}
		sort.Sort(nginx.UpstreamServerByAddrPort(value.Backends))
		aUpstreams = append(aUpstreams, value)
	}
	sort.Sort(nginx.UpstreamByNameServers(aUpstreams))

	aServers := make([]nginx.Server, 0, len(servers))
	for _, value := range servers {
		sort.Sort(nginx.LocationByPath(value.Locations))
		aServers = append(aServers, value)
	}
	sort.Sort(nginx.ServerByName(aServers))

	return aUpstreams, aServers
}

// getEndpoints returns a list of <endpoint ip>:<port> for a given service/target port combination.
func (lbc *loadBalancerController) getEndpoints(s *api.Service, servicePort intstr.IntOrString) []nginx.UpstreamServer {
	ep, err := lbc.endpLister.GetServiceEndpoints(s)
	if err != nil {
		glog.Warningf("unexpected error obtaining service endpoints: %v", err)
		return []nginx.UpstreamServer{}
	}

	var upsServers []nginx.UpstreamServer

	for _, ss := range ep.Subsets {
		for _, epPort := range ss.Ports {
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
	}
}

// Run starts the loadbalancer controller.
func (lbc *loadBalancerController) Run() {
	glog.Infof("starting NGINX loadbalancer controller")
	go lbc.nginx.Start()
	go lbc.registerHandlers()

	go lbc.configController.Run(lbc.stopCh)
	go lbc.ingController.Run(lbc.stopCh)
	go lbc.endpController.Run(lbc.stopCh)
	go lbc.svcController.Run(lbc.stopCh)

	// periodic check for changes in configuration
	go wait.Until(lbc.sync, 5*time.Second, wait.NeverStop)

	time.Sleep(5 * time.Second)

	<-lbc.stopCh
	glog.Infof("shutting down NGINX loadbalancer controller")
}
