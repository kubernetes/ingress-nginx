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

package process

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	ps "github.com/mitchellh/go-ps"
	"github.com/ncabatoff/process-exporter/proc"
	"k8s.io/klog"
)

// IsRespawnIfRequired checks if error type is exec.ExitError or not
func IsRespawnIfRequired(err error) bool {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	waitStatus := exitError.Sys().(syscall.WaitStatus)
	klog.Warningf(`
-------------------------------------------------------------------------------
NGINX master process died (%v): %v
-------------------------------------------------------------------------------
`, waitStatus.ExitStatus(), err)
	return true
}

// WaitUntilPortIsAvailable waits until there is no NGINX master or worker
// process/es listening in a particular port.
func WaitUntilPortIsAvailable(port int) {
	// we wait until the workers are killed
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("0.0.0.0:%v", port), 1*time.Second)
		if err != nil {
			break
		}
		conn.Close()
		// kill nginx worker processes
		fs, err := proc.NewFS("/proc", false)
		if err != nil {
			klog.Errorf("unexpected error reading /proc information: %v", err)
			continue
		}

		procs, _ := fs.FS.AllProcs()
		for _, p := range procs {
			pn, err := p.Comm()
			if err != nil {
				klog.Errorf("unexpected error obtaining process information: %v", err)
				continue
			}

			if pn == "nginx" {
				osp, err := os.FindProcess(p.PID)
				if err != nil {
					klog.Errorf("unexpected error obtaining process information: %v", err)
					continue
				}
				osp.Signal(syscall.SIGQUIT)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// IsNginxRunning returns true if a process with the name 'nginx' is found
func IsNginxRunning() bool {
	processes, _ := ps.Processes()
	for _, p := range processes {
		if p.Executable() == "nginx" {
			return true
		}
	}
	return false
}
