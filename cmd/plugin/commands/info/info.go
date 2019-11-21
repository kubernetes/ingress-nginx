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

package info

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
)

// CreateCommand creates and returns this cobra subcommand
func CreateCommand(flags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the ingress-nginx service",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := cmd.Flags().GetString("service")
			if err != nil {
				return err
			}

			util.PrintError(info(flags, service))
			return nil
		},
	}

	cmd.Flags().String("service", util.DefaultIngressServiceName, "The name of the ingress-nginx service")
	return cmd
}

func info(flags *genericclioptions.ConfigFlags, serviceName string) error {
	service, err := request.GetServiceByName(flags, serviceName, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Service cluster IP address: %v\n", service.Spec.ClusterIP)
	fmt.Printf("LoadBalancer IP|CNAME: %v\n", service.Spec.LoadBalancerIP)
	return nil
}
