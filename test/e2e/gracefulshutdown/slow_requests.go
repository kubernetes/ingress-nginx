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

package gracefulshutdown

import (
	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("[Shutdown] Graceful shutdown with pending request", func() {
	f := framework.NewDefaultFramework("shutdown-slow-requests")

	ginkgo.BeforeEach(func() {
		f.NewSlowEchoDeployment()
		f.UpdateNginxConfigMapData("worker-shutdown-timeout", "50s")
	})
	/* @rikatz  - This seems to be failing on GH Actions and needs to be re-checked and re-verified
	ginkgo.It("should let slow requests finish before shutting down", func() {
		host := "graceful-shutdown"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, nil))
		f.WaitForNginxConfiguration(
			func(conf string) bool {
				return strings.Contains(conf, "worker_shutdown_timeout")
			})

		done := make(chan bool)
		go func() {
			defer func() { done <- true }()
			defer ginkgo.GinkgoRecover()

			f.HTTPTestClient().
				GET("/sleep/30").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)
		}()

		framework.Sleep()
		f.DeleteNGINXPod(60)
		<-done
	}) */
})
