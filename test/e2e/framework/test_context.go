/*
Copyright 2016 The Kubernetes Authors.

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
	"flag"
	"os"

	"github.com/onsi/ginkgo/config"

	"k8s.io/client-go/tools/clientcmd"
)

// TestContextType describes the client context to use in communications with the Kubernetes API.
type TestContextType struct {
	KubeHost    string
	KubeConfig  string
	KubeContext string
}

// TestContext is the global client context for tests.
var TestContext TestContextType

// RegisterCommonFlags registers flags common to all e2e test suites.
func RegisterCommonFlags() {
	// Turn on verbose by default to get spec names
	config.DefaultReporterConfig.Verbose = true

	// Turn on EmitSpecProgress to get spec progress (especially on interrupt)
	config.GinkgoConfig.EmitSpecProgress = true

	// Randomize specs as well as suites
	config.GinkgoConfig.RandomizeAllSpecs = true

	// Default SlowSpecThreshold is 5 seconds.
	// Too low for the kind of operations we need to tests
	config.DefaultReporterConfig.SlowSpecThreshold = 20

	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "http://127.0.0.1:8080", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kubernetes. If unset, will use value from 'current-context'")
}

// RegisterParseFlags registers and parses flags for the test binary.
func RegisterParseFlags() {
	RegisterCommonFlags()
	flag.Parse()
}
