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
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"

	"k8s.io/ingress/core/pkg/ingress"
)

// AddOrUpdateCertAndKey creates a .pem file wth the cert and the key with the specified name
func AddOrUpdateCertAndKey(name string, cert, key, ca []byte) (*ingress.SSLCert, error) {
	pemName := fmt.Sprintf("%v.pem", name)
	pemFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, pemName)

	tempPemFile, err := ioutil.TempFile(ingress.DefaultSSLDirectory, pemName)

	glog.V(3).Infof("Creating temp file %v for Keypair: %v", tempPemFile.Name(), pemName)
	if err != nil {
		return nil, fmt.Errorf("could not create temp pem file %v: %v", pemFileName, err)
	}

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
		return nil, err
	}

	pemBlock, _ := pem.Decode(pemCerts)
	if pemBlock == nil {
		return nil, fmt.Errorf("No valid PEM formatted block found")
	}

	pemCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	cn := []string{pemCert.Subject.CommonName}
	if len(pemCert.DNSNames) > 0 {
		cn = append(cn, pemCert.DNSNames...)
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
			return nil, fmt.Errorf("Could not open file %v for writing additional CA chains: %v", pemFileName, err)
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
			PemSHA:      pemSHA1(pemFileName),
			CN:          cn,
		}, nil
	}

	return &ingress.SSLCert{
		PemFileName: pemFileName,
		PemSHA:      pemSHA1(pemFileName),
		CN:          cn,
	}, nil
}

// AddCertAuth creates a .pem file with the specified CAs to be used in Cert Authentication
// If it's already exists, it's clobbered.
func AddCertAuth(name string, ca []byte) (*ingress.SSLCert, error) {

	caName := fmt.Sprintf("ca-%v.pem", name)
	caFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, caName)

	pemCABlock, _ := pem.Decode(ca)
	if pemCABlock == nil {
		return nil, fmt.Errorf("No valid PEM formatted block found")
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
		PemSHA:      pemSHA1(caFileName),
	}, nil
}

// SearchDHParamFile iterates all the secrets mounted inside the /etc/nginx-ssl directory
// in order to find a file with the name dhparam.pem. If such file exists it will
// returns the path. If not it just returns an empty string
func SearchDHParamFile(baseDir string) string {
	files, _ := ioutil.ReadDir(baseDir)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dhPath := fmt.Sprintf("%v/%v/dhparam.pem", baseDir, file.Name())
		if _, err := os.Stat(dhPath); err == nil {
			glog.Infof("using file '%v' for parameter ssl_dhparam", dhPath)
			return dhPath
		}
	}

	glog.Warning("no file dhparam.pem found in secrets")
	return ""
}

// pemSHA1 returns the SHA1 of a pem file. This is used to
// reload NGINX in case a secret with a SSL certificate changed.
func pemSHA1(filename string) string {
	hasher := sha1.New()
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil))
}

const (
	snakeOilPem = "/etc/ssl/certs/ssl-cert-snakeoil.pem"
	snakeOilKey = "/etc/ssl/private/ssl-cert-snakeoil.key"
)

// GetFakeSSLCert returns the snake oil ssl certificate created by the command
// make-ssl-cert generate-default-snakeoil --force-overwrite
func GetFakeSSLCert() ([]byte, []byte) {
	cert, err := ioutil.ReadFile(snakeOilPem)
	if err != nil {
		return nil, nil
	}

	key, err := ioutil.ReadFile(snakeOilKey)
	if err != nil {
		return nil, nil
	}

	return cert, key
}
