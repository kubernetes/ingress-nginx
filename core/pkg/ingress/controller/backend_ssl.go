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

package controller

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/glog"

	api "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/net/ssl"
)

// syncSecret keeps in sync Secrets used by Ingress rules with the files on
// disk to allow copy of the content of the secret to disk to be used
// by external processes.
func (ic *GenericController) syncSecret(key string) {
	glog.V(3).Infof("starting syncing of secret %v", key)

	cert, err := ic.getPemCertificate(key)
	if err != nil {
		glog.Warningf("error obtaining PEM from secret %v: %v", key, err)
		return
	}

	// create certificates and add or update the item in the store
	cur, exists := ic.sslCertTracker.Get(key)
	if exists {
		s := cur.(*ingress.SSLCert)
		if reflect.DeepEqual(s, cert) {
			// no need to update
			return
		}
		glog.Infof("updating secret %v in the local store", key)
		ic.sslCertTracker.Update(key, cert)
		return
	}

	glog.Infof("adding secret %v to the local store", key)
	ic.sslCertTracker.Add(key, cert)
}

// getPemCertificate receives a secret, and creates a ingress.SSLCert as return.
// It parses the secret and verifies if it's a keypair, or a 'ca.crt' secret only.
func (ic *GenericController) getPemCertificate(secretName string) (*ingress.SSLCert, error) {
	secretInterface, exists, err := ic.secrLister.Store.GetByKey(secretName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving secret %v: %v", secretName, err)
	}
	if !exists {
		return nil, fmt.Errorf("secret named %v does not exist", secretName)
	}

	secret := secretInterface.(*api.Secret)
	cert, okcert := secret.Data[api.TLSCertKey]
	key, okkey := secret.Data[api.TLSPrivateKeyKey]

	ca := secret.Data["ca.crt"]

	nsSecName := strings.Replace(secretName, "/", "-", -1)

	var s *ingress.SSLCert
	if okcert && okkey {
		s, err = ssl.AddOrUpdateCertAndKey(nsSecName, cert, key, ca)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating pem file %v", err)
		}
		glog.V(3).Infof("found certificate and private key, configuring %v as a TLS Secret (CN: %v)", secretName, s.CN)
	} else if ca != nil {
		glog.V(3).Infof("found only ca.crt, configuring %v as an Certificate Authentication secret", secretName)
		s, err = ssl.AddCertAuth(nsSecName, ca)
		if err != nil {
			return nil, fmt.Errorf("unexpected error creating pem file %v", err)
		}
	} else {
		return nil, fmt.Errorf("no keypair or CA cert could be found in %v", secretName)
	}

	if err != nil {
		return nil, err
	}

	s.Name = secret.Name
	s.Namespace = secret.Namespace
	return s, nil
}

// sslCertTracker holds a store of referenced Secrets in Ingress rules
type sslCertTracker struct {
	cache.ThreadSafeStore
}

func newSSLCertTracker() *sslCertTracker {
	return &sslCertTracker{
		cache.NewThreadSafeStore(cache.Indexers{}, cache.Indices{}),
	}
}

func (s *sslCertTracker) DeleteAll(key string) {
	s.Delete(key)
}
