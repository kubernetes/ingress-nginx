//go:build linux
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
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	libcontainercgroups "github.com/opencontainers/runc/libcontainer/cgroups"
)

// NumCPU returns the number of logical CPUs usable by the current process.
// If CPU cgroups limits are configured, use cfs_quota_us / cfs_period_us
// as formula
//
//	https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
func NumCPU() int {
	return NumCPUWithCustomPath("")
}

func NumCPUWithCustomPath(path string) int {
	cpus := runtime.NumCPU()

	cgroupVersionCheckPath := path

	if cgroupVersionCheckPath == "" {
		cgroupVersionCheckPath = "/sys/fs/cgroup/"
	}

	cgroupVersion := GetCgroupVersion(cgroupVersionCheckPath)
	cpuQuota := int64(-1)
	cpuPeriod := int64(-1)

	if cgroupVersion == 1 {
		cgroupPath := ""
		if path == "" {
			cgroupPathRd, err := libcontainercgroups.FindCgroupMountpoint("", "cpu")
			if err != nil {
				return cpus
			}
			cgroupPath = cgroupPathRd
		} else {
			cgroupPath = path
		}
		cpuQuota = readCgroupFileToInt64(cgroupPath, "cpu.cfs_quota_us")
		cpuPeriod = readCgroupFileToInt64(cgroupPath, "cpu.cfs_period_us")
	} else if cgroupVersion == 2 {
		cgroupPath := "/sys/fs/cgroup/"
		if path != "" {
			cgroupPath = path
		}
		cpuQuota, cpuPeriod = readCgroup2FileToInt64Tuple(cgroupPath, "cpu.max")
	}

	if cpuQuota == -1 || cpuPeriod == -1 {
		return cpus
	}

	return int(math.Ceil(float64(cpuQuota) / float64(cpuPeriod)))
}

func GetCgroupVersion(cgroupPath string) int64 {
	// /sys/fs/cgroup/cgroup.controllers will not exist with cgroupsv1
	if _, err := os.Stat(filepath.Join(cgroupPath, "cgroup.controllers")); err == nil {
		return 2
	}

	return 1
}

func readCgroup2StringToInt64Tuple(cgroupString string) (quota, period int64) {
	// file contents looks like: $MAX $PERIOD
	// $MAX can have value "max" indicating no limit
	// it is possible for $PERIOD to be unset

	values := strings.Fields(cgroupString)

	if values[0] == "max" {
		return -1, -1
	}

	cpuQuota, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return -1, -1
	}

	if len(values) == 1 {
		return cpuQuota, 100000
	}

	cpuPeriod, err := strconv.ParseInt(values[1], 10, 64)
	if err != nil {
		return -1, -1
	}

	return cpuQuota, cpuPeriod
}

func readCgroup2FileToInt64Tuple(cgroupPath, cgroupFile string) (quota, period int64) {
	contents, err := os.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return -1, -1
	}

	return readCgroup2StringToInt64Tuple(string(contents))
}

func readCgroupStringToInt64(contents string) int64 {
	strValue := strings.TrimSpace(contents)
	if value, err := strconv.ParseInt(strValue, 10, 64); err == nil {
		return value
	}

	return -1
}

func readCgroupFileToInt64(cgroupPath, cgroupFile string) int64 {
	contents, err := os.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return -1
	}

	return readCgroupStringToInt64(string(contents))
}
