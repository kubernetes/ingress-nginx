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
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	certutil "k8s.io/client-go/util/cert"

	"k8s.io/ingress-nginx/internal/file"
)

// generateRSACerts generates a self signed certificate using a self generated ca
func generateRSACerts(host string) (*keyPair, *keyPair, error) {
	ca, err := newCA("self-sign-ca")
	if err != nil {
		return nil, nil, err
	}

	key, err := certutil.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create a server private key: %v", err)
	}

	config := certutil.Config{
		CommonName: host,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	cert, err := certutil.NewSignedCert(config, key, ca.Cert, ca.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to sign the server certificate: %v", err)
	}

	return &keyPair{
		Key:  key,
		Cert: cert,
	}, ca, nil
}

func TestAddOrUpdateCertAndKey(t *testing.T) {
	fs := newFS(t)

	cert, _, err := generateRSACerts("echoheaders")
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	name := fmt.Sprintf("test-%v", time.Now().UnixNano())

	c := certutil.EncodeCertPEM(cert.Cert)
	k := certutil.EncodePrivateKeyPEM(cert.Key)

	ngxCert, err := AddOrUpdateCertAndKey(name, c, k, []byte{}, fs)
	if err != nil {
		t.Fatalf("unexpected error checking SSL certificate: %v", err)
	}

	if ngxCert.PemFileName == "" {
		t.Fatalf("expected path to pem file but returned empty")
	}

	if len(ngxCert.CN) == 0 {
		t.Fatalf("expected at least one cname but none returned")
	}

	if ngxCert.CN[0] != "echoheaders" {
		t.Fatalf("expected cname echoheaders but %v returned", ngxCert.CN[0])
	}
}

func TestCACert(t *testing.T) {
	fs := newFS(t)

	cert, CA, err := generateRSACerts("echoheaders")
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}

	name := fmt.Sprintf("test-%v", time.Now().UnixNano())

	c := certutil.EncodeCertPEM(cert.Cert)
	k := certutil.EncodePrivateKeyPEM(cert.Key)
	ca := certutil.EncodeCertPEM(CA.Cert)

	ngxCert, err := AddOrUpdateCertAndKey(name, c, k, ca, fs)
	if err != nil {
		t.Fatalf("unexpected error checking SSL certificate: %v", err)
	}
	if ngxCert.CAFileName == "" {
		t.Fatalf("expected a valid CA file name")
	}
}

func TestGetFakeSSLCert(t *testing.T) {
	k, c := GetFakeSSLCert()
	if len(k) == 0 {
		t.Fatalf("expected a valid key")
	}
	if len(c) == 0 {
		t.Fatalf("expected a valid certificate")
	}
}

func TestAddCertAuth(t *testing.T) {
	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("unexpected error creating filesystem: %v", err)
	}

	cn := "demo-ca"
	_, ca, err := generateRSACerts(cn)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}
	c := certutil.EncodeCertPEM(ca.Cert)
	ic, err := AddCertAuth(cn, c, fs)
	if err != nil {
		t.Fatalf("unexpected error creating SSL certificate: %v", err)
	}
	if ic.CAFileName == "" {
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

	name := fmt.Sprintf("test-%v", time.Now().UnixNano())

	c := certutil.EncodeCertPEM(cert.Cert)
	k := certutil.EncodePrivateKeyPEM(cert.Key)

	ngxCert, err := CreateSSLCert(name, c, k, []byte{})
	if err != nil {
		t.Fatalf("unexpected error checking SSL certificate: %v", err)
	}

	var certKeyBuf bytes.Buffer
	certKeyBuf.Write(c)
	certKeyBuf.Write([]byte("\n"))
	certKeyBuf.Write(k)

	if ngxCert.PemCertKey != certKeyBuf.String() {
		t.Fatalf("expected concatenated PEM cert and key but returned %v", ngxCert.PemCertKey)
	}

	if len(ngxCert.CN) == 0 {
		t.Fatalf("expected at least one cname but none returned")
	}

	if ngxCert.CN[0] != "echoheaders" {
		t.Fatalf("expected cname echoheaders but %v returned", ngxCert.CN[0])
	}
}

type keyPair struct {
	Key  *rsa.PrivateKey
	Cert *x509.Certificate
}

func newCA(name string) (*keyPair, error) {
	key, err := certutil.NewPrivateKey()
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
