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

	pembBock, _ := pem.Decode(pemCerts)
	if pembBock == nil {
		return nil, fmt.Errorf("No valid PEM formatted block found")
	}

	pemCert, err := x509.ParseCertificate(pembBock.Bytes)
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

		caName := fmt.Sprintf("ca-%v.pem", name)
		caFileName := fmt.Sprintf("%v/%v", ingress.DefaultSSLDirectory, caName)
		f, err := os.Create(caFileName)
		if err != nil {
			return nil, fmt.Errorf("could not create ca pem file %v: %v", caFileName, err)
		}
		defer f.Close()
		_, err = f.Write(ca)
		if err != nil {
			return nil, fmt.Errorf("could not create ca pem file %v: %v", caFileName, err)
		}
		f.Write([]byte("\n"))

		return &ingress.SSLCert{
			CAFileName:  caFileName,
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
