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

	"github.com/golang/glog"
	"github.com/imdario/mergo"

	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/net/ssl"
)

// syncSecret keeps in sync Secrets used by Ingress rules with the files on
// disk to allow copy of the content of the secret to disk to be used
// by external processes.
func (s k8sStore) syncSecret(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	glog.V(3).Infof("starting syncing of secret %v", key)

	// TODO: getPemCertificate should not write to disk to avoid unnecessary overhead
	cert, err := s.getPemCertificate(key)
	if err != nil {
		glog.Warningf("error obtaining PEM from secret %v: %v", key, err)
		return
	}

	// create certificates and add or update the item in the store
	cur, err := s.GetLocalSSLCert(key)
	if err == nil {
		if cur.Equal(cert) {
			// no need to update
			return
		}
		glog.Infof("updating secret %v in the local store", key)
		s.sslStore.Update(key, cert)
		// this update must trigger an update
		// (like an update event from a change in Ingress)
		s.sendDummyEvent()
		return
	}

	glog.Infof("adding secret %v to the local store", key)
	s.sslStore.Add(key, cert)
	// this update must trigger an update
	// (like an update event from a change in Ingress)
	s.sendDummyEvent()
}

// getPemCertificate receives a secret, and creates a ingress.SSLCert as return.
// It parses the secret and verifies if it's a keypair, or a 'ca.crt' secret only.
func (s k8sStore) getPemCertificate(secretName string) (*ingress.SSLCert, error) {
	secret, err := s.listers.Secret.ByKey(secretName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving secret %v: %v", secretName, err)
	}

	cert, okcert := secret.Data[apiv1.TLSCertKey]
	key, okkey := secret.Data[apiv1.TLSPrivateKeyKey]
	ca := secret.Data["ca.crt"]

	// namespace/secretName -> namespace-secretName
	nsSecName := strings.Replace(secretName, "/", "-", -1)

	var sslCert *ingress.SSLCert
	if okcert && okkey {
		if cert == nil {
			return nil, fmt.Errorf("secret %v has no 'tls.crt'", secretName)
		}
		if key == nil {
			return nil, fmt.Errorf("secret %v has no 'tls.key'", secretName)
		}

		// If 'ca.crt' is also present, it will allow this secret to be used in the
		// 'nginx.ingress.kubernetes.io/auth-tls-secret' annotation
		sslCert, err = ssl.AddOrUpdateCertAndKey(nsSecName, cert, key, ca, s.filesystem)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating pem file: %v", err)
		}

		glog.V(3).Infof("found 'tls.crt' and 'tls.key', configuring %v as a TLS Secret (CN: %v)", secretName, sslCert.CN)
		if ca != nil {
			glog.V(3).Infof("found 'ca.crt', secret %v can also be used for Certificate Authentication", secretName)
		}
	} else if ca != nil {
		sslCert, err = ssl.AddCertAuth(nsSecName, ca, s.filesystem)

		if err != nil {
			return nil, fmt.Errorf("unexpected error creating pem file: %v", err)
		}

		// makes this secret in 'syncSecret' to be used for Certificate Authentication
		// this does not enable Certificate Authentication
		glog.V(3).Infof("found only 'ca.crt', configuring %v as an Certificate Authentication Secret", secretName)

	} else {
		return nil, fmt.Errorf("no keypair or CA cert could be found in %v", secretName)
	}

	sslCert.Name = secret.Name
	sslCert.Namespace = secret.Namespace

	return sslCert, nil
}

func (s k8sStore) checkSSLChainIssues() {
	for _, item := range s.ListLocalSSLCerts() {
		secretName := k8s.MetaNamespaceKey(item)
		secret, err := s.GetLocalSSLCert(secretName)
		if err != nil {
			continue
		}

		if secret.FullChainPemFileName != "" {
			// chain already checked
			continue
		}

		data, err := ssl.FullChainCert(secret.PemFileName, s.filesystem)
		if err != nil {
			glog.Errorf("unexpected error generating SSL certificate with full intermediate chain CA certs: %v", err)
			continue
		}

		fullChainPemFileName := fmt.Sprintf("%v/%v-%v-full-chain.pem", file.DefaultSSLDirectory, secret.Namespace, secret.Name)

		file, err := s.filesystem.Create(fullChainPemFileName)
		if err != nil {
			glog.Errorf("unexpected error creating SSL certificate file %v: %v", fullChainPemFileName, err)
			continue
		}

		_, err = file.Write(data)
		if err != nil {
			glog.Errorf("unexpected error creating SSL certificate: %v", err)
			continue
		}

		dst := &ingress.SSLCert{}

		err = mergo.MergeWithOverwrite(dst, secret)
		if err != nil {
			glog.Errorf("unexpected error creating SSL certificate: %v", err)
			continue
		}

		dst.FullChainPemFileName = fullChainPemFileName

		glog.Infof("updating local copy of ssl certificate %v with missing intermediate CA certs", secretName)
		s.sslStore.Update(secretName, dst)
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
