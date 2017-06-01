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

package main

import (
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	hc := newHAProxyController()
	errCh := make(chan error)
	go handleSignal(hc, errCh)
	hc.Start()
	code := 0
	err := <-errCh
	if err != nil {
		glog.Warningf("Error stopping Ingress: %v", err)
		code++
	}
	glog.Infof("Exiting (%v)", code)
	os.Exit(code)
}

func handleSignal(hc *haproxyController, err chan error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	glog.Infof("Shutting down with signal %v", <-sig)
	err <- hc.Stop()
}
