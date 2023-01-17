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
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
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
func (s *k8sStore) syncSecret(key string, usingVault bool) {
	s.syncSecretMu.Lock()
	defer s.syncSecretMu.Unlock()

	klog.V(3).InfoS("Syncing Secret", "secret-path", key, "vault", usingVault)

	cert, err := s.getPemCertificate(key, usingVault)
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
		klog.InfoS("Updating secret in local store", "secret-path", key)
		s.sslStore.Update(key, cert)
		// this update must trigger an update
		// (like an update event from a change in Ingress)
		s.sendDummyEvent()
		return
	}

	klog.InfoS("Adding secret to local store", "secret-path", key)
	s.sslStore.Add(key, cert)
	// this update must trigger an update
	// (like an update event from a change in Ingress)
	s.sendDummyEvent()
}

// getVaultCertificate returns the cert, key and ca from Vault using the
// provided path and the Vault's data from environment variables.
func getVaultCertificate(secretPath string) (cert []byte, key []byte, ca []byte, onlyca bool, err error) {
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultPathDelimIndex := strings.LastIndex(secretPath, "/")
	vaultPath := secretPath[:vaultPathDelimIndex]
	// Check if vaultPath variable starts with '/' to build the API request
	vaultKey := secretPath[vaultPathDelimIndex+1:]
	klog.V(3).InfoS("Retrieving certificate from Vault using secret vault path  and key", "vault-path", vaultPath, "key", vaultKey)
	vaultAddress := "https://" + os.Getenv("VAULT_HOSTS") + ":" + os.Getenv("VAULT_PORT") + "/v1" + vaultPath
	// Prepare client
	klog.InfoS("Making an http request to vault against url", "vault-address", vaultAddress)
	req, err := http.NewRequest("GET", vaultAddress, nil)
	if err != nil {
		return cert, key, ca, onlyca, fmt.Errorf("error reading request: %v ", err)
	}

	req.Header.Set("X-Vault-Token", vaultToken)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return cert, key, ca, onlyca, fmt.Errorf("error reading response: %v ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return cert, key, ca, onlyca, fmt.Errorf("error reading body: %v ", err)
	}

	type Response struct {
		Data interface{}
	}
	var secret Response

	if err := json.Unmarshal(body, &secret); err != nil {
		return cert, key, ca, onlyca, fmt.Errorf("error reading vault response: %v ", err)
	}

	if secret.Data == nil {
		klog.Error("Cannot get data from Vault")
		return cert, key, ca, onlyca, nil
	}

	data := secret.Data.(map[string]interface{})

	// Check if there is a CA cert
	caCert, caCertOK := data[vaultKey+"_ca"].(string)
	if caCertOK {
		klog.V(3).Info("Found CA certificate in the key, not looking for anything else")
		ca = []byte(foldCrtPem(caCert))
		cert = nil
		key = nil
		onlyca = true
		return cert, key, ca, onlyca, nil
	}

	rawCert, rawCertOk := data[vaultKey+"_crt"].(string)
	if !rawCertOk {
		return cert, key, ca, onlyca, fmt.Errorf("no cert found")
	}

	cert = []byte(foldCrtPem(rawCert))
	rawKey, rawKeyOk := data[vaultKey+"_key"].(string)
	if !rawKeyOk {
		return cert, key, ca, onlyca, fmt.Errorf("no key found")
	}
	key = []byte(foldKeyPem(rawKey))

	re := regexp.MustCompile(`-----BEGIN CERTIFICATE-----.*?-----END CERTIFICATE-----`)
	rawCerts := re.FindAllString(rawCert, -1)
	rawCa := rawCerts[len(rawCerts)-1]
	ca = []byte(foldCrtPem(rawCa))

	return cert, key, ca, onlyca, nil
}

func foldCrtPem(pem string) string {
	pem = strings.ReplaceAll(pem, "-----BEGIN CERTIFICATE-----", "\n-----BEGIN CERTIFICATE-----\n")
	pem = strings.ReplaceAll(pem, "-----END CERTIFICATE-----", "\n-----END CERTIFICATE-----")
	return pem
}

