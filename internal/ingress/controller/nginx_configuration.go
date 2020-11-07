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
	"fmt"
	"sort"

	networking "k8s.io/api/networking/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/k8s"
)

// extractServers build a list of unique hostnames from a list of ingresses
func extractServers(ingresses []*ingress.Ingress) []string {
	servers := sets.NewString()

	for _, ingressDefinition := range ingresses {
		annotations := ingressDefinition.ParsedAnnotations

		for _, rule := range ingressDefinition.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = defServerName
			}

			// Ingresses marked as canary cannot create a server
			if annotations.Canary.Enabled {
				continue
			}

			if servers.Has(host) {
				continue
			}

			servers.Insert(host)
		}
	}

	if !servers.Has(defServerName) {
		servers.Insert(defServerName)
	}

	sorterServers := servers.List()
	sort.Strings(sorterServers)

	return sorterServers
}

func buildServerLocations(servers []string, ingresses []*ingress.Ingress) map[string]map[string][]*ingress.Location {
	var pathPrefix = networking.PathTypePrefix

	withLocations := make(map[string]map[string][]*ingress.Location)

	for _, hostname := range servers {
		withLocations[hostname] = map[string][]*ingress.Location{}

		locations := serverLocations(hostname, ingresses)
		for _, location := range locations {
			withLocations[hostname][location.Path] = []*ingress.Location{location}

			annotationsToLocation(location, nil)

			if *location.PathType != pathPrefix {
				continue
			}

			// normalize path. Must end in /
			// add a new location of type Exact
			// this is required to pass conformance tests
			// If a location is defined by a prefix string that ends with the slash character, and requests are processed by one of
			// proxy_pass, fastcgi_pass, uwsgi_pass, scgi_pass, memcached_pass, or grpc_pass, then the special processing is performed.
			// In response to a request with URI equal to // this string, but without the trailing slash, a permanent redirect with the
			// code 301 will be returned to the requested URI with the slash appended.
			// If this is not desired, an exact match of the URI and location could be defined like this:
			//
			// location /user/ {
			//     proxy_pass http://user.example.com;
			// }
			// location = /user {
			//     proxy_pass http://login.example.com;
			// }
		}

		if _, isRootMapped := withLocations[hostname][rootLocation]; !isRootMapped {
			// the server contains at least one path but not /
			// we need to configure one and point to the default backend.
		}
	}

	return withLocations
}

func buildServerConfiguration(proxySSLLocationOnly bool,
	serverLocations map[string]map[string][]*ingress.Location,
	ingresses []*ingress.Ingress,
	ingressByKey func(key string) (*annotations.Ingress, error)) map[string]*ingress.Server {

	servers := map[string]*ingress.Server{}

	for serverName, pathLocations := range serverLocations {
		_, ok := servers[serverName]
		if !ok {
			servers[serverName] = &ingress.Server{}
		}

		server := servers[serverName]

		for _, locations := range pathLocations {
			for _, location := range locations {
				ingKey := location.Ingress

				anns, err := ingressByKey(ingKey)
				if err != nil {
					// TODO: logs
					continue
				}

				if server.AuthTLSError == "" && anns.CertificateAuth.AuthTLSError != "" {
					server.AuthTLSError = anns.CertificateAuth.AuthTLSError
				}

				if server.CertificateAuth.CAFileName == "" {
					server.CertificateAuth = anns.CertificateAuth
					if server.CertificateAuth.Secret != "" && server.CertificateAuth.CAFileName == "" {
						klog.V(3).InfoS("Secret has no 'ca.crt' key, mutual authentication disabled for Ingress", "hostname", server.CertificateAuth.Secret, "ingress", ingKey)
					}
				} else {
					klog.V(3).InfoS("Server already configured for mutual authentication", "hostname", "ingress", ingKey)
				}

				if proxySSLLocationOnly {
					continue
				}

				if server.ProxySSL.CAFileName == "" {
					server.ProxySSL = anns.ProxySSL
					if server.ProxySSL.Secret != "" && server.ProxySSL.CAFileName == "" {
						klog.V(3).InfoS("Secret has no 'ca.crt' key, client cert authentication disabled", "secret", server.ProxySSL.Secret, "ingress", ingKey)
					}
				} else {
					klog.V(3).InfoS("Server already configured for client cert authentication", "hostname", server.Hostname, "ingress", ingKey)
				}
			}
		}
	}

	return servers
}

