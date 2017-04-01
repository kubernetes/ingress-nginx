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

package storage

import (
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
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

type keyFunc func(interface{}) (string, error)

type cloudLister interface {
	List() ([]interface{}, error)
}

// CloudListingPool wraps InMemoryPool but relists from the cloud periodically.
type CloudListingPool struct {
	// A lock to protect against concurrent mutation of the pool
	lock sync.Mutex
	// The pool that is re-populated via re-list from cloud, and written to
	// from controller
	*InMemoryPool
	// An interface that lists objects from the cloud.
	lister cloudLister
	// A function capable of producing a key for a given object.
	// This key must match the key used to store the same object in the user of
	// this cache.
	keyGetter keyFunc
}

// ReplenishPool lists through the cloudLister and inserts into the pool. This
// is especially useful in scenarios like deleting an Ingress while the
// controller is restarting. As long as the resource exists in the shared
// memory pool, it is visible to the caller and they can take corrective
// actions, eg: backend pool deletes backends with non-matching node ports
// in its sync method.
func (c *CloudListingPool) ReplenishPool() {
	c.lock.Lock()
	defer c.lock.Unlock()
	glog.V(4).Infof("Replenishing pool")

	// We must list with the lock, because the controller also lists through
	// Snapshot(). It's ok if the controller takes a snpshot, we list, we
	// delete, because we have delete based on the most recent state. Worst
	// case we thrash. It's not ok if we list, the controller lists and
	// creates a backend, and we delete that backend based on stale state.
	items, err := c.lister.List()
	if err != nil {
		glog.Warningf("Failed to list: %v", err)
		return
	}

	for i := range items {
		key, err := c.keyGetter(items[i])
		if err != nil {
			glog.V(5).Infof("CloudListingPool: %v", err)
			continue
		}
		c.InMemoryPool.Add(key, items[i])
	}
}

// Snapshot just snapshots the underlying pool.
func (c *CloudListingPool) Snapshot() map[string]interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.InMemoryPool.Snapshot()
}

// Add simply adds to the underlying pool.
func (c *CloudListingPool) Add(key string, obj interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.InMemoryPool.Add(key, obj)
}

// Delete just deletes from underlying pool.
func (c *CloudListingPool) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.InMemoryPool.Delete(key)
}

// NewCloudListingPool replenishes the InMemoryPool through a background
// goroutine that lists from the given cloudLister.
func NewCloudListingPool(k keyFunc, lister cloudLister, relistPeriod time.Duration) *CloudListingPool {
	cl := &CloudListingPool{
		InMemoryPool: NewInMemoryPool(),
		lister:       lister,
		keyGetter:    k,
	}
	glog.V(4).Infof("Starting pool replenish goroutine")
	go wait.Until(cl.ReplenishPool, relistPeriod, make(chan struct{}))
	return cl
}
