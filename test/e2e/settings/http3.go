/*
Copyright 2025 The Kubernetes Authors.

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
	"context"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("http3", func() {
	f := framework.NewDefaultFramework("http3")
	host := "http3.com"

	ginkgo.BeforeEach(func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--enable-quic")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")
		annotations := map[string]string{}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, "listen 443 default_server reuseport quic;")
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "listen 443 quic;")
			})
	})

	ginkgo.It("should enable HTTP/3 on a custom port", func() {
		err := f.UpdateIngressControllerDeployment(func(deployment *appsv1.Deployment) error {
			args := deployment.Spec.Template.Spec.Containers[0].Args
			args = append(args, "--quic-port=4321")
			deployment.Spec.Template.Spec.Containers[0].Args = args
			_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller deployment flags")

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, "listen 4321 default_server reuseport quic;")
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "listen 4321 quic;")
			})
	})

	ginkgo.It("should have default http3_hq value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_hq off;")
		})
	})

	ginkgo.It("should set http3_hq value", func() {
		f.UpdateNginxConfigMapData("http3-hq", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_hq on;")
		})
	})

	ginkgo.It("should have default http3_max_concurrent_streams value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_max_concurrent_streams 128;")
		})
	})

	ginkgo.It("should set http3_max_concurrent_streams value", func() {
		f.UpdateNginxConfigMapData("http3-max-concurrent-streams", "256")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_max_concurrent_streams 256;")
		})
	})

	ginkgo.It("should have default http3_stream_buffer_size value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_stream_buffer_size 64k;")
		})
	})

	ginkgo.It("should set http3_stream_buffer_size value", func() {
		f.UpdateNginxConfigMapData("http3-stream-buffer-size", "128k")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "http3_stream_buffer_size 128k;")
		})
	})

	ginkgo.It("should have default quic_active_connection_id_limit value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_active_connection_id_limit 2;")
		})
	})

	ginkgo.It("should set quic_active_connection_id_limit value", func() {
		f.UpdateNginxConfigMapData("quic-active-connection-id-limit", "16")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_active_connection_id_limit 16;")
		})
	})

	ginkgo.It("should have default quic_bpf value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_bpf off;")
		})
	})

	ginkgo.It("should set quic_bpf value", func() {
		f.UpdateNginxConfigMapData("quic-bpf", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_bpf on;")
		})
	})

	ginkgo.It("should have default quic_gso value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_gso off;")
		})
	})

	ginkgo.It("should set quic_gso value", func() {
		f.UpdateNginxConfigMapData("quic-gso", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_gso on;")
		})
	})

	ginkgo.It("should have default quic_retry value", func() {
		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_retry off;")
		})
	})

	ginkgo.It("should set quic_retry value", func() {
		f.UpdateNginxConfigMapData("quic-retry", "true")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "quic_retry on;")
		})
	})
})
