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
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
)

// ExecIngressPod executes a command inside the first container in ingress controller running pod
func (f *Framework) ExecIngressPod(command string) (string, error) {
	pod, err := getIngressNGINXPod(f.IngressController.Namespace, f.KubeClientSet)
	if err != nil {
		return "", err
	}

	return f.ExecCommand(pod, command)
}

// ExecCommand executes a command inside a the first container in a running pod
func (f *Framework) ExecCommand(pod *v1.Pod, command string) (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v exec --namespace %s %s --container nginx-ingress-controller -- %s", KubectlPath, pod.Namespace, pod.Name, command))
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

// NewIngressController deploys a new NGINX Ingress controller in a namespace
func (f *Framework) NewIngressController(namespace string) error {
	// Creates an nginx deployment
	cmd := exec.Command("./wait-for-nginx.sh", namespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unexpected error waiting for ingress controller deployment: %v.\nLogs:\n%v", err, string(out))
	}

	return nil
}

var (
	proxyRegexp = regexp.MustCompile("Starting to serve on .*:([0-9]+)")
)

// KubectlProxy creates a proxy to kubernetes apiserver
func (f *Framework) KubectlProxy(port int) (int, *exec.Cmd, error) {
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s proxy --accept-hosts=.* --address=0.0.0.0 --port=%d", KubectlPath, port))
	stdout, stderr, err := startCmdAndStreamOutput(cmd)
	if err != nil {
		return -1, nil, err
	}

	defer stdout.Close()
	defer stderr.Close()

	buf := make([]byte, 128)
	var n int
	if n, err = stdout.Read(buf); err != nil {
		return -1, cmd, fmt.Errorf("Failed to read from kubectl proxy stdout: %v", err)
	}

	output := string(buf[:n])
	match := proxyRegexp.FindStringSubmatch(output)
	if len(match) == 2 {
		if port, err := strconv.Atoi(match[1]); err == nil {
			return port, cmd, nil
		}
	}

	return -1, cmd, fmt.Errorf("Failed to parse port from proxy stdout: %s", output)
}

func startCmdAndStreamOutput(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}

	Logf("Asynchronously running '%s'", strings.Join(cmd.Args, " "))
	err = cmd.Start()
	return
}
