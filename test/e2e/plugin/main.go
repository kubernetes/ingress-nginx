/*
Copyright 2019 The Kubernetes Authors.

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

package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("kubectl plugin", func() {
	f := framework.NewDefaultFramework("plugin")
	host := "foo.com"
	path := "/a/really/long/random/path/very/convoluted"
	annotations := map[string]string{}

	BeforeEach(func() {
		err := framework.UpdateDeployment(f.KubeClientSet, f.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, "--enable-dynamic-certificates")
				args = append(args, "--enable-ssl-chain-completion=false")
				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.Namespace).Update(deployment)

				return err
			})
		Expect(err).NotTo(HaveOccurred())

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "http-svc", 80, &annotations)
		f.EnsureIngress(ing)
		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "ok, res = pcall(require, \"certificate\")") && strings.Contains(cfg, host)
			})
	})

	Context("the backends subcommand", func() {
		It("should list the backend servers", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "backends", "--list", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())

			// Should be 2: the default and the echo deployment
			numUpstreams := len(strings.Split(strings.Trim(string(output), "\n"), "\n"))
			Expect(numUpstreams).Should(Equal(2))
		})
	})

	Context("the certs subcommand", func() {
		It("should find and return an extant cert", func() {
			ing, err := f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace).Get(host, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			ing.Spec.TLS = []extensions.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: host,
				},
			}
			_, err = framework.CreateIngressTLSSecret(f.KubeClientSet,
				ing.Spec.TLS[0].Hosts,
				ing.Spec.TLS[0].SecretName,
				ing.Namespace)
			Expect(err).ToNot(HaveOccurred())
			_, err = f.KubeClientSet.ExtensionsV1beta1().Ingresses(f.Namespace).Update(ing)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(5 * time.Second)

			cmd := exec.Command("/kubectl-ingress_nginx", "certs", "--host", host, "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())
			Expect(string(output)).Should(ContainSubstring("BEGIN CERTIFICATE"))
		})

		It("should not find a cert that doesn't exist", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "certs", "--host", "nonexistenthost.com", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())
			Expect(strings.Trim(string(output), "\n")).Should(Equal("No cert found for host nonexistenthost.com"))
		})
	})

	Context("the conf subcommand", func() {
		It("should find the server block for a host that exists", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "conf", "--host", host, "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())
			Expect(strings.Trim(string(output), "\n")).ShouldNot(Equal(""))
		})

		It("should find nothing for a host that doesn't exist", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "conf", "--host", "nonexistenthost.com", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())
			Expect(strings.Trim(string(output), "\n")).Should(Equal("Host nonexistenthost.com was not found in the controller's nginx.conf"))
		})
	})

	Context("the exec subcommand", func() {
		It("should execute a simple command inside the pod", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "exec", "--namespace", f.Namespace, "--", "ls", "/")
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())

			reference, err := f.ExecIngressPod("ls /")
			Expect(err).Should(BeNil())
			Expect(string(output)).Should(Equal(string(reference)))
		})
	})

	Context("the general subcommand", func() {
		It("should return valid JSON", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "general", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())

			var f interface{}
			unmarshalErr := json.Unmarshal([]byte(output), &f)
			Expect(unmarshalErr).Should(BeNil())
		})
	})

	Context("the info subcommand", func() {
		It("should get the correct cluster IP", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "info", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())

			lines := strings.Split(strings.Trim(string(output), "\n"), "\n")
			Expect(len(lines)).Should(Equal(2))
			Expect(lines[0]).Should(Equal(fmt.Sprintf("Service cluster IP address: %v", f.GetNginxIP())))
		})
	})

	Context("the ingresses subcommand", func() {
		It("should find the ingress definition", func() {
			cmd := exec.Command("/kubectl-ingress_nginx", "ingresses", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)
			Expect(err).Should(BeNil())

			lines := strings.Split(strings.Trim(string(output), "\n"), "\n")
			Expect(len(lines)).Should(Equal(2))

			columns := make([]string, 0)
			for _, c := range strings.Split(lines[1], " ") {
				s := strings.Trim(c, " ")
				if s != "" {
					columns = append(columns, s)
				}
			}
			Expect(len(columns)).Should(Or(Equal(7), Equal(6)))

			Expect(columns[0]).Should(Equal("foo.com"))  // INGRESS
			Expect(columns[1]).Should(Equal(host + "/")) // HOST+PATH

			i := len(columns) - 6                          // If it exists, skip ADDRESSES
			Expect(columns[2+i]).Should(Equal("NO"))       // TLS
			Expect(columns[3+i]).Should(Equal("http-svc")) // SERVICE NAME
			Expect(columns[4+i]).Should(Equal("80"))       // SERVICE PORT
			Expect(columns[5+i]).Should(Equal("1"))        // ENDPOINTS
		})
	})

	Context("the logs subcommand", func() {
		It("should find a request log", func() {
			resp, _, errs := gorequest.New().
				Get(f.GetURL(framework.HTTP)+path).
				Set("Host", host).
				End()

			Expect(errs).Should(BeEmpty())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))

			cmd := exec.Command("/kubectl-ingress_nginx", "logs", "--namespace", f.Namespace)
			output, err := f.ExecAnyCommand(cmd)

			Expect(err).Should(BeNil())
			Expect(string(output)).Should(ContainSubstring(path))
		})
	})
})
