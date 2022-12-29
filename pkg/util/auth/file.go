/*
Copyright 2022 The Kubernetes Authors.

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
	"crypto/sha1" // #nosec
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	api "k8s.io/api/core/v1"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/pkg/util/file"
)

// DumpSecretAuthFile dumps the content of a secret into a file
// in the expected format for the specified authorization
func DumpSecretAuthFile(filename string, secret *api.Secret) ([]byte, error) {
	secretVal, ok := secret.Data["auth"]
	if !ok {
		return nil, ing_errors.LocationDenied{
			Reason: fmt.Errorf("the secret %s does not contain a key with value auth", secret.Name),
		}
	}

	// TODO: Stop writing secret locally
	if err := WriteSecretFile(filename, secretVal); err != nil {
		return nil, err
	}
	return secretVal, nil
}

// DumpSecretAuthFile dumps the content of a secret into a file
// with each key being the user and the value as password
func DumpSecretAuthMap(filename string, secret *api.Secret) ([]byte, error) {
	builder := &strings.Builder{}
	for user, pass := range secret.Data {
		builder.WriteString(user)
		builder.WriteString(":")
		builder.WriteString(string(pass))
		builder.WriteString("\n")
	}

	secretVal := []byte(builder.String())

	if err := WriteSecretFile(filename, secretVal); err != nil {
		return nil, err
	}
	return secretVal, nil
}

func WriteSecretFile(filename string, secretVal []byte) error {
	err := os.WriteFile(filename, secretVal, file.ReadWriteByUser)
	if err != nil {
		return fmt.Errorf("unexpected error creating password file: %w", err)
	}
	return nil
}

func SecretSHA1(content []byte) string {
	hasher := sha1.New() // #nosec
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}
