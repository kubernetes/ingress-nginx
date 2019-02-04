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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	ing_errors "k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: extensions.IngressSpec{
			Backend: &extensions.IngressBackend{
				ServiceName: "default-backend",
				ServicePort: intstr.FromInt(80),
			},
			Rules: []extensions.IngressRule{
				{
					Host: "foo.bar.com",
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
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
		Data: map[string][]byte{"auth": []byte("foo:$apr1$OFG3Xybp$ckL0FHDAkoXYIlH9.cysT0")},
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

func TestDumpSecret(t *testing.T) {
	tmpfile, dir, s := dummySecretContent(t)
	defer os.RemoveAll(dir)

	sd := s.Data
	s.Data = nil

	err := dumpSecret(tmpfile, s)
	if err == nil {
		t.Errorf("Expected error with secret without auth")
	}

	s.Data = sd
	err = dumpSecret(tmpfile, s)
	if err != nil {
		t.Errorf("Unexpected error creating htpasswd file %v: %v", tmpfile, err)
	}
}
