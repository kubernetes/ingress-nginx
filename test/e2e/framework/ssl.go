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

package framework

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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// CreateIngressTLSSecret creates a secret containing TLS certificates for the given Ingress.
// If a secret with the same name already pathExists in the namespace of the
// Ingress, it's updated.
func CreateIngressTLSSecret(client kubernetes.Interface, hosts []string, secretName, namespace string) (host string, rootCA, privKey []byte, err error) {
	var k, c bytes.Buffer
	host = strings.Join(hosts, ",")
	if err = generateRSACerts(host, true, &k, &c); err != nil {
		return
	}
	cert := c.Bytes()
	key := k.Bytes()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			v1.TLSCertKey:       cert,
			v1.TLSPrivateKeyKey: key,
		},
	}
	var s *v1.Secret
	if s, err = client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{}); err == nil {
		s.Data = secret.Data
		_, err = client.CoreV1().Secrets(namespace).Update(s)
	} else {
		_, err = client.CoreV1().Secrets(namespace).Create(secret)
	}
	return host, cert, key, err
}

// generateRSACerts generates a basic self signed certificate using a key length
// of rsaBits, valid for validFor time.
func generateRSACerts(host string, isCA bool, keyOut, certOut io.Writer) error {
	if len(host) == 0 {
		return fmt.Errorf("require a non-empty host for client hello")
	}
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
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
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("failed creating keay: %v", err)
	}
	return nil
}
