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
	go_flag "flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"k8s.io/contrib/ingress/controllers/gce/controller"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	kubectl_util "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/golang/glog"
)

// Entrypoint of GLBC. Example invocation:
// 1. In a pod:
// glbc --delete-all-on-quit
// 2. Dry run (on localhost):
// $ kubectl proxy --api-prefix="/"
// $ glbc --proxy="http://localhost:proxyport"

const (
	// lbApiPort is the port on which the loadbalancer controller serves a
	// minimal api (/healthz, /delete-all-and-quit etc).
	lbApiPort = 8081

	// A delimiter used for clarity in naming GCE resources.
	clusterNameDelimiter = "--"

	// Arbitrarily chosen alphanumeric character to use in constructing resource
	// names, eg: to avoid cases where we end up with a name ending in '-'.
	alphaNumericChar = "0"

	// Current docker image version. Only used in debug logging.
	imageVersion = "glbc:0.6.2"
)

var (
	flags = flag.NewFlagSet(
		`gclb: gclb --runngin-in-cluster=false --default-backend-node-port=123`,
		flag.ExitOnError)

	clusterName = flags.String("cluster-uid", controller.DefaultClusterUID,
		`Optional, used to tag cluster wide, shared loadbalancer resources such
		 as instance groups. Use this flag if you'd like to continue using the
		 same resources across a pod restart. Note that this does not need to
		 match the name of you Kubernetes cluster, it's just an arbitrary name
		 used to tag/lookup cloud resources.`)

	inCluster = flags.Bool("running-in-cluster", true,
		`Optional, if this controller is running in a kubernetes cluster, use the
		 pod secrets for creating a Kubernetes client.`)

	// TODO: Consolidate this flag and running-in-cluster. People already use
	// the first one to mean "running in dev", unfortunately.
	useRealCloud = flags.Bool("use-real-cloud", false,
		`Optional, if set a real cloud client is created. Only matters with
		 --running-in-cluster=false, i.e a real cloud is always used when this
		 controller is running on a Kubernetes node.`)

	resyncPeriod = flags.Duration("sync-period", 30*time.Second,
		`Relist and confirm cloud resources this often.`)

	deleteAllOnQuit = flags.Bool("delete-all-on-quit", false,
		`If true, the controller will delete all Ingress and the associated
		external cloud resources as it's shutting down. Mostly used for
		testing. In normal environments the controller should only delete
		a loadbalancer if the associated Ingress is deleted.`)

	defaultSvc = flags.String("default-backend-service", "kube-system/default-http-backend",
		`Service used to serve a 404 page for the default backend. Takes the form
		namespace/name. The controller uses the first node port of this Service for
		the default backend.`)

	healthCheckPath = flags.String("health-check-path", "/",
		`Path used to health-check a backend service. All Services must serve
		a 200 page on this path. Currently this is only configurable globally.`)

	watchNamespace = flags.String("watch-namespace", api.NamespaceAll,
		`Namespace to watch for Ingress/Services/Endpoints.`)

	verbose = flags.Bool("verbose", false,
		`If true, logs are displayed at V(4), otherwise V(2).`)
)

func registerHandlers(lbc *controller.LoadBalancerController) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := lbc.CloudClusterManager.IsHealthy(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Cluster unhealthy: %v", err)))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/delete-all-and-quit", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Retry failures during shutdown.
		lbc.Stop(true)
	})

	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", lbApiPort), nil))
}

func handleSigterm(lbc *controller.LoadBalancerController, deleteAll bool) {
	// Multiple SIGTERMs will get dropped
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	glog.Infof("Received SIGTERM, shutting down")

	// TODO: Better retires than relying on restartPolicy.
	exitCode := 0
	if err := lbc.Stop(deleteAll); err != nil {
		glog.Infof("Error during shutdown %v", err)
		exitCode = 1
	}
	glog.Infof("Exiting with %v", exitCode)
	os.Exit(exitCode)
}

// main function for GLBC.
func main() {
	// TODO: Add a healthz endpoint
	var kubeClient *client.Client
	var err error
	var clusterManager *controller.ClusterManager
	flags.Parse(os.Args)
	clientConfig := kubectl_util.DefaultClientConfig(flags)

	// Set glog verbosity levels
	if *verbose {
		go_flag.Lookup("logtostderr").Value.Set("true")
		go_flag.Set("v", "4")
	}
	glog.Infof("Starting GLBC image: %v", imageVersion)
	if *defaultSvc == "" {
		glog.Fatalf("Please specify --default-backend")
	}

	// Create kubeclient
	if *inCluster {
		if kubeClient, err = client.NewInCluster(); err != nil {
			glog.Fatalf("Failed to create client: %v.", err)
		}
	} else {
		config, err := clientConfig.ClientConfig()
		if err != nil {
			glog.Fatalf("error connecting to the client: %v", err)
		}
		kubeClient, err = client.New(config)
	}
	// Wait for the default backend Service. There's no pretty way to do this.
	parts := strings.Split(*defaultSvc, "/")
	if len(parts) != 2 {
		glog.Fatalf("Default backend should take the form namespace/name: %v",
			*defaultSvc)
	}
	defaultBackendNodePort, err := getNodePort(kubeClient, parts[0], parts[1])
	if err != nil {
		glog.Fatalf("Could not configure default backend %v: %v",
			*defaultSvc, err)
	}

	if *inCluster || *useRealCloud {
		// Create cluster manager
		clusterManager, err = controller.NewClusterManager(
			*clusterName, defaultBackendNodePort, *healthCheckPath)
		if err != nil {
			glog.Fatalf("%v", err)
		}
	} else {
		// Create fake cluster manager
		clusterManager = controller.NewFakeClusterManager(*clusterName).ClusterManager
	}

	// Start loadbalancer controller
	lbc, err := controller.NewLoadBalancerController(kubeClient, clusterManager, *resyncPeriod, *watchNamespace)
	if err != nil {
		glog.Fatalf("%v", err)
	}
	if clusterManager.ClusterNamer.ClusterName != "" {
		glog.V(3).Infof("Cluster name %+v", clusterManager.ClusterNamer.ClusterName)
	}
	go registerHandlers(lbc)
	go handleSigterm(lbc, *deleteAllOnQuit)

	lbc.Run()
	for {
		glog.Infof("Handled quit, awaiting pod deletion.")
		time.Sleep(30 * time.Second)
	}
}

// getNodePort waits for the Service, and returns it's first node port.
func getNodePort(client *client.Client, ns, name string) (nodePort int64, err error) {
	var svc *api.Service
	glog.V(3).Infof("Waiting for %v/%v", ns, name)
	wait.Poll(1*time.Second, 5*time.Minute, func() (bool, error) {
		svc, err = client.Services(ns).Get(name)
		if err != nil {
			return false, nil
		}
		for _, p := range svc.Spec.Ports {
			if p.NodePort != 0 {
				nodePort = int64(p.NodePort)
				glog.V(3).Infof("Node port %v", nodePort)
				break
			}
		}
		return true, nil
	})
	return
}
