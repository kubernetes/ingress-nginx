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

package ocsp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ocsp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("OCSP", func() {
	f := framework.NewDefaultFramework("ocsp")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should enable OCSP and contain stapling information in the connection", func() {
		host := "www.example.com"

		f.UpdateNginxConfigMapData("enable-ocsp", "true")

		err := prepareCertificates(f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		leafCert, err := os.ReadFile("leaf.pem")
		assert.Nil(ginkgo.GinkgoT(), err)

		leafKey, err := os.ReadFile("leaf-key.pem")
		assert.Nil(ginkgo.GinkgoT(), err)

		intermediateCa, err := os.ReadFile("intermediate_ca.pem")
		assert.Nil(ginkgo.GinkgoT(), err)

		var pemCertBuffer bytes.Buffer
		pemCertBuffer.Write(leafCert)
		pemCertBuffer.Write([]byte("\n"))
		pemCertBuffer.Write(intermediateCa)

		f.EnsureSecret(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ing.Spec.TLS[0].SecretName,
				Namespace: f.Namespace,
			},
			Data: map[string][]byte{
				corev1.TLSCertKey:       pemCertBuffer.Bytes(),
				corev1.TLSPrivateKeyKey: leafKey,
			},
		})

		cfsslDB, err := os.ReadFile("empty.db")
		assert.Nil(ginkgo.GinkgoT(), err)

		f.EnsureConfigMap(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ocspserve",
				Namespace: f.Namespace,
			},
			BinaryData: map[string][]byte{
				"empty.db":       cfsslDB,
				"db-config.json": []byte(`{"driver":"sqlite3","data_source":"/data/empty.db"}`),
			},
		})

		d, s := ocspserveDeployment(f.Namespace)
		f.EnsureDeployment(d)
		f.EnsureService(s)

		err = framework.WaitForEndpoints(f.KubeClientSet, framework.DefaultTimeout, "ocspserve", f.Namespace, 1)
		assert.Nil(ginkgo.GinkgoT(), err, "waiting for endpoints to become ready")

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, "certificate.is_ocsp_stapling_enabled = true")
		})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name %v`, host))
			})

		tlsConfig := &tls.Config{ServerName: host, InsecureSkipVerify: true}
		f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Raw()

		// give time the lua request to the OCSP
		// URL to finish and update the cache
		framework.Sleep()

		// TODO: is possible to avoid second request?
		resp := f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Raw()

		state := resp.TLS
		assert.NotNil(ginkgo.GinkgoT(), state.OCSPResponse, "unexpected connection without OCSP response")

		var issuerCertificate *x509.Certificate
		var leafAuthorityKeyID string
		for index, certificate := range state.PeerCertificates {
			if index == 0 {
				leafAuthorityKeyID = string(certificate.AuthorityKeyId)
				continue
			}

			if leafAuthorityKeyID == string(certificate.SubjectKeyId) {
				issuerCertificate = certificate
			}
		}

		response, err := ocsp.ParseResponse(state.OCSPResponse, issuerCertificate)
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Equal(ginkgo.GinkgoT(), ocsp.Good, response.Status)
	})
})

const configTemplate = `
{
	"signing": {
		"default": {
			"ocsp_url": "http://ocspserve.%v.svc.cluster.local",
			"expiry": "219000h",
			"usages": [
				"signing",
				"key encipherment",
				"client auth"
			]
		},
		"profiles": {
			"ocsp": {
				"usages": ["digital signature", "ocsp signing"],
				"expiry": "8760h"
			},
			"intermediate": {
				"usages": ["cert sign", "crl sign"],
				"expiry": "219000h",
				"ca_constraint": {
					"is_ca": true
				}
			},
			"server": {
				"usages": ["signing", "key encipherment", "server auth"],
				"expiry": "8760h"
			},
			"client": {
				"usages": ["signing", "key encipherment", "client auth"],
				"expiry": "8760h"
			}
		}
	}
}
`

func prepareCertificates(namespace string) error {
	config := fmt.Sprintf(configTemplate, namespace)
	err := os.WriteFile("cfssl_config.json", []byte(config), 0644)
	if err != nil {
		return fmt.Errorf("creating cfssl_config.json file: %v", err)
	}

	cpCmd := exec.Command("cp", "-rf", "template.db", "empty.db")
	err = cpCmd.Run()
	if err != nil {
		return fmt.Errorf("copying sqlite file: %v", err)
	}

	commands := []string{
		"cfssl gencert -initca ca_csr.json | cfssljson -bare ca",
		"cfssl gencert -ca ca.pem -ca-key ca-key.pem -config=cfssl_config.json -profile=intermediate intermediate_ca_csr.json | cfssljson -bare intermediate_ca",
		"cfssl gencert -ca intermediate_ca.pem -ca-key intermediate_ca-key.pem -config=cfssl_config.json -profile=ocsp ocsp_csr.json | cfssljson -bare ocsp",
	}

	for _, command := range commands {
		ginkgo.By(fmt.Sprintf("running %v", command))
		out, err := exec.Command("bash", "-c", command).CombinedOutput()
		if err != nil {
			framework.Logf("Command error: %v\n%v\n%v", command, err, string(out))
			return err
		}
	}

	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	command := "cfssl serve -db-config=db-config.json -ca-key=intermediate_ca-key.pem -ca=intermediate_ca.pem -config=cfssl_config.json -responder=ocsp.pem -responder-key=ocsp-key.pem"
	ginkgo.By(fmt.Sprintf("running %v", command))
	serve := exec.CommandContext(ctx, "bash", "-c", command)
	if err := serve.Start(); err != nil {
		framework.Logf("Command start error: %v\n%v", command, err)
		return err
	}

	framework.Sleep()

	command = "cfssl gencert -remote=localhost -profile=server leaf_csr.json | cfssljson -bare leaf"
	ginkgo.By(fmt.Sprintf("running %v", command))
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		framework.Logf("Command error: %v\n%v\n%v", command, err, string(out))
		return err
	}

	err = serve.Process.Signal(syscall.SIGTERM)
	if err != nil {
		framework.Logf("Command error: %v", err)
		return err
	}

	command = "cfssl ocsprefresh -ca intermediate_ca.pem -responder=ocsp.pem -responder-key=ocsp-key.pem -db-config=db-config.json"
	ginkgo.By(fmt.Sprintf("running %v", command))
	out, err = exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		framework.Logf("Command error: %v\n%v\n%v", command, err, string(out))
		return err
	}

	/*
		Example:
		cfssl ocspserve -port=8080 -db-config=db-config.json
		openssl ocsp -issuer intermediate_ca.pem -no_nonce -cert leaf.pem -CAfile ca.pem -text -url http://localhost:8080
	*/

	return nil
}

func ocspserveDeployment(namespace string) (*appsv1.Deployment, *corev1.Service) {
	name := "ocspserve"
	return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: framework.NewInt32(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": name,
						},
					},
					Spec: corev1.PodSpec{
						TerminationGracePeriodSeconds: framework.NewInt64(0),
						Containers: []corev1.Container{
							{
								Name:  name,
								Image: "registry.k8s.io/ingress-nginx/e2e-test-cfssl@sha256:d02c1e18f573449966999fc850f1fed3d37621bf77797562cbe77ebdb06a66ea",
								Command: []string{
									"/bin/bash",
									"-c",
									"cfssl ocspserve -port=80 -address=0.0.0.0 -db-config=/data/db-config.json -loglevel=0",
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										ContainerPort: 80,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      name,
										MountPath: "/data",
										ReadOnly:  true,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: name,
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: name,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Port:       80,
						TargetPort: intstr.FromInt(80),
						Protocol:   corev1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"app": name,
				},
			},
		}
}
