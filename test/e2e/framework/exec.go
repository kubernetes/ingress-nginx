/*
Copyright 2017 The Kubernetes Authors.

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
	"bytes"
	"fmt"
	"os/exec"

	"k8s.io/api/core/v1"
)

// ExecCommand executes a command inside a the first container in a running pod
func (f *Framework) ExecCommand(pod *v1.Pod, command string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	args := fmt.Sprintf("kubectl exec --namespace %v %v --container nginx-ingress-controller -- %v", pod.Namespace, pod.Name, command)
	cmd := exec.Command("/bin/bash", "-c", args)
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not execute: %v", err)
	}

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	return execOut.String(), nil
}

// NewIngressController deploys a new NGINX Ingress controller in a namespace
func (f *Framework) NewIngressController(namespace string) error {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	cmd := exec.Command("test/e2e/wait-for-nginx.sh", namespace)
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not execute: %v", err)
	}

	if execErr.Len() > 0 {
		return fmt.Errorf("stderr: %v", execErr.String())
	}

	return nil
}
