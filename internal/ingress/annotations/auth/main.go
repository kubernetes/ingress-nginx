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
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

var (
	authTypeRegex = regexp.MustCompile(`basic|digest|oidc`)
	// AuthDirectory default directory used to store files
	// to authenticate request
	AuthDirectory = "/etc/ingress-controller/auth"
)

const (
	fileAuth = "auth-file"
	mapAuth  = "auth-map"
)

// Config returns authentication configuration for an Ingress rule
type Config struct {
	Type          string `json:"type"`
	Realm         string `json:"realm"`
	File          string `json:"file"`
	Secured       bool   `json:"secured"`
	FileSHA       string `json:"fileSha"`
	Secret        string `json:"secret"`
	SecretType    string `json:"secretType"`
	SecretData    string `json:"secretData"`
	ConfigMap     string `json:"configMap"`
	ConfigMapData string `json:"configMapData"`
}

// Equal tests for equality between two Config types
func (bd1 *Config) Equal(bd2 *Config) bool {
	if bd1 == bd2 {
		return true
	}
	if bd1 == nil || bd2 == nil {
		return false
	}
	if bd1.Type != bd2.Type {
		return false
	}
	if bd1.Realm != bd2.Realm {
		return false
	}
	if bd1.File != bd2.File {
		return false
	}
	if bd1.Secured != bd2.Secured {
		return false
	}
	if bd1.FileSHA != bd2.FileSHA {
		return false
	}
	if bd1.Secret != bd2.Secret {
		return false
	}
	if bd1.SecretData != bd2.SecretData {
		return false
	}
	if bd1.ConfigMapData != bd2.ConfigMapData {
		return false
	}
	return true
}

type auth struct {
	r             resolver.Resolver
	authDirectory string
}

// NewParser creates a new authentication annotation parser
func NewParser(authDirectory string, r resolver.Resolver) parser.IngressAnnotation {
	return auth{r, authDirectory}
}

// Parse parses the annotations contained in the ingress
// rule used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func (a auth) Parse(ing *networking.Ingress) (interface{}, error) {
	at, err := parser.GetStringAnnotation("auth-type", ing)
	if err != nil {
		return nil, err
	}

	if !authTypeRegex.MatchString(at) {
		return nil, ing_errors.NewLocationDenied("invalid authentication type")
	}

	s, err := parser.GetStringAnnotation("auth-secret", ing)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error reading secret name from annotation"),
		}
	}

	sns, sname, err := cache.SplitMetaNamespaceKey(s)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error reading secret name from annotation"),
		}
	}

	if sns == "" {
		sns = ing.Namespace
	}

	secretName := fmt.Sprintf("%v/%v", sns, sname)
	secret, err := a.r.GetSecret(secretName)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrapf(err, "unexpected error reading secret %v", secretName),
		}
	}

	var realm, passFile, secretType, cmName, fileSHA string
	var configMap *api.ConfigMap
	secretBufSlice := make([]byte, 0)
	secretBuffer := bytes.NewBuffer(secretBufSlice)
	cmBufSlice := make([]byte, 0)
	cmBuffer := bytes.NewBuffer(cmBufSlice)
	if at == "oidc" {
		cm, err := parser.GetStringAnnotation("auth-config", ing)

		if err == nil {
			cmns, cmname, err := cache.SplitMetaNamespaceKey(cm)
			if err != nil {
				return nil, ing_errors.LocationDenied{
					Reason: errors.Wrap(err, "error parsing configmap name from annotation"),
				}
			}
			if cmns == "" {
				cmns = ing.Namespace
			}
			cmName = fmt.Sprintf("%v/%v", cmns, cmname)
			configMap, err = a.r.GetConfigMap(cmName)
			if err != nil {
				return nil, ing_errors.LocationDenied{
					Reason: errors.Wrapf(err, "unexpected error reading configmap %v", cmName),
				}
			}
			dumpCMAuthBuffer(cmBuffer, configMap)
		} else if err == ing_errors.ErrMissingAnnotations {
			cmName = ""
		} else {
			return nil, ing_errors.LocationDenied{
				Reason: errors.Wrap(err, "error reading configmap name from annotation"),
			}
		}
		dumpSecretAuthBuffer(secretBuffer, secret)
	} else {
		realm, _ = parser.GetStringAnnotation("auth-realm", ing)

		passFile = fmt.Sprintf("%v/%v-%v.passwd", a.authDirectory, ing.GetNamespace(), ing.GetName())

		var secretType string
		secretType, err = parser.GetStringAnnotation("auth-secret-type", ing)
		if err != nil {
			secretType = "auth-file"
		}
		if secretType == "auth-file" {
			err = dumpSecretAuthFile(passFile, secret)
			if err != nil {
				return nil, err
			}
		} else if secretType == "auth-map" {
			err = dumpSecretAuthMap(passFile, secret)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, ing_errors.LocationDenied{
				Reason: errors.Wrap(err, "invalid auth-secret-type in annotation, must be 'auth-file' or 'auth-map'"),
			}
		}
		fileSHA = file.SHA1(passFile)
	}

	return &Config{
		Type:          at,
		Realm:         realm,
		File:          passFile,
		Secured:       true,
		FileSHA:       fileSHA,
		Secret:        secretName,
		SecretType:    secretType,
		SecretData:    secretBuffer.String(),
		ConfigMap:     cmName,
		ConfigMapData: cmBuffer.String(),
	}, nil
}

// dumpSecret dumps the content of a secret into a file
// in the expected format for the specified authorization
func dumpSecretAuthFile(filename string, secret *api.Secret) error {
	val, ok := secret.Data["auth"]
	if !ok {
		return ing_errors.LocationDenied{
			Reason: errors.Errorf("the secret %v does not contain a key with value auth", secret.Name),
		}
	}

	err := ioutil.WriteFile(filename, val, file.ReadWriteByUser)
	if err != nil {
		return ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "unexpected error creating password file"),
		}
	}

	return nil
}

func dumpSecretAuthMap(filename string, secret *api.Secret) error {
	builder := &strings.Builder{}
	for user, pass := range secret.Data {
		builder.WriteString(user)
		builder.WriteString(":")
		builder.WriteString(string(pass))
		builder.WriteString("\n")
	}

	err := ioutil.WriteFile(filename, []byte(builder.String()), file.ReadWriteByUser)
	if err != nil {
		return ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "unexpected error creating password file"),
		}
	}

	return nil
}

func dumpSecretAuthBuffer(buffer *bytes.Buffer, secret *api.Secret) {
	builder := &strings.Builder{}
	for key, value := range secret.Data {
		builder.WriteString(key)
		builder.WriteString(":")
		builder.WriteString(string(value))
		builder.WriteString("\n")
	}

	buffer.Write([]byte(builder.String()))
}

func dumpCMAuthBuffer(buffer *bytes.Buffer, cm *api.ConfigMap) {
	builder := &strings.Builder{}
	for key, value := range cm.Data {
		builder.WriteString(key)
		builder.WriteString(":")
		str := base64.StdEncoding.EncodeToString([]byte(value))
		builder.WriteString(str)
		builder.WriteString("\n")
	}
	buffer.Write([]byte(builder.String()))
}
