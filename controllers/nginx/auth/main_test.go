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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func buildIngress() *extensions.Ingress {
	defaultBackend := extensions.IngressBackend{
		ServiceName: "default-backend",
		ServicePort: intstr.FromInt(80),
	}

	return &extensions.Ingress{
		ObjectMeta: api.ObjectMeta{
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

type secretsClient struct {
	unversioned.Interface
}

func mockClient() *testclient.Fake {
	secretObj := &api.Secret{
		ObjectMeta: api.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo-secret",
		},
	}

	return testclient.NewSimpleFake(secretObj)
}

func TestAnnotations(t *testing.T) {
	ing := buildIngress()

	_, err := ingAnnotations(ing.GetAnnotations()).authType()
	if err == nil {
		t.Error("Expected a validation error")
	}
	realm := ingAnnotations(ing.GetAnnotations()).realm()
	if realm != defAuthRealm {
		t.Error("Expected default realm")
	}

	_, err = ingAnnotations(ing.GetAnnotations()).secretName()
	if err == nil {
		t.Error("Expected a validation error")
	}

	data := map[string]string{}
	data[authType] = "demo"
	data[authSecret] = "demo-secret"
	data[authRealm] = "demo"
	ing.SetAnnotations(data)

	_, err = ingAnnotations(ing.GetAnnotations()).authType()
	if err == nil {
		t.Error("Expected a validation error")
	}

	realm = ingAnnotations(ing.GetAnnotations()).realm()
	if realm != "demo" {
		t.Errorf("Expected demo as realm but returned %s", realm)
	}

	secret, err := ingAnnotations(ing.GetAnnotations()).secretName()
	if err != nil {
		t.Error("Unexpec error %v", err)
	}
	if secret != "demo-secret" {
		t.Errorf("Expected demo-secret as realm but returned %s", secret)
	}
}

func TestIngressWithoutAuth(t *testing.T) {
	ing := buildIngress()
	client := mockClient()
	_, err := Parse(client, ing)
	if err == nil {
		t.Error("Expected error with ingress without annotations")
	}

	if err != ErrMissingAuthType {
		t.Errorf("Expected MissingAuthType error but returned %v", err)
	}
}

func TestIngressBasicAuth(t *testing.T) {
}

func TestIngressDigestAuth(t *testing.T) {
}

func TestIngressMissingAnnotation(t *testing.T) {
}
