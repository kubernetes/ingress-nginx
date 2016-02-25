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

package nginx

import (
	"os"
	"os/exec"

	"github.com/golang/glog"
)

const (
	nginxEvent = "NGINX"
)

// Start starts a nginx (master process) and waits. If the process ends
// we need to kill the controller process and return the reason.
func (ngx *NginxManager) Start() {
	glog.Info("Starting nginx...")
	cmd := exec.Command("nginx")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		glog.Errorf("nginx error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		glog.Errorf("nginx error: %v", err)
	}
}

// Reload the master process receives the signal to reload configuration, it checks
// the syntax validity of the new configuration file and tries to apply the
// configuration provided in it. If this is a success, the master process starts
// new worker processes and sends messages to old worker processes, requesting them
// to shut down. Otherwise, the master process rolls back the changes and continues
// to work with the old configuration. Old worker processes, receiving a command to
// shut down, stop accepting new connections and continue to service current requests
// until all such requests are serviced. After that, the old worker processes exit.
// http://nginx.org/en/docs/beginners_guide.html#control
func (ngx *NginxManager) Reload(cfg *nginxConfiguration, servicesL4 []Service) {
	ngx.reloadLock.Lock()
	defer ngx.reloadLock.Unlock()

	if err := ngx.writeCfg(cfg, servicesL4); err != nil {
		glog.Errorf("Failed to write new nginx configuration. Avoiding reload: %v", err)
		return
	}

	if err := ngx.shellOut("nginx -s reload"); err == nil {
		glog.Info("Change in configuration detected. Reloading...")
	}
}

// shellOut executes a command and returns its combined standard output and standard
// error in case of an error in the execution
func (ngx *NginxManager) shellOut(cmd string) error {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		glog.Errorf("Failed to execute %v: %v", cmd, string(out))
		return err
	}

	return nil
}
