/*
Copyright 2015 The Kubernetes Authors.

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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"

	"k8s.io/ingress/core/pkg/ingress/controller"
)

func main() {
	// start a new nginx controller
	ngx := newNGINXController()
	// create a custom Ingress controller using NGINX as backend
	ic := controller.NewIngressController(ngx)

	// start the controller
	go ic.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Received SIGTERM, shutting down")

	if err := ic.Stop(); err != nil {
		glog.Errorf("unexpected error shutting down the ingress controller: %v", err)
	}

	glog.Infof("stopping nginx gracefully...")
	ngx.Stop()

	timer := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-timer.C:
			if !isNginxRunning(ngx.ports.Status) {
				glog.Infof("nginx stopped...")
				glog.Infof("Exiting with code 0")
				os.Exit(0)
			}
		}
	}
}
