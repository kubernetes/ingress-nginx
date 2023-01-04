/*
Copyright 2022 The Kubernetes Authors.

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

package process

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	klog "k8s.io/klog/v2"
)

type exiter func(code int)

// HandleSigterm receives a ProcessController interface and deals with
// the graceful shutdown
func HandleSigterm(ngx ProcessController, delay int, exit exiter) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	klog.InfoS("Received SIGTERM, shutting down")

	exitCode := 0
	if err := ngx.Stop(); err != nil {
		klog.Warningf("Error during shutdown: %v", err)
		exitCode = 1
	}

	klog.Infof("Handled quit, delaying controller exit for %d seconds", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	klog.InfoS("Exiting", "code", exitCode)
	exit(exitCode)
}
