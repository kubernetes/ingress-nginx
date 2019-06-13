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

package lint

import (
	"fmt"

	"github.com/spf13/cobra"

	appsv1 "k8s.io/api/apps/v1"
	networking "k8s.io/api/networking/v1beta1"
	kmeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/ingress-nginx/cmd/plugin/lints"
	"k8s.io/ingress-nginx/cmd/plugin/request"
	"k8s.io/ingress-nginx/cmd/plugin/util"
	"k8s.io/ingress-nginx/version"
)

// CreateCommand creates and returns this cobra subcommand
func CreateCommand(flags *genericclioptions.ConfigFlags) *cobra.Command {
	var opts *lintOptions
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Inspect kubernetes resources for possible issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.Validate()
			if err != nil {
				return err
			}

			fmt.Println("Checking ingresses...")
			err = ingresses(*opts)
			if err != nil {
				util.PrintError(err)
			}
			fmt.Println("Checking deployments...")
			err = deployments(*opts)
			if err != nil {
				util.PrintError(err)
			}

			return nil
		},
	}

	opts = addCommonOptions(flags, cmd)

	cmd.AddCommand(createSubcommand(flags, []string{"ingresses", "ingress", "ing"}, "Check ingresses for possible issues", ingresses))
	cmd.AddCommand(createSubcommand(flags, []string{"deployments", "deployment", "dep"}, "Check deployments for possible issues", deployments))

	return cmd
}

func createSubcommand(flags *genericclioptions.ConfigFlags, names []string, short string, f func(opts lintOptions) error) *cobra.Command {
	var opts *lintOptions
	cmd := &cobra.Command{
		Use:     names[0],
		Aliases: names[1:],
		Short:   short,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.Validate()
			if err != nil {
				return err
			}
			util.PrintError(f(*opts))
			return nil
		},
	}

	opts = addCommonOptions(flags, cmd)

	return cmd
}

func addCommonOptions(flags *genericclioptions.ConfigFlags, cmd *cobra.Command) *lintOptions {
	out := lintOptions{
		flags: flags,
	}
	cmd.Flags().BoolVar(&out.allNamespaces, "all-namespaces", false, "Check resources in all namespaces")
	cmd.Flags().BoolVar(&out.showAll, "show-all", false, "Show all resources, not just the ones with problems")
	cmd.Flags().BoolVarP(&out.verbose, "verbose", "v", false, "Show extra information about the lints")
	cmd.Flags().StringVarP(&out.versionFrom, "from-version", "f", "0.0.0", "Use lints added for versions starting with this one")
	cmd.Flags().StringVarP(&out.versionTo, "to-version", "t", version.RELEASE, "Use lints added for versions up to and including this one")

	return &out
}

type lintOptions struct {
	flags         *genericclioptions.ConfigFlags
	allNamespaces bool
	showAll       bool
	verbose       bool
	versionFrom   string
	versionTo     string
}

func (opts *lintOptions) Validate() error {
	_, _, _, err := util.ParseVersionString(opts.versionFrom)
	if err != nil {
		return err
	}

	_, _, _, err = util.ParseVersionString(opts.versionTo)
	if err != nil {
		return err
	}

	return nil
}

type lint interface {
	Check(obj kmeta.Object) bool
	Message() string
	Link() string
	Version() string
}

func checkObjectArray(lints []lint, objects []kmeta.Object, opts lintOptions) {
	usedLints := make([]lint, 0)
	for _, lint := range lints {
		lintVersion := lint.Version()
		if lint.Version() == "" {
			lintVersion = "0.0.0"
		}
		if util.InVersionRangeInclusive(opts.versionFrom, lintVersion, opts.versionTo) {
			usedLints = append(usedLints, lint)
		}
	}

	for _, obj := range objects {
		objName := obj.GetName()
		if opts.allNamespaces {
			objName = obj.GetNamespace() + "/" + obj.GetName()
		}

		failedLints := make([]lint, 0)
		for _, lint := range usedLints {
			if lint.Check(obj) {
				failedLints = append(failedLints, lint)
			}
		}

		if len(failedLints) != 0 {
			fmt.Printf("✗ %v\n", objName)
			for _, lint := range failedLints {
				fmt.Printf("  - %v\n", lint.Message())
				if opts.verbose && lint.Version() != "" {
					fmt.Printf("      Lint added for version %v\n", lint.Version())
				}
				if opts.verbose && lint.Link() != "" {
					fmt.Printf("      %v\n", lint.Link())
				}
			}
			fmt.Println("")
			continue
		}

		if opts.showAll {
			fmt.Printf("✓ %v\n", objName)
		}
	}
}

func ingresses(opts lintOptions) error {
	var ings []networking.Ingress
	var err error
	if opts.allNamespaces {
		ings, err = request.GetIngressDefinitions(opts.flags, "")
	} else {
		ings, err = request.GetIngressDefinitions(opts.flags, util.GetNamespace(opts.flags))
	}
	if err != nil {
		return err
	}

	var iLints []lints.IngressLint = lints.GetIngressLints()
	genericLints := make([]lint, len(iLints))
	for i := range iLints {
		genericLints[i] = iLints[i]
	}

	objects := make([]kmeta.Object, 0)
	for i := range ings {
		objects = append(objects, &ings[i])
	}

	checkObjectArray(genericLints, objects, opts)
	return nil
}

func deployments(opts lintOptions) error {
	var deps []appsv1.Deployment
	var err error
	if opts.allNamespaces {
		deps, err = request.GetDeployments(opts.flags, "")
	} else {
		deps, err = request.GetDeployments(opts.flags, util.GetNamespace(opts.flags))
	}
	if err != nil {
		return err
	}

	var iLints []lints.DeploymentLint = lints.GetDeploymentLints()
	genericLints := make([]lint, len(iLints))
	for i := range iLints {
		genericLints[i] = iLints[i]
	}

	objects := make([]kmeta.Object, 0)
	for i := range deps {
		objects = append(objects, &deps[i])
	}

	checkObjectArray(genericLints, objects, opts)
	return nil
}
