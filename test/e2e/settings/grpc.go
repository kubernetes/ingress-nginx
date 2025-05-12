/*
Copyright 2024 The Kubernetes Authors.

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
	"crypto/tls"
	"fmt"
	"strings"

	pb "github.com/moul/pb/grpcbin/go-grpc"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const echoHost = "echo"

var _ = framework.DescribeSetting("GRPC", func() {
	f := framework.NewDefaultFramework("grpc-buffersize", framework.WithHTTPBunEnabled())

	ginkgo.It("should set the correct GRPC Buffer Size", func() {
		f.SetNginxConfigMapData(map[string]string{
			"grpc-buffer-size-kb": "8",
		})

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, "grpc_buffer_size 8k")
			})

		f.NewGRPCBinDeployment()

		host := echoHost

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin-test",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: fmt.Sprintf("grpcbin.%v.svc.cluster.local", f.Namespace),
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       9000,
						TargetPort: intstr.FromInt(9000),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9000, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpc://upstream_balancer;")
			})

		conn, err := grpc.NewClient(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         echoHost,
					InsecureSkipVerify: true, //nolint:gosec // Ignore certificate validation in testing
				}),
			),
		)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating a connection")
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := context.Background()

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"].Values[0], "application/grpc")
		assert.Equal(ginkgo.GinkgoT(), metadata[":authority"].Values[0], host)
	})
})
