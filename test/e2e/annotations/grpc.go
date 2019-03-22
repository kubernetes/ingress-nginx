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

	pb "github.com/moul/pb/grpcbin/go-grpc"
	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("backend-protocol - GRPC", func() {
	f := framework.NewDefaultFramework("grpc")

	ginkgo.It("should use grpc_pass in the configuration file", func() {
		f.NewGRPCFortuneTellerDeployment()

		host := "grpc"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "fortune-teller", 50051, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %v", host))
			})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass") &&
					strings.Contains(server, "grpc_set_header") &&
					!strings.Contains(server, "proxy_pass")
			})
	})

	ginkgo.It("should return OK for service with backend protocol GRPC", func() {
		f.NewGRPCBinDeployment()

		host := "echo"

		svc := &core.Service{
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
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"].Values[0], "application/grpc")
	})

	ginkgo.It("authorization metadata should be overwritten by external auth response headers", func() {
		f.NewGRPCBinDeployment()
		f.NewHttpbinDeployment()

		host := "echo"

		svc := &core.Service{
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

		err := framework.WaitForEndpoints(f.KubeClientSet, framework.DefaultTimeout, framework.HTTPBinService, f.Namespace, 1)
		assert.Nil(ginkgo.GinkgoT(), err)

		e, err := f.KubeClientSet.CoreV1().Endpoints(f.Namespace).Get(context.TODO(), framework.HTTPBinService, metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err)

		assert.GreaterOrEqual(ginkgo.GinkgoT(), len(e.Subsets), 1, "expected at least one endpoint")
		assert.GreaterOrEqual(ginkgo.GinkgoT(), len(e.Subsets[0].Addresses), 1, "expected at least one address ready in the endpoint")

		httpbinIP := e.Subsets[0].Addresses[0].IP

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-url":              fmt.Sprintf("http://%s/response-headers?authorization=foo", httpbinIP),
			"nginx.ingress.kubernetes.io/auth-response-headers": "Authorization",
			"nginx.ingress.kubernetes.io/backend-protocol":      "GRPC",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9000, annotations)

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
		ctx := metadata.AppendToOutgoingContext(context.Background(),
			"authorization", "bar")

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), "foo", metadata["authorization"].Values[0])
	})

	ginkgo.It("should return OK for service with backend protocol GRPCS", func() {
		f.NewGRPCBinDeployment()

		host := "echo"

		svc := &core.Service{
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
						Port:       9001,
						TargetPort: intstr.FromInt(9001),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPCS",
			"nginx.ingress.kubernetes.io/configuration-snippet": fmt.Sprintf(`
			   # without this setting NGINX sends echo instead
			   grpc_ssl_name      		grpcbin.%v.svc.cluster.local;
			   grpc_ssl_server_name		on;
			   grpc_ssl_ciphers 		HIGH:!aNULL:!MD5;
			`, f.Namespace),
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9001, annotations)
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
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"].Values[0], "application/grpc")
	})
})
