/*
Copyright 2020 The Kubernetes Authors.

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

package controller

import (
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress"
)

// OK to merge canary ingresses iff there exists one or more ingresses to potentially merge into
func nonCanaryIngressExists(ingresses []*ingress.Ingress, canaryIngresses []*ingress.Ingress) bool {
	return len(ingresses)-len(canaryIngresses) > 0
}

// ensure that the following conditions are met
// 1) names of backends do not match and canary doesn't merge into itself
// 2) primary name is not the default upstream
// 3) the primary has a server
func canMergeBackend(primary *ingress.Backend, alternative *ingress.Backend) bool {
	return alternative != nil && primary.Name != alternative.Name && primary.Name != defUpstreamName && !primary.NoServer
}

// Performs the merge action and checks to ensure that one two alternative backends do not merge into each other
func mergeAlternativeBackend(priUps *ingress.Backend, altUps *ingress.Backend) bool {
	if priUps.NoServer {
		klog.Warningf("unable to merge alternative backend %v into primary backend %v because %v is a primary backend",
			altUps.Name, priUps.Name, priUps.Name)
		return false
	}

	for _, ab := range priUps.AlternativeBackends {
		if ab == altUps.Name {
			klog.V(2).Infof("skip merge alternative backend %v into %v, it's already present", altUps.Name, priUps.Name)
			return true
		}
	}

	priUps.AlternativeBackends =
		append(priUps.AlternativeBackends, altUps.Name)

	return true
}

// Compares an Ingress of a potential alternative backend's rules with each existing server and finds matching host + path pairs.
// If a match is found, we know that this server should back the alternative backend and add the alternative backend
// to a backend's alternative list.
// If no match is found, then the serverless backend is deleted.
func mergeAlternativeBackends(ing *ingress.Ingress, upstreams map[string]*ingress.Backend,
	servers map[string]*ingress.Server) {

	// merge catch-all alternative backends
	if ing.Spec.Backend != nil {
		upsName := upstreamName(ing.Namespace, ing.Spec.Backend.ServiceName, ing.Spec.Backend.ServicePort)

		altUps := upstreams[upsName]

		if altUps == nil {
			klog.Warningf("alternative backend %s has already been removed", upsName)
		} else {

			merged := false
			altEqualsPri := false

			for _, loc := range servers[defServerName].Locations {
				priUps := upstreams[loc.Backend]
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, servers[defServerName].Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(priUps, altUps)
				}
			}

			if !altEqualsPri && !merged {
				klog.Warningf("unable to find real backend for alternative backend %v. Deleting.", altUps.Name)
				delete(upstreams, altUps.Name)
			}
		}
	}

	for _, rule := range ing.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			upsName := upstreamName(ing.Namespace, path.Backend.ServiceName, path.Backend.ServicePort)

			altUps := upstreams[upsName]

			if altUps == nil {
				klog.Warningf("alternative backend %s has already been removed", upsName)
				continue
			}

			merged := false
			altEqualsPri := false

			server, ok := servers[rule.Host]
			if !ok {
				klog.Errorf("cannot merge alternative backend %s into hostname %s that does not exist",
					altUps.Name,
					rule.Host)

				continue
			}

			// find matching paths
			for _, loc := range server.Locations {
				priUps := upstreams[loc.Backend]
				altEqualsPri = altUps.Name == priUps.Name
				if altEqualsPri {
					klog.Warningf("alternative upstream %s in Ingress %s/%s is primary upstream in Other Ingress for location %s%s!",
						altUps.Name, ing.Namespace, ing.Name, server.Hostname, loc.Path)
					break
				}

				if canMergeBackend(priUps, altUps) && loc.Path == path.Path && *loc.PathType == *path.PathType {
					klog.V(2).Infof("matching backend %v found for alternative backend %v",
						priUps.Name, altUps.Name)

					merged = mergeAlternativeBackend(priUps, altUps)
				}
			}

			if !altEqualsPri && !merged {
				klog.Warningf("unable to find real backend for alternative backend %v. Deleting.", altUps.Name)
				delete(upstreams, altUps.Name)
			}
		}
	}
}
