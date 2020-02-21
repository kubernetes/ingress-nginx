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

package exec

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/ingress-nginx/cmd/plugin/kubectl"
	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// CreateCommand creates and returns this cobra subcommand
func CreateCommand(flags *genericclioptions.ConfigFlags) *cobra.Command {
	opts := execFlags{}
	var pod, deployment, selector *string

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a command inside an ingress-nginx pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintError(exec(flags, *pod, *deployment, *selector, args, opts))
			return nil
		},
	}
	pod = util.AddPodFlag(cmd)
	deployment = util.AddDeploymentFlag(cmd)
	selector = util.AddSelectorFlag(cmd)
	cmd.Flags().BoolVarP(&opts.TTY, "tty", "t", false, "Stdin is a TTY")
	cmd.Flags().BoolVarP(&opts.Stdin, "stdin", "i", false, "Pass stdin to the container")

	return cmd
}

type execFlags struct {
	TTY   bool
	Stdin bool
}

func exec(flags *genericclioptions.ConfigFlags, podName string, deployment string, selector string, cmd []string, opts execFlags) error {
	pod, err := request.ChoosePod(flags, podName, deployment, selector)
	if err != nil {
		return err
	}

	args := []string{"exec"}
	if opts.TTY {
		args = append(args, "-t")
	}
	if opts.Stdin {
		args = append(args, "-i")
	}

	args = append(args, []string{"-n", pod.Namespace, pod.Name, "--"}...)
	args = append(args, cmd...)
	return kubectl.Exec(flags, args)
}
