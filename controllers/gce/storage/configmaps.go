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

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/cache"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// UIDVault stores UIDs.
type UIDVault interface {
	Get() (string, bool, error)
	Put(string) error
	Delete() error
}

// uidDataKey is the key used in config maps to store the UID.
const uidDataKey = "uid"

// ConfigMapVault stores cluster UIDs in config maps.
// It's a layer on top of ConfigMapStore that just implements the utils.uidVault
// interface.
type ConfigMapVault struct {
	ConfigMapStore cache.Store
	namespace      string
	name           string
}

// Get retrieves the cluster UID from the cluster config map.
// If this method returns an error, it's guaranteed to be apiserver flake.
// If the error is a not found error it sets the boolean to false and
// returns and error of nil instead.
func (c *ConfigMapVault) Get() (string, bool, error) {
	key := fmt.Sprintf("%v/%v", c.namespace, c.name)
	item, found, err := c.ConfigMapStore.GetByKey(key)
	if err != nil || !found {
		return "", false, err
	}
	cfg := item.(*api.ConfigMap)
	if k, ok := cfg.Data[uidDataKey]; ok {
		return k, true, nil
	}
	return "", false, fmt.Errorf("Found config map %v but it doesn't contain uid key: %+v", key, cfg.Data)
}

// Put stores the given UID in the cluster config map.
func (c *ConfigMapVault) Put(uid string) error {
	apiObj := &api.ConfigMap{
		ObjectMeta: api.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Data: map[string]string{uidDataKey: uid},
	}
	cfgMapKey := fmt.Sprintf("%v/%v", c.namespace, c.name)

	item, exists, err := c.ConfigMapStore.GetByKey(cfgMapKey)
	if err == nil && exists {
		data := item.(*api.ConfigMap).Data
		if k, ok := data[uidDataKey]; ok && k == uid {
			return nil
		} else if ok {
			glog.Infof("Configmap %v has key %v but wrong value %v, updating", cfgMapKey, k, uid)
		}

		if err := c.ConfigMapStore.Update(apiObj); err != nil {
			return fmt.Errorf("Failed to update %v: %v", cfgMapKey, err)
		}
	} else if err := c.ConfigMapStore.Add(apiObj); err != nil {
		return fmt.Errorf("Failed to add %v: %v", cfgMapKey, err)
	}
	glog.Infof("Successfully stored uid %q in config map %v", uid, cfgMapKey)
	return nil
}

// Delete deletes the cluster UID storing config map.
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
func NewConfigMapVault(c *client.Client, uidNs, uidConfigMapName string) *ConfigMapVault {
	return &ConfigMapVault{NewConfigMapStore(c), uidNs, uidConfigMapName}
}

// NewFakeConfigMapVault is an implementation of the ConfigMapStore that doesn't
// persist configmaps. Only used in testing.
func NewFakeConfigMapVault(ns, name string) *ConfigMapVault {
	return &ConfigMapVault{cache.NewStore(cache.MetaNamespaceKeyFunc), ns, name}
}

// ConfigMapStore wraps the store interface. Implementations usually persist
// contents of the store transparently.
type ConfigMapStore interface {
	cache.Store
}

// ApiServerConfigMapStore only services Add and GetByKey from apiserver.
// TODO: Implement all the other store methods and make this a write
// through cache.
type ApiServerConfigMapStore struct {
	ConfigMapStore
	client *client.Client
}

// Add adds the given config map to the apiserver's store.
func (a *ApiServerConfigMapStore) Add(obj interface{}) error {
	cfg := obj.(*api.ConfigMap)
	_, err := a.client.ConfigMaps(cfg.Namespace).Create(cfg)
	return err
}

// Update updates the existing config map object.
func (a *ApiServerConfigMapStore) Update(obj interface{}) error {
	cfg := obj.(*api.ConfigMap)
	_, err := a.client.ConfigMaps(cfg.Namespace).Update(cfg)
	return err
}

// Delete deletes the existing config map object.
func (a *ApiServerConfigMapStore) Delete(obj interface{}) error {
	cfg := obj.(*api.ConfigMap)
	return a.client.ConfigMaps(cfg.Namespace).Delete(cfg.Name)
}

// GetByKey returns the config map for a given key.
// The key must take the form namespace/name.
func (a *ApiServerConfigMapStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	nsName := strings.Split(key, "/")
	if len(nsName) != 2 {
		return nil, false, fmt.Errorf("Failed to get key %v, unexpecte format, expecting ns/name", key)
	}
	ns, name := nsName[0], nsName[1]
	cfg, err := a.client.ConfigMaps(ns).Get(name)
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
func NewConfigMapStore(c *client.Client) ConfigMapStore {
	return &ApiServerConfigMapStore{ConfigMapStore: cache.NewStore(cache.MetaNamespaceKeyFunc), client: c}
}
