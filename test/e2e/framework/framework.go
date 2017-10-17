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

	"k8s.io/api/core/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	podName = "test-ingress-controller"
)

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
	APIExtensionsClientSet apiextcs.Interface

	// Namespace in which all test resources should reside
	Namespace *v1.Namespace

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach for
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

	f.KubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Building a namespace api object")
	f.Namespace, err = CreateKubeNamespace(f.BaseName, f.KubeClientSet)
	Expect(err).NotTo(HaveOccurred())
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	By("Deleting test namespace")
	err := DeleteKubeNamespace(f.KubeClientSet, f.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())

	By("Waiting for test namespace to no longer exist")
	err = WaitForKubeNamespaceNotExist(f.KubeClientSet, f.Namespace.Name)
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
	s, err := f.KubeClientSet.CoreV1().Services("ingress-nginx").Get("ingress-nginx", meta_v1.GetOptions{})
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
