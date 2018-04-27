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
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	net_url "net/url"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// CreateIngressTLSSecret creates or updates a Secret containing a TLS
// certificate for the given Ingress and returns a TLS configuration suitable
// for HTTP clients to use against that particular Ingress.
func CreateIngressTLSSecret(client kubernetes.Interface, hosts []string, secretName, namespace string) (*tls.Config, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("require a non-empty host for client hello")
	}

	var k, c bytes.Buffer
	host := strings.Join(hosts, ",")
	if err := generateRSACert(host, true, &k, &c); err != nil {
		return nil, err
	}

	cert, key := c.Bytes(), k.Bytes()
	newSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			v1.TLSCertKey:       cert,
			v1.TLSPrivateKeyKey: key,
		},
	}

	var apierr error
	curSecret, err := client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err == nil && curSecret != nil {
		curSecret.Data = newSecret.Data
		_, apierr = client.CoreV1().Secrets(namespace).Update(curSecret)
	} else {
		_, apierr = client.CoreV1().Secrets(namespace).Create(newSecret)
	}
	if apierr != nil {
		return nil, apierr
	}

	serverName := hosts[0]
	return tlsConfig(serverName, cert)
}

// WaitForTLS waits until the TLS handshake with a given server completes successfully.
func WaitForTLS(url string, tlsConfig *tls.Config) error {
	return wait.Poll(Poll, 30*time.Second, matchTLSServerName(url, tlsConfig))
}

// generateRSACert generates a basic self signed certificate using a key length
// of rsaBits, valid for validFor time.
func generateRSACert(host string, isCA bool, keyOut, certOut io.Writer) error {
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

// tlsConfig returns a client TLS configuration for the given server name and
// CA certificate (PEM).
func tlsConfig(serverName string, pemCA []byte) (*tls.Config, error) {
	rootCAPool := x509.NewCertPool()
	if !rootCAPool.AppendCertsFromPEM(pemCA) {
		return nil, fmt.Errorf("error creating CA certificate pool (%s)", serverName)
	}
	return &tls.Config{
		ServerName: serverName,
		RootCAs:    rootCAPool,
	}, nil
}

// matchTLSServerName connects to the network address corresponding to the
// given URL using the given TLS configuration and returns whether the TLS
// handshake completed successfully.
func matchTLSServerName(url string, tlsConfig *tls.Config) wait.ConditionFunc {
	return func() (ready bool, err error) {
		u, err := net_url.Parse(url)
		if err != nil {
			return
		}

		conn, err := tls.Dial("tcp", u.Host, tlsConfig)
		if err != nil {
			return false, nil
		}
		conn.Close()

		ready = true
		return
	}
}
