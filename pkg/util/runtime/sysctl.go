/*
Copyright 2018 The Kubernetes Authors.

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

package runtime

import (
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	"k8s.io/klog/v2"
)

// SysctlSomaxconn returns the maximum number of connections that can be queued
// for acceptance (value of net.core.somaxconn)
// http://nginx.org/en/docs/http/ngx_http_core_module.html#listen
func SysctlSomaxconn() int {
	maxConns, err := getSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		klog.V(3).InfoS("Using default net.core.somaxconn", "value", maxConns)
		return 511
	}

	return maxConns
}

// getSysctl returns the value for the specified sysctl setting
func getSysctl(sysctl string) (int, error) {
	data, err := os.ReadFile(path.Join("/proc/sys", sysctl))
	if err != nil {
		return -1, err
	}

	val, err := strconv.Atoi(strings.Trim(string(data), " \n"))
	if err != nil {
		return -1, err
	}

	return val, nil
}

// RlimitMaxNumFiles returns hard limit for RLIMIT_NOFILE
func RlimitMaxNumFiles() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		klog.ErrorS(err, "Error reading system maximum number of open file descriptors (RLIMIT_NOFILE)")
		return 0
	}
	return int(rLimit.Max)
}
