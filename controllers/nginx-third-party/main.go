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
	"flag"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	healthPort = 10249
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)

	defaultSvc = flags.String("default-backend-service", "",
		`Service used to serve a 404 page for the default backend. Takes the form
    namespace/name. The controller uses the first node port of this Service for
    the default backend.`)

	customErrorSvc = flags.String("custom-error-service", "",
		`Service used that will receive the errors from nginx and serve a custom error page. 
    Takes the form namespace/name. The controller uses the first node port of this Service 
    for the backend.`)

	resyncPeriod = flags.Duration("sync-period", 30*time.Second,
		`Relist and confirm cloud resources this often.`)

	watchNamespace = flags.String("watch-namespace", api.NamespaceAll,
		`Namespace to watch for Ingress. Default is to watch all namespaces`)

	healthzPort = flags.Int("healthz-port", healthPort, "port for healthz endpoint.")
)

func main() {
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	if *defaultSvc == "" {
		glog.Fatalf("Please specify --default-backend")
	}

	kubeClient, err := unversioned.NewInCluster()
	if err != nil {
		glog.Fatalf("failed to create client: %v", err)
	}

	lbInfo, _ := getLBDetails(kubeClient)
	defSvc := getService(kubeClient, *defaultSvc)
	defError := getService(kubeClient, *customErrorSvc)

	// Start loadbalancer controller
	lbc, err := NewLoadBalancerController(kubeClient, *resyncPeriod, defSvc, defError, *watchNamespace, lbInfo)
	if err != nil {
		glog.Fatalf("%v", err)
	}

	lbc.Run()

	for {
		glog.Infof("Handled quit, awaiting pod deletion.")
		time.Sleep(30 * time.Second)
	}
}

// lbInfo contains runtime information about the pod and replication controller
type lbInfo struct {
	RCNamespace  string
	RCName       string
	Podname      string
	PodIP        string
	PodNamespace string
}
