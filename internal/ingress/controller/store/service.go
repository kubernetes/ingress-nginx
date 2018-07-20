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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// ServiceLister makes a Store that lists Services.
type ServiceLister struct {
	cache.Store
}

// ByKey returns the Service matching key in the local Service Store.
func (sl *ServiceLister) ByKey(key string) (*apiv1.Service, error) {
	s, exists, err := sl.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NotExistsError(key)
	}
	return s.(*apiv1.Service), nil
}
