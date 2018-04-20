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

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/golang/glog"
	"github.com/pkg/errors"

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

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after. If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle

	IngressController *ingressController
}

type ingressController struct {
	HTTPURL  string
	HTTPSURL string

	Namespace string
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
	ingressNamespace, err := CreateKubeNamespace(f.BaseName, f.KubeClientSet)
	Expect(err).NotTo(HaveOccurred())

	f.IngressController = &ingressController{
		Namespace: ingressNamespace,
	}

	By("Starting new ingress controller")
	err = f.NewIngressController(f.IngressController.Namespace)
	Expect(err).NotTo(HaveOccurred())

	err = WaitForPodsReady(f.KubeClientSet, 5*time.Minute, 1, f.IngressController.Namespace, metav1.ListOptions{
		LabelSelector: "app=ingress-nginx",
	})
	Expect(err).NotTo(HaveOccurred())

	By("Building NGINX HTTP URL")
	HTTPURL, err := f.GetNginxURL(HTTP)
	Expect(err).NotTo(HaveOccurred())

	f.IngressController.HTTPURL = HTTPURL

	By("Building NGINX HTTPS URL")
	HTTPSURL, err := f.GetNginxURL(HTTPS)
	Expect(err).NotTo(HaveOccurred())

	f.IngressController.HTTPSURL = HTTPSURL

	// we wait for any change in the informers and SSL certificate generation
	time.Sleep(5 * time.Second)
}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	By("Waiting for test namespace to no longer exist")
	err := DeleteKubeNamespace(f.KubeClientSet, f.IngressController.Namespace)
	Expect(err).NotTo(HaveOccurred())
}

// IngressNginxDescribe wrapper function for ginkgo describe. Adds namespacing.
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
	s, err := f.KubeClientSet.
		CoreV1().
		Services(f.IngressController.Namespace).
		Get("ingress-nginx", metav1.GetOptions{})
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
	return wait.Poll(Poll, time.Minute*5, f.matchNginxConditions(name, matcher))
}

// WaitForNginxConfiguration waits until the nginx configuration contains a particular configuration
func (f *Framework) WaitForNginxConfiguration(matcher func(cfg string) bool) error {
	return wait.Poll(Poll, time.Minute*5, f.matchNginxConditions("", matcher))
}

// NginxLogs returns the logs of the nginx ingress controller pod running
func (f *Framework) NginxLogs() (string, error) {
	l, err := f.KubeClientSet.CoreV1().Pods(f.IngressController.Namespace).List(metav1.ListOptions{
		LabelSelector: "app=ingress-nginx",
	})
	if err != nil {
		return "", err
	}

	for _, pod := range l.Items {
		if strings.HasPrefix(pod.GetName(), "nginx-ingress-controller") {
			if isRunning, err := podRunningReady(&pod); err == nil && isRunning {
				return f.Logs(&pod)
			}
		}
	}

	return "", fmt.Errorf("no nginx ingress controller pod is running (logs)")
}

func (f *Framework) matchNginxConditions(name string, matcher func(cfg string) bool) wait.ConditionFunc {
	return func() (bool, error) {
		l, err := f.KubeClientSet.CoreV1().Pods(f.IngressController.Namespace).List(metav1.ListOptions{
			LabelSelector: "app=ingress-nginx",
		})
		if err != nil {
			return false, err
		}

		if len(l.Items) == 0 {
			return false, nil
		}

		var cmd string
		if name == "" {
			cmd = fmt.Sprintf("cat /etc/nginx/nginx.conf")
		} else {
			cmd = fmt.Sprintf("cat /etc/nginx/nginx.conf | awk '/## start server %v/,/## end server %v/'", name, name)
		}

		var pod *v1.Pod

		for _, p := range l.Items {
			if strings.HasPrefix(p.GetName(), "nginx-ingress-controller") {
				if isRunning, err := podRunningReady(&p); err == nil && isRunning {
					pod = &p
					break
				}
			}
		}

		if pod == nil {
			return false, nil
		}

		o, err := f.ExecCommand(pod, cmd)
		if err != nil {
			return false, err
		}

		var match bool
		errs := InterceptGomegaFailures(func() {
			if glog.V(10) && len(o) > 0 {
				glog.Infof("nginx.conf:\n%v", o)
			}

			if matcher(o) {
				match = true
			}
		})

		if match {
			return true, nil
		}

		if len(errs) > 0 {
			glog.V(2).Infof("Errors waiting for conditions: %v", errs)
		}

		return false, nil
	}
}

