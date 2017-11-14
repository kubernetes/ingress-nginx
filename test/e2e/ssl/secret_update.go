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

package ssl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

var _ = framework.IngressNginxDescribe("SSL", func() {
	f := framework.NewDefaultFramework("ssl")

	BeforeEach(func() {
		err := f.NewEchoDeployment()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should not appear references to secret updates not used in ingress rules", func() {
		host := "ssl-update"

		dummySecret, err := f.EnsureSecret(&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: f.Namespace.Name,
			},
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		ing, err := f.EnsureIngress(&v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host,
				Namespace: f.Namespace.Name,
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						Hosts:      []string{host},
						SecretName: host,
					},
				},
				Rules: []v1beta1.IngressRule{
					{
						Host: host,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/",
										Backend: v1beta1.IngressBackend{
											ServiceName: "http-svc",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(ing).ToNot(BeNil())

		_, _, _, err = createIngressTLSSecret(f.KubeClientSet, ing)
		Expect(err).ToNot(HaveOccurred())

		err = f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "server_name ssl-update") &&
					strings.Contains(server, "listen 443")
			})
		Expect(err).ToNot(HaveOccurred())

		log, err := f.NginxLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(log).ToNot(BeEmpty())

		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace.Name)))
		time.Sleep(30 * time.Second)
		dummySecret.Data["some-key"] = []byte("some value")
		f.KubeClientSet.CoreV1().Secrets(f.Namespace.Name).Update(dummySecret)
		time.Sleep(30 * time.Second)
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("starting syncing of secret %v/dummy", f.Namespace.Name)))
		Expect(log).ToNot(ContainSubstring(fmt.Sprintf("error obtaining PEM from secret %v/dummy", f.Namespace.Name)))
	})
})

// createIngressTLSSecret creates a secret containing TLS certificates for the given Ingress.
// If a secret with the same name already pathExists in the namespace of the
// Ingress, it's updated.
func createIngressTLSSecret(kubeClient kubernetes.Interface, ing *v1beta1.Ingress) (host string, rootCA, privKey []byte, err error) {
	var k, c bytes.Buffer
	tls := ing.Spec.TLS[0]
	host = strings.Join(tls.Hosts, ",")
	if err = generateRSACerts(host, true, &k, &c); err != nil {
		return
	}
	cert := c.Bytes()
	key := k.Bytes()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: tls.SecretName,
		},
		Data: map[string][]byte{
			v1.TLSCertKey:       cert,
			v1.TLSPrivateKeyKey: key,
		},
	}
	var s *v1.Secret
	if s, err = kubeClient.CoreV1().Secrets(ing.Namespace).Get(tls.SecretName, metav1.GetOptions{}); err == nil {
		s.Data = secret.Data
		_, err = kubeClient.CoreV1().Secrets(ing.Namespace).Update(s)
	} else {
		_, err = kubeClient.CoreV1().Secrets(ing.Namespace).Create(secret)
	}
	return host, cert, key, err
}

// generateRSACerts generates a basic self signed certificate using a key length
// of rsaBits, valid for validFor time.
func generateRSACerts(host string, isCA bool, keyOut, certOut io.Writer) error {
	if len(host) == 0 {
		return fmt.Errorf("Require a non-empty host for client hello")
	}
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("Failed to generate key: %v", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "default",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("Failed to create certificate: %s", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("Failed creating cert: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("Failed creating keay: %v", err)
	}
	return nil
}
