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

package k8s

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	testclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
)

func TestParseNameNS(t *testing.T) {

	tests := []struct {
		title  string
		input  string
		ns     string
		name   string
		expErr bool
	}{
		{"empty string", "", "", "", true},
		{"demo", "demo", "", "", true},
		{"kube-system", "kube-system", "", "", true},
		{"default/kube-system", "default/kube-system", "default", "kube-system", false},
	}

	for _, test := range tests {
		ns, name, err := ParseNameNS(test.input)
		if test.expErr {
			if err == nil {
				t.Errorf("%v: expected error but retuned nil", test.title)
			}
			continue
		}
		if test.ns != ns {
			t.Errorf("%v: expected %v but retuned %v", test.title, test.ns, ns)
		}
		if test.name != name {
			t.Errorf("%v: expected %v but retuned %v", test.title, test.name, name)
		}
	}
}

func TestIsValidService(t *testing.T) {
	fk := testclient.NewSimpleClientset(&api.Service{
		ObjectMeta: api.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo",
		},
	})

	_, err := IsValidService(fk, "")
	if err == nil {
		t.Errorf("expected error but retuned nil")
	}
	s, err := IsValidService(fk, "default/demo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s == nil {
		t.Errorf("expected a Service but retuned nil")
	}

	fk = testclient.NewSimpleClientset()
	s, err = IsValidService(fk, "default/demo")
	if err == nil {
		t.Errorf("expected an error but retuned nil")
	}
	if s != nil {
		t.Errorf("unexpected Service returned: %v", s)
	}
}

func TestIsValidSecret(t *testing.T) {
	fk := testclient.NewSimpleClientset(&api.Secret{
		ObjectMeta: api.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo",
		},
	})

	_, err := IsValidSecret(fk, "")
	if err == nil {
		t.Errorf("expected error but retuned nil")
	}
	s, err := IsValidSecret(fk, "default/demo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s == nil {
		t.Errorf("expected a Secret but retuned nil")
	}

	fk = testclient.NewSimpleClientset()
	s, err = IsValidSecret(fk, "default/demo")
	if err == nil {
		t.Errorf("expected an error but retuned nil")
	}
	if s != nil {
		t.Errorf("unexpected Secret returned: %v", s)
	}
}