func foldKeyPem(pem string) string {
	pem = strings.ReplaceAll(pem, "-----BEGIN RSA PRIVATE KEY-----", "\n-----BEGIN RSA PRIVATE KEY-----\n")
	pem = strings.ReplaceAll(pem, "-----END RSA PRIVATE KEY-----", "\n-----END RSA PRIVATE KEY-----")
	return pem
}

// getPemCertificate receives a secret, and creates a ingress.SSLCert as return.
// It parses the secret and verifies if it's a keypair, or a 'ca.crt' secret only.
func (s *k8sStore) getPemCertificate(secretName string, usingVault bool) (*ingress.SSLCert, error) {
	var cert, key, ca, crl, auth []byte
	var okcert, okkey, onlyca bool
	var uid, sslCertName, sslCertNamespace string
	var err error
	// We  check if we have got a vault secret annotation and proceed accordingly
	if usingVault {
		// UID is needed for checking certificates in a later step, but it has to be consistent and not random
		secretNameHex := hex.EncodeToString([]byte(secretName))

		uid = fmt.Sprintf("%s-%s-%s-%s-%s", secretNameHex[:8], secretNameHex[9:13], secretNameHex[14:18], secretNameHex[19:23], secretNameHex[24:36])
		klog.InfoS("Trying secret as a Vault Path", "secret", secretName)

		sslCertName = secretName
		// Check if secretName has a namespace defined
		sslCertNamespace = "vault"

		if !strings.HasPrefix(sslCertName, "/") {
			return nil, fmt.Errorf("Error in path %s, Vault paths must start with '/'", sslCertName)
		}

		cert, key, ca, onlyca, err = getVaultCertificate(sslCertName)
		if err != nil {
			return nil, fmt.Errorf("missing certificates in Vault's Path %s", sslCertName)
		}
		okcert = true
		okkey = true
	} else {
		secret, err := s.listers.Secret.ByKey(secretName)
		if err != nil {
			return nil, fmt.Errorf("Missing kubernetes secret: %s", secretName)
		}
		cert, okcert = secret.Data[apiv1.TLSCertKey]
		key, okkey = secret.Data[apiv1.TLSPrivateKeyKey]
		ca = secret.Data["ca.crt"]

		crl = secret.Data["ca.crl"]
		auth = secret.Data["auth"]

		uid = string(secret.UID)

		sslCertName = secret.Name
		sslCertNamespace = secret.Namespace

	}

	// namespace/secretName -> namespace-secretName
	nsSecName := strings.Replace(secretName, "/", "-", -1)

	var sslCert *ingress.SSLCert
	if okcert && okkey && !onlyca {
		if cert == nil {
			return nil, fmt.Errorf("key 'tls.crt' missing from Secret %q", secretName)
		}

		if key == nil {
			return nil, fmt.Errorf("key 'tls.key' missing from Secret %q", secretName)
		}

		// if there is no key we have to build the sslCert with the CA
		sslCert, err = ssl.CreateSSLCert(cert, key, uid)
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
	} else if len(ca) > 0 {
		sslCert, err = ssl.CreateCACert(ca)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating SSL Cert: %v", err)
		}

		err = ssl.ConfigureCACert(nsSecName, ca, sslCert)
		if err != nil {
			return nil, fmt.Errorf("error configuring CA certificate: %v", err)
		}

		sslCert.CASHA = file.SHA1(sslCert.CAFileName)

		if len(crl) > 0 {
			err = ssl.ConfigureCRL(nsSecName, crl, sslCert)
			if err != nil {
				return nil, err
			}
		}
		// makes this secret in 'syncSecret' to be used for Certificate Authentication
		// this does not enable Certificate Authentication
		klog.V(3).InfoS("Configuring Secret for TLS authentication", "secret", secretName)
	} else {
		if auth != nil {
			return nil, ErrSecretForAuth
		}

		return nil, fmt.Errorf("secret %q contains no keypair or CA certificate", secretName)
	}

	sslCert.Name = sslCertName
	sslCert.Namespace = sslCertNamespace

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

// sendDummyEvent sends a dummy event to trigger an update
// This is used in when a secret change
func (s *k8sStore) sendDummyEvent() {
	klog.V(3).Infof("Sending the update event with dummy")
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
