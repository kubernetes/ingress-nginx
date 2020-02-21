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

package logs

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
	o := logsFlags{}
	var pod, deployment, selector *string

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get the kubernetes logs for an ingress-nginx pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			util.PrintError(logs(flags, *pod, *deployment, *selector, o))
			return nil
		},
	}
	pod = util.AddPodFlag(cmd)
	deployment = util.AddDeploymentFlag(cmd)
	selector = util.AddSelectorFlag(cmd)

	cmd.Flags().BoolVarP(&o.Follow, "follow", "f", o.Follow, "Specify if the logs should be streamed.")
	cmd.Flags().BoolVar(&o.Timestamps, "timestamps", o.Timestamps, "Include timestamps on each line in the log output")
	cmd.Flags().Int64Var(&o.LimitBytes, "limit-bytes", o.LimitBytes, "Maximum bytes of logs to return. Defaults to no limit.")
	cmd.Flags().BoolVarP(&o.Previous, "previous", "p", o.Previous, "If true, print the logs for the previous instance of the container in a pod if it exists.")
	cmd.Flags().Int64Var(&o.Tail, "tail", o.Tail, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	cmd.Flags().StringVar(&o.SinceTime, "since-time", o.SinceTime, "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	cmd.Flags().StringVar(&o.SinceSeconds, "since", o.SinceSeconds, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")

	return cmd
}

type logsFlags struct {
	SinceTime    string
	SinceSeconds string
	Follow       bool
	Previous     bool
	Timestamps   bool
	LimitBytes   int64
	Tail         int64
	Selector     string
}

func (o *logsFlags) toStrings() []string {
	r := []string{}
	if o.SinceTime != "" {
		r = append(r, "--since-time", o.SinceTime)
	}
	if o.SinceSeconds != "" {
		r = append(r, "--since", o.SinceSeconds)
	}
	if o.Follow {
		r = append(r, "--follow")
	}
	if o.Previous {
		r = append(r, "--previous")
	}
	if o.Timestamps {
		r = append(r, "--timestamps")
	}
	if o.LimitBytes != 0 {
		r = append(r, "--limit-bytes", fmt.Sprintf("%v", o.LimitBytes))
	}
	if o.Tail != 0 {
		r = append(r, "--tail", fmt.Sprintf("%v", o.Tail))
	}

	return r
}

func logs(flags *genericclioptions.ConfigFlags, podName string, deployment string, selector string, opts logsFlags) error {
	pod, err := request.ChoosePod(flags, podName, deployment, selector)
	if err != nil {
		return err
	}

	cmd := []string{"logs", "-n", pod.Namespace, pod.Name}
	cmd = append(cmd, opts.toStrings()...)
	return kubectl.Exec(flags, cmd)
}
