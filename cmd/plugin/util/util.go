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
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// The default deployment and service names for ingress-nginx
const (
	DefaultIngressDeploymentName = "ingress-nginx-controller"
	DefaultIngressServiceName    = "ingress-nginx-controller"
)

// IssuePrefix is the github url that we can append an issue number to to link to it
const IssuePrefix = "https://github.com/kubernetes/ingress-nginx/issues/"

var versionRegex = regexp.MustCompile(`(\d)+\.(\d)+\.(\d)+.*`)

// PrintError receives an error value and prints it if it exists
func PrintError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

// ParseVersionString returns the major, minor, and patch numbers of a version string
func ParseVersionString(v string) (int, int, int, error) {
	parts := versionRegex.FindStringSubmatch(v)

	if len(parts) != 4 {
		return 0, 0, 0, fmt.Errorf("could not parse %v as a version string (like 0.20.3)", v)
	}

	major, _ := strconv.Atoi(parts[1])
	minor, _ := strconv.Atoi(parts[2])
	patch, _ := strconv.Atoi(parts[3])

	return major, minor, patch, nil
}

// InVersionRangeInclusive checks that the middle version is between the other two versions
func InVersionRangeInclusive(start, v, stop string) bool {
	return !isVersionLessThan(v, start) && !isVersionLessThan(stop, v)
}

func isVersionLessThan(a, b string) bool {
	aMajor, aMinor, aPatch, err := ParseVersionString(a)
	if err != nil {
		panic(err)
	}

	bMajor, bMinor, bPatch, err := ParseVersionString(b)
	if err != nil {
		panic(err)
	}

	if aMajor != bMajor {
		return aMajor < bMajor
	}

	if aMinor != bMinor {
		return aMinor < bMinor
	}

	return aPatch < bPatch
}

// PodInDeployment returns whether a pod is part of a deployment with the given name
// a pod is considered to be in {deployment} if it is owned by a replicaset with a name of format {deployment}-otherchars
func PodInDeployment(pod apiv1.Pod, deployment string) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Controller == nil || !*owner.Controller || owner.Kind != "ReplicaSet" {
			continue
		}

		if strings.Count(owner.Name, "-") != strings.Count(deployment, "-")+1 {
			continue
		}

		if strings.HasPrefix(owner.Name, deployment+"-") {
			return true
		}
	}
	return false
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

// AddSelectorFlag adds a --selector flag to a cobra command
func AddSelectorFlag(cmd *cobra.Command) *string {
	v := ""
	cmd.Flags().StringVarP(&v, "selector", "l", "", "Selector (label query) of the ingress-nginx pod")
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
