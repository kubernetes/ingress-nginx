//go:build !linux
// +build !linux

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

	ginkgo.It("detects cgroups is not avaliable", func() {
		assert.True(ginkgo.GinkgoT(), !runtime.IsCgroupAvaliable())
	})
})
