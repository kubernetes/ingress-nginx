/*
Copyright 2019 The Kubernetes Authors.

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

package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// PodExecString takes a pod and a command, uses kubectl exec to run the command in the pod
// and returns stdout as a string
func PodExecString(flags *genericclioptions.ConfigFlags, pod *apiv1.Pod, args []string) (string, error) {
	args = append([]string{"exec", "-n", pod.Namespace, pod.Name}, args...)
	return ExecToString(flags, args)
}

// ExecToString runs a kubectl subcommand and returns stdout as a string
func ExecToString(flags *genericclioptions.ConfigFlags, args []string) (string, error) {
	kArgs := getKubectlConfigFlags(flags)
	kArgs = append(kArgs, args...)

	buf := bytes.NewBuffer(make([]byte, 0))
	err := execToWriter(append([]string{"kubectl"}, kArgs...), buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Exec replaces the current process with a kubectl invocation
func Exec(flags *genericclioptions.ConfigFlags, args []string) error {
	kArgs := getKubectlConfigFlags(flags)
	kArgs = append(kArgs, args...)
	return execCommand(append([]string{"kubectl"}, kArgs...))
}

// Replaces the currently running process with the given command
func execCommand(args []string) error {
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	args[0] = path

	env := os.Environ()
	return syscall.Exec(path, args, env)
}

// Runs a command and returns stdout
func execToWriter(args []string, writer io.Writer) error {
	cmd := exec.Command(args[0], args[1:]...)

	op, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go io.Copy(writer, op)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// getKubectlConfigFlags serializes the parsed flag struct back into a series of command line args
// that can then be passed to kubectl. The mirror image of
// https://github.com/kubernetes/cli-runtime/blob/master/pkg/genericclioptions/config_flags.go#L251
func getKubectlConfigFlags(flags *genericclioptions.ConfigFlags) []string {
	out := []string{}
	o := &out

	appendStringFlag(o, flags.KubeConfig, "kubeconfig")
	appendStringFlag(o, flags.CacheDir, "cache-dir")
	appendStringFlag(o, flags.CertFile, "client-certificate")
	appendStringFlag(o, flags.KeyFile, "client-key")
	appendStringFlag(o, flags.BearerToken, "token")
	appendStringFlag(o, flags.Impersonate, "as")
	appendStringArrayFlag(o, flags.ImpersonateGroup, "as-group")
	appendStringFlag(o, flags.Username, "username")
	appendStringFlag(o, flags.Password, "password")
	appendStringFlag(o, flags.ClusterName, "cluster")
	appendStringFlag(o, flags.AuthInfoName, "user")
	//appendStringFlag(o, flags.Namespace, "namespace")
	appendStringFlag(o, flags.Context, "context")
	appendStringFlag(o, flags.APIServer, "server")
	appendBoolFlag(o, flags.Insecure, "insecure-skip-tls-verify")
	appendStringFlag(o, flags.CAFile, "certificate-authority")
	appendStringFlag(o, flags.Timeout, "request-timeout")

	return out
}

func appendStringFlag(out *[]string, in *string, flag string) {
	if in != nil && *in != "" {
		*out = append(*out, fmt.Sprintf("--%v=%v", flag, *in))
	}
}

func appendBoolFlag(out *[]string, in *bool, flag string) {
	if in != nil {
		*out = append(*out, fmt.Sprintf("--%v=%v", flag, *in))
	}
}

func appendStringArrayFlag(out *[]string, in *[]string, flag string) {
	if in != nil && len(*in) > 0 {
		*out = append(*out, fmt.Sprintf("--%v=%v'", flag, strings.Join(*in, ",")))
	}
}
