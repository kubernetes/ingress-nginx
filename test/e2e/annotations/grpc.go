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

package annotations

import (
	"crypto/tls"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/moul/pb/grpcbin/go-grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Annotations - GRPC", func() {
	f := framework.NewDefaultFramework("grpc")

	It("should use grpc_pass in the configuration file", func() {
		f.NewGRPCFortuneTellerDeployment()

		host := "grpc"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "fortune-teller", 50051, &annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring(fmt.Sprintf("server_name %v", host)))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return Expect(server).Should(ContainSubstring("grpc_pass")) &&
					Expect(server).Should(ContainSubstring("grpc_set_header")) &&
					Expect(server).ShouldNot(ContainSubstring("proxy_pass"))
			})
	})

	It("should return OK for service with backend protocol GRPC", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "grpcb.in",
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

		annotations := &map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin", 9000, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpc://upstream_balancer;")
			})

		conn, _ := grpc.Dial(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         "echo",
					InsecureSkipVerify: true,
				}),
			),
		)
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := context.Background()

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		Expect(err).Should(BeNil())

		metadata := res.GetMetadata()
		Expect(metadata["x-original-uri"].Values[0]).Should(Equal("/grpcbin.GRPCBin/HeadersUnary"))
		Expect(metadata["content-type"].Values[0]).Should(Equal("application/grpc"))
	})

	It("should return OK for service with backend protocol GRPCS", func() {
		host := "echo"

		svc := &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: "grpcb.in",
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       9001,
						TargetPort: intstr.FromInt(9001),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := &map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPCS",
			"nginx.ingress.kubernetes.io/configuration-snippet": `
			   # without this setting NGINX sends echo instead
			   grpc_ssl_name      		grpcb.in;
			   grpc_ssl_server_name		on;
			   grpc_ssl_ciphers 		HIGH:!aNULL:!MD5;
			`,
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin", 9001, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpcs://upstream_balancer;")
			})

		conn, _ := grpc.Dial(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         "echo",
					InsecureSkipVerify: true,
				}),
			),
		)
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := context.Background()

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		Expect(err).Should(BeNil())

		metadata := res.GetMetadata()
		Expect(metadata["x-original-uri"].Values[0]).Should(Equal("/grpcbin.GRPCBin/HeadersUnary"))
		Expect(metadata["content-type"].Values[0]).Should(Equal("application/grpc"))
	})
})
