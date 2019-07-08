/*
Copyright 2015 The Kubernetes Authors.

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
	"crypto"
	"crypto/rand"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/pkg/util/filesystem"

	"k8s.io/ingress-nginx/internal/file"
)

// generateRSACerts generates a self signed certificate using a self generated ca
func generateRSACerts(host string) (*keyPair, *keyPair, error) {
	ca, err := newCA("self-sign-ca")
	if err != nil {
		return nil, nil, err
	}

	key, err := newPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create a server private key: %v", err)
	}

	config := certutil.Config{
		CommonName: host,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	cert, err := newSignedCert(config, key, ca.Cert, ca.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign the server certificate: %v", err)
	}

	return &keyPair{
		Key:  key,
		Cert: cert,
	}, ca, nil
}

func TestStoreSSLCertOnDisk(t *testing.T) {
	fs := newFS(t)

	cert, _, err := generateRSACerts("echoheaders")
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	name := fmt.Sprintf("test-%v", time.Now().UnixNano())

	c := encodeCertPEM(cert.Cert)
	k := encodePrivateKeyPEM(cert.Key)

	sslCert, err := CreateSSLCert(c, k)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	err = StoreSSLCertOnDisk(fs, name, sslCert)
	if err != nil {
		t.Fatalf("unexpected error storing SSL certificate: %v", err)
	}

	if sslCert.PemFileName == "" {
		t.Fatalf("expected path to pem file but returned empty")
	}

	if len(sslCert.CN) == 0 {
		t.Fatalf("expected at least one cname but none returned")
	}

	if sslCert.CN[0] != "echoheaders" {
		t.Fatalf("expected cname echoheaders but %v returned", sslCert.CN[0])
	}
}

func TestCACert(t *testing.T) {
	fs := newFS(t)

	cert, CA, err := generateRSACerts("echoheaders")
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	name := fmt.Sprintf("test-%v", time.Now().UnixNano())

	c := encodeCertPEM(cert.Cert)
	k := encodePrivateKeyPEM(cert.Key)
	ca := encodeCertPEM(CA.Cert)

	sslCert, err := CreateSSLCert(c, k)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	err = StoreSSLCertOnDisk(fs, name, sslCert)
	if err != nil {
		t.Fatalf("unexpected error storing SSL certificate: %v", err)
	}

	if sslCert.CAFileName != "" {
		t.Fatalf("expected CA file name to be empty")
	}

	err = ConfigureCACertWithCertAndKey(fs, name, ca, sslCert)
	if err != nil {
		t.Fatalf("unexpected error configuring CA certificate: %v", err)
	}

	if sslCert.CAFileName == "" {
		t.Fatalf("expected a valid CA file name")
	}
}

func TestGetFakeSSLCert(t *testing.T) {
	fs := newFS(t)

	sslCert := GetFakeSSLCert(fs)

	if len(sslCert.PemCertKey) == 0 {
		t.Fatalf("expected PemCertKey to not be empty")
	}

	if len(sslCert.PemFileName) == 0 {
		t.Fatalf("expected PemFileName to not be empty")
	}

	if len(sslCert.CN) != 2 {
		t.Fatalf("expected 2 entries in CN, but got %v", len(sslCert.CN))
	}

	if sslCert.CN[0] != "Kubernetes Ingress Controller Fake Certificate" {
		t.Fatalf("expected common name to be \"Kubernetes Ingress Controller Fake Certificate\" but got %v", sslCert.CN[0])
	}

	if sslCert.CN[1] != "ingress.local" {
		t.Fatalf("expected a DNS name \"ingress.local\" but got: %v", sslCert.CN[1])
	}
}

func TestConfigureCACert(t *testing.T) {
	fs := newFS(t)

	cn := "demo-ca"
	_, ca, err := generateRSACerts(cn)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}
	c := encodeCertPEM(ca.Cert)

	sslCert, err := CreateCACert(c)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}
	if sslCert.CAFileName != "" {
		t.Fatalf("expected CAFileName to be empty")
	}
	if sslCert.Certificate == nil {
		t.Fatalf("expected Certificate to be set")
	}

	err = ConfigureCACert(fs, cn, c, sslCert)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}
	if sslCert.CAFileName == "" {
		t.Fatalf("expected a valid CA file name")
	}
}

func newFS(t *testing.T) file.Filesystem {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error creating filesystem: %v", err)
	}
	return fs
}

func TestCreateSSLCert(t *testing.T) {
	cert, _, err := generateRSACerts("echoheaders")
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	c := encodeCertPEM(cert.Cert)
	k := encodePrivateKeyPEM(cert.Key)

	sslCert, err := CreateSSLCert(c, k)
	if err != nil {
		t.Fatalf("unexpected error checking SSL certificate: %v", err)
	}

	var certKeyBuf bytes.Buffer
	certKeyBuf.Write(c)
	certKeyBuf.Write([]byte("\n"))
	certKeyBuf.Write(k)

	if sslCert.PemCertKey != certKeyBuf.String() {
		t.Fatalf("expected concatenated PEM cert and key but returned %v", sslCert.PemCertKey)
	}

	if len(sslCert.CN) == 0 {
		t.Fatalf("expected at least one CN but none returned")
	}

	if sslCert.CN[0] != "echoheaders" {
		t.Fatalf("expected cname echoheaders but %v returned", sslCert.CN[0])
	}
}

type keyPair struct {
	Key  *rsa.PrivateKey
	Cert *x509.Certificate
}

func newCA(name string) (*keyPair, error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a private key for a new CA: %v", err)
	}
	config := certutil.Config{
		CommonName: name,
	}
	cert, err := certutil.NewSelfSignedCACert(config, key)
	if err != nil {
		return nil, fmt.Errorf("unable to create a self-signed certificate for a new CA: %v", err)
	}
	return &keyPair{
		Key:  key,
		Cert: cert,
	}, nil
}

func TestIsValidHostname(t *testing.T) {
	cases := map[string]struct {
		Hostname string
		CN       []string
		Valid    bool
	}{
		"when there is no common names": {
			"foo.bar",
			[]string{},
			false,
		},
		"when there is a match for foo.bar": {
			"foo.bar",
			[]string{"foo.bar"},
			true,
		},
		"when there is a wildcard match for foo.bar": {
			"foo.bar",
			[]string{"*.bar"},
			true,
		},
		"when there is a wrong wildcard for *.bar": {
			"invalid.foo.bar",
			[]string{"*.bar"},
			false,
		},
	}

	for k, tc := range cases {
		valid := IsValidHostname(tc.Hostname, tc.CN)
		if valid != tc.Valid {
			t.Errorf("%s: expected '%v' but returned %v", k, tc.Valid, valid)
		}
	}
}

const (
	duration365d = time.Hour * 24 * 365
	rsaKeySize   = 2048
)

// newPrivateKey creates an RSA private key
func newPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
}

// newSignedCert creates a signed certificate using the given CA certificate and key
func newSignedCert(cfg certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// encodePublicKeyPEM returns PEM-encoded public data
func encodePublicKeyPEM(key *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return []byte{}, err
	}
	block := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(&block), nil
}

// encodePrivateKeyPEM returns PEM-encoded private key data
func encodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

// encodeCertPEM returns PEM-endcoded certificate data
func encodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  certutil.CertificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

func fakeCertificate(t *testing.T, fs filesystem.Filesystem) []byte {
	cert, key := getFakeHostSSLCert("localhost")
	fd, err := fs.Create("/key.crt")
	if err != nil {
		t.Errorf("failed to write test key: %v", err)
	}
	fd.Write(cert)
	fd, err = fs.Create("/key.key")
	if err != nil {
		t.Errorf("failed to write test key: %v", err)
	}
	fd.Write(key)
	return cert
}

func dialTestServer(port string, rootCertificates ...[]byte) error {
	roots := x509.NewCertPool()
	for _, cert := range rootCertificates {
		ok := roots.AppendCertsFromPEM(cert)
		if !ok {
			return fmt.Errorf("failed to add root certificate")
		}
	}
	resp, err := tls.Dial("tcp", "localhost:"+port, &tls.Config{
		RootCAs: roots,
	})
	if err != nil {
		return err
	}
	if resp.Handshake() != nil {
		return fmt.Errorf("TLS handshake should succeed: %v", err)
	}
	return nil
}

func TestTLSKeyReloader(t *testing.T) {
	fs := filesystem.NewFakeFs()
	cert := fakeCertificate(t, fs)

	watcher := TLSListener{
		certificatePath: "/key.crt",
		keyPath:         "/key.key",
		fs:              fs,
		lock:            sync.Mutex{},
	}
	watcher.load()

	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	s.Config.TLSConfig = watcher.TLSConfig()
	s.Listener = tls.NewListener(s.Listener, s.Config.TLSConfig)
	go s.Start()
	defer s.Close()
	port := strings.Split(s.Listener.Addr().String(), ":")[1]

	t.Run("without the trusted certificate", func(t *testing.T) {
		if dialTestServer(port) == nil {
			t.Errorf("TLS dial should fail")
		}
	})

	t.Run("with the certificate trustes as root certificate", func(t *testing.T) {
		if err := dialTestServer(port, cert); err != nil {
			t.Errorf("TLS dial should succeed, got error: %v", err)
		}
	})

	t.Run("with a new certificate", func(t *testing.T) {
		newCert := fakeCertificate(t, fs)
		t.Run("when the certificate is not reloaded", func(t *testing.T) {
			if dialTestServer(port, newCert) == nil {
				t.Errorf("TLS dial should fail")
			}
		})
		// simulate watch.NewFileWatcher to call the load function
		watcher.load()
		t.Run("when the certificate is reloaded", func(t *testing.T) {
			if err := dialTestServer(port, newCert); err != nil {
				t.Errorf("TLS dial should succeed, got error: %v", err)
			}
		})
	})

}
