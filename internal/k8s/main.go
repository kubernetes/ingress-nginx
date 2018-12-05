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
	"fmt"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
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
	node, err := kubeClient.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Error getting node %v: %v", name, err)
		return ""
	}

	if useInternalIP {
		for _, address := range node.Status.Addresses {
			if address.Type == apiv1.NodeInternalIP {
				if address.Address != "" {
					return address.Address
				}
			}
		}
	}

	for _, address := range node.Status.Addresses {
		if address.Type == apiv1.NodeExternalIP {
			if address.Address != "" {
				return address.Address
			}
		}
	}

	return ""
}

// PodInfo contains runtime information about the pod running the Ingres controller
type PodInfo struct {
	Name      string
	Namespace string
	NodeIP    string
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

	pod, _ := kubeClient.CoreV1().Pods(podNs).Get(podName, metav1.GetOptions{})
	if pod == nil {
		return nil, fmt.Errorf("unable to get POD information")
	}

	return &PodInfo{
		Name:      podName,
		Namespace: podNs,
		NodeIP:    GetNodeIPOrName(kubeClient, pod.Spec.NodeName, true),
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
