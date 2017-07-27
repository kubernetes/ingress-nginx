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

	"github.com/golang/glog"

	api_v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/ingress/controllers/gce/loadbalancers"
)

// secretLoaders returns a type containing all the secrets of an Ingress.
type tlsLoader interface {
	load(ing *extensions.Ingress) (*loadbalancers.TLSCerts, error)
	validate(certs *loadbalancers.TLSCerts) error
}

// TODO: Add better cert validation.
type noOPValidator struct{}

func (n *noOPValidator) validate(certs *loadbalancers.TLSCerts) error {
	return nil
}

// apiServerTLSLoader loads TLS certs from the apiserver.
type apiServerTLSLoader struct {
	noOPValidator
	client kubernetes.Interface
}

func (t *apiServerTLSLoader) load(ing *extensions.Ingress) (*loadbalancers.TLSCerts, error) {
	if len(ing.Spec.TLS) == 0 {
		return nil, nil
	}
	// GCE L7s currently only support a single cert.
	if len(ing.Spec.TLS) > 1 {
		glog.Warningf("Ignoring %d certs and taking the first for ingress %v/%v",
			len(ing.Spec.TLS)-1, ing.Namespace, ing.Name)
	}
	secretName := ing.Spec.TLS[0].SecretName
	// TODO: Replace this for a secret watcher.
	glog.V(3).Infof("Retrieving secret for ing %v with name %v", ing.Name, secretName)
	secret, err := t.client.Core().Secrets(ing.Namespace).Get(secretName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	cert, ok := secret.Data[api_v1.TLSCertKey]
	if !ok {
		return nil, fmt.Errorf("secret %v has no private key", secretName)
	}
	key, ok := secret.Data[api_v1.TLSPrivateKeyKey]
	if !ok {
		return nil, fmt.Errorf("secret %v has no cert", secretName)
	}
	certs := &loadbalancers.TLSCerts{Key: string(key), Cert: string(cert)}
	if err := t.validate(certs); err != nil {
		return nil, err
	}
	return certs, nil
}

// TODO: Add support for file loading so we can support HTTPS default backends.

// fakeTLSSecretLoader fakes out TLS loading.
type fakeTLSSecretLoader struct {
	noOPValidator
	fakeCerts map[string]*loadbalancers.TLSCerts
}

func (f *fakeTLSSecretLoader) load(ing *extensions.Ingress) (*loadbalancers.TLSCerts, error) {
	if len(ing.Spec.TLS) == 0 {
		return nil, nil
	}
	for name, cert := range f.fakeCerts {
		if ing.Spec.TLS[0].SecretName == name {
			return cert, nil
		}
	}
	return nil, fmt.Errorf("couldn't find secret for ingress %v", ing.Name)
}
