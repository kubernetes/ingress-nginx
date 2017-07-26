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
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"k8s.io/kubernetes/pkg/util/sysctl"

	"github.com/golang/glog"
)

// sysctlSomaxconn returns the value of net.core.somaxconn, i.e.
// maximum number of connections that can be queued for acceptance
// http://nginx.org/en/docs/http/ngx_http_core_module.html#listen
func sysctlSomaxconn() int {
	maxConns, err := sysctl.New().GetSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		glog.V(3).Infof("system net.core.somaxconn=%v (using system default)", maxConns)
		return 511
	}

	return maxConns
}

// sysctlFSFileMax returns the value of fs.file-max, i.e.
// maximum number of open file descriptors
func sysctlFSFileMax() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		glog.Errorf("unexpected error reading system maximum number of open file descriptors (RLIMIT_NOFILE): %v", err)
		// returning 0 means don't render the value
		return 0
	}
	return int(rLimit.Max)
}

func diff(b1, b2 []byte) ([]byte, error) {
	f1, err := ioutil.TempFile("", "a")
	if err != nil {
		return nil, err
	}
	defer f1.Close()
	defer os.Remove(f1.Name())

	f2, err := ioutil.TempFile("", "b")
	if err != nil {
		return nil, err
	}
	defer f2.Close()
	defer os.Remove(f2.Name())

	f1.Write(b1)
	f2.Write(b2)

	out, _ := exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	return out, nil
}
