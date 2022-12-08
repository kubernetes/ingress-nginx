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
	"fmt"
	"testing"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func newEndpointSliceLister(t *testing.T) *EndpointSliceLister {
	t.Helper()

	return &EndpointSliceLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
}

func TestEndpointSliceLister(t *testing.T) {
	t.Run("the key does not exist", func(t *testing.T) {
		el := newEndpointSliceLister(t)

		key := "namespace/svcname"
		_, err := el.MatchByKey(key)

		if err == nil {
			t.Error("expected an error but nothing has been returned")
		}

		if _, ok := err.(NotExistsError); !ok {
			t.Errorf("expected NotExistsError, got %v", err)
		}
	})
	t.Run("the key exists", func(t *testing.T) {
		el := newEndpointSliceLister(t)

		key := "namespace/svcname"
		endpointSlice := &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      "anothername-foo",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "svcname",
				},
			},
		}
		el.Add(endpointSlice)
		endpointSlice = &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      "svcname-bar",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "othersvc",
				},
			},
		}
		el.Add(endpointSlice)
		endpointSlice = &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      "svcname-buz",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "svcname",
				},
			},
		}
		el.Add(endpointSlice)
		eps, err := el.MatchByKey(key)

		if err != nil {
			t.Errorf("unexpeted error %v", err)
		}
		if err == nil && len(eps) != 1 {
			t.Errorf("expected one slice %v, error, got %d slices", endpointSlice, len(eps))
		}
		if len(eps) > 0 && eps[0].GetName() != endpointSlice.GetName() {
			t.Errorf("expected %v, error, got %v", endpointSlice.GetName(), eps[0].GetName())
		}
	})
	t.Run("svc long name", func(t *testing.T) {
		el := newEndpointSliceLister(t)
		ns := "namespace"
		ns2 := "another-ns"
		svcName := "test-backend-http-test-http-test-http-test-http-test-http-truncated"
		svcName2 := "another-long-svc-name-for-test-inhttp-test-http-test-http-truncated"
		key := fmt.Sprintf("%s/%s", ns, svcName)
		endpointSlice := &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      "test-backend-http-test-http-test-http-test-http-test-http-bar88",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: svcName,
				},
			},
		}
		el.Add(endpointSlice)
		endpointSlice2 := &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns2,
				Name:      "another-long-svc-name-for-test-inhttp-test-http-test-http-bar88",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: svcName2,
				},
			},
		}
		el.Add(endpointSlice2)
		eps, err := el.MatchByKey(key)
		if err != nil {
			t.Errorf("unexpeted error %v", err)
		}
		if len(eps) != 1 {
			t.Errorf("expected one slice %v, error, got %d slices", endpointSlice, len(eps))
		}
		if len(eps) == 1 && eps[0].Labels[discoveryv1.LabelServiceName] != svcName {
			t.Errorf("expected slice %v, error, got %v slices", endpointSlice, eps[0])
		}
	})
}
