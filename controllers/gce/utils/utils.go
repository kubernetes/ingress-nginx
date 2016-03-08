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

package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"regexp"
)

const (
	// Add used to record additions in a sync pool.
	Add = iota
	// Remove used to record removals from a sync pool.
	Remove
	// Sync used to record syncs of a sync pool.
	Sync
	// Get used to record Get from a sync pool.
	Get
	// Create used to recrod creations in a sync pool.
	Create
	// Update used to record updates in a sync pool.
	Update
	// Delete used to record deltions from a sync pool.
	Delete
	// AddInstances used to record a call to AddInstances.
	AddInstances
	// RemoveInstances used to record a call to RemoveInstances.
	RemoveInstances

	// This allows sharing of backends across loadbalancers.
	backendPrefix = "k8s-be"
	backendRegex  = "k8s-be-([0-9]+).*"

	// Prefix used for instance groups involved in L7 balancing.
	igPrefix = "k8s-ig"

	// Suffix used in the l7 firewall rule. There is currently only one.
	// Note that this name is used by the cloudprovider lib that inserts its
	// own k8s-fw prefix.
	globalFirewallSuffix = "l7"

	// A delimiter used for clarity in naming GCE resources.
	clusterNameDelimiter = "--"

	// Arbitrarily chosen alphanumeric character to use in constructing resource
	// names, eg: to avoid cases where we end up with a name ending in '-'.
	alphaNumericChar = "0"

	// Names longer than this are truncated, because of GCE restrictions.
	nameLenLimit = 62

	// DefaultBackendKey is the key used to transmit the defaultBackend through
	// a urlmap. It's not a valid subdomain, and it is a catch all path.
	// TODO: Find a better way to transmit this, once we've decided on default
	// backend semantics (i.e do we want a default per host, per lb etc).
	DefaultBackendKey = "DefaultBackend"

	// K8sAnnotationPrefix is the prefix used in annotations used to record
	// debug information in the Ingress annotations.
	K8sAnnotationPrefix = "ingress.kubernetes.io"
)

// Namer handles centralized naming for the cluster.
type Namer struct {
	ClusterName string
}

// Truncate truncates the given key to a GCE length limit.
func (n *Namer) Truncate(key string) string {
	if len(key) > nameLenLimit {
		// GCE requires names to end with an albhanumeric, but allows characters
		// like '-', so make sure the trucated name ends legally.
		return fmt.Sprintf("%v%v", key[:nameLenLimit], alphaNumericChar)
	}
	return key
}

func (n *Namer) decorateName(name string) string {
	if n.ClusterName == "" {
		return name
	}
	return n.Truncate(fmt.Sprintf("%v%v%v", name, clusterNameDelimiter, n.ClusterName))
}

// NameBelongsToCluster checks if a given name is tagged with this cluster's UID.
func (n *Namer) NameBelongsToCluster(name string) bool {
	if !strings.HasPrefix(name, "k8s-") {
		glog.V(4).Infof("%v not part of cluster", name)
		return false
	}
	parts := strings.Split(name, clusterNameDelimiter)
	if len(parts) == 1 {
		if n.ClusterName == "" {
			return true
		}
		return false
	}
	if len(parts) > 2 {
		glog.Warningf("Too many parts to name %v, ignoring", name)
		return false
	}
	return parts[1] == n.ClusterName
}

// BeName constructs the name for a backend.
func (n *Namer) BeName(port int64) string {
	return n.decorateName(fmt.Sprintf("%v-%d", backendPrefix, port))
}

// BePort retrieves the port from the given backend name.
func (n *Namer) BePort(beName string) (string, error) {
	r, err := regexp.Compile(backendRegex)
	if err != nil {
		return "", err
	}
	match := r.FindStringSubmatch(beName)
	if len(match) < 2 {
		return "", fmt.Errorf("Unable to lookup port for %v", beName)
	}
	_, err = strconv.Atoi(match[1])
	if err != nil {
		return "", fmt.Errorf("Unexpected regex match: %v", beName)
	}
	return match[1], nil
}

// IGName constructs the name for an Instance Group.
func (n *Namer) IGName() string {
	// Currently all ports are added to a single instance group.
	return n.decorateName(igPrefix)
}

// FrSuffix constructs the glbc specific suffix for the FirewallRule.
func (n *Namer) FrSuffix() string {
	// The entire cluster only needs a single firewall rule.
	if n.ClusterName == "" {
		return globalFirewallSuffix
	}
	return n.Truncate(fmt.Sprintf("%v%v%v", globalFirewallSuffix, clusterNameDelimiter, n.ClusterName))
}

// FrName constructs the full firewall rule name, this is the name assigned by
// the cloudprovider lib + suffix from glbc, so we don't mix this rule with a
// rule created for L4 loadbalancing.
func (n *Namer) FrName(suffix string) string {
	return fmt.Sprintf("k8s-fw-%s", suffix)
}

// LBName constructs a loadbalancer name from the given key. The key is usually
// the namespace/name of a Kubernetes Ingress.
func (n *Namer) LBName(key string) string {
	// TODO: Pipe the clusterName through, for now it saves code churn to just
	// grab it globally, especially since we haven't decided how to handle
	// namespace conflicts in the Ubernetes context.
	parts := strings.Split(key, clusterNameDelimiter)
	scrubbedName := strings.Replace(key, "/", "-", -1)
	if n.ClusterName == "" || parts[len(parts)-1] == n.ClusterName {
		return scrubbedName
	}
	return n.Truncate(fmt.Sprintf("%v%v%v", scrubbedName, clusterNameDelimiter, n.ClusterName))
}

// GCEURLMap is a nested map of hostname->path regex->backend
type GCEURLMap map[string]map[string]*compute.BackendService

// GetDefaultBackend performs a destructive read and returns the default
// backend of the urlmap.
func (g GCEURLMap) GetDefaultBackend() *compute.BackendService {
	var d *compute.BackendService
	var exists bool
	if h, ok := g[DefaultBackendKey]; ok {
		if d, exists = h[DefaultBackendKey]; exists {
			delete(h, DefaultBackendKey)
		}
		delete(g, DefaultBackendKey)
	}
	return d
}

// String implements the string interface for the GCEURLMap.
func (g GCEURLMap) String() string {
	msg := ""
	for host, um := range g {
		msg += fmt.Sprintf("%v\n", host)
		for url, be := range um {
			msg += fmt.Sprintf("\t%v: ", url)
			if be == nil {
				msg += fmt.Sprintf("No backend\n")
			} else {
				msg += fmt.Sprintf("%v\n", be.Name)
			}
		}
	}
	return msg
}

// PutDefaultBackend performs a destructive write replacing the
// default backend of the url map with the given backend.
func (g GCEURLMap) PutDefaultBackend(d *compute.BackendService) {
	g[DefaultBackendKey] = map[string]*compute.BackendService{
		DefaultBackendKey: d,
	}
}

// IsHTTPErrorCode checks if the given error matches the given HTTP Error code.
// For this to work the error must be a googleapi Error.
func IsHTTPErrorCode(err error, code int) bool {
	apiErr, ok := err.(*googleapi.Error)
	return ok && apiErr.Code == code
}

// CompareLinks returns true if the 2 self links are equal.
func CompareLinks(l1, l2 string) bool {
	// TODO: These can be partial links
	return l1 == l2 && l1 != ""
}

// FakeIngressRuleValueMap is a convenience type used by multiple submodules
// that share the same testing methods.
type FakeIngressRuleValueMap map[string]string
