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
	"strings"

	"github.com/mitchellh/copystructure"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations"
	"k8s.io/ingress-nginx/internal/ingress/annotations/log"
	"k8s.io/ingress-nginx/internal/ingress/annotations/proxy"
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

func buildServerLocations(
	enableAccessLogForDefaultBackend bool,
	defaults proxy.Config,
	defaultUpstream *ingress.Backend,
	servers []string,
	ingresses []*ingress.Ingress,
	ingressByKey func(key string) (*annotations.Ingress, error)) map[string]map[string][]*ingress.Location {

	withLocations := make(map[string]map[string][]*ingress.Location)

	for _, hostname := range servers {
		withLocations[hostname] = map[string][]*ingress.Location{}

		locations := serverLocations(hostname, ingresses)
		for _, location := range locations {
			ing, _ := ingressByKey(location.Ingress)
			annotationsToLocation(location, ing)

			if *location.PathType != pathTypePrefix {
				continue
			}

			if location.Path == rootLocation {
				continue
			}

			if needsRewrite(location) || location.Rewrite.UseRegex {
				// TODO: review. we cannot change the path.
				withLocations[hostname][location.Path] = []*ingress.Location{location}
				klog.Warningf("Ingress path %v in Ingress %v for host %v cannot be processed", location.Path, location.Backend, hostname)
				continue
			}

			// copy location before any change
			el, err := copystructure.Copy(location)
			if err != nil {
				klog.ErrorS(err, "copying location")
			}

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

			// normalize path. Must end in /
			location.Path = normalizePrefixPath(location.Path)
			withLocations[hostname][location.Path] = []*ingress.Location{location}

			// add exact location
			exactLocation := el.(*ingress.Location)
			exactLocation.PathType = &pathTypeExact

			if _, ok := withLocations[hostname][exactLocation.Path]; !ok {
				withLocations[hostname][exactLocation.Path] = []*ingress.Location{exactLocation}
			}
		}

		// the server contains at least one path but not /
		// we need to configure one and point to the default backend.
		if _, isRootMapped := withLocations[hostname][rootLocation]; !isRootMapped {
			var svcKey string
			if defaultUpstream.Service != nil {
				svcKey = k8s.MetaNamespaceKey(defaultUpstream.Service)
			}

			withLocations[hostname][rootLocation] = append(withLocations[hostname][rootLocation], &ingress.Location{
				Path:         rootLocation,
				PathType:     &pathTypePrefix,
				IsDefBackend: true,
				Backend:      defaultUpstream.Name,
				Proxy:        defaults,
				Service:      svcKey,
				Port:         defaultUpstream.Port,
				Logs: log.Config{
					Access: enableAccessLogForDefaultBackend,
				},
			})
		}
	}

	return withLocations
}

func buildServerConfiguration(
	proxySSLLocationOnly bool,
	serverLocations map[string]map[string][]*ingress.Location,
	ingresses []*ingress.Ingress,
	ingressByKey func(key string) (*annotations.Ingress, error)) map[string]*ingress.Server {

	servers := map[string]*ingress.Server{}

	for serverName, pathLocations := range serverLocations {
		if _, ok := servers[serverName]; !ok {
			servers[serverName] = &ingress.Server{
				Hostname: serverName,
			}
		}

		server := servers[serverName]

		for _, locations := range pathLocations {
			for _, location := range locations {
				ingKey := location.Ingress

				anns, err := ingressByKey(ingKey)
				if err != nil {
					klog.ErrorS(err, "searching ingress by key")
					continue
				}

				server.Locations = append(server.Locations, location)

				// one location configured Redirect from-to-www
				if location.Redirect.FromToWWW {
					server.RedirectFromToWWW = true
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

	// return a sorted list by Path and number of locations
	sort.SliceStable(locations, func(i, j int) bool {
		return locations[i].Path > locations[j].Path
	})

	sort.SliceStable(locations, func(i, j int) bool {
		return len(locations[i].Path) > len(locations[j].Path)
	})

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

func normalizePrefixPath(path string) string {
	if path == rootLocation {
		return rootLocation
	}

	if !strings.HasSuffix(path, "/") {
		return fmt.Sprintf("%v/", path)
	}

	return path
}

func needsRewrite(location *ingress.Location) bool {
	if len(location.Rewrite.Target) > 0 && location.Rewrite.Target != location.Path {
		return true
	}
	return false
}

func buildUpstreamsWithDefaultBackend(
	upstreams map[string]*ingress.Backend,
	servers map[string]*ingress.Server,
	getServiceEndpoints func(key string) (*corev1.Endpoints, error)) []*ingress.Backend {
	upstreamsWithDefaultBackend := make([]*ingress.Backend, 0, len(upstreams))

	for _, upstream := range upstreams {
		upstreamsWithDefaultBackend = append(upstreamsWithDefaultBackend, upstream)
		if upstream.Name == defUpstreamName {
			continue
		}

		isHTTPSfrom := []*ingress.Server{}

		for _, server := range servers {
			for _, location := range server.Locations {
				// use default backend
				if !shouldCreateUpstreamForLocationDefaultBackend(upstream, location) {
					continue
				}

				if len(location.DefaultBackend.Spec.Ports) == 0 {
					klog.Errorf("Custom default backend service %v/%v has no ports. Ignoring", location.DefaultBackend.Namespace, location.DefaultBackend.Name)
					continue
				}

				sp := location.DefaultBackend.Spec.Ports[0]
				endps := getEndpoints(location.DefaultBackend, &sp, apiv1.ProtocolTCP, getServiceEndpoints)
				// custom backend is valid only if contains at least one endpoint
				if len(endps) > 0 {
					name := fmt.Sprintf("custom-default-backend-%v", location.DefaultBackend.GetName())
					klog.V(3).Infof("Creating \"%v\" upstream based on default backend annotation", name)

					nb := upstream.DeepCopy()
					nb.Name = name
					nb.Endpoints = endps

					upstreamsWithDefaultBackend = append(upstreamsWithDefaultBackend, nb)

					location.DefaultBackendUpstreamName = name

					if len(upstream.Endpoints) == 0 {
						klog.V(3).Infof("Upstream %q has no active Endpoint, so using custom default backend for location %q in server %q (Service \"%v/%v\")",
							upstream.Name, location.Path, server.Hostname, location.DefaultBackend.Namespace, location.DefaultBackend.Name)

						location.Backend = name
					}
				}

				if server.SSLPassthrough {
					if location.Path == rootLocation {
						if location.Backend == defUpstreamName {
							klog.Warningf("Server %q has no default backend, ignoring SSL Passthrough.", server.Hostname)
							continue
						}
						isHTTPSfrom = append(isHTTPSfrom, server)
					}
				}
			}
		}

		if len(isHTTPSfrom) > 0 {
			upstream.SSLPassthrough = true
		}
	}

	// sort upstreams by name
	sort.SliceStable(upstreamsWithDefaultBackend, func(a, b int) bool {
		return upstreamsWithDefaultBackend[a].Name < upstreamsWithDefaultBackend[b].Name
	})

	return upstreamsWithDefaultBackend
}
