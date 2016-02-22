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
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/record"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
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
	lbTcpServices = "tcpservices"

	k8sAnnotationPrefix = "nginx-ingress.kubernetes.io"
)

var (
	keyFunc = framework.DeletionHandlingMetaNamespaceKeyFunc
)

// loadBalancerController watches the kubernetes api and adds/removes services
// from the loadbalancer
type loadBalancerController struct {
	client           *client.Client
	ingController    *framework.Controller
	configController *framework.Controller
	ingLister        StoreToIngressLister
	configLister     StoreToConfigMapLister
	recorder         record.EventRecorder
	ingQueue         *taskQueue
	configQueue      *taskQueue
	stopCh           chan struct{}
	ngx              *nginx.NginxManager
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

func (a annotations) getTcpServices() (string, bool) {
	val, ok := a[fmt.Sprintf("%v/%v", k8sAnnotationPrefix, lbTcpServices)]
	return val, ok
}

// NewLoadBalancerController creates a controller for nginx loadbalancer
func NewLoadBalancerController(kubeClient *client.Client, resyncPeriod time.Duration, defaultSvc, customErrorSvc nginx.Service, namespace string, lbInfo *lbInfo) (*loadBalancerController, error) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(""))

	lbc := loadBalancerController{
		client: kubeClient,
		stopCh: make(chan struct{}),
		recorder: eventBroadcaster.NewRecorder(
			api.EventSource{Component: "nginx-lb-controller"}),
		lbInfo: lbInfo,
	}
	lbc.ingQueue = NewTaskQueue(lbc.syncIngress)
	lbc.configQueue = NewTaskQueue(lbc.syncConfig)

	lbc.ngx = nginx.NewManager(kubeClient, defaultSvc, customErrorSvc)

	// Ingress watch handlers
	pathHandlers := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addIng := obj.(*extensions.Ingress)
			lbc.recorder.Eventf(addIng, api.EventTypeNormal, "ADD", fmt.Sprintf("Adding ingress %s/%s", addIng.Namespace, addIng.Name))
			lbc.ingQueue.enqueue(obj)
		},
		DeleteFunc: lbc.ingQueue.enqueue,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				glog.V(2).Infof("Ingress %v changed, syncing", cur.(*extensions.Ingress).Name)
			}
			lbc.ingQueue.enqueue(cur)
		},
	}
	lbc.ingLister.Store, lbc.ingController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc:  ingressListFunc(lbc.client, namespace),
			WatchFunc: ingressWatchFunc(lbc.client, namespace),
		},
		&extensions.Ingress{}, resyncPeriod, pathHandlers)

	// Config watch handlers
	configHandlers := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			lbc.configQueue.enqueue(obj)
		},
		DeleteFunc: lbc.configQueue.enqueue,
		UpdateFunc: func(old, cur interface{}) {
			if !reflect.DeepEqual(old, cur) {
				glog.V(2).Infof("nginx rc changed, syncing")
				lbc.configQueue.enqueue(cur)
			}
		},
	}

	lbc.configLister.Store, lbc.configController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func(api.ListOptions) (runtime.Object, error) {
				rc, err := kubeClient.ReplicationControllers(lbInfo.RCNamespace).Get(lbInfo.RCName)
				return &api.ReplicationControllerList{
					Items: []api.ReplicationController{*rc},
				}, err
			},
			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
				options.LabelSelector = labels.SelectorFromSet(labels.Set{"name": lbInfo.RCName})
				return kubeClient.ReplicationControllers(lbInfo.RCNamespace).Watch(options)
			},
		},
		&api.ReplicationController{}, resyncPeriod, configHandlers)

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

