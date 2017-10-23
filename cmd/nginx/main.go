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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/ingress-nginx/pkg/nginx/controller"

	"github.com/golang/glog"
)

func main() {
	// start a new nginx controller
	ngx := controller.NewNGINXController()

	go handleSigterm(ngx)
	// start the controller
	ngx.Start()
	// wait
	glog.Infof("shutting down Ingress controller...")
	for {
		glog.Infof("Handled quit, awaiting pod deletion")
		time.Sleep(30 * time.Second)
	}
}

func handleSigterm(ngx *controller.NGINXController) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	running := false
	srv := &http.Server{Addr: ":61234"}
	for value := range signalChan {
		glog.V(4).Infof("Received %s, shutting down\n", value.String())
		switch value {
		case syscall.SIGUSR1:
			go func() {
				if running {
					glog.V(1).Info("SIGUSR1 hited, http://:61234 pprof service running")
					return
				}
				glog.Info("SIGUSR1 hited, to start http//:61234 pprof service")
				running = true
				if err := srv.ListenAndServe(); err != nil {
					glog.Errorf("start pprof service error: %v", err)
				}
				running = false
			}()
		case syscall.SIGUSR2:
			glog.Info("SIGUSR2 hited, to stop http pprof service")
			if running {
				//go 1.8 support srv.Shutdown
				err := srv.Shutdown(nil)
				if err != nil {
					glog.Errorf("pprof stop httpserver error: %s", err)
				}
			}
		case syscall.SIGSTOP:
			exitCode := 0
			if running {
				//go 1.8 support srv.Shutdown
				err := srv.Shutdown(nil)
				if err != nil {
					glog.Errorf("pprof stop httpserver error: %s", err)
				}
			}

			if err := ngx.Stop(); err != nil {
				glog.Infof("Error during shutdown %v", err)
				exitCode = 1
			}

			glog.Infof("Exiting with %v", exitCode)
			os.Exit(exitCode)
		}
	}
}
