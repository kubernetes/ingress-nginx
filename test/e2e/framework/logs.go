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

// Logs returns the log entries of a given Pod.
func Logs(pod *v1.Pod) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	if len(pod.Spec.Containers) != 1 {
		return "", fmt.Errorf("could not determine which container to use")
	}

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v logs --namespace %s %s", KubectlPath, pod.Namespace, pod.Name))
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not execute '%s %s': %v", cmd.Path, cmd.Args, err)
	}

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	return execOut.String(), nil
}
