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

package lua

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	logDynamicConfigSuccess = "Dynamic reconfiguration succeeded"
	logDynamicConfigFailure = "Dynamic reconfiguration failed"
	logRequireBackendReload = "Configuration changes detected, backend reload required"
	logBackendReloadSuccess = "Backend successfully reloaded"
	logSkipBackendReload    = "Changes handled by the dynamic configuration, skipping backend reload"
	logInitialConfigSync    = "Initial synchronization of the NGINX configuration"
)

var _ = framework.IngressNginxDescribe("Dynamic Configuration", func() {
	f := framework.NewDefaultFramework("dynamic-configuration")

	BeforeEach(func() {
		err := enableDynamicConfiguration(f.IngressController.Namespace, f.KubeClientSet)
		Expect(err).NotTo(HaveOccurred())

		err = f.NewEchoDeploymentWithReplicas(1)
		Expect(err).NotTo(HaveOccurred())

		host := "foo.com"
		ing, err := ensureIngress(f, host)
		Expect(err).NotTo(HaveOccurred())
		Expect(ing).NotTo(BeNil())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())

		// give some time for Lua to sync the backend
		time.Sleep(5 * time.Second)

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(ContainSubstring(logDynamicConfigFailure))
		Expect(log).To(ContainSubstring(logDynamicConfigSuccess))
	})

	Context("when only backends change", func() {
		It("should handle endpoints only changes", func() {
			resp, _, errs := gorequest.New().
				Get(fmt.Sprintf("%s?id=endpoints_only_changes", f.IngressController.HTTPURL)).
				Set("Host", "foo.com").
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			replicas := 2
			err := framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", replicas, nil)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(5 * time.Second)

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=endpoints_only_changes")
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring(logDynamicConfigSuccess))
			Expect(restOfLogs).ToNot(ContainSubstring(logDynamicConfigFailure))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring(logRequireBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logBackendReloadSuccess))
			Expect(restOfLogs).To(ContainSubstring(logSkipBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logInitialConfigSync))
		})

		It("should be able to update endpoints even when the update POST size(request body) > size(client_body_buffer_size)", func() {
			// Update client-body-buffer-size to 1 byte
			err := f.UpdateNginxConfigMapData("client-body-buffer-size", "1")
			Expect(err).NotTo(HaveOccurred())

			replicas := 0
			err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", replicas, nil)
			Expect(err).NotTo(HaveOccurred())

			replicas = 4
			err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", replicas, nil)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(5 * time.Second)

			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", "foo.com").
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "POST /configuration/backends HTTP/1.1")
			restOfLogs := log[index:]

			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())

			By("POSTing new backends to Lua endpoint")
			// NOTE(elvinefendi) now that we disabled access log for this endpoint we have to find a different way to assert this
			// or maybe delete this test completely and just rely on unit testing of Lua middleware?
			//Expect(restOfLogs).To(ContainSubstring("a client request body is buffered to a temporary file"))
			Expect(restOfLogs).ToNot(ContainSubstring("dynamic-configuration: unable to read valid request body"))
		})

		It("should handle annotation changes", func() {
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			resp, _, errs := gorequest.New().
				Get(fmt.Sprintf("%s?id=should_handle_annotation_changes", f.IngressController.HTTPURL)).
				Set("Host", "foo.com").
				End()
			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/load-balance"] = "round_robin"
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(5 * time.Second)

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=should_handle_annotation_changes")
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring(logDynamicConfigSuccess))
			Expect(restOfLogs).ToNot(ContainSubstring(logDynamicConfigFailure))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring(logRequireBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logBackendReloadSuccess))
			Expect(restOfLogs).To(ContainSubstring(logSkipBackendReload))
			Expect(restOfLogs).ToNot(ContainSubstring(logInitialConfigSync))
		})
	})

	It("should handle a non backend update", func() {
		ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		ingress.Spec.TLS = []extensions.IngressTLS{
			{
				Hosts:      []string{"foo.com"},
				SecretName: "foo.com",
			},
		}

		_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ingress.Spec.TLS[0].Hosts,
			ingress.Spec.TLS[0].SecretName,
			ingress.Namespace)
		Expect(err).ToNot(HaveOccurred())

		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
		Expect(err).ToNot(HaveOccurred())

		By("generating the respective ssl listen directive")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "server_name foo.com") &&
					strings.Contains(server, "listen 443")
			})
		Expect(err).ToNot(HaveOccurred())

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		By("reloading Nginx")
		Expect(log).To(ContainSubstring(logBackendReloadSuccess))

		By("POSTing new backends to Lua endpoint")
		Expect(log).To(ContainSubstring(logDynamicConfigSuccess))

		By("still be proxying requests through Lua balancer")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not fail requests when upstream-hash-by annotation is set", func() {
		ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/upstream-hash-by"] = "$query_string"
		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
		Expect(err).ToNot(HaveOccurred())

		err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", 2, nil)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(10 * time.Second)

		resp, body, errs := gorequest.New().
			Get(fmt.Sprintf("%s?a-unique-request-uri", f.IngressController.HTTPURL)).
			Set("Host", "foo.com").
			End()
		Expect(len(errs)).Should(Equal(0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		hostnamePattern := regexp.MustCompile(`Hostname: ([a-zA-Z0-9\-]+)`)
		upstreamName := hostnamePattern.FindAllStringSubmatch(body, -1)[0][1]

		for i := 0; i < 5; i++ {
			resp, body, errs := gorequest.New().
				Get(fmt.Sprintf("%s?a-unique-request-uri", f.IngressController.HTTPURL)).
				Set("Host", "foo.com").
				End()
			Expect(len(errs)).Should(Equal(0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			newUpstreamName := hostnamePattern.FindAllStringSubmatch(body, -1)[0][1]
			Expect(newUpstreamName).Should(Equal(upstreamName))
		}
	})

	Context("when session affinity annotation is present", func() {
		It("should use sticky sessions when ingress rules are configured", func() {
			err := framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 2, nil)
			Expect(err).NotTo(HaveOccurred())

			cookieName := "STICKYSESSION"

			By("Updating affinity annotation on ingress")
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			ingress.ObjectMeta.Annotations = map[string]string{
				"nginx.ingress.kubernetes.io/affinity":            "cookie",
				"nginx.ingress.kubernetes.io/session-cookie-name": cookieName,
			}
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
			Expect(err).ToNot(HaveOccurred())

			By("Increasing the number of service replicas")
			err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", 2, nil)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(5 * time.Second)

			By("Making a first request")
			host := "foo.com"
			resp, body, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			hostnamePattern := regexp.MustCompile(`Hostname: ([a-zA-Z0-9\-]+)`)
			upstreamName := hostnamePattern.FindAllStringSubmatch(body, -1)[0][1]

			cookies := (*http.Response)(resp).Cookies()
			sessionCookie, err := getCookie(cookieName, cookies)
			Expect(err).ToNot(HaveOccurred())

			Expect(sessionCookie.Domain).Should(Equal(host))

			By("Making many requests with the previous session cookie")
			for i := 0; i < 5; i++ {
				resp, _, errs = gorequest.New().
					Get(f.IngressController.HTTPURL).
					AddCookie(sessionCookie).
					Set("Host", host).
					End()
				Expect(len(errs)).Should(BeNumerically("==", 0))
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))

				newCookies := (*http.Response)(resp).Cookies()
				_, err := getCookie(cookieName, newCookies)
				By("Omitting cookie in all subsequent requests")
				Expect(err).To(HaveOccurred())

				By("By proxying to the same upstream")
				newUpstreamName := hostnamePattern.FindAllStringSubmatch(body, -1)[0][1]
				Expect(newUpstreamName).Should(Equal(upstreamName))
			}
		})

		It("should NOT use sticky sessions when a default backend and no ingress rules configured", func() {
			By("Updating affinity annotation and rules on ingress")
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			ingress.Spec = extensions.IngressSpec{
				Backend: &extensions.IngressBackend{
					ServiceName: "http-svc",
					ServicePort: intstr.FromInt(80),
				},
			}
			ingress.ObjectMeta.Annotations = map[string]string{
				"nginx.ingress.kubernetes.io/affinity": "cookie",
			}
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(5 * time.Second)

			By("Making a request")
			host := "foo.com"
			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			By("Ensuring no cookies are set")
			cookies := (*http.Response)(resp).Cookies()
			Expect(len(cookies)).Should(BeNumerically("==", 0))
		})
	})
})

func enableDynamicConfiguration(namespace string, kubeClientSet kubernetes.Interface) error {
	return framework.UpdateDeployment(kubeClientSet, namespace, "nginx-ingress-controller", 1,
		func(deployment *appsv1beta1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--enable-dynamic-configuration")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := kubeClientSet.AppsV1beta1().Deployments(namespace).Update(deployment)
			if err != nil {
				return err
			}

			return nil
		})
}

func ensureIngress(f *framework.Framework, host string) (*extensions.Ingress, error) {
	return f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, &map[string]string{
		"nginx.ingress.kubernetes.io/load-balance": "ewma",
	}))
}

func getCookie(name string, cookies []*http.Cookie) (*http.Cookie, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return &http.Cookie{}, fmt.Errorf("Cookie does not exist")
}
