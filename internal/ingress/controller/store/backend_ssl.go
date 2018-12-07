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

package store

import (
	"fmt"
	"strings"

	"github.com/imdario/mergo"
	"k8s.io/klog"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/net/ssl"
)

// syncSecret synchronizes the content of a TLS Secret (certificate(s), secret
// key) with the filesystem. The resulting files can be used by NGINX.
func (s *k8sStore) syncSecret(key string) {
	s.syncSecretMu.Lock()
	defer s.syncSecretMu.Unlock()

	klog.V(3).Infof("Syncing Secret %q", key)

	// TODO: getPemCertificate should not write to disk to avoid unnecessary overhead
	cert, err := s.getPemCertificate(key)
	if err != nil {
		if !isErrSecretForAuth(err) {
			klog.Warningf("Error obtaining X.509 certificate: %v", err)
		}
		return
	}

	// create certificates and add or update the item in the store
	cur, err := s.GetLocalSSLCert(key)
	if err == nil {
		if cur.Equal(cert) {
			// no need to update
			return
		}
		klog.Infof("Updating Secret %q in the local store", key)
		s.sslStore.Update(key, cert)
		// this update must trigger an update
		// (like an update event from a change in Ingress)
		s.sendDummyEvent()
		return
	}

	klog.Infof("Adding Secret %q to the local store", key)
	s.sslStore.Add(key, cert)
	// this update must trigger an update
	// (like an update event from a change in Ingress)
	s.sendDummyEvent()
}

// getPemCertificate receives a secret, and creates a ingress.SSLCert as return.
// It parses the secret and verifies if it's a keypair, or a 'ca.crt' secret only.
func (s *k8sStore) getPemCertificate(secretName string) (*ingress.SSLCert, error) {
	secret, err := s.listers.Secret.ByKey(secretName)
	if err != nil {
		return nil, err
	}

	cert, okcert := secret.Data[apiv1.TLSCertKey]
	key, okkey := secret.Data[apiv1.TLSPrivateKeyKey]
	ca := secret.Data["ca.crt"]

	auth := secret.Data["auth"]

	// namespace/secretName -> namespace-secretName
	nsSecName := strings.Replace(secretName, "/", "-", -1)

	var sslCert *ingress.SSLCert
	if okcert && okkey {
		if cert == nil {
			return nil, fmt.Errorf("key 'tls.crt' missing from Secret %q", secretName)
		}
		if key == nil {
			return nil, fmt.Errorf("key 'tls.key' missing from Secret %q", secretName)
		}

		if s.isDynamicCertificatesEnabled {
			sslCert, err = ssl.CreateSSLCert(nsSecName, cert, key, ca)
			if err != nil {
				return nil, fmt.Errorf("unexpected error creating SSL Cert: %v", err)
			}
		} else {
			// If 'ca.crt' is also present, it will allow this secret to be used in the
			// 'nginx.ingress.kubernetes.io/auth-tls-secret' annotation
			sslCert, err = ssl.AddOrUpdateCertAndKey(nsSecName, cert, key, ca, s.filesystem)
			if err != nil {
				return nil, fmt.Errorf("unexpected error creating pem file: %v", err)
			}
		}

		msg := fmt.Sprintf("Configuring Secret %q for TLS encryption (CN: %v)", secretName, sslCert.CN)
		if ca != nil {
			msg += " and authentication"
		}
		klog.V(3).Info(msg)

	} else if ca != nil {
		sslCert, err = ssl.AddCertAuth(nsSecName, ca, s.filesystem)

		if err != nil {
			return nil, err
		}

		// makes this secret in 'syncSecret' to be used for Certificate Authentication
		// this does not enable Certificate Authentication
		klog.V(3).Infof("Configuring Secret %q for TLS authentication", secretName)

	} else {
		if auth != nil {
			return nil, ErrSecretForAuth
		}

		return nil, fmt.Errorf("secret %q contains no keypair or CA certificate", secretName)
	}

	sslCert.Name = secret.Name
	sslCert.Namespace = secret.Namespace

	return sslCert, nil
}

func (s *k8sStore) checkSSLChainIssues() {
	for _, item := range s.ListLocalSSLCerts() {
		secrKey := k8s.MetaNamespaceKey(item)
		secret, err := s.GetLocalSSLCert(secrKey)
		if err != nil {
			continue
		}

		if secret.FullChainPemFileName != "" {
			// chain already checked
			continue
		}

		data, err := ssl.FullChainCert(secret.PemFileName, s.filesystem)
		if err != nil {
			klog.Errorf("Error generating CA certificate chain for Secret %q: %v", secrKey, err)
			continue
		}

		fullChainPemFileName := fmt.Sprintf("%v/%v-%v-full-chain.pem", file.DefaultSSLDirectory, secret.Namespace, secret.Name)

		file, err := s.filesystem.Create(fullChainPemFileName)
		if err != nil {
			klog.Errorf("Error creating SSL certificate file for Secret %q: %v", secrKey, err)
			continue
		}

		_, err = file.Write(data)
		if err != nil {
			klog.Errorf("Error creating SSL certificate for Secret %q: %v", secrKey, err)
			continue
		}

		dst := &ingress.SSLCert{}

		err = mergo.MergeWithOverwrite(dst, secret)
		if err != nil {
			klog.Errorf("Error creating SSL certificate for Secret %q: %v", secrKey, err)
			continue
		}

		dst.FullChainPemFileName = fullChainPemFileName

		klog.Infof("Updating local copy of SSL certificate %q with missing intermediate CA certs", secrKey)
		s.sslStore.Update(secrKey, dst)
		// this update must trigger an update
		// (like an update event from a change in Ingress)
		s.sendDummyEvent()
	}
}

// sendDummyEvent sends a dummy event to trigger an update
// This is used in when a secret change
func (s *k8sStore) sendDummyEvent() {
	s.updateCh.In() <- Event{
		Type: UpdateEvent,
		Obj: &extensions.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy",
				Namespace: "dummy",
			},
		},
	}
}

// ErrSecretForAuth error to indicate a secret is used for authentication
var ErrSecretForAuth = fmt.Errorf("secret is used for authentication")

func isErrSecretForAuth(e error) bool {
	return e == ErrSecretForAuth
}
