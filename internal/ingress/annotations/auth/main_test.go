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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *networking.Ingress {
	defaultBackend := networking.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &networking.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: networking.IngressSpec{
			Backend: &networking.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []networking.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:    "/foo",
									Backend: defaultBackend,
								},
							},
						},
					},
				},
			},
		},
	}
}

type mockSecret struct {
	resolver.Mock
}

func (m mockSecret) GetSecret(name string) (*api.Secret, error) {
	if name != "default/demo-secret" {
		return nil, errors.Errorf("there is no secret with name %v", name)
	}

	return &api.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo-secret",
		},
		Data: map[string][]byte{"auth": []byte("bXlsaXR0bGVzZWNyZXQK")},
	}, nil
}

func TestIngressWithoutAuth(t *testing.T) {
	ing := buildIngress()
	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)
	_, err := NewParser(dir, &mockSecret{}).Parse(ing)
	if err == nil {
		t.Error("Expected error with ingress without annotations")
	}
}

func TestIngressAuthBadAuthType(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("auth-type")] = "invalid"
	ing.SetAnnotations(data)

	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)

	expected := ing_errors.NewLocationDenied("invalid authentication type")
	_, err := NewParser(dir, &mockSecret{}).Parse(ing)
	if err.Error() != expected.Error() {
		t.Errorf("expected '%v' but got '%v'", expected, err)
	}
}

func TestInvalidIngressAuthNoSecret(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("auth-type")] = "basic"
	ing.SetAnnotations(data)

	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)

	expected := ing_errors.LocationDenied{
		Reason: errors.New("error reading secret name from annotation: ingress rule without annotations"),
	}
	_, err := NewParser(dir, &mockSecret{}).Parse(ing)
	if err.Error() != expected.Reason.Error() {
		t.Errorf("expected '%v' but got '%v'", expected, err)
	}
}

func TestIngressAuth(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("auth-type")] = "basic"
	data[parser.GetAnnotationWithPrefix("auth-secret")] = "demo-secret"
	data[parser.GetAnnotationWithPrefix("auth-realm")] = "-realm-"
	ing.SetAnnotations(data)

	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)

	i, err := NewParser(dir, &mockSecret{}).Parse(ing)
	if err != nil {
		t.Errorf("Uxpected error with ingress: %v", err)
	}
	auth, ok := i.(*Config)
	if !ok {
		t.Errorf("expected a BasicDigest type")
	}
	if auth.Type != "basic" {
		t.Errorf("Expected basic as auth type but returned %s", auth.Type)
	}
	if auth.Realm != "-realm-" {
		t.Errorf("Expected -realm- as realm but returned %s", auth.Realm)
	}
	if !auth.Secured {
		t.Errorf("Expected true as secured but returned %v", auth.Secured)
	}
}

func TestIngressAuthWithoutSecret(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("auth-type")] = "basic"
	data[parser.GetAnnotationWithPrefix("auth-secret")] = "invalid-secret"
	data[parser.GetAnnotationWithPrefix("auth-realm")] = "-realm-"
	ing.SetAnnotations(data)

	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)

	_, err := NewParser(dir, mockSecret{}).Parse(ing)
	if err == nil {
		t.Errorf("expected an error with invalid secret name")
	}
}

func TestIngressAuthInvalidSecretKey(t *testing.T) {
	ing := buildIngress()

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix("auth-type")] = "basic"
	data[parser.GetAnnotationWithPrefix("auth-secret")] = "demo-secret"
	data[parser.GetAnnotationWithPrefix("auth-secret-type")] = "invalid-type"
	data[parser.GetAnnotationWithPrefix("auth-realm")] = "-realm-"
	ing.SetAnnotations(data)

	_, dir, _ := dummySecretContent(t)
	defer os.RemoveAll(dir)

	_, err := NewParser(dir, mockSecret{}).Parse(ing)
	if err == nil {
		t.Errorf("expected an error with invalid secret name")
	}
}

func dummySecretContent(t *testing.T) (string, string, *api.Secret) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%v", time.Now().Unix()))
	if err != nil {
		t.Error(err)
	}

	tmpfile, err := ioutil.TempFile("", "example-")
	if err != nil {
		t.Error(err)
	}
	defer tmpfile.Close()
	s, _ := mockSecret{}.GetSecret("default/demo-secret")
	return tmpfile.Name(), dir, s
}

func TestDumpSecretAuthFile(t *testing.T) {
	tmpfile, dir, s := dummySecretContent(t)
	defer os.RemoveAll(dir)

	sd := s.Data
	s.Data = nil

	err := dumpSecretAuthFile(tmpfile, s)
	if err == nil {
		t.Errorf("Expected error with secret without auth")
	}

	s.Data = sd
	err = dumpSecretAuthFile(tmpfile, s)
	if err != nil {
		t.Errorf("Unexpected error creating htpasswd file %v: %v", tmpfile, err)
	}
}

