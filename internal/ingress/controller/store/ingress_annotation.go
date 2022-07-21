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

package store

import (
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

// IngressWithAnnotationsLister makes a Store that lists Ingress rules with annotations already parsed
type IngressWithAnnotationsLister struct {
	cache.Store
}

// ByKey returns the Ingress with annotations matching key in the local store or an error
func (il IngressWithAnnotationsLister) ByKey(key string) (*ingress.Ingress, error) {
	i, exists, err := il.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return i.(*ingress.Ingress), nil
}