// syncIngress manages Ingress create/updates/deletes.
func (lbc *loadBalancerController) syncIngress(key string) {
	glog.V(2).Infof("Syncing Ingress %v", key)

	obj, ingExists, err := lbc.ingLister.Store.GetByKey(key)
	if err != nil {
		lbc.ingQueue.requeue(key, err)
		return
	}

	if !ingExists {
		glog.Errorf("Ingress not found: %v", key)
		return
	}

	// this means some Ingress rule changed. There is no need to reload nginx but
	// we need to update the rules to use invoking "POST /update-ingress" with the
	// list of Ingress rules
	ingList := lbc.ingLister.Store.List()
	if err := lbc.ngx.SyncIngress(ingList); err != nil {
		lbc.ingQueue.requeue(key, err)
		return
	}

	ing := *obj.(*extensions.Ingress)
	if err := lbc.updateIngressStatus(ing); err != nil {
		lbc.recorder.Eventf(&ing, api.EventTypeWarning, "Status", err.Error())
		lbc.ingQueue.requeue(key, err)
	}
	return
}

// syncConfig manages changes in nginx configuration.
func (lbc *loadBalancerController) syncConfig(key string) {
	// we only need to sync the nginx rc
	if key != fmt.Sprintf("%v/%v", lbc.lbInfo.RCNamespace, lbc.lbInfo.RCName) {
		return
	}

	obj, configExists, err := lbc.configLister.Store.GetByKey(key)
	if err != nil {
		lbc.configQueue.requeue(key, err)
		return
	}

	if !configExists {
		glog.Errorf("Configutation not found: %v", key)
		return
	}

	glog.V(2).Infof("Syncing config %v", key)

	rc := *obj.(*api.ReplicationController)
	ngxCfgAnn, _ := annotations(rc.Annotations).getNginxConfig()
	tcpSvcAnn, _ := annotations(rc.Annotations).getTcpServices()

	ngxConfig, err := lbc.ngx.ReadConfig(ngxCfgAnn)
	if err != nil {
		glog.Warningf("%v", err)
	}

	// TODO: tcp services can change (new item in the annotation list)
	// TODO: skip get everytime
	tcpServices := getTcpServices(lbc.client, tcpSvcAnn)
	lbc.ngx.Reload(ngxConfig, tcpServices)

	return
}

// updateIngressStatus updates the IP and annotations of a loadbalancer.
// The annotations are parsed by kubectl describe.
func (lbc *loadBalancerController) updateIngressStatus(ing extensions.Ingress) error {
	ingClient := lbc.client.Extensions().Ingress(ing.Namespace)

	ip := lbc.lbInfo.PodIP
	currIng, err := ingClient.Get(ing.Name)
	if err != nil {
		return err
	}
	currIng.Status = extensions.IngressStatus{
		LoadBalancer: api.LoadBalancerStatus{
			Ingress: []api.LoadBalancerIngress{
				{IP: ip},
			},
		},
	}

	glog.Infof("Updating loadbalancer %v/%v with IP %v", ing.Namespace, ing.Name, ip)
	lbc.recorder.Eventf(currIng, api.EventTypeNormal, "CREATE", "ip: %v", ip)
	return nil
}

func (lbc *loadBalancerController) registerHandlers() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := lbc.ngx.IsHealthy(); err != nil {
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

// Stop stops the loadbalancer controller.
func (lbc *loadBalancerController) Stop() {
	// Stop is invoked from the http endpoint.
	lbc.stopLock.Lock()
	defer lbc.stopLock.Unlock()

	// Only try draining the workqueue if we haven't already.
	if !lbc.shutdown {
		close(lbc.stopCh)
		glog.Infof("Shutting down controller queues")
		lbc.ingQueue.shutdown()
		lbc.configQueue.shutdown()
		lbc.shutdown = true
	}
}

// Run starts the loadbalancer controller.
func (lbc *loadBalancerController) Run() {
	glog.Infof("Starting nginx loadbalancer controller")
	go lbc.ngx.Start()
	go lbc.registerHandlers()

	go lbc.configController.Run(lbc.stopCh)
	go lbc.configQueue.run(time.Second, lbc.stopCh)

	// Initial nginx configuration.
	lbc.syncConfig(lbc.lbInfo.RCName)

	time.Sleep(5 * time.Second)

	go lbc.ingController.Run(lbc.stopCh)
	go lbc.ingQueue.run(time.Second, lbc.stopCh)

	<-lbc.stopCh
	glog.Infof("Shutting down nginx loadbalancer controller")
}
