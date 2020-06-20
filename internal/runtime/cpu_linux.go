// +build linux

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
	"io/ioutil"
	"math"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	libcontainercgroups "github.com/opencontainers/runc/libcontainer/cgroups"
)

// NumCPU returns the number of logical CPUs usable by the current process.
// If CPU cgroups limits are configured, use cfs_quota_us / cfs_period_us
// as formula
//  https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
func NumCPU() int {
	cpus := runtime.NumCPU()

	cgroupPath, err := libcontainercgroups.FindCgroupMountpoint("", "cpu")
	if err != nil {
		return cpus
	}

	cpuQuota := readCgroupFileToInt64(cgroupPath, "cpu.cfs_quota_us")
	cpuPeriod := readCgroupFileToInt64(cgroupPath, "cpu.cfs_period_us")

	if cpuQuota == -1 || cpuPeriod == -1 {
		return cpus
	}

	return int(math.Ceil(float64(cpuQuota) / float64(cpuPeriod)))
}

func readCgroupFileToInt64(cgroupPath, cgroupFile string) int64 {
	contents, err := ioutil.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return -1
	}

	strValue := strings.TrimSpace(string(contents))
	if value, err := strconv.ParseInt(strValue, 10, 64); err == nil {
		return value
	}

	return -1
}
