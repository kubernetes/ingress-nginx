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

package settings

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/onsi/ginkgo"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("global-options", func() {
	f := framework.NewDefaultFramework("global-options")

	ginkgo.It("should have worker_rlimit_nofile option", func() {
		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, fmt.Sprintf("worker_rlimit_nofile %d;", rlimitMaxNumFiles()-1024))

		})
	})

	ginkgo.It("should have worker_rlimit_nofile option and be independent on amount of worker processes", func() {
		f.SetNginxConfigMapData(map[string]string{
			"worker-processes": "11",
		})

		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, "worker_processes 11;") &&
				strings.Contains(server, fmt.Sprintf("worker_rlimit_nofile %d;", rlimitMaxNumFiles()-1024))
		})
	})
})

// rlimitMaxNumFiles returns hard limit for RLIMIT_NOFILE
func rlimitMaxNumFiles() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0
	}
	return int(rLimit.Max)
}
