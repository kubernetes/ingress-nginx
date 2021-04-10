/*
Copyright 2017 The Kubernetes Authors.

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

package store

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"testing"
)

func newEndpointLister(t *testing.T) *EndpointLister {
	t.Helper()

	return &EndpointLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
}

func TestEndpointLister(t *testing.T) {
	t.Run("the key does not exist", func(t *testing.T) {
		el := newEndpointLister(t)

		key := "namespace/endpoint"
		_, err := el.ByKey(key)

		if err == nil {
			t.Error("expected an error but nothing has been returned")
		}

		if _, ok := err.(NotExistsError); !ok {
			t.Errorf("expected NotExistsError, got %v", err)
		}
	})

	t.Run("the key exists", func(t *testing.T) {
		el := newEndpointLister(t)

		key := "namespace/endpoint"
		endpoint := &apiv1.Endpoints{ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "endpoint"}}

		el.Add(endpoint)

		e, err := el.ByKey(key)

		if err != nil {
			t.Errorf("unexpeted error %v", err)
		}

		if e != endpoint {
			t.Errorf("expected %v, error, got %v", e, endpoint)
		}
	})
}
