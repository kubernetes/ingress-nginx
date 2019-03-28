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

// Package commands holds the CLI glue mapping textual commands/args to method calls.
package commands

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/commands/edit"
	"sigs.k8s.io/kustomize/pkg/commands/misc"
	"sigs.k8s.io/kustomize/pkg/factory"
	"sigs.k8s.io/kustomize/pkg/fs"
)

// NewDefaultCommand returns the default (aka root) command for kustomize command.
func NewDefaultCommand(f *factory.KustFactory) *cobra.Command {
	fsys := fs.MakeRealFS()
	stdOut := os.Stdout

	c := &cobra.Command{
		Use:   "kustomize",
		Short: "kustomize manages declarative configuration of Kubernetes",
		Long: `
kustomize manages declarative configuration of Kubernetes.

See https://sigs.k8s.io/kustomize
`,
	}

	c.AddCommand(
		// TODO: Make consistent API for newCmd* functions.
		build.NewCmdBuild(stdOut, fsys, f.ResmapF, f.TransformerF),
		edit.NewCmdEdit(fsys, f.ValidatorF, f.UnstructF),
		misc.NewCmdConfig(fsys),
		misc.NewCmdVersion(stdOut),
	)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})
	return c
}
