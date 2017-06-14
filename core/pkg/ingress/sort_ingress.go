/*
Copyright 2016 The Kubernetes Authors.

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

package ingress

import (
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// BackendByNameServers sorts upstreams by name
type BackendByNameServers []*Backend

func (c BackendByNameServers) Len() int      { return len(c) }
func (c BackendByNameServers) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c BackendByNameServers) Less(i, j int) bool {

	return c[i].Name < c[j].Name
}

// EndpointByAddrPort sorts endpoints by address and port
type EndpointByAddrPort []Endpoint

func (c EndpointByAddrPort) Len() int      { return len(c) }
func (c EndpointByAddrPort) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c EndpointByAddrPort) Less(i, j int) bool {
	iName := c[i].Address
	jName := c[j].Address
	if iName != jName {
		return iName < jName
	}

	iU := c[i].Port
	jU := c[j].Port
	return iU < jU
}

// ServerByName sorts servers by name
type ServerByName []*Server

func (c ServerByName) Len() int      { return len(c) }
func (c ServerByName) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c ServerByName) Less(i, j int) bool {
	return c[i].Hostname < c[j].Hostname
}

// LocationByPath sorts location by path in descending order
// Location / is the last one
type LocationByPath []*Location

func (c LocationByPath) Len() int      { return len(c) }
func (c LocationByPath) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c LocationByPath) Less(i, j int) bool {
	return c[i].Path > c[j].Path
}

// SSLCert describes a SSL certificate to be used in a server
type SSLCert struct {
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	// CAFileName contains the path to the file with the root certificate
	CAFileName string `json:"caFileName"`
	// PemFileName contains the path to the file with the certificate and key concatenated
	PemFileName string `json:"pemFileName"`
	// PemSHA contains the sha1 of the pem file.
	// This is used to detect changes in the secret that contains the certificates
	PemSHA string `json:"pemSha"`
	// CN contains all the common names defined in the SSL certificate
	CN []string `json:"cn"`
	// ExpiresTime contains the expiration of this SSL certificate in timestamp format
	ExpireTime time.Time `json:"expires"`
}

// GetObjectKind implements the ObjectKind interface as a noop
func (s SSLCert) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
