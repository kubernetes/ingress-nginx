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

package settings

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Status Update [Status]", func() {
	f := framework.NewDefaultFramework("status-update")
	host := "status-update"
	address := getHostIP()

	BeforeEach(func() {
	})

	AfterEach(func() {
	})

	It("should update status field after client-go reconnection", func() {
		port, cmd, err := f.KubectlProxy(0)
		Expect(err).NotTo(HaveOccurred(), "unexpected error starting kubectl proxy")

		err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				args := deployment.Spec.Template.Spec.Containers[0].Args
				args = append(args, fmt.Sprintf("--apiserver-host=http://%s:%d", address.String(), port))
				args = append(args, "--publish-status-address=1.1.0.0")
				// flags --publish-service and --publish-status-address are mutually exclusive
				var index int
				for k, v := range args {
					if strings.Index(v, "--publish-service") != -1 {
						index = k
						break
					}
				}
				if index > -1 {
					args[index] = ""
				}

				deployment.Spec.Template.Spec.Containers[0].Args = args
				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.IngressController.Namespace).Update(deployment)
				return err
			})
		Expect(err).NotTo(HaveOccurred(), "unexpected error updating ingress controller deployment flags")

		f.NewEchoDeploymentWithReplicas(1)

		ing := f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, fmt.Sprintf("server_name %s", host))
			})

		framework.Logf("waiting for leader election and initial status update")
		time.Sleep(30 * time.Second)

		err = cmd.Process.Kill()
		Expect(err).NotTo(HaveOccurred(), "unexpected error terminating kubectl proxy")

		ing, err = f.KubeClientSet.Extensions().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred(), "unexpected error getting %s/%v Ingress", f.IngressController.Namespace, host)

		ing.Status.LoadBalancer.Ingress = []apiv1.LoadBalancerIngress{}
		_, err = f.KubeClientSet.Extensions().Ingresses(f.IngressController.Namespace).UpdateStatus(ing)
		Expect(err).NotTo(HaveOccurred(), "unexpected error cleaning Ingress status")
		time.Sleep(10 * time.Second)

		err = f.KubeClientSet.CoreV1().
			ConfigMaps(f.IngressController.Namespace).
			Delete("ingress-controller-leader-nginx", &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred(), "unexpected error deleting leader election configmap")

		_, cmd, err = f.KubectlProxy(port)
		Expect(err).NotTo(HaveOccurred(), "unexpected error starting kubectl proxy")
		defer func() {
			if cmd != nil {
				err := cmd.Process.Kill()
				Expect(err).NotTo(HaveOccurred(), "unexpected error terminating kubectl proxy")
			}
		}()

		err = wait.Poll(10*time.Second, framework.DefaultTimeout, func() (done bool, err error) {
			ing, err = f.KubeClientSet.Extensions().Ingresses(f.IngressController.Namespace).Get(host, metav1.GetOptions{})
			if err != nil {
				return false, err
			}

			if len(ing.Status.LoadBalancer.Ingress) != 1 {
				return false, nil
			}

			return true, nil
		})
		Expect(err).NotTo(HaveOccurred(), "unexpected error waiting for ingress status")
		Expect(ing.Status.LoadBalancer.Ingress).Should(Equal([]apiv1.LoadBalancerIngress{
			{IP: "1.1.0.0"},
		}))
	})
})

func getHostIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
