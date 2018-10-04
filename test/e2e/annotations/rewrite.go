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

package annotations

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/parnurzeal/gorequest"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - Rewrite", func() {
	f := framework.NewDefaultFramework("rewrite")

	BeforeEach(func() {
		err := f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should rewrite request URL", func() {
		By("setting rewrite-target annotation")

		host := "rewrite.foo.com"
		annotations := map[string]string{"nginx.ingress.kubernetes.io/rewrite-target": "/"}
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:8080/", host)

		ing := framework.NewSingleIngress(host, "/something", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `rewrite "(?i)/something/(.*)" /$1 break;`) &&
					strings.Contains(server, `rewrite "(?i)/something$" / break;`)
			})
		Expect(err).NotTo(HaveOccurred())

		By("sending request to Ingress rule path (lowercase)")

		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))

		By("sending request to Ingress rule path (mixed case)")

		resp, body, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/SomeThing").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))
	})

	It("should write rewrite logs", func() {
		By("setting enable-rewrite-log annotation")

		host := "rewrite.bar.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target":     "/",
			"nginx.ingress.kubernetes.io/enable-rewrite-log": "true",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err := f.EnsureIngress(ing)

		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "rewrite_log on;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/something").
			Set("Host", host).
			End()

		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		logs, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).To(ContainSubstring(`"(?i)/something$" matches "/something", client:`))
		Expect(logs).To(ContainSubstring(`rewritten data: "/", args: "",`))
	})

	It("should use correct longest path match", func() {
		host := "rewrite.bar.com"

		By("creating a regular ingress definition")
		ing := framework.NewSingleIngress("kube-lego", "/.well-known/acme/challenge", host, f.IngressController.Namespace, "http-svc", 80, &map[string]string{})
		_, err := f.EnsureIngress(ing)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "/.well-known/acme/challenge")
			})
		Expect(err).NotTo(HaveOccurred())

		By("making a request to the non-rewritten location")
		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/.well-known/acme/challenge").
			Set("Host", host).
			End()
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:8080/.well-known/acme/challenge", host)
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))

		By(`creating an ingress definition with the rewrite-target annotation set on the "/" location`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		rewriteIng := framework.NewSingleIngress("rewrite-index", "/", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err = f.EnsureIngress(rewriteIng)
		Expect(err).NotTo(HaveOccurred())
		Expect(rewriteIng).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "location ~* ^/ {") && strings.Contains(server, `location ~* "^/.well-known/acme/challenge" {`)
			})
		Expect(err).NotTo(HaveOccurred())

		By("making a second request to the non-rewritten location")
		resp, body, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/.well-known/acme/challenge").
			Set("Host", host).
			End()
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))
	})

	It("should use ~* location modifier if regex annotation is present", func() {
		host := "rewrite.bar.com"

		By("creating a regular ingress definition")
		ing := framework.NewSingleIngress("foo", "/foo", host, f.IngressController.Namespace, "http-svc", 80, &map[string]string{})
		_, err := f.EnsureIngress(ing)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "location /foo {")
			})
		Expect(err).NotTo(HaveOccurred())

		By(`creating an ingress definition with the use-regex amd rewrite-target annotation`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		ing = framework.NewSingleIngress("regex", "/foo.+", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err = f.EnsureIngress(ing)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/foo" {`) && strings.Contains(server, `location ~* "^/foo.+\/?(?<baseuri>.*)" {`)
			})
		Expect(err).NotTo(HaveOccurred())

		By("ensuring '/foo' matches '~* ^/foo'")
		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo").
			Set("Host", host).
			End()
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:8080/foo", host)
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))

		By("ensuring '/foo/bar' matches '~* ^/foo.+'")
		resp, body, errs = gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo/bar").
			Set("Host", host).
			End()
		expectBodyRequestURI = fmt.Sprintf("request_uri=http://%v:8080/new/backend", host)
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))
	})

	It("should fail to use longest match for documented warning", func() {
		host := "rewrite.bar.com"

		By("creating a regular ingress definition")
		ing := framework.NewSingleIngress("foo", "/foo/bar/bar", host, f.IngressController.Namespace, "http-svc", 80, &map[string]string{})
		_, err := f.EnsureIngress(ing)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		By(`creating an ingress definition with the use-regex annotation`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		ing = framework.NewSingleIngress("regex", "/foo/bar/[a-z]{3}", host, f.IngressController.Namespace, "http-svc", 80, &annotations)
		_, err = f.EnsureIngress(ing)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/foo/bar/bar" {`) && strings.Contains(server, `location ~* "^/foo/bar/[a-z]{3}\/?(?<baseuri>.*)" {`)
			})
		Expect(err).NotTo(HaveOccurred())

		By("check that '/foo/bar/bar' does not match the longest exact path")
		resp, body, errs := gorequest.New().
			Get(f.IngressController.HTTPURL+"/foo/bar/bar").
			Set("Host", host).
			End()
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:8080/new/backend", host)
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		Expect(body).Should(ContainSubstring(expectBodyRequestURI))
	})
})
