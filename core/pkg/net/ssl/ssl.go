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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress/core/pkg/file"
	"k8s.io/ingress/core/pkg/ingress"
)

var (
	oidExtensionSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
)

// AddOrUpdateCertAndKey creates a .pem file wth the cert and the key with the specified name
func AddOrUpdateCertAndKey(name string, cert, key, ca []byte) (*ingress.SSLCert, error) {
	pemName := fmt.Sprintf("%v.pem", name)
	pemFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, pemName)

	tempPemFile, err := ioutil.TempFile(ingress.DefaultSSLDirectory, pemName)

	if err != nil {
		return nil, fmt.Errorf("could not create temp pem file %v: %v", pemFileName, err)
	}
	glog.V(3).Infof("Creating temp file %v for Keypair: %v", tempPemFile.Name(), pemName)

	_, err = tempPemFile.Write(cert)
	if err != nil {
		return nil, fmt.Errorf("could not write to pem file %v: %v", tempPemFile.Name(), err)
	}
	_, err = tempPemFile.Write([]byte("\n"))
	if err != nil {
		return nil, fmt.Errorf("could not write to pem file %v: %v", tempPemFile.Name(), err)
	}
	_, err = tempPemFile.Write(key)
	if err != nil {
		return nil, fmt.Errorf("could not write to pem file %v: %v", tempPemFile.Name(), err)
	}

	err = tempPemFile.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close temp pem file %v: %v", tempPemFile.Name(), err)
	}

	pemCerts, err := ioutil.ReadFile(tempPemFile.Name())
	if err != nil {
		_ = os.Remove(tempPemFile.Name())
		return nil, err
	}

	pemBlock, _ := pem.Decode(pemCerts)
	if pemBlock == nil {
		_ = os.Remove(tempPemFile.Name())
		return nil, fmt.Errorf("no valid PEM formatted block found")
	}

	// If the file does not start with 'BEGIN CERTIFICATE' it's invalid and must not be used.
	if pemBlock.Type != "CERTIFICATE" {
		_ = os.Remove(tempPemFile.Name())
		return nil, fmt.Errorf("certificate %v contains invalid data, and must be created with 'kubectl create secret tls'", name)
	}

	pemCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		_ = os.Remove(tempPemFile.Name())
		return nil, err
	}

	//Ensure that certificate and private key have a matching public key
	if _, err := tls.X509KeyPair(cert, key); err != nil {
		_ = os.Remove(tempPemFile.Name())
		return nil, err
	}

	cn := sets.NewString(pemCert.Subject.CommonName)
	for _, dns := range pemCert.DNSNames {
		if !cn.Has(dns) {
			cn.Insert(dns)
		}
	}

	if len(pemCert.Extensions) > 0 {
		glog.V(3).Info("parsing ssl certificate extensions")
		for _, ext := range getExtension(pemCert, oidExtensionSubjectAltName) {
			dns, _, _, err := parseSANExtension(ext.Value)
			if err != nil {
				glog.Warningf("unexpected error parsing certificate extensions: %v", err)
				continue
			}

			for _, dns := range dns {
				if !cn.Has(dns) {
					cn.Insert(dns)
				}
			}
		}
	}

	err = os.Rename(tempPemFile.Name(), pemFileName)
	if err != nil {
		return nil, fmt.Errorf("could not move temp pem file %v to destination %v: %v", tempPemFile.Name(), pemFileName, err)
	}

	if len(ca) > 0 {
		bundle := x509.NewCertPool()
		bundle.AppendCertsFromPEM(ca)
		opts := x509.VerifyOptions{
			Roots: bundle,
		}

		_, err := pemCert.Verify(opts)
		if err != nil {
			oe := fmt.Sprintf("failed to verify certificate chain: \n\t%s\n", err)
			return nil, errors.New(oe)
		}

		caFile, err := os.OpenFile(pemFileName, os.O_RDWR|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("could not open file %v for writing additional CA chains: %v", pemFileName, err)
		}

		defer caFile.Close()
		_, err = caFile.Write([]byte("\n"))
		if err != nil {
			return nil, fmt.Errorf("could not append CA to cert file %v: %v", pemFileName, err)
		}
		caFile.Write(ca)
		caFile.Write([]byte("\n"))

		return &ingress.SSLCert{
			CAFileName:  pemFileName,
			PemFileName: pemFileName,
			PemSHA:      file.SHA1(pemFileName),
			CN:          cn.List(),
			ExpireTime:  pemCert.NotAfter,
		}, nil
	}

	return &ingress.SSLCert{
		PemFileName: pemFileName,
		PemSHA:      file.SHA1(pemFileName),
		CN:          cn.List(),
		ExpireTime:  pemCert.NotAfter,
	}, nil
}

func getExtension(c *x509.Certificate, id asn1.ObjectIdentifier) []pkix.Extension {
	var exts []pkix.Extension
	for _, ext := range c.Extensions {
		if ext.Id.Equal(id) {
			exts = append(exts, ext)
		}
	}
	return exts
}

