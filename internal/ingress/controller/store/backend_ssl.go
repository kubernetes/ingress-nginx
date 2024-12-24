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

	apiv1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/pkg/apis/ingress"

	klog "k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/net/ssl"

	"k8s.io/ingress-nginx/pkg/util/file"
)

// syncSecret synchronizes the content of a TLS Secret (certificate(s), secret
// key) with the filesystem. The resulting files can be used by NGINX.
func (s *k8sStore) syncSecret(key string) {
	s.syncResource(key, "Secret", s.getPemCertificate)
}

// syncSecret synchronizes the content of a client TLS Secret (certificate, secret
// key) with the filesystem. The resulting files can be used by NGINX.
func (s *k8sStore) syncClientCertSecret(key string) {
	s.syncResource(key, "Secret", s.getClientPemCertificate)
}

// syncCAConfigMap synchronizes the content of a ConfigMap including CAs
// with the filesystem. The resulting file can be used by NGINX.
func (s *k8sStore) syncCAConfigMap(key string) {
	s.syncResource(key, "ConfigMap", s.getCAPEMCertificate)
}

// syncResource synchronizes the content of a K8s resource (Secret, ConfigMap...)
func (s *k8sStore) syncResource(key, resourceKind string, getPermCert func(string) (*ingress.SSLCert, error)) {
	s.syncSecretMu.Lock()
	defer s.syncSecretMu.Unlock()

	klog.V(3).InfoS(fmt.Sprintf("Syncing %s", resourceKind), "name", key)

	cert, err := getPermCert(key)
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
		klog.InfoS("Updating cert in local store", "name", key)
		s.sslStore.Update(key, cert)
		// this update must trigger an update
		// (like an update event from a change in Ingress)
		s.sendDummyEvent()
		return
	}

	klog.InfoS("Adding cert to local store", "name", key)
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

	crl := secret.Data["ca.crl"]

	auth := secret.Data["auth"]

	// namespace/secretName -> namespace-secretName
	nsSecName := strings.ReplaceAll(secretName, "/", "-")

	var sslCert *ingress.SSLCert
	switch {
	case okcert && okkey:
		if cert == nil {
			return nil, fmt.Errorf("key 'tls.crt' missing from Secret %q", secretName)
		}

		if key == nil {
			return nil, fmt.Errorf("key 'tls.key' missing from Secret %q", secretName)
		}

		sslCert, err = ssl.CreateSSLCert(cert, key, string(secret.UID))
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating SSL Cert: %v", err)
		}

		if len(ca) > 0 {
			caCert, err := ssl.CheckCACert(ca)
			if err != nil {
				return nil, fmt.Errorf("parsing CA certificate: %v", err)
			}

			path, err := ssl.StoreSSLCertOnDisk(nsSecName, sslCert)
			if err != nil {
				return nil, fmt.Errorf("error while storing certificate and key: %v", err)
			}

			sslCert.PemFileName = path
			sslCert.CACertificate = caCert
			sslCert.CAFileName = path
			sslCert.CASHA = file.SHA1(path)

			err = ssl.ConfigureCACertWithCertAndKey(nsSecName, ca, sslCert)
			if err != nil {
				return nil, fmt.Errorf("error configuring CA certificate: %v", err)
			}

			if len(crl) > 0 {
				err = ssl.ConfigureCRL(nsSecName, crl, sslCert)
				if err != nil {
					return nil, fmt.Errorf("error configuring CRL certificate: %v", err)
				}
			}
		}

		msg := fmt.Sprintf("Configuring Secret %q for TLS encryption (CN: %v)", secretName, sslCert.CN)
		if ca != nil {
			msg += " and authentication"
		}

		if crl != nil {
			msg += " and CRL"
		}

		klog.V(3).InfoS(msg)
	case len(ca) > 0:
		sslCert, err = createCASSLCert(secretName, "Secret", string(secret.UID), ca, crl)
		if err != nil {
			return nil, err
		}
	default:
		if auth != nil {
			return nil, ErrSecretForAuth
		}

		return nil, fmt.Errorf("secret %q contains no keypair or CA certificate", secretName)
	}

	sslCert.Name = secret.Name
	sslCert.Namespace = secret.Namespace

	// the default SSL certificate needs to be present on disk
	if secretName == s.defaultSSLCertificate {
		path, err := ssl.StoreSSLCertOnDisk(nsSecName, sslCert)
		if err != nil {
			return nil, fmt.Errorf("storing default SSL Certificate: %w", err)
		}

		sslCert.PemFileName = path
	}

	return sslCert, nil
}

