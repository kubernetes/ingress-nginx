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

package util

import (
	"fmt"
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// The default deployment and service names for ingress-nginx
const (
	DefaultIngressDeploymentName = "nginx-ingress-controller"
	DefaultIngressServiceName    = "ingress-nginx"
)

// PrintError receives an error value and prints it if it exists
func PrintError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func printWithError(s string, e error) {
	if e != nil {
		fmt.Println(e)
	}
	fmt.Print(s)
}

func printOrError(s string, e error) error {
	if e != nil {
		return e
	}
	fmt.Print(s)
	return nil
}

// AddPodFlag adds a --pod flag to a cobra command
func AddPodFlag(cmd *cobra.Command) *string {
	v := ""
	cmd.Flags().StringVar(&v, "pod", "", "Query a particular ingress-nginx pod")
	return &v
}

// AddDeploymentFlag adds a --deployment flag to a cobra command
func AddDeploymentFlag(cmd *cobra.Command) *string {
	v := ""
	cmd.Flags().StringVar(&v, "deployment", DefaultIngressDeploymentName, "The name of the ingress-nginx deployment")
	return &v
}

// GetNamespace takes a set of kubectl flag values and returns the namespace we should be operating in
func GetNamespace(flags *genericclioptions.ConfigFlags) string {
	namespace, _, err := flags.ToRawKubeConfigLoader().Namespace()
	if err != nil || len(namespace) == 0 {
		namespace = apiv1.NamespaceDefault
	}
	return namespace
}