func parseSANExtension(value []byte) (dnsNames, emailAddresses []string, ipAddresses []net.IP, err error) {
	// RFC 5280, 4.2.1.6

	// SubjectAltName ::= GeneralNames
	//
	// GeneralNames ::= SEQUENCE SIZE (1..MAX) OF GeneralName
	//
	// GeneralName ::= CHOICE {
	//      otherName                       [0]     OtherName,
	//      rfc822Name                      [1]     IA5String,
	//      dNSName                         [2]     IA5String,
	//      x400Address                     [3]     ORAddress,
	//      directoryName                   [4]     Name,
	//      ediPartyName                    [5]     EDIPartyName,
	//      uniformResourceIdentifier       [6]     IA5String,
	//      iPAddress                       [7]     OCTET STRING,
	//      registeredID                    [8]     OBJECT IDENTIFIER }
	var seq asn1.RawValue
	var rest []byte
	if rest, err = asn1.Unmarshal(value, &seq); err != nil {
		return
	} else if len(rest) != 0 {
		err = errors.New("x509: trailing data after X.509 extension")
		return
	}
	if !seq.IsCompound || seq.Tag != 16 || seq.Class != 0 {
		err = asn1.StructuralError{Msg: "bad SAN sequence"}
		return
	}

	rest = seq.Bytes
	for len(rest) > 0 {
		var v asn1.RawValue
		rest, err = asn1.Unmarshal(rest, &v)
		if err != nil {
			return
		}
		switch v.Tag {
		case 1:
			emailAddresses = append(emailAddresses, string(v.Bytes))
		case 2:
			dnsNames = append(dnsNames, string(v.Bytes))
		case 7:
			switch len(v.Bytes) {
			case net.IPv4len, net.IPv6len:
				ipAddresses = append(ipAddresses, v.Bytes)
			default:
				err = errors.New("x509: certificate contained IP address of length " + strconv.Itoa(len(v.Bytes)))
				return
			}
		}
	}

	return
}

// AddCertAuth creates a .pem file with the specified CAs to be used in Cert Authentication
// If it's already exists, it's clobbered.
func AddCertAuth(name string, ca []byte) (*ingress.SSLCert, error) {

	caName := fmt.Sprintf("ca-%v.pem", name)
	caFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, caName)

	pemCABlock, _ := pem.Decode(ca)
	if pemCABlock == nil {
		return nil, fmt.Errorf("no valid PEM formatted block found")
	}
	// If the first certificate does not start with 'BEGIN CERTIFICATE' it's invalid and must not be used.
	if pemCABlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("CA file %v contains invalid data, and must be created only with PEM formated certificates", name)
	}

	_, err := x509.ParseCertificate(pemCABlock.Bytes)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(caFileName, ca, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write CA file %v: %v", caFileName, err)
	}

	glog.V(3).Infof("Created CA Certificate for authentication: %v", caFileName)
	return &ingress.SSLCert{
		CAFileName:  caFileName,
		PemFileName: caFileName,
		PemSHA:      file.SHA1(caFileName),
	}, nil
}

// AddOrUpdateDHParam creates a dh parameters file with the specified name
func AddOrUpdateDHParam(name string, dh []byte) (string, error) {
	pemName := fmt.Sprintf("%v.pem", name)
	pemFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, pemName)

	tempPemFile, err := ioutil.TempFile(ingress.DefaultSSLDirectory, pemName)

	glog.V(3).Infof("Creating temp file %v for DH param: %v", tempPemFile.Name(), pemName)
	if err != nil {
		return "", fmt.Errorf("could not create temp pem file %v: %v", pemFileName, err)
	}

	_, err = tempPemFile.Write(dh)
	if err != nil {
		return "", fmt.Errorf("could not write to pem file %v: %v", tempPemFile.Name(), err)
	}

	err = tempPemFile.Close()
	if err != nil {
		return "", fmt.Errorf("could not close temp pem file %v: %v", tempPemFile.Name(), err)
	}

	pemCerts, err := ioutil.ReadFile(tempPemFile.Name())
	if err != nil {
		_ = os.Remove(tempPemFile.Name())
		return "", err
	}

	pemBlock, _ := pem.Decode(pemCerts)
	if pemBlock == nil {
		_ = os.Remove(tempPemFile.Name())
		return "", fmt.Errorf("no valid PEM formatted block found")
	}

	// If the file does not start with 'BEGIN DH PARAMETERS' it's invalid and must not be used.
	if pemBlock.Type != "DH PARAMETERS" {
		_ = os.Remove(tempPemFile.Name())
		return "", fmt.Errorf("certificate %v contains invalid data", name)
	}

	err = os.Rename(tempPemFile.Name(), pemFileName)
	if err != nil {
		return "", fmt.Errorf("could not move temp pem file %v to destination %v: %v", tempPemFile.Name(), pemFileName, err)
	}

	return pemFileName, nil
}

// GetFakeSSLCert creates a Self Signed Certificate
// Based in the code https://golang.org/src/crypto/tls/generate_cert.go
func GetFakeSSLCert() ([]byte, []byte) {

	var priv interface{}
	var err error

	priv, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		glog.Fatalf("failed to generate fake private key: %s", err)
	}

	notBefore := time.Now()
	// This certificate is valid for 365 days
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		glog.Fatalf("failed to generate fake serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
			CommonName:   "Kubernetes Ingress Controller Fake Certificate",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"ingress.local"},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.(*rsa.PrivateKey).PublicKey, priv)
	if err != nil {
		glog.Fatalf("Failed to create fake certificate: %s", err)
	}

	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv.(*rsa.PrivateKey))})

	return cert, key
}
