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
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"k8s.io/ingress-nginx/test/e2e/framework"
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

		time.Sleep(5 * time.Second)

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())

		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			Set("Host", host).
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(ContainSubstring("could not dynamically reconfigure"))
		Expect(log).To(ContainSubstring("first sync of Nginx configuration"))
	})

	AfterEach(func() {
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

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, "id=endpoints_only_changes")
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring("dynamic reconfiguration succeeded"))
			Expect(restOfLogs).ToNot(ContainSubstring("could not dynamically reconfigure"))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring("backend reload required"))
			Expect(restOfLogs).ToNot(ContainSubstring("ingress backend successfully reloaded"))
			Expect(restOfLogs).To(ContainSubstring("skipping reload"))
			Expect(restOfLogs).ToNot(ContainSubstring("first sync of Nginx configuration"))
		})

		It("should handle annotation changes", func() {
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())

			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/load-balance"] = "round_robin"
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(5 * time.Second)
			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())
			index := strings.Index(log, fmt.Sprintf("reason: 'UPDATE' Ingress %s/foo.com", f.IngressController.Namespace))
			restOfLogs := log[index:]

			By("POSTing new backends to Lua endpoint")
			Expect(restOfLogs).To(ContainSubstring("dynamic reconfiguration succeeded"))
			Expect(restOfLogs).ToNot(ContainSubstring("could not dynamically reconfigure"))

			By("skipping Nginx reload")
			Expect(restOfLogs).ToNot(ContainSubstring("backend reload required"))
			Expect(restOfLogs).ToNot(ContainSubstring("ingress backend successfully reloaded"))
			Expect(restOfLogs).To(ContainSubstring("skipping reload"))
			Expect(restOfLogs).ToNot(ContainSubstring("first sync of Nginx configuration"))
		})
	})

	It("should handle a non backend update", func() {
		ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		ingress.Spec.TLS = []v1beta1.IngressTLS{
			{
				Hosts:      []string{"foo.com"},
				SecretName: "foo.com",
			},
		}

		_, _, _, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
			ingress.Spec.TLS[0].Hosts,
			ingress.Spec.TLS[0].SecretName,
			ingress.Namespace)
		Expect(err).ToNot(HaveOccurred())

		_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Update(ingress)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)
		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		By("reloading Nginx")
		Expect(log).To(ContainSubstring("ingress backend successfully reloaded"))

		By("POSTing new backends to Lua endpoint")
		Expect(log).To(ContainSubstring("dynamic reconfiguration succeeded"))

		By("still be proxying requests through Lua balancer")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "proxy_pass http://upstream_balancer;")
			})
		Expect(err).NotTo(HaveOccurred())

		By("generating the respective ssl listen directive")
		err = f.WaitForNginxServer("foo.com",
			func(server string) bool {
				return strings.Contains(server, "server_name foo.com") &&
					strings.Contains(server, "listen 443")
			})
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when session affinity annotation is present", func() {
		It("should use sticky sessions when ingress rules are configured", func() {
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
			time.Sleep(5 * time.Second)

			By("Increasing the number of service replicas")
			err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "http-svc", 2, nil)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			By("Making a first request")
			host := "foo.com"
			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			cookies := (*http.Response)(resp).Cookies()
			sessionCookie, err := getCookie(cookieName, cookies)
			Expect(err).ToNot(HaveOccurred())

			By("Making a second request with the previous session cookie")
			resp, _, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				AddCookie(sessionCookie).
				Set("Host", host).
				End()
			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			By("Making a third request with no cookie")
			resp, _, errs = gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()

			Expect(len(errs)).Should(BeNumerically("==", 0))
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			log, err := f.NginxLogs()
			Expect(err).ToNot(HaveOccurred())
			Expect(log).ToNot(BeEmpty())

			By("Checking that upstreams are sticky when session cookie is used")
			index := strings.Index(log, fmt.Sprintf("reason: 'UPDATE' Ingress %s/foo.com", f.IngressController.Namespace))
			reqLogs := log[index:]
			re := regexp.MustCompile(`\d{1,3}(?:\.\d{1,3}){3}(?::\d{1,5})`)
			upstreams := re.FindAllString(reqLogs, -1)
			Expect(len(upstreams)).Should(BeNumerically("==", 3))
			Expect(upstreams[0]).To(Equal(upstreams[1]))
			Expect(upstreams[1]).ToNot(Equal(upstreams[2]))
		})

		It("should NOT use sticky sessions when a default backend and no ingress rules configured", func() {
			By("Updating affinity annotation and rules on ingress")
			ingress, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.IngressController.Namespace).Get("foo.com", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			ingress.Spec = v1beta1.IngressSpec{
				Backend: &v1beta1.IngressBackend{
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

			time.Sleep(5 * time.Second)

			return nil
		})
}

func ensureIngress(f *framework.Framework, host string) (*extensions.Ingress, error) {
	return f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, &map[string]string{
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
