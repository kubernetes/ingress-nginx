/*
Copyright 2014 The Kubernetes Authors.

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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/ingress-nginx/internal/file"
)

const (
	// Poll how often to poll for conditions
	Poll = 2 * time.Second

	// Default time to wait for operations to complete
	defaultTimeout = 30 * time.Second
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

// Logf logs to the INFO logs.
func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

// Failf logs to the INFO logs and fails the test.
func Failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	Fail(nowStamp()+": "+msg, 1)
}

// Skipf logs to the INFO logs and skips the test.
func Skipf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	Skip(nowStamp() + ": " + msg)
}

// RestclientConfig deserializes the contents of a kubeconfig file into a Config object.
func RestclientConfig(config, context string) (*api.Config, error) {
	Logf(">>> config: %s\n", config)
	if config == "" {
		return nil, fmt.Errorf("config file must be specified to load client config")
	}
	c, err := clientcmd.LoadFromFile(config)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err.Error())
	}
	if context != "" {
		Logf(">>> context: %s\n", context)
		c.CurrentContext = context
	}
	return c, nil
}

// LoadConfig deserializes the contents of a kubeconfig file into a REST configuration.
func LoadConfig(config, context string) (*rest.Config, error) {
	c, err := RestclientConfig(config, context)
	if err != nil {
		return nil, err
	}
	return clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{}).ClientConfig()
}

// RunID unique identifier of the e2e run
var RunID = uuid.NewUUID()

// CreateKubeNamespace creates a new namespace in the cluster
func CreateKubeNamespace(baseName string, c kubernetes.Interface) (string, error) {
	ts := time.Now().UnixNano()
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-%v-", baseName, ts),
		},
	}
	// Be robust about making the namespace creation call.
	var got *v1.Namespace
	err := wait.PollImmediate(Poll, defaultTimeout, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(ns)
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		Logf("Created namespace: %v", got.Name)
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return got.Name, nil
}

// DeleteKubeNamespace deletes a namespace and all the objects inside
func DeleteKubeNamespace(c kubernetes.Interface, namespace string) error {
	return c.CoreV1().Namespaces().Delete(namespace, metav1.NewDeleteOptions(0))
}

// ExpectNoError tests whether an error occured.
func ExpectNoError(err error, explain ...interface{}) {
	if err != nil {
		Logf("Unexpected error occurred: %v", err)
	}
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), explain...)
}

// WaitForKubeNamespaceNotExist waits until a namespaces is not present in the cluster
func WaitForKubeNamespaceNotExist(c kubernetes.Interface, namespace string) error {
	return wait.PollImmediate(Poll, time.Minute*5, namespaceNotExist(c, namespace))
}

func namespaceNotExist(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}
}

// WaitForNoPodsInNamespace waits until there are no pods running in a namespace
func WaitForNoPodsInNamespace(c kubernetes.Interface, namespace string) error {
	return wait.PollImmediate(Poll, time.Minute*5, noPodsInNamespace(c, namespace))
}

func noPodsInNamespace(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		items, err := c.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		if len(items.Items) == 0 {
			return true, nil
		}
		return false, nil
	}
}

// WaitForPodRunningInNamespace waits a default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodRunningInNamespace(c kubernetes.Interface, pod *v1.Pod) error {
	if pod.Status.Phase == v1.PodRunning {
		return nil
	}
	return waitTimeoutForPodRunningInNamespace(c, pod.Name, pod.Namespace, defaultTimeout)
}

func waitTimeoutForPodRunningInNamespace(c kubernetes.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, defaultTimeout, podRunning(c, podName, namespace))
}

// WaitForSecretInNamespace waits a default amount of time for the specified secret is present in a particular namespace
func WaitForSecretInNamespace(c kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(1*time.Second, time.Minute*2, secretInNamespace(c, namespace, name))
}

func secretInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		s, err := c.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, err
		}
		if err != nil {
			return false, err
		}

		if s != nil {
			return true, nil
		}
		return false, nil
	}
}

// WaitForFileInFS waits a default amount of time for the specified file is present in the filesystem
func WaitForFileInFS(file string, fs file.Filesystem) error {
	return wait.PollImmediate(1*time.Second, time.Minute*2, fileInFS(file, fs))
}

func fileInFS(file string, fs file.Filesystem) wait.ConditionFunc {
	return func() (bool, error) {
		stat, err := fs.Stat(file)
		if err != nil {
			return false, err
		}

		if stat == nil {
			return false, fmt.Errorf("file %v does not exists", file)
		}

		if stat.Size() > 0 {
			return true, nil
		}

		return false, fmt.Errorf("the file %v exists but it is empty", file)
	}
}

// WaitForNoIngressInNamespace waits until there is no ingress object in a particular namespace
func WaitForNoIngressInNamespace(c kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(1*time.Second, time.Minute*2, noIngressInNamespace(c, namespace, name))
}

func noIngressInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		ing, err := c.ExtensionsV1beta1().Ingresses(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		if ing == nil {
			return true, nil
		}
		return false, nil
	}
}

// WaitForIngressInNamespace waits until a particular ingress object exists namespace
func WaitForIngressInNamespace(c kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(1*time.Second, time.Minute*2, ingressInNamespace(c, namespace, name))
}

func ingressInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		ing, err := c.ExtensionsV1beta1().Ingresses(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, err
		}
		if err != nil {
			return false, err
		}

		if ing != nil {
			return true, nil
		}
		return false, nil
	}
}

func podRunning(c kubernetes.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, fmt.Errorf("pod ran to completion")
		}
		return false, nil
	}
}

// NewInt32 converts int32 to a pointer
func NewInt32(val int32) *int32 {
	p := new(int32)
	*p = val
	return p
}

// NewInt64 converts int64 to a pointer
func NewInt64(val int64) *int64 {
	p := new(int64)
	*p = val
	return p
}
