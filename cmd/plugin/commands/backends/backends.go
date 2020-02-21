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

package backends

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/ingress-nginx/cmd/plugin/kubectl"
	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// CreateCommand creates and returns this cobra subcommand
func CreateCommand(flags *genericclioptions.ConfigFlags) *cobra.Command {
	var pod, deployment, selector *string
	cmd := &cobra.Command{
		Use:   "backends",
		Short: "Inspect the dynamic backend information of an ingress-nginx instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			backend, err := cmd.Flags().GetString("backend")
			if err != nil {
				return err
			}
			onlyList, err := cmd.Flags().GetBool("list")
			if err != nil {
				return err
			}
			if onlyList && backend != "" {
				return fmt.Errorf("--list and --backend cannot both be specified")
			}

			util.PrintError(backends(flags, *pod, *deployment, *selector, backend, onlyList))
			return nil
		},
	}

	pod = util.AddPodFlag(cmd)
	deployment = util.AddDeploymentFlag(cmd)
	selector = util.AddSelectorFlag(cmd)

	cmd.Flags().String("backend", "", "Output only the information for the given backend")
	cmd.Flags().Bool("list", false, "Output a newline-separated list of backend names")

	return cmd
}

func backends(flags *genericclioptions.ConfigFlags, podName string, deployment string, selector string, backend string, onlyList bool) error {
	var command []string
	if onlyList {
		command = []string{"/dbg", "backends", "list"}
	} else if backend != "" {
		command = []string{"/dbg", "backends", "get", backend}
	} else {
		command = []string{"/dbg", "backends", "all"}
	}

	pod, err := request.ChoosePod(flags, podName, deployment, selector)
	if err != nil {
		return err
	}

	out, err := kubectl.PodExecString(flags, &pod, command)
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}
