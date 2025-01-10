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

package store

import (
	"fmt"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

type getEpssForServiceFunc = func(key string) ([]*discoveryv1.EndpointSlice, error)

// EndpointSliceLister makes a Store that lists Endpoints.
type EndpointSliceLister struct {
	cache.Store
	endpointSliceIndex getEpssForServiceFunc
}

// MatchByKey returns the EndpointsSlices of the Service matching key in the local Endpoint Store.
func (s *EndpointSliceLister) MatchByKey(key string) ([]*discoveryv1.EndpointSlice, error) {
	epss, err := s.endpointSliceIndex(key)
	if err != nil {
		return nil, err
	}

	if len(epss) == 0 {
		return nil, NotExistsError(key)
	}
	return epss, nil
}

func epssIndexer() cache.Indexers {
	return cache.Indexers{
		discoveryv1.LabelServiceName: func(obj interface{}) ([]string, error) {
			eps, ok := obj.(*discoveryv1.EndpointSlice)
			if !ok {
				// Skip object as it is not an endpointslice
				return nil, nil
			}

			parentService, ok := eps.Labels[discoveryv1.LabelServiceName]
			if !ok {
				// There is no parent service and thus we cannot match this endpointslice to any service
				// As far as i'm aware, this is only possible if you create epps objects by hand
				return nil, nil
			}

			key := fmt.Sprintf("%s/%s", eps.Namespace, parentService)

			return []string{key}, nil
		},
	}
}

func epssForServiceFuncFromIndexer(indexer cache.Indexer) getEppsForServiceFunc {
	return func(key string) ([]*discoveryv1.EndpointSlice, error) {
		objs, err := indexer.ByIndex(discoveryv1.LabelServiceName, key)
		if err != nil {
			return nil, err
		}

		epss := make([]*discoveryv1.EndpointSlice, 0, len(objs))
		for _, obj := range objs {
			eps, ok := obj.(*discoveryv1.EndpointSlice)
			if !ok {
				continue
			}

			epss = append(epss, eps)
		}

		return epss, nil
	}
}