// serverLocations builds a list of locations for a hostname
func serverLocations(hostname string, ingresses []*ingress.Ingress) []*ingress.Location {
	var pathPrefix = networking.PathTypePrefix

	locations := []*ingress.Location{}

	for _, ingressDefinition := range ingresses {
		if ingressDefinition.Spec.Backend != nil {
			locations = append(locations, &ingress.Location{
				Path:     "_defaultBackend",
				PathType: &pathPrefix,
			})
		}

		for _, rule := range ingressDefinition.Spec.Rules {
			if hostname != defServerName && hostname != rule.Host {
				continue
			}

			if rule.HTTP == nil {
				continue
			}

			for _, httpPath := range rule.HTTP.Paths {
				if containsPath(httpPath.Path, httpPath.PathType, locations) {
					continue
				}

				location := &ingress.Location{
					Path:         httpPath.Path,
					PathType:     httpPath.PathType,
					Ingress:      k8s.MetaNamespaceKey(ingressDefinition),
					Service:      fmt.Sprintf("%v/%v", ingressDefinition.Namespace, httpPath.Backend.ServiceName),
					Port:         httpPath.Backend.ServicePort,
					IsDefBackend: false,
				}

				locations = append(locations, location)
			}
		}
	}

	return locations
}

// containsPath checks if a list of locations contains a path with a particular PathType
func containsPath(path string, pathType *networking.PathType, locations []*ingress.Location) bool {
	for _, location := range locations {
		if location.Path == path && samePathType(location.PathType, pathType) {
			return true
		}
	}

	return false
}

// samePathType returns if two PathType instances are equal or not
func samePathType(pt1, pt2 *networking.PathType) bool {
	return apiequality.Semantic.DeepEqual(pt1, pt2)
}

// annotationsToLocation sets annotation values to location object
func annotationsToLocation(location *ingress.Location, ingressAnnotations *annotations.Ingress) {
	if location == nil {
		return
	}

	if ingressAnnotations == nil {
		return
	}

	location.BackendProtocol = ingressAnnotations.BackendProtocol
	location.BasicDigestAuth = ingressAnnotations.BasicDigestAuth
	location.ClientBodyBufferSize = ingressAnnotations.ClientBodyBufferSize
	location.ConfigurationSnippet = ingressAnnotations.ConfigurationSnippet
	location.Connection = ingressAnnotations.Connection
	location.CorsConfig = ingressAnnotations.CorsConfig
	location.CustomHTTPErrors = ingressAnnotations.CustomHTTPErrors
	location.Denied = ingressAnnotations.Denied
	location.EnableGlobalAuth = ingressAnnotations.EnableGlobalAuth
	location.ExternalAuth = ingressAnnotations.ExternalAuth
	location.FastCGI = ingressAnnotations.FastCGI
	location.HTTP2PushPreload = ingressAnnotations.HTTP2PushPreload
	location.InfluxDB = ingressAnnotations.InfluxDB
	location.Logs = ingressAnnotations.Logs
	location.Mirror = ingressAnnotations.Mirror
	location.ModSecurity = ingressAnnotations.ModSecurity
	location.Opentracing = ingressAnnotations.Opentracing
	location.Proxy = ingressAnnotations.Proxy
	location.ProxySSL = ingressAnnotations.ProxySSL
	location.RateLimit = ingressAnnotations.RateLimit
	location.Redirect = ingressAnnotations.Redirect
	location.Rewrite = ingressAnnotations.Rewrite
	location.Satisfy = ingressAnnotations.Satisfy
	location.UpstreamVhost = ingressAnnotations.UpstreamVhost
	location.UsePortInRedirects = ingressAnnotations.UsePortInRedirects
	location.Whitelist = ingressAnnotations.Whitelist
	location.XForwardedPrefix = ingressAnnotations.XForwardedPrefix

	location.DefaultBackend = ingressAnnotations.DefaultBackend
	location.DefaultBackendUpstreamName = defUpstreamName
}
