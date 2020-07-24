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

var _ = framework.IngressNginxDescribe("Configure OpenTracing", func() {
	f := framework.NewDefaultFramework("enable-opentracing")
	enableOpentracing := "enable-opentracing"
	zipkinCollectorHost := "zipkin-collector-host"
	opentracingOperationName := "opentracing-operation-name"
	opentracingLocationOperationName := "opentracing-location-operation-name"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.AfterEach(func() {
	})

	ginkgo.It("should not exists opentracing directive", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "false")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "opentracing on")
			})
	})

	ginkgo.It("should exists opentracing directive when is enabled", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "true")
		f.UpdateNginxConfigMapData(zipkinCollectorHost, "127.0.0.1")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "opentracing on")
			})
	})

	ginkgo.It("should not exists opentracing_operation_name directive when is empty", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "true")
		f.UpdateNginxConfigMapData(zipkinCollectorHost, "127.0.0.1")
		f.UpdateNginxConfigMapData(opentracingOperationName, "")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "opentracing_operation_name")
			})
	})

	ginkgo.It("should exists opentracing_operation_name directive when is configured", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "true")
		f.UpdateNginxConfigMapData(zipkinCollectorHost, "127.0.0.1")
		f.UpdateNginxConfigMapData(opentracingOperationName, "HTTP $request_method $uri")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "opentracing_operation_name \"HTTP $request_method $uri\"")
			})
	})

	ginkgo.It("should not exists opentracing_location_operation_name directive when is empty", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "true")
		f.UpdateNginxConfigMapData(zipkinCollectorHost, "127.0.0.1")
		f.UpdateNginxConfigMapData(opentracingLocationOperationName, "")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return !strings.Contains(cfg, "opentracing_location_operation_name")
			})
	})

	ginkgo.It("should exists opentracing_location_operation_name directive when is configured", func() {
		f.UpdateNginxConfigMapData(enableOpentracing, "true")
		f.UpdateNginxConfigMapData(zipkinCollectorHost, "127.0.0.1")
		f.UpdateNginxConfigMapData(opentracingLocationOperationName, "HTTP $request_method $uri")

		f.EnsureIngress(framework.NewSingleIngress(enableOpentracing, "/", enableOpentracing, f.Namespace, "http-svc", 80, nil))

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "opentracing_location_operation_name \"HTTP $request_method $uri\"")
			})
	})
})
