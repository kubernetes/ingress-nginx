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

package auth

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	authType   = "ingress-nginx.kubernetes.io/auth-type"
	authSecret = "ingress-nginx.kubernetes.io/auth-secret"
	authRealm  = "ingress-nginx.kubernetes.io/auth-realm"

	defAuthRealm = "Authentication Required"
)

var (
	authTypeRegex = "basic|digest"
	authDir       = "/etc/nginx/auth"

	// ErrInvalidAuthType is return in case of unsupported authentication type
	ErrInvalidAuthType = errors.New("invalid authentication type")

	// ErrMissingAuthType is return when the annotation for authentication is missing
	ErrMissingAuthType = errors.New("authentication type is missing")

	// ErrMissingSecretName is returned when the name of the secret is missing
	ErrMissingSecretName = errors.New("secret name is missing")
)

// Nginx returns authentication configuration for an Ingress rule
type Nginx struct {
	Type   string
	Secret *api.Secret
	Realm  string
	File   string
}

type ingAnnotations map[string]string

func (a ingAnnotations) authType() (string, error) {
	val, ok := a[authType]
	if !ok {
		return "", ErrMissingAuthType
	}

	if val != "basic" || val != "digest" {
		return "", ErrInvalidAuthType
	}

	return val, nil
}

func (a ingAnnotations) realm() string {
	val, ok := a[authRealm]
	if !ok {
		return defAuthRealm
	}

	return val
}

func (a ingAnnotations) secretName() (string, error) {
	val, ok := a[authSecret]
	if !ok {
		return "", ErrMissingSecretName
	}

	return val, nil
}

// Parse parses the annotations contained in the ingress rule
// used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func Parse(kubeClient client.Interface, ing *extensions.Ingress) (*Nginx, error) {
	at, err := ingAnnotations(ing.GetAnnotations()).authType()
	if err != nil {
		return nil, err
	}

	s, err := ingAnnotations(ing.GetAnnotations()).secretName()
	if err != nil {
		return nil, err
	}

	secret, err := kubeClient.Secrets(ing.Namespace).Get(s)
	if err != nil {
		return nil, err
	}

	realm := ingAnnotations(ing.GetAnnotations()).realm()

	passFile := fmt.Sprintf("%v/%v-%v.passwd", authDir, ing.GetNamespace(), ing.GetName())
	err = dumpSecret(passFile, at, secret)
	if err != nil {
		return nil, err
	}

	n := &Nginx{
		Type:   at,
		Secret: secret,
		Realm:  realm,
		File:   passFile,
	}

	return n, nil
}

// dumpSecret dumps the content of a secret into a file
// in the expected format for the specified authorization
func dumpSecret(filename, auth string, secret *api.Secret) error {
	buf := bytes.NewBuffer([]byte{})

	for key, value := range secret.Data {
		fmt.Fprintf(buf, "%v:%s\n", key, value)
	}

	return ioutil.WriteFile(filename, buf.Bytes(), 600)
}
