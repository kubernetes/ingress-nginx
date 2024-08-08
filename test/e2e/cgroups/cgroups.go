/*
Copyright 2020 The Kubernetes Authors.

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

package cgroups

import (
	"log"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"

	"k8s.io/ingress-nginx/pkg/util/runtime"
)

var _ = framework.IngressNginxDescribeSerial("[CGroups] cgroups", func() {
	f := framework.NewDefaultFramework("cgroups")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("detects cgroups version v1", func() {
		cgroupPath := "/testing/sys/fs/cgroup/"
		if err := os.MkdirAll(cgroupPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		quotaFile, err := os.Create(filepath.Join(cgroupPath, "cpu.cfs_quota_us"))
		if err != nil {
			log.Fatal(err)
		}

		periodFile, err := os.Create(filepath.Join(cgroupPath, "cpu.cfs_period_us"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = quotaFile.WriteString("4")
		if err != nil {
			log.Fatal(err)
		}

		err = quotaFile.Sync()
		if err != nil {
			log.Fatal(err)
		}

		_, err = periodFile.WriteString("2")
		if err != nil {
			log.Fatal(err)
		}

		err = periodFile.Sync()
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(ginkgo.GinkgoT(), runtime.GetCgroupVersion(cgroupPath), int64(1))
		assert.Equal(ginkgo.GinkgoT(), runtime.NumCPUWithCustomPath(cgroupPath), 2)

		os.Remove(filepath.Join(cgroupPath, "cpu.cfs_quota_us"))
		os.Remove(filepath.Join(cgroupPath, "cpu.cfs_period_us"))
	})

	ginkgo.It("detect cgroups version v2", func() {
		cgroupPath := "/testing/sys/fs/cgroup/"
		if err := os.MkdirAll(cgroupPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		_, err := os.Create(filepath.Join(cgroupPath, "cgroup.controllers"))
		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Create(filepath.Join(cgroupPath, "cpu.max"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = file.WriteString("4 2")
		if err != nil {
			log.Fatal(err)
		}

		err = file.Sync()
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(ginkgo.GinkgoT(), runtime.GetCgroupVersion(cgroupPath), int64(2))
		assert.Equal(ginkgo.GinkgoT(), runtime.NumCPUWithCustomPath(cgroupPath), 2)

		os.Remove(filepath.Join(cgroupPath, "cpu.max"))
		os.Remove(filepath.Join(cgroupPath, "cgroup.controllers"))
	})
})
