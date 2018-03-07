/*
Copyright 2017 Jetstack Ltd.
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

package framework

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/golang/glog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// RequestScheme define a scheme used in a test request.
type RequestScheme string

// These are valid test request schemes.
const (
	HTTP  RequestScheme = "http"
	HTTPS RequestScheme = "https"
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	BaseName string

	// A Kubernetes and Service Catalog client
	KubeClientSet          kubernetes.Interface
	KubeConfig             *restclient.Config
	APIExtensionsClientSet apiextcs.Interface

	// Namespace in which all test resources should reside
	Namespace *v1.Namespace

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after. If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle

	NginxHTTPURL  string
	NginxHTTPSURL string
}

// NewDefaultFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewDefaultFramework(baseName string) *Framework {
	f := &Framework{
		BaseName: baseName,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	f.cleanupHandle = AddCleanupAction(f.AfterEach)

	By("Creating a kubernetes client")
	kubeConfig, err := LoadConfig(TestContext.KubeConfig, TestContext.KubeContext)
	Expect(err).NotTo(HaveOccurred())

	f.KubeConfig = kubeConfig
	f.KubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Building a namespace api object")
	f.Namespace, err = CreateKubeNamespace(f.BaseName, f.KubeClientSet)
	Expect(err).NotTo(HaveOccurred())

	By("Building NGINX HTTP URL")
	f.NginxHTTPURL, err = f.GetNginxURL(HTTP)
	Expect(err).NotTo(HaveOccurred())

	By("Building NGINX HTTPS URL")
	f.NginxHTTPSURL, err = f.GetNginxURL(HTTPS)
	Expect(err).NotTo(HaveOccurred())
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	By("Deleting test namespace")
	err := DeleteKubeNamespace(f.KubeClientSet, f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for test namespace to no longer exist")
	err = WaitForNoPodsInNamespace(f.KubeClientSet, f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())
}

// IngressNginxDescribe wrapper function for ginkgo describe.  Adds namespacing.
func IngressNginxDescribe(text string, body func()) bool {
	return Describe("[nginx-ingress] "+text, body)
}

// GetNginxIP returns the IP address of the minikube cluster
// where the NGINX ingress controller is running
func (f *Framework) GetNginxIP() (string, error) {
	out, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetNginxPort returns the number of TCP port where NGINX is running
func (f *Framework) GetNginxPort(name string) (int, error) {
	s, err := f.KubeClientSet.CoreV1().Services("ingress-nginx").Get("ingress-nginx", metav1.GetOptions{})
	if err != nil {
		return -1, err
	}

	for _, p := range s.Spec.Ports {
		if p.NodePort != 0 && p.Name == name {
			return int(p.NodePort), nil
		}
	}

	return -1, err
}

// GetNginxURL returns the URL should be used to make a request to NGINX
func (f *Framework) GetNginxURL(scheme RequestScheme) (string, error) {
	ip, err := f.GetNginxIP()
	if err != nil {
		return "", err
	}

	port, err := f.GetNginxPort(fmt.Sprintf("%v", scheme))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v://%v:%v", scheme, ip, port), nil
}

// WaitForNginxServer waits until the nginx configuration contains a particular server section
func (f *Framework) WaitForNginxServer(name string, matcher func(cfg string) bool) error {
	// initial wait to allow the update of the ingress controller
	time.Sleep(5 * time.Second)
	return wait.PollImmediate(Poll, time.Minute*2, f.matchNginxConditions(name, matcher))
}

// WaitForNginxConfiguration waits until the nginx configuration contains a particular configuration
func (f *Framework) WaitForNginxConfiguration(matcher func(cfg string) bool) error {
	// initial wait to allow the update of the ingress controller
	time.Sleep(5 * time.Second)
	return wait.PollImmediate(Poll, time.Minute*2, f.matchNginxConditions("", matcher))
}

// NginxLogs returns the logs of the nginx ingress controller pod running
func (f *Framework) NginxLogs() (string, error) {
	l, err := f.KubeClientSet.CoreV1().Pods("ingress-nginx").List(metav1.ListOptions{
		LabelSelector: "app=ingress-nginx",
	})
	if err != nil {
		return "", err
	}

	if len(l.Items) == 0 {
		return "", fmt.Errorf("no nginx ingress controller pod is running")
	}

	for _, pod := range l.Items {
		if strings.HasPrefix(pod.GetName(), "nginx-ingress-controller") &&
			len(pod.Status.ContainerStatuses) > 0 &&
			pod.Status.ContainerStatuses[0].State.Running != nil {
			return f.Logs(&pod)
		}
	}

	return "", fmt.Errorf("no nginx ingress controller pod is running")
}

func (f *Framework) matchNginxConditions(name string, matcher func(cfg string) bool) wait.ConditionFunc {
	return func() (bool, error) {
		l, err := f.KubeClientSet.CoreV1().Pods("ingress-nginx").List(metav1.ListOptions{
			LabelSelector: "app=ingress-nginx",
		})
		if err != nil {
			return false, err
		}

		if len(l.Items) == 0 {
			return false, fmt.Errorf("no nginx ingress controller pod is running")
		}

		var cmd string
		if name == "" {
			cmd = fmt.Sprintf("cat /etc/nginx/nginx.conf")
		} else {
			cmd = fmt.Sprintf("cat /etc/nginx/nginx.conf | awk '/## start server %v/,/## end server %v/'", name, name)
		}

		var pod *v1.Pod
	Loop:
		for _, p := range l.Items {
			if strings.HasPrefix(p.GetName(), "nginx-ingress-controller") {
				for _, cs := range p.Status.ContainerStatuses {
					if cs.State.Running != nil && cs.Name == "nginx-ingress-controller" {
						pod = &p
						break Loop
					}
				}
			}
		}

		if pod == nil {
			return false, fmt.Errorf("no nginx ingress controller pod is running")
		}

		o, err := f.ExecCommand(pod, cmd)
		if err != nil {
			return false, err
		}

		var match bool
		errs := InterceptGomegaFailures(func() {
			if matcher(o) {
				match = true
			}
		})

		glog.V(2).Infof("Errors waiting for conditions: %v", errs)

		if match {
			return true, nil
		}

		return false, nil
	}
}
