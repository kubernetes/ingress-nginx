/*
Copyright 2021 The Kubernetes Authors.

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
	"context"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribeSerial("[Flag] watch namespace selector", func() {
	f := framework.NewDefaultFramework("namespace-selector")
	notMatchedHost, matchedHost := "bar", "foo"
	var notMatchedNs string
	var matchedNs string

	// create a test namespace, under which create an ingress and backend deployment
	prepareTestIngress := func(baseName string, host string, labels map[string]string) string {
		ns, err := framework.CreateKubeNamespaceWithLabel(f.BaseName, labels, f.KubeClientSet)
		assert.Nil(ginkgo.GinkgoT(), err, "creating test namespace")
		f.NewEchoDeployment(framework.WithDeploymentNamespace(ns))
		ing := framework.NewSingleIngressWithIngressClass(host, "/", host, ns, framework.EchoService, f.IngressClass, 80, nil)
		f.EnsureIngress(ing)
		return ns
	}

	cleanupNamespace := func(ns string) {
		err := framework.DeleteKubeNamespace(f.KubeClientSet, ns)
		assert.Nil(ginkgo.GinkgoT(), err, "deleting temporarily created namespace")
	}

	ginkgo.BeforeEach(func() {
		notMatchedNs = prepareTestIngress(notMatchedHost, notMatchedHost, nil) // create namespace without label "foo=bar"
		matchedNs = prepareTestIngress(matchedHost, matchedHost, map[string]string{"foo": "bar"})
	})

	ginkgo.AfterEach(func() {
		cleanupNamespace(notMatchedNs)
		cleanupNamespace(matchedNs)
	})

	ginkgo.Context("With specific watch-namespace-selector flags", func() {

		ginkgo.It("should ingore Ingress of namespace without label foo=bar and accept those of namespace with label foo=bar", func() {

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return !strings.Contains(cfg, "server_name bar") &&
					strings.Contains(cfg, "server_name foo")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", matchedHost).
				Expect().
				Status(http.StatusOK)

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", notMatchedHost).
				Expect().
				Status(http.StatusNotFound)

			// should accept Ingress when namespace labeled with foo=bar
			ns, err := f.KubeClientSet.CoreV1().Namespaces().Get(context.TODO(), notMatchedNs, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err)

			if ns.Labels == nil {
				ns.Labels = make(map[string]string)
			}
			ns.Labels["foo"] = "bar"

			_, err = f.KubeClientSet.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err, "labeling not matched namespace")

			// update ingress to trigger reconciliation
			ing, err := f.KubeClientSet.NetworkingV1().Ingresses(notMatchedNs).Get(context.TODO(), notMatchedHost, metav1.GetOptions{})
			assert.Nil(ginkgo.GinkgoT(), err, "retrieve test ingress")
			if ing.Labels == nil {
				ing.Labels = make(map[string]string)
			}
			ing.Labels["foo"] = "bar"

			_, err = f.KubeClientSet.NetworkingV1().Ingresses(notMatchedNs).Update(context.TODO(), ing, metav1.UpdateOptions{})
			assert.Nil(ginkgo.GinkgoT(), err, "updating ingress")

			f.WaitForNginxConfiguration(func(cfg string) bool {
				return strings.Contains(cfg, "server_name bar")
			})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", notMatchedHost).
				Expect().
				Status(http.StatusOK)
		})
	})
})
