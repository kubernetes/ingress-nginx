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

package storage

import (
	"k8s.io/kubernetes/pkg/client/cache"
)

// Snapshotter is an interface capable of providing a consistent snapshot of
// the underlying storage implementation of a pool. It does not guarantee
// thread safety of snapshots, so they should be treated as read only unless
// the implementation specifies otherwise.
type Snapshotter interface {
	Snapshot() map[string]interface{}
	cache.ThreadSafeStore
}

// InMemoryPool is used as a cache for cluster resource pools.
type InMemoryPool struct {
	cache.ThreadSafeStore
}

// Snapshot returns a read only copy of the k:v pairs in the store.
// Caller beware: Violates traditional snapshot guarantees.
func (p *InMemoryPool) Snapshot() map[string]interface{} {
	snap := map[string]interface{}{}
	for _, key := range p.ListKeys() {
		if item, ok := p.Get(key); ok {
			snap[key] = item
		}
	}
	return snap
}

// NewInMemoryPool creates an InMemoryPool.
func NewInMemoryPool() *InMemoryPool {
	return &InMemoryPool{
		cache.NewThreadSafeStore(cache.Indexers{}, cache.Indices{})}
}
