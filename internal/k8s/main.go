/*
Copyright 2015 The Kubernetes Authors.

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

package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog"

	apiv1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ParseNameNS parses a string searching a namespace and name
func ParseNameNS(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
}

// GetNodeIPOrName returns the IP address or the name of a node in the cluster
func GetNodeIPOrName(kubeClient clientset.Interface, name string, useInternalIP bool) string {
	node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Error getting node %v: %v", name, err)
		return ""
	}

	defaultOrInternalIP := ""
	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeInternalIP {
			if address.Address != "" {
				defaultOrInternalIP = address.Address
				break
			}
		}
	}

	if useInternalIP {
		return defaultOrInternalIP
	}

	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeExternalIP {
			if address.Address != "" {
				return address.Address
			}
		}
	}

	return defaultOrInternalIP
}

// PodInfo contains runtime information about the pod running the Ingres controller
type PodInfo struct {
	Name      string
	Namespace string
	// Labels selectors of the running pod
	// This is used to search for other Ingress controller pods
	Labels map[string]string
}

// GetPodDetails returns runtime information about the pod:
// name, namespace and IP of the node where it is running
func GetPodDetails(kubeClient clientset.Interface) (*PodInfo, error) {
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNs == "" {
		return nil, fmt.Errorf("unable to get POD information (missing POD_NAME or POD_NAMESPACE environment variable")
	}

	pod, _ := kubeClient.CoreV1().Pods(podNs).Get(context.TODO(), podName, metav1.GetOptions{})
	if pod == nil {
		return nil, fmt.Errorf("unable to get POD information")
	}

	return &PodInfo{
		Name:      podName,
		Namespace: podNs,
		Labels:    pod.GetLabels(),
	}, nil
}

// MetaNamespaceKey knows how to make keys for API objects which implement meta.Interface.
func MetaNamespaceKey(obj interface{}) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Warning(err)
	}

	return key
}

// IsNetworkingIngressAvailable indicates if package "k8s.io/api/networking/v1beta1" is available or not
var IsNetworkingIngressAvailable bool

// IsIngressV1Ready indicates if the running Kubernetes version is at least v1.18.0
var IsIngressV1Ready bool

// IngressClass indicates the class of the Ingress to use as filter
var IngressClass *networkingv1beta1.IngressClass

// IngressNGINXController defines the valid value of IngressClass
// Controller field for ingress-nginx
const IngressNGINXController = "k8s.io/ingress-nginx"

// NetworkingIngressAvailable checks if the package "k8s.io/api/networking/v1beta1"
// is available or not and if Ingress V1 is supported (k8s >= v1.18.0)
func NetworkingIngressAvailable(client clientset.Interface) (bool, bool) {
	// check kubernetes version to use new ingress package or not
	version114, _ := version.ParseGeneric("v1.14.0")
	version118, _ := version.ParseGeneric("v1.18.0")

	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return false, false
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		klog.Errorf("unexpected error parsing running Kubernetes version: %v", err)
		return false, false
	}

	return runningVersion.AtLeast(version114), runningVersion.AtLeast(version118)
}

// default path type is Prefix to not break existing definitions
var defaultPathType = networkingv1beta1.PathTypePrefix

// SetDefaultNGINXPathType sets a default PathType when is not defined.
func SetDefaultNGINXPathType(ing *networkingv1beta1.Ingress) {
	for _, rule := range ing.Spec.Rules {
		if rule.IngressRuleValue.HTTP == nil {
			continue
		}

		for idx := range rule.IngressRuleValue.HTTP.Paths {
			p := &rule.IngressRuleValue.HTTP.Paths[idx]
			if p.PathType == nil {
				p.PathType = &defaultPathType
			}

			if *p.PathType == networkingv1beta1.PathTypeImplementationSpecific {
				p.PathType = &defaultPathType
			}
		}
	}
}
