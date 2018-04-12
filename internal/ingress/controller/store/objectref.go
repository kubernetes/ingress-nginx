/*
Copyright 2018 The Kubernetes Authors.

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
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ObjectRefMap is a map of references from object(s) to object (1:n). It is
// used to keep track of which data objects (Secrets) are used within Ingress
// objects.
type ObjectRefMap interface {
	Insert(consumer string, ref ...string)
	Delete(consumer string)
	Len() int
	Has(ref string) bool
	HasConsumer(consumer string) bool
	Reference(ref string) []string
	ReferencedBy(consumer string) []string
}

type objectRefMap struct {
	sync.Mutex
	v map[string]sets.String
}

// NewObjectRefMap returns a new ObjectRefMap.
func NewObjectRefMap() ObjectRefMap {
	return &objectRefMap{
		v: make(map[string]sets.String),
	}
}

// Insert adds a consumer to one or more referenced objects.
func (o *objectRefMap) Insert(consumer string, ref ...string) {
	o.Lock()
	defer o.Unlock()

	for _, r := range ref {
		if _, ok := o.v[r]; !ok {
			o.v[r] = sets.NewString(consumer)
			continue
		}
		o.v[r].Insert(consumer)
	}
}

// Delete deletes a consumer from all referenced objects.
func (o *objectRefMap) Delete(consumer string) {
	o.Lock()
	defer o.Unlock()

	for ref, consumers := range o.v {
		consumers.Delete(consumer)
		if consumers.Len() == 0 {
			delete(o.v, ref)
		}
	}
}

// Len returns the count of referenced objects.
func (o *objectRefMap) Len() int {
	return len(o.v)
}

// Has returns whether the given object is referenced by any other object.
func (o *objectRefMap) Has(ref string) bool {
	o.Lock()
	defer o.Unlock()

	if _, ok := o.v[ref]; ok {
		return true
	}
	return false
}

// HasConsumer returns whether the store contains the given consumer.
func (o *objectRefMap) HasConsumer(consumer string) bool {
	o.Lock()
	defer o.Unlock()

	for _, consumers := range o.v {
		if consumers.Has(consumer) {
			return true
		}
	}
	return false
}

// Reference returns all objects referencing the given object.
func (o *objectRefMap) Reference(ref string) []string {
	o.Lock()
	defer o.Unlock()

	consumers, ok := o.v[ref]
	if !ok {
		return make([]string, 0)
	}
	return consumers.List()
}

// ReferencedBy returns all objects referenced by the given object.
func (o *objectRefMap) ReferencedBy(consumer string) []string {
	o.Lock()
	defer o.Unlock()

	refs := make([]string, 0)
	for ref, consumers := range o.v {
		if consumers.Has(consumer) {
			refs = append(refs, ref)
		}
	}
	return refs
}
