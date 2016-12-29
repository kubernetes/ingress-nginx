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

package cache

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/util/sets"
)

func TestStoreToIngressLister(t *testing.T) {
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	ids := sets.NewString("foo", "bar", "baz")
	for id := range ids {
		store.Add(&extensions.Ingress{ObjectMeta: api.ObjectMeta{Name: id}})
	}
	sml := StoreToIngressLister{store}

	gotIngress := sml.List()
	got := make([]string, len(gotIngress))
	for ix := range gotIngress {
		ing, ok := gotIngress[ix].(*extensions.Ingress)
		if !ok {
			t.Errorf("expected an Ingress type")
		}
		got[ix] = ing.Name
	}
	if !ids.HasAll(got...) || len(got) != len(ids) {
		t.Errorf("expected %v, got %v", ids, got)
	}
}

func TestStoreToSecretsLister(t *testing.T) {
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	ids := sets.NewString("foo", "bar", "baz")
	for id := range ids {
		store.Add(&api.Secret{ObjectMeta: api.ObjectMeta{Name: id}})
	}
	sml := StoreToSecretsLister{store}

	gotIngress := sml.List()
	got := make([]string, len(gotIngress))
	for ix := range gotIngress {
		s, ok := gotIngress[ix].(*api.Secret)
		if !ok {
			t.Errorf("expected a Secret type")
		}
		got[ix] = s.Name
	}
	if !ids.HasAll(got...) || len(got) != len(ids) {
		t.Errorf("expected %v, got %v", ids, got)
	}
}

func TestStoreToConfigmapLister(t *testing.T) {
	store := cache.NewStore(cache.MetaNamespaceKeyFunc)
	ids := sets.NewString("foo", "bar", "baz")
	for id := range ids {
		store.Add(&api.ConfigMap{ObjectMeta: api.ObjectMeta{Name: id}})
	}
	sml := StoreToConfigmapLister{store}

	gotIngress := sml.List()
	got := make([]string, len(gotIngress))
	for ix := range gotIngress {
		m, ok := gotIngress[ix].(*api.ConfigMap)
		if !ok {
			t.Errorf("expected an Ingress type")
		}
		got[ix] = m.Name
	}
	if !ids.HasAll(got...) || len(got) != len(ids) {
		t.Errorf("expected %v, got %v", ids, got)
	}
}
