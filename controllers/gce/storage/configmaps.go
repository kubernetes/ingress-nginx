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
	"fmt"
	"strings"
	"sync"

	"github.com/golang/glog"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	// UidDataKey is the key used in config maps to store the UID.
	UidDataKey = "uid"
	// ProviderDataKey is the key used in config maps to store the Provider
	// UID which we use to ensure unique firewalls.
	ProviderDataKey = "provider-uid"
)

// ConfigMapVault stores cluster UIDs in config maps.
// It's a layer on top of ConfigMapStore that just implements the utils.uidVault
// interface.
type ConfigMapVault struct {
	storeLock      sync.Mutex
	ConfigMapStore cache.Store
	namespace      string
	name           string
}

// Get retrieves the value associated to the provided 'key' from the cluster config map.
// If this method returns an error, it's guaranteed to be apiserver flake.
// If the error is a not found error it sets the boolean to false and
// returns and error of nil instead.
func (c *ConfigMapVault) Get(key string) (string, bool, error) {
	keyStore := fmt.Sprintf("%v/%v", c.namespace, c.name)
	item, found, err := c.ConfigMapStore.GetByKey(keyStore)
	if err != nil || !found {
		return "", false, err
	}
	data := item.(*api_v1.ConfigMap).Data
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	if k, ok := data[key]; ok {
		return k, true, nil
	}
	glog.Infof("Found config map %v but it doesn't contain key %v: %+v", keyStore, key, data)
	return "", false, nil
}

// Put inserts a key/value pair in the cluster config map.
// If the key already exists, the value provided is stored.
func (c *ConfigMapVault) Put(key, val string) error {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()
	apiObj := &api_v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
	}
	cfgMapKey := fmt.Sprintf("%v/%v", c.namespace, c.name)

	item, exists, err := c.ConfigMapStore.GetByKey(cfgMapKey)
	if err == nil && exists {
		data := item.(*api_v1.ConfigMap).Data
		existingVal, ok := data[key]
		if ok && existingVal == val {
			// duplicate, no need to update.
			return nil
		}
		data[key] = val
		apiObj.Data = data
		if existingVal != val {
			glog.Infof("Configmap %v has key %v but wrong value %v, updating to %v", cfgMapKey, key, existingVal, val)
		} else {
			glog.Infof("Configmap %v will be updated with %v = %v", cfgMapKey, key, val)
		}
		if err := c.ConfigMapStore.Update(apiObj); err != nil {
			return fmt.Errorf("failed to update %v: %v", cfgMapKey, err)
		}
	} else {
		apiObj.Data = map[string]string{key: val}
		if err := c.ConfigMapStore.Add(apiObj); err != nil {
			return fmt.Errorf("failed to add %v: %v", cfgMapKey, err)
		}
	}
	glog.Infof("Successfully stored key %v = %v in config map %v", key, val, cfgMapKey)
	return nil
}

// Delete deletes the ConfigMapStore.
func (c *ConfigMapVault) Delete() error {
	cfgMapKey := fmt.Sprintf("%v/%v", c.namespace, c.name)
	item, _, err := c.ConfigMapStore.GetByKey(cfgMapKey)
	if err == nil {
		return c.ConfigMapStore.Delete(item)
	}
	glog.Warningf("Couldn't find item %v in vault, unable to delete", cfgMapKey)
	return nil
}

// NewConfigMapVault creates a config map client.
// This client is essentially meant to abstract out the details of
// configmaps and the API, and just store/retrieve a single value, the cluster uid.
func NewConfigMapVault(c kubernetes.Interface, uidNs, uidConfigMapName string) *ConfigMapVault {
	return &ConfigMapVault{
		ConfigMapStore: NewConfigMapStore(c),
		namespace:      uidNs,
		name:           uidConfigMapName}
}

// NewFakeConfigMapVault is an implementation of the ConfigMapStore that doesn't
// persist configmaps. Only used in testing.
func NewFakeConfigMapVault(ns, name string) *ConfigMapVault {
	return &ConfigMapVault{
		ConfigMapStore: cache.NewStore(cache.MetaNamespaceKeyFunc),
		namespace:      ns,
		name:           name}
}

// ConfigMapStore wraps the store interface. Implementations usually persist
// contents of the store transparently.
type ConfigMapStore interface {
	cache.Store
}

// APIServerConfigMapStore only services Add and GetByKey from apiserver.
// TODO: Implement all the other store methods and make this a write
// through cache.
type APIServerConfigMapStore struct {
	ConfigMapStore
	client kubernetes.Interface
}

// Add adds the given config map to the apiserver's store.
func (a *APIServerConfigMapStore) Add(obj interface{}) error {
	cfg := obj.(*api_v1.ConfigMap)
	_, err := a.client.Core().ConfigMaps(cfg.Namespace).Create(cfg)
	return err
}

// Update updates the existing config map object.
func (a *APIServerConfigMapStore) Update(obj interface{}) error {
	cfg := obj.(*api_v1.ConfigMap)
	_, err := a.client.Core().ConfigMaps(cfg.Namespace).Update(cfg)
	return err
}

// Delete deletes the existing config map object.
func (a *APIServerConfigMapStore) Delete(obj interface{}) error {
	cfg := obj.(*api_v1.ConfigMap)
	return a.client.Core().ConfigMaps(cfg.Namespace).Delete(cfg.Name, &metav1.DeleteOptions{})
}

// GetByKey returns the config map for a given key.
// The key must take the form namespace/name.
func (a *APIServerConfigMapStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	nsName := strings.Split(key, "/")
	if len(nsName) != 2 {
		return nil, false, fmt.Errorf("failed to get key %v, unexpecte format, expecting ns/name", key)
	}
	ns, name := nsName[0], nsName[1]
	cfg, err := a.client.Core().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		// Translate not found errors to found=false, err=nil
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return cfg, true, nil
}

// NewConfigMapStore returns a config map store capable of persisting updates
// to apiserver.
func NewConfigMapStore(c kubernetes.Interface) ConfigMapStore {
	return &APIServerConfigMapStore{ConfigMapStore: cache.NewStore(cache.MetaNamespaceKeyFunc), client: c}
}