func (f *Framework) getNginxConfigMap() (*v1.ConfigMap, error) {
	if f.KubeClientSet == nil {
		return nil, fmt.Errorf("KubeClientSet not initialized")
	}

	config, err := f.KubeClientSet.
		CoreV1().
		ConfigMaps(f.IngressController.Namespace).
		Get("nginx-configuration", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return config, err
}

// GetNginxConfigMapData gets ingress-nginx's nginx-configuration map's data
func (f *Framework) GetNginxConfigMapData() (map[string]string, error) {
	config, err := f.getNginxConfigMap()
	if err != nil {
		return nil, err
	}
	if config.Data == nil {
		config.Data = map[string]string{}
	}

	return config.Data, err
}

// SetNginxConfigMapData sets ingress-nginx's nginx-configuration configMap data
func (f *Framework) SetNginxConfigMapData(cmData map[string]string) error {
	// Needs to do a Get and Set, Update will not take just the Data field
	// or a configMap that is not the very last revision
	config, err := f.getNginxConfigMap()
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("Unable to get nginx-configuration configMap")
	}

	config.Data = cmData

	_, err = f.KubeClientSet.
		CoreV1().
		ConfigMaps(f.IngressController.Namespace).
		Update(config)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	return err
}

// UpdateNginxConfigMapData updates single field in ingress-nginx's nginx-configuration map data
func (f *Framework) UpdateNginxConfigMapData(key string, value string) error {
	config, err := f.GetNginxConfigMapData()
	if err != nil {
		return err
	}

	config[key] = value

	return f.SetNginxConfigMapData(config)
}

// UpdateDeployment runs the given updateFunc on the deployment and waits for it to be updated
func UpdateDeployment(kubeClientSet kubernetes.Interface, namespace string, name string, replicas int, updateFunc func(d *appsv1beta1.Deployment) error) error {
	deployment, err := kubeClientSet.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if updateFunc != nil {
		if err := updateFunc(deployment); err != nil {
			return err
		}
	}

	if *deployment.Spec.Replicas != int32(replicas) {
		glog.Infof("updating replica count from %v to %v...", *deployment.Spec.Replicas, replicas)
		deployment, err := kubeClientSet.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Replicas = NewInt32(int32(replicas))
		_, err = kubeClientSet.AppsV1beta1().Deployments(namespace).Update(deployment)
		if err != nil {
			return errors.Wrapf(err, "scaling the number of replicas to %v", replicas)
		}
	}

	err = WaitForPodsReady(kubeClientSet, 5*time.Minute, replicas, namespace, metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(deployment.Spec.Template.ObjectMeta.Labels)).String(),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to wait for nginx-ingress-controller replica count to be %v", replicas)
	}

	return nil
}

// NewSingleIngress creates a simple ingress rule
func NewSingleIngress(name, path, host, ns string, annotations *map[string]string) *extensions.Ingress {
	if annotations == nil {
		annotations = &map[string]string{}
	}

	ing := &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: *annotations,
		},
		Spec: extensions.IngressSpec{
			TLS: []extensions.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: host,
				},
			},
			Rules: []extensions.IngressRule{
				{
					Host: host,
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Path: path,
									Backend: extensions.IngressBackend{
										ServiceName: "http-svc",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ing
}
