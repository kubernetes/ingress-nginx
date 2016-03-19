/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package main

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	errMissingPodInfo = fmt.Errorf("Unable to get POD information")
)

// StoreToIngressLister makes a Store that lists Ingress.
type StoreToIngressLister struct {
	cache.Store
}

// getLBDetails returns runtime information about the pod (name, IP) and replication
// controller or daemonset (namespace and name).
// This is required to watch for changes in annotations or configuration (ConfigMap)
func getLBDetails(kubeClient *unversioned.Client) (*lbInfo, error) {
	podIP := os.Getenv("POD_IP")
	podName := os.Getenv("POD_NAME")
	podNs := os.Getenv("POD_NAMESPACE")

	pod, _ := kubeClient.Pods(podNs).Get(podName)
	if pod == nil {
		return nil, errMissingPodInfo
	}

	return &lbInfo{
		PodIP:        podIP,
		Podname:      podName,
		PodNamespace: podNs,
	}, nil
}

func isValidService(kubeClient *unversioned.Client, name string) error {
	if name == "" {
		return fmt.Errorf("Empty string is not a valid service name")
	}

	parts := strings.Split(name, "/")
	if len(parts) != 2 {
		return fmt.Errorf("Invalid name format (namespace/name) in service '%v'", name)
	}

	_, err := kubeClient.Services(parts[0]).Get(parts[1])
	return err
}

func isHostValid(host string, cns []string) bool {
	for _, cn := range cns {
		if matchHostnames(cn, host) {
			return true
		}
	}

	return false
}

func matchHostnames(pattern, host string) bool {
	host = strings.TrimSuffix(host, ".")
	pattern = strings.TrimSuffix(pattern, ".")

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

func parseNsName(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
}
