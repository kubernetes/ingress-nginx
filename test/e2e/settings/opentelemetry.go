/*
Copyright 2020 The Kubernetes Authors.

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

package settings

import (
	"strings"

	"github.com/onsi/ginkgo"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	enableOpenTelemetry = "enable-opentelemetry"

	openTelemetryOperationName = "opentelemetry-operation-name"
)

var _ = framework.IngressNginxDescribe("Configure OpenTelemetry", func() {
	f := framework.NewDefaultFramework("enable-opentelemetry")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(
			framework.WithDeploymentModule("opentelemetry", "ingress-nginx/opentelemetry"),
		)
	})

	ginkgo.AfterEach(func() {
	})

	ginkgo.It("should not exists opentelemetry directive", func() {
		config := map[string]string{}
		config[enableOpenTelemetry] = "false"
		f.SetNginxConfigMapData(config)

		f.EnsureIngress(framework.NewSingleIngress(enableOpenTelemetry, "/", enableOpenTelemetry, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "opentelemetry on")
			})
	})

	ginkgo.It("should exists opentelemetry directive when is enabled", func() {
		config := map[string]string{}
		config[enableOpenTelemetry] = "true"
		f.SetNginxConfigMapData(config)

		f.EnsureIngress(framework.NewSingleIngress(enableOpenTelemetry, "/", enableOpenTelemetry, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "opentelemetry on")
			})
	})

	ginkgo.It("should not exists opentelemetry_operation_name directive when is empty", func() {
		config := map[string]string{}
		config[enableOpenTelemetry] = "true"
		config[openTelemetryOperationName] = ""
		f.SetNginxConfigMapData(config)

		f.EnsureIngress(framework.NewSingleIngress(enableOpenTelemetry, "/", enableOpenTelemetry, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "opentelemetry_operation_name")
			})
	})

	ginkgo.It("should exists opentelemetry_operation_name directive when is configured", func() {
		config := map[string]string{}
		config[enableOpenTelemetry] = "true"
		config[openTelemetryOperationName] = "HTTP $request_method $uri"
		f.SetNginxConfigMapData(config)

		f.EnsureIngress(framework.NewSingleIngress(enableOpenTelemetry, "/", enableOpenTelemetry, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `opentelemetry_operation_name "HTTP $request_method $uri"`)
			})
	})
})