func TestDumpSecretAuthMap(t *testing.T) {
	tmpfile, dir, s := dummySecretContent(t)
	defer os.RemoveAll(dir)

	err := dumpSecretAuthMap(tmpfile, s)
	if err != nil {
		t.Errorf("Unexpected error creating htpasswd file %v: %v", tmpfile, err)
	}
}

func TestDumpSecretAuthBuffer(t *testing.T) {
	buf := make([]byte, 0)
	testBuf := bytes.NewBuffer(buf)
	secret := &api.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo-secret",
		},
		Data: map[string][]byte{"auth": []byte("bXlsaXR0bGVzZWNyZXQK")},
	}
	dumpSecretAuthBuffer(testBuf, secret)
	var inputString string
	for key, value := range secret.Data {
		inputString = key + ":" + string(value) + "\n"
	}
	assert.Equal(t, inputString, testBuf.String())
}

func TestDumpCMAuthBuffer(t *testing.T) {
	testCases := map[string]struct {
		configMap      api.ConfigMap
		expectedCMData string
	}{
		"Single key-value": {
			configMap: api.ConfigMap{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: api.NamespaceDefault,
					Name:      "demo-config",
				},
				Data: map[string]string{
					"bar": "default",
				},
			},
			expectedCMData: "bar:ZGVmYXVsdA==\n",
		},
		"Single key-value with multiline value": {
			configMap: api.ConfigMap{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: api.NamespaceDefault,
					Name:      "demo-config",
				},
				Data: map[string]string{
					"bar": "first:default\nsecond:non-default\nthird:different",
				},
			},
			expectedCMData: "bar:Zmlyc3Q6ZGVmYXVsdApzZWNvbmQ6bm9uLWRlZmF1bHQKdGhpcmQ6ZGlmZmVyZW50\n",
		},
		"Multiple key-value": {
			configMap: api.ConfigMap{
				ObjectMeta: meta_v1.ObjectMeta{
					Namespace: api.NamespaceDefault,
					Name:      "demo-config",
				},
				Data: map[string]string{
					"bar": "default",
					"foo": "non-default",
				},
			},
			expectedCMData: "bar:ZGVmYXVsdA==\nfoo:bm9uLWRlZmF1bHQ=\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(*testing.T) {
			buf := make([]byte, 0)
			testBuf := bytes.NewBuffer(buf)
			dumpCMAuthBuffer(testBuf, &tc.configMap)
			if len(tc.configMap.Data) == 1 {
				assert.Equal(t, tc.expectedCMData, testBuf.String())
			} else {
				for key, inputData := range tc.configMap.Data {
					base64InputData := base64.StdEncoding.EncodeToString([]byte(inputData))
					assert.True(t, strings.Contains(tc.expectedCMData, key+":"+base64InputData+"\n"))
				}

			}
		})

	}

}

func (m mockSecret) GetConfigMap(name string) (*api.ConfigMap, error) {
	if name != "default/demo-config" {
		return nil, errors.Errorf("there is no configmap with name %v", name)
	}

	return &api.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo-config",
		},
		Data: map[string]string{
			"bar":     "default",
			"foo":     "non-default",
			"whatnot": "",
		},
	}, nil
}

func TestOIDCPluginAuth(t *testing.T) {
	testCases := map[string]struct {
		annotations        map[string]string
		expectedSecret     string
		expectedSecretData string
		expectedCMName     string
		expectedCMData     []string
	}{
		"Basic case": {
			annotations: map[string]string{
				parser.GetAnnotationWithPrefix("auth-type"):   "oidc",
				parser.GetAnnotationWithPrefix("auth-secret"): "demo-secret",
				parser.GetAnnotationWithPrefix("auth-config"): "demo-config",
			},
			expectedSecret:     "default/demo-secret",
			expectedSecretData: "auth:bXlsaXR0bGVzZWNyZXQK\n",
			expectedCMName:     "default/demo-config",
			expectedCMData: []string{
				"whatnot:\n",
				"bar:ZGVmYXVsdA==\n",
				"foo:bm9uLWRlZmF1bHQ=\n",
			},
		},
		"Only secret is defined": {
			annotations: map[string]string{
				parser.GetAnnotationWithPrefix("auth-type"):   "oidc",
				parser.GetAnnotationWithPrefix("auth-secret"): "demo-secret",
			},
			expectedSecret:     "default/demo-secret",
			expectedSecretData: "auth:bXlsaXR0bGVzZWNyZXQK\n",
			expectedCMName:     "",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(*testing.T) {
			ing := buildIngress()
			ing.SetAnnotations(tc.annotations)
			config, err := NewParser("", &mockSecret{}).Parse(ing)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCMName, config.(*Config).ConfigMap)
			assert.Equal(t, tc.expectedSecret, config.(*Config).Secret)
			assert.Equal(t, tc.expectedSecretData, config.(*Config).SecretData)
			if len(tc.expectedCMData) != 0 {
				for _, expectedCMData := range tc.expectedCMData {
					assert.True(t, strings.Contains(config.(*Config).ConfigMapData, expectedCMData))
				}
			} else {
				assert.Empty(t, config.(*Config).ConfigMapData)
			}
		})
	}
}
