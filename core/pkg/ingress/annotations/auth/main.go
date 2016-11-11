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

package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

const (
	authType   = "ingress.kubernetes.io/auth-type"
	authSecret = "ingress.kubernetes.io/auth-secret"
	authRealm  = "ingress.kubernetes.io/auth-realm"

	// DefAuthDirectory default directory used to store files
	// to authenticate request
	DefAuthDirectory = "/etc/ingress-controller/auth"
)

func init() {
	// TODO: check permissions required
	os.MkdirAll(DefAuthDirectory, 0655)
}

var (
	authTypeRegex = regexp.MustCompile(`basic|digest`)

	// ErrInvalidAuthType is return in case of unsupported authentication type
	ErrInvalidAuthType = errors.New("invalid authentication type")

	// ErrMissingSecretName is returned when the name of the secret is missing
	ErrMissingSecretName = errors.New("secret name is missing")

	// ErrMissingAuthInSecret is returned when there is no auth key in secret data
	ErrMissingAuthInSecret = errors.New("the secret does not contains the auth key")
)

// BasicDigest returns authentication configuration for an Ingress rule
type BasicDigest struct {
	Type    string
	Realm   string
	File    string
	Secured bool
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func ParseAnnotations(ing *extensions.Ingress, authDir string, fn func(string) (*api.Secret, error)) (*BasicDigest, error) {
	if ing.GetAnnotations() == nil {
		return &BasicDigest{}, parser.ErrMissingAnnotations
	}

	at, err := parser.GetStringAnnotation(authType, ing)
	if err != nil {
		return &BasicDigest{}, err
	}

	if !authTypeRegex.MatchString(at) {
		return &BasicDigest{}, ErrInvalidAuthType
	}

	s, err := parser.GetStringAnnotation(authSecret, ing)
	if err != nil {
		return &BasicDigest{}, err
	}

	secret, err := fn(fmt.Sprintf("%v/%v", ing.Namespace, s))
	if err != nil {
		return &BasicDigest{}, err
	}

	realm, _ := parser.GetStringAnnotation(authRealm, ing)

	passFile := fmt.Sprintf("%v/%v-%v.passwd", authDir, ing.GetNamespace(), ing.GetName())
	err = dumpSecret(passFile, secret)
	if err != nil {
		return &BasicDigest{}, err
	}

	return &BasicDigest{
		Type:    at,
		Realm:   realm,
		File:    passFile,
		Secured: true,
	}, nil
}

// dumpSecret dumps the content of a secret into a file
// in the expected format for the specified authorization
func dumpSecret(filename string, secret *api.Secret) error {
	val, ok := secret.Data["auth"]
	if !ok {
		return ErrMissingAuthInSecret
	}

	// TODO: check permissions required
	return ioutil.WriteFile(filename, val, 0777)
}