// getClientPemCertificate receives a secret, and creates an ingress.SSLCert as return.
// It parses the secret and verifies they keypair.
func (s *k8sStore) getClientPemCertificate(secretName string) (*ingress.SSLCert, error) {
	secret, err := s.listers.Secret.ByKey(secretName)
	if err != nil {
		return nil, err
	}

	cert, okcert := secret.Data[apiv1.TLSCertKey]
	if !okcert || cert == nil {
		return nil, fmt.Errorf("key 'tls.crt' missing from Secret %q", secretName)
	}

	key, okkey := secret.Data[apiv1.TLSPrivateKeyKey]
	if !okkey || key == nil {
		return nil, fmt.Errorf("key 'tls.key' missing from Secret %q", secretName)
	}

	// namespace/secretName -> namespace-secretName
	nsSecName := strings.ReplaceAll(secretName, "/", "-")

	sslCert, err := ssl.CreateSSLCert(cert, key, string(secret.UID))
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating SSL Cert: %v", err)
	}

	path, err := ssl.StoreSSLCertOnDisk(nsSecName, sslCert)
	if err != nil {
		return nil, fmt.Errorf("error while storing certificate and key: %v", err)
	}
	sslCert.PemFileName = path

	klog.V(3).InfoS(fmt.Sprintf("Configuring Secret %q for TLS encryption (CN: %v)", secretName, sslCert.CN))

	sslCert.Name = secret.Name
	sslCert.Namespace = secret.Namespace

	return sslCert, nil
}

// getCAPEMCertificate receives a configMap, and creates an ingress.SSLCert as return.
// It parses the configMap
func (s *k8sStore) getCAPEMCertificate(configMapName string) (*ingress.SSLCert, error) {
	configmap, err := s.listers.ConfigMap.ByKey(configMapName)
	if err != nil {
		return nil, err
	}

	ca := configmap.Data["ca.crt"]
	crl := configmap.Data["ca.crl"]

	return createCASSLCert(configMapName, "ConfigMap", string(configmap.UID), []byte(ca), []byte(crl))
}

func createCASSLCert(resourceName, resourceKind, resourceUID string, ca, crl []byte) (*ingress.SSLCert, error) {
	nsName := strings.ReplaceAll(resourceName, "/", "-")

	sslCert, err := ssl.CreateCACert(ca)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating SSL Cert: %v", err)
	}

	sslCert.UID = resourceUID

	err = ssl.ConfigureCACert(nsName, ca, sslCert)
	if err != nil {
		return nil, fmt.Errorf("error configuring CA certificate: %v", err)
	}

	sslCert.CASHA = file.SHA1(sslCert.CAFileName)

	if len(crl) > 0 {
		err = ssl.ConfigureCRL(nsName, crl, sslCert)
		if err != nil {
			return nil, err
		}
	}
	klog.V(3).InfoS("Configuring certs for TLS authentication", resourceKind, resourceName)

	return sslCert, nil
}

// sendDummyEvent sends a dummy event to trigger an update
// This is used in when a secret change
func (s *k8sStore) sendDummyEvent() {
	s.updateCh.In() <- Event{
		Type: UpdateEvent,
		Obj: &networking.Ingress{
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
