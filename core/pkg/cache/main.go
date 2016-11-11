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

import "k8s.io/kubernetes/pkg/client/cache"

// StoreToIngressLister makes a Store that lists Ingress.
type StoreToIngressLister struct {
	cache.Store
}

// StoreToSecretsLister makes a Store that lists Secrets.
type StoreToSecretsLister struct {
	cache.Store
}

// StoreToConfigmapLister makes a Store that lists Configmap.
type StoreToConfigmapLister struct {
	cache.Store
}
