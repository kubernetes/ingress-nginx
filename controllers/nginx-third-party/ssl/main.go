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

package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/glog"
)

// Certificate contains the cert, key and the list of valid hostnames
type Certificate struct {
	Cert    string
	Key     string
	Cname   []string
	Valid   bool
	Default bool
}

// CreateSSLCerts reads the content of the /etc/nginx-ssl directory and
// verifies the cert and key extracting the common names for this pair
func CreateSSLCerts(baseDir string) []Certificate {
	sslCerts := []Certificate{}

	glog.Infof("inspecting directory %v for SSL certificates\n", baseDir)
	files, _ := ioutil.ReadDir(baseDir)
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		// the name of the secret could be different than the certificate file
		cert, key, err := getCert(fmt.Sprintf("%v/%v", baseDir, file.Name()))
		if err != nil {
			glog.Errorf("error checking certificate: %v", err)
			continue
		}

		hosts, err := checkSSLCertificate(cert, key)
		if err == nil {
			sslCert := Certificate{
				Cert:  cert,
				Key:   key,
				Cname: hosts,
				Valid: true,
			}

			if file.Name() == "default" {
				sslCert.Default = true
			}

			sslCerts = append(sslCerts, sslCert)
		} else {
			glog.Errorf("error checking certificate: %v", err)
		}
	}

	if len(sslCerts) == 1 {
		sslCerts[0].Default = true
	}

	glog.Infof("ssl certificates found: %v", sslCerts)

	return sslCerts
}

// checkSSLCertificate check if the certificate and key file are valid
// returning the result of the validation and the list of hostnames
// contained in the common name/s
func checkSSLCertificate(certFile, keyFile string) ([]string, error) {
	_, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		glog.Errorf("Error checking certificate and key file %v/%v: %v", certFile, keyFile, err)
		return []string{}, err
	}

	pemCerts, err := ioutil.ReadFile(certFile)
	if err != nil {
		return []string{}, err
	}

	var block *pem.Block
	block, pemCerts = pem.Decode(pemCerts)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		glog.Errorf("Error checking certificate and key file %v/%v: %v", certFile, keyFile, err)
		return []string{}, err
	}

	cn := []string{cert.Subject.CommonName}
	if len(cert.DNSNames) > 0 {
		cn = append(cn, cert.DNSNames...)
	}

	glog.Infof("DNS %v %v\n", cn, len(cn))
	return cn, nil
}

func verifyHostname(certFile, host string) bool {
	pemCerts, err := ioutil.ReadFile(certFile)
	if err != nil {
		return false
	}

	var block *pem.Block
	block, pemCerts = pem.Decode(pemCerts)

	cert, err := x509.ParseCertificate(block.Bytes)

	err = cert.VerifyHostname(host)
	if err == nil {
		return true
	}

	return false
}

// GetSSLHost checks if in one of the secrets that contains SSL
// certificates could be used for the specified server name
func GetSSLHost(serverName string, certs []Certificate) Certificate {
	for _, sslCert := range certs {
		if verifyHostname(sslCert.Cert, serverName) {
			return sslCert
		}
	}

	return Certificate{}
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

// getCert returns the pair cert-key if exists or an error
func getCert(certDir string) (cert string, key string, err error) {
	// we search for a file with extension crt
	filepath.Walk(certDir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(".crt", f.Name())
			if err == nil && r {
				cert = f.Name()
				return nil
			}
		}

		return nil
	})

	cert = fmt.Sprintf("%v/%v", certDir, cert)
	if _, err := os.Stat(cert); os.IsNotExist(err) {
		return "", "", fmt.Errorf("No certificate found in directory %v: %v", certDir, err)
	}

	key = strings.Replace(cert, ".crt", ".key", 1)

	if _, err := os.Stat(key); os.IsNotExist(err) {
		return "", "", fmt.Errorf("No certificate key found in directory %v: %v", certDir, err)
	}

	return
}
