/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package nginx

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"

	"k8s.io/contrib/ingress/controllers/nginx/nginx/config"
)

// SSLCert describes a SSL certificate to be used in NGINX
type SSLCert struct {
	CertFileName string
	KeyFileName  string
	// PemFileName contains the path to the file with the certificate and key concatenated
	PemFileName string
	// PemSHA contains the sha1 of the pem file.
	// This is used to detect changes in the secret that contains the certificates
	PemSHA string
	// CN contains all the common names defined in the SSL certificate
	CN []string
}

// AddOrUpdateCertAndKey creates a .pem file wth the cert and the key with the specified name
func (nginx *Manager) AddOrUpdateCertAndKey(name string, cert string, key string) (SSLCert, error) {
	temporaryPemFileName := fmt.Sprintf("%v.pem", name)
	pemFileName := fmt.Sprintf("%v/%v.pem", config.SSLDirectory, name)

	temporaryPemFile, err := ioutil.TempFile("", temporaryPemFileName)
	if err != nil {
		return SSLCert{}, fmt.Errorf("Couldn't create temp pem file %v: %v", temporaryPemFile.Name(), err)
	}

	_, err = temporaryPemFile.WriteString(fmt.Sprintf("%v\n%v", cert, key))
	if err != nil {
		return SSLCert{}, fmt.Errorf("Couldn't write to pem file %v: %v", temporaryPemFile.Name(), err)
	}

	err = temporaryPemFile.Close()
	if err != nil {
		return SSLCert{}, fmt.Errorf("Couldn't close temp pem file %v: %v", temporaryPemFile.Name(), err)
	}

	cn, err := nginx.commonNames(temporaryPemFile.Name())
	if err != nil {
		os.Remove(temporaryPemFile.Name())
		return SSLCert{}, err
	}

	err = os.Rename(temporaryPemFile.Name(), pemFileName)
	if err != nil {
		os.Remove(temporaryPemFile.Name())
		return SSLCert{}, fmt.Errorf("Couldn't move temp pem file %v to destination %v: %v", temporaryPemFile.Name(), pemFileName, err)
	}

	return SSLCert{
		CertFileName: cert,
		KeyFileName:  key,
		PemFileName:  pemFileName,
		PemSHA:       nginx.pemSHA1(pemFileName),
		CN:           cn,
	}, nil
}

// commonNames checks if the certificate and key file are valid
// returning the result of the validation and the list of hostnames
// contained in the common name/s
func (nginx *Manager) commonNames(pemFileName string) ([]string, error) {
	pemCerts, err := ioutil.ReadFile(pemFileName)
	if err != nil {
		return []string{}, err
	}

	block, _ := pem.Decode(pemCerts)
	if block == nil {
		return []string{}, fmt.Errorf("No valid PEM formatted block found")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return []string{}, err
	}

	cn := []string{cert.Subject.CommonName}
	if len(cert.DNSNames) > 0 {
		cn = append(cn, cert.DNSNames...)
	}

	glog.V(3).Infof("found %v common names: %v\n", cn, len(cn))
	return cn, nil
}

// SearchDHParamFile iterates all the secrets mounted inside the /etc/nginx-ssl directory
// in order to find a file with the name dhparam.pem. If such file exists it will
// returns the path. If not it just returns an empty string
func (nginx *Manager) SearchDHParamFile(baseDir string) string {
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

func (nginx *Manager) pemSHA1(filename string) string {
	hasher := sha1.New()
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil))
}
