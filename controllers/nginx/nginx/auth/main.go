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

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	authType   = "ingress.kubernetes.io/auth-type"
	authSecret = "ingress.kubernetes.io/auth-secret"
	authRealm  = "ingress.kubernetes.io/auth-realm"

	defAuthRealm = "Authentication Required"

	// DefAuthDirectory default directory used to store files
	// to authenticate request in NGINX
	DefAuthDirectory = "/etc/nginx/auth"
)

func init() {
	// TODO: check permissions required
	os.MkdirAll(DefAuthDirectory, 0655)
}

var (
	authTypeRegex = regexp.MustCompile(`basic|digest`)

	// ErrInvalidAuthType is return in case of unsupported authentication type
	ErrInvalidAuthType = errors.New("invalid authentication type")

	// ErrMissingAuthType is return when the annotation for authentication is missing
	ErrMissingAuthType = errors.New("authentication type is missing")

	// ErrMissingSecretName is returned when the name of the secret is missing
	ErrMissingSecretName = errors.New("secret name is missing")

	// ErrMissingAuthInSecret is returned when there is no auth key in secret data
	ErrMissingAuthInSecret = errors.New("the secret does not contains the auth key")

	// ErrMissingAnnotations is returned when the ingress rule
	// does not contains annotations related with authentication
	ErrMissingAnnotations = errors.New("missing authentication annotations")
)

// Nginx returns authentication configuration for an Ingress rule
type Nginx struct {
	Type    string
	Realm   string
	File    string
	Secured bool
}

type ingAnnotations map[string]string

func (a ingAnnotations) authType() (string, error) {
	val, ok := a[authType]
	if !ok {
		return "", ErrMissingAuthType
	}

	if !authTypeRegex.MatchString(val) {
		glog.Warningf("%v is not a valid authentication type", val)
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

// ParseAnnotations parses the annotations contained in the ingress
// rule used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func ParseAnnotations(kubeClient client.Interface, ing *extensions.Ingress, authDir string) (*Nginx, error) {
	if ing.GetAnnotations() == nil {
		return &Nginx{}, ErrMissingAnnotations
	}

	at, err := ingAnnotations(ing.GetAnnotations()).authType()
	if err != nil {
		return &Nginx{}, err
	}

	s, err := ingAnnotations(ing.GetAnnotations()).secretName()
	if err != nil {
		return &Nginx{}, err
	}

	secret, err := kubeClient.Secrets(ing.Namespace).Get(s)
	if err != nil {
		return &Nginx{}, err
	}

	realm := ingAnnotations(ing.GetAnnotations()).realm()

	passFile := fmt.Sprintf("%v/%v-%v.passwd", authDir, ing.GetNamespace(), ing.GetName())
	err = dumpSecret(passFile, secret)
	if err != nil {
		return &Nginx{}, err
	}

	return &Nginx{
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
