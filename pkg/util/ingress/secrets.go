/*
Copyright 2022 The Kubernetes Authors.
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

package ingress

import (
	"os"

	authfile "k8s.io/ingress-nginx/pkg/util/auth"
	"k8s.io/ingress-nginx/pkg/util/file"

	"k8s.io/ingress-nginx/internal/ingress/annotations/auth"
	"k8s.io/ingress-nginx/internal/ingress/controller/config"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/klog/v2"
)

// CheckAndWriteAuthSecrets is a util function that gets two Server arrays and write those on FS
// TODO: UNIT TEST!!! And e2e tests!
func CheckAndWriteAuthSecrets(newServers []*ingress.Server, oldServers []*ingress.Server) bool {
	var changed bool

	tempAuthOld := make(map[string]auth.Config)
	tempAuthNew := make(map[string]auth.Config)

	// Get added auth secrets:
	for _, v := range newServers {
		for _, location := range v.Locations {
			if location.BasicDigestAuth.FileSHA != "" {
				tempAuthNew[location.BasicDigestAuth.File] = location.BasicDigestAuth
			}
		}
	}

	// Get old auth secrets:
	for _, v := range oldServers {
		for _, location := range v.Locations {
			if location.BasicDigestAuth.FileSHA != "" {
				tempAuthOld[location.BasicDigestAuth.File] = location.BasicDigestAuth
			}
		}
	}

	// Check for removed secrets/files.
	for authFile := range tempAuthOld {
		if _, ok := tempAuthNew[authFile]; !ok {
			changed = true
			if err := os.Remove(authFile); err != nil && !os.IsNotExist(err) {
				klog.Warningf("failed removing old auth file %s: %s", authFile, err)
				continue
			}
		}
	}

	// Check on newMap if the files already exists and SHA matches, otherwise create/update
	for newAuthFile, newAuthConfig := range tempAuthNew {
		// If the auth secret didn't existed or existed but is not equal, then rewrite
		if oldAuthConfig, ok := tempAuthOld[newAuthFile]; !ok || oldAuthConfig.FileSHA != newAuthConfig.FileSHA {
			changed = true
			if err := authfile.WriteSecretFile(newAuthFile, newAuthConfig.SecretContent); err != nil {
				klog.Warningf("failed adding/updating auth file %s: %s", newAuthFile, err)
				continue
			}
		}
	}

	return changed
}

type certOperation struct {
	cert      config.CertificateFile
	operation CertOperationType
}

type CertOperationType string

const (
	certAdd    CertOperationType = "ADD"
	certRemove CertOperationType = "REMOVE"
)

// CheckAndWriteDeltaCertificates takes two maps of certificateFiles, add/remove those on Filesystem and return if somethins was changes
// TODO: UNIT TEST!!! And e2e tests!
func CheckAndWriteDeltaCertificates(oldCerts map[string]config.CertificateFile, newCerts map[string]config.CertificateFile) bool {
	var changed bool

	tempMap := make(map[string]certOperation)
	// Get added certificates:
	for k, v := range newCerts {
		kold, ok := oldCerts[k]
		// if non existent on the old one, add to the tempMap as a newCert
		if !ok {
			// format of added file will be +;Filename being ; a separator a + the create operation
			tempMap[k] = certOperation{
				operation: certAdd,
				cert:      v,
			}
		} else {
			// if existent in the old map, let's see if it changed checking the checksum
			if v.Checksum != kold.Checksum {
				tempMap[k] = certOperation{
					operation: certAdd,
					cert:      v,
				}
			}
		}
	}

	// Now remove old certificates
	for k := range oldCerts {
		if _, ok := newCerts[k]; !ok {
			// if NOK, add on the tempMap an operation to remove it
			// removal does not need the bytes or the SHA
			tempMap[k] = certOperation{
				operation: certRemove,
				cert:      config.CertificateFile{},
			}
		}
	}

	// If so, let's do the required operations and mark the reload as required
	if len(tempMap) > 0 {
		changed = true
		for k, v := range tempMap {
			if v.operation == certRemove {
				klog.Infof("removing cert %s", k)
				if err := os.Remove(k); err != nil {
					klog.Warningf("failed removing old certificate %s: %s", k, err)
				}
				continue
			}
			klog.Infof("adding/updating cert %s", k)
			err := os.WriteFile(k, []byte(v.cert.Content), file.ReadWriteByUser)
			if err != nil {
				klog.Errorf("could not create PEM certificate file %v: %v", k, err)
				continue
			}
		}
	}
	return changed
}
