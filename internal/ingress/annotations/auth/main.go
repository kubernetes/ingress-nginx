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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
	"k8s.io/ingress-nginx/pkg/util/file"
)

const (
	authSecretTypeAnnotation = "auth-secret-type" //#nosec G101
	authRealmAnnotation      = "auth-realm"
	authTypeAnnotation       = "auth-type"
	// This should be exported as it is imported by other packages
	AuthSecretAnnotation = "auth-secret" //#nosec G101
)

var (
	authTypeRegex       = regexp.MustCompile(`basic|digest`)
	authSecretTypeRegex = regexp.MustCompile(`auth-file|auth-map`)

	// AuthDirectory default directory used to store files
	// to authenticate request
	AuthDirectory = "/etc/ingress-controller/auth"
)

var AuthSecretConfig = parser.AnnotationConfig{
	Validator:     parser.ValidateRegex(parser.BasicCharsRegex, true),
	Scope:         parser.AnnotationScopeLocation,
	Risk:          parser.AnnotationRiskMedium, // Medium as it allows a subset of chars
	Documentation: `This annotation defines the name of the Secret that contains the usernames and passwords which are granted access to the paths defined in the Ingress rules. `,
}

var authSecretAnnotations = parser.Annotation{
	Group: "authentication",
	Annotations: parser.AnnotationFields{
		AuthSecretAnnotation: AuthSecretConfig,
		authSecretTypeAnnotation: {
			Validator: parser.ValidateRegex(authSecretTypeRegex, true),
			Scope:     parser.AnnotationScopeLocation,
			Risk:      parser.AnnotationRiskLow,
			Documentation: `This annotation what is the format of auth-secret value. Can be "auth-file" that defines the content of an htpasswd file, or "auth-map" where each key
			is a user and each value is the password.`,
		},
		authRealmAnnotation: {
			Validator:     parser.ValidateRegex(parser.CharsWithSpace, false),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskMedium, // Medium as it allows a subset of chars
			Documentation: `This annotation defines the realm (message) that should be shown to user when authentication is requested.`,
		},
		authTypeAnnotation: {
			Validator:     parser.ValidateRegex(authTypeRegex, true),
			Scope:         parser.AnnotationScopeLocation,
			Risk:          parser.AnnotationRiskLow,
			Documentation: `This annotation defines the basic authentication type. Should be "basic" or "digest"`,
		},
	},
}

const (
	fileAuth = "auth-file"
	mapAuth  = "auth-map"
)

// Config returns authentication configuration for an Ingress rule
type Config struct {
	Type       string `json:"type"`
	Realm      string `json:"realm"`
	File       string `json:"file"`
	Secured    bool   `json:"secured"`
	FileSHA    string `json:"fileSha"`
	Secret     string `json:"secret"`
	SecretType string `json:"secretType"`
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
	return true
}

type auth struct {
	r                resolver.Resolver
	authDirectory    string
	annotationConfig parser.Annotation
}

// NewParser creates a new authentication annotation parser
func NewParser(authDirectory string, r resolver.Resolver) parser.IngressAnnotation {
	return auth{
		r:                r,
		authDirectory:    authDirectory,
		annotationConfig: authSecretAnnotations,
	}
}

// Parse parses the annotations contained in the ingress
// rule used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func (a auth) Parse(ing *networking.Ingress) (interface{}, error) {
	at, err := parser.GetStringAnnotation(authTypeAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		return nil, err
	}

	if !authTypeRegex.MatchString(at) {
		return nil, ing_errors.NewLocationDenied("invalid authentication type")
	}

	var secretType string
	secretType, err = parser.GetStringAnnotation(authSecretTypeAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		if ing_errors.IsValidationError(err) {
			return nil, err
		}
		secretType = fileAuth
	}

	s, err := parser.GetStringAnnotation(AuthSecretAnnotation, ing, a.annotationConfig.Annotations)
	if err != nil {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("error reading secret name from annotation: %w", err),
		}
	}

	sns, sname, err := cache.SplitMetaNamespaceKey(s)
	if err != nil {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("error reading secret name from annotation: %w", err),
		}
	}

	if sns == "" {
		sns = ing.Namespace
	}
	secCfg := a.r.GetSecurityConfiguration()
	// We don't accept different namespaces for secrets.
	if !secCfg.AllowCrossNamespaceResources && sns != ing.Namespace {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("cross namespace usage of secrets is not allowed"),
		}
	}

	name := fmt.Sprintf("%v/%v", sns, sname)
	secret, err := a.r.GetSecret(name)
	if err != nil {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("unexpected error reading secret %s: %w", name, err),
		}
	}

	realm, err := parser.GetStringAnnotation(authRealmAnnotation, ing, a.annotationConfig.Annotations)
	if ing_errors.IsValidationError(err) {
		return nil, err
	}

	passFileName := fmt.Sprintf("%v-%v-%v.passwd", ing.GetNamespace(), ing.UID, secret.UID)
	passFilePath := filepath.Join(a.authDirectory, passFileName)

	// Ensure password file name does not contain any path traversal characters.
	if a.authDirectory != filepath.Dir(passFilePath) || passFileName != filepath.Base(passFilePath) {
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("invalid password file name: %s", passFileName),
		}
	}

	switch secretType {
	case fileAuth:
		err = dumpSecretAuthFile(passFilePath, secret)
		if err != nil {
			return nil, err
		}
	case mapAuth:
		err = dumpSecretAuthMap(passFilePath, secret)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("invalid auth-secret-type in annotation, must be 'auth-file' or 'auth-map': %w", err),
		}
	}

	return &Config{
		Type:       at,
		Realm:      realm,
		File:       passFilePath,
		Secured:    true,
		FileSHA:    file.SHA1(passFilePath),
		Secret:     name,
		SecretType: secretType,
	}, nil
}

// dumpSecret dumps the content of a secret into a file
// in the expected format for the specified authorization
func dumpSecretAuthFile(filename string, secret *api.Secret) error {
	val, ok := secret.Data["auth"]
	if !ok {
		return ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("the secret %s does not contain a key with value auth", secret.Name),
		}
	}

	err := os.WriteFile(filename, val, file.ReadWriteByUser)
	if err != nil {
		return ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("unexpected error creating password file: %w", err),
		}
	}

	return nil
}

func dumpSecretAuthMap(filename string, secret *api.Secret) error {
	builder := &strings.Builder{}
	for user, pass := range secret.Data {
		builder.WriteString(user)
		builder.WriteString(":")
		builder.Write(pass)
		builder.WriteString("\n")
	}

	err := os.WriteFile(filename, []byte(builder.String()), file.ReadWriteByUser)
	if err != nil {
		return ing_errors.LocationDeniedError{
			Reason: fmt.Errorf("unexpected error creating password file: %w", err),
		}
	}

	return nil
}

func (a auth) GetDocumentation() parser.AnnotationFields {
	return a.annotationConfig.Annotations
}

func (a auth) Validate(anns map[string]string) error {
	maxrisk := parser.StringRiskToRisk(a.r.GetSecurityConfiguration().AnnotationsRiskLevel)
	return parser.CheckAnnotationRisk(anns, maxrisk, authSecretAnnotations.Annotations)
}
