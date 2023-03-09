/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/ingress-nginx/internal/k8s"
	"k8s.io/ingress-nginx/internal/net/ssl"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
	"k8s.io/klog/v2"
)

func GetRemovedHosts(rucfg, newcfg *ingress.Configuration) []string {
	oldSet := sets.NewString()
	newSet := sets.NewString()

	for _, s := range rucfg.Servers {
		if !oldSet.Has(s.Hostname) {
			oldSet.Insert(s.Hostname)
		}
	}

	for _, s := range newcfg.Servers {
		if !newSet.Has(s.Hostname) {
			newSet.Insert(s.Hostname)
		}
	}

	return oldSet.Difference(newSet).List()
}

// GetRemovedCertificateSerialNumbers extracts the difference of certificates between two configurations
func GetRemovedCertificateSerialNumbers(rucfg, newcfg *ingress.Configuration) []string {
	oldCertificates := sets.NewString()
	newCertificates := sets.NewString()

	for _, server := range rucfg.Servers {
		if server.SSLCert == nil {
			continue
		}
		identifier := server.SSLCert.Identifier()
		if identifier != "" {
			if !oldCertificates.Has(identifier) {
				oldCertificates.Insert(identifier)
			}
		}
	}

	for _, server := range newcfg.Servers {
		if server.SSLCert == nil {
			continue
		}
		identifier := server.SSLCert.Identifier()
		if identifier != "" {
			if !newCertificates.Has(identifier) {
				newCertificates.Insert(identifier)
			}
		}
	}

	return oldCertificates.Difference(newCertificates).List()
}

// GetRemovedIngresses extracts the difference of ingresses between two configurations
func GetRemovedIngresses(rucfg, newcfg *ingress.Configuration) []string {
	oldIngresses := sets.NewString()
	newIngresses := sets.NewString()

	for _, server := range rucfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == nil {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !oldIngresses.Has(ingKey) {
				oldIngresses.Insert(ingKey)
			}
		}
	}

	for _, server := range newcfg.Servers {
		for _, location := range server.Locations {
			if location.Ingress == nil {
				continue
			}

			ingKey := k8s.MetaNamespaceKey(location.Ingress)
			if !newIngresses.Has(ingKey) {
				newIngresses.Insert(ingKey)
			}
		}
	}

	return oldIngresses.Difference(newIngresses).List()
}

// IsDynamicConfigurationEnough returns whether a Configuration can be
// dynamically applied, without reloading the backend.
func IsDynamicConfigurationEnough(newcfg *ingress.Configuration, oldcfg *ingress.Configuration) bool {
	copyOfRunningConfig := *oldcfg
	copyOfPcfg := *newcfg

	copyOfRunningConfig.Backends = []*ingress.Backend{}
	copyOfPcfg.Backends = []*ingress.Backend{}

	clearL4serviceEndpoints(&copyOfRunningConfig)
	clearL4serviceEndpoints(&copyOfPcfg)

	clearCertificates(&copyOfRunningConfig)
	clearCertificates(&copyOfPcfg)

	return copyOfRunningConfig.Equal(&copyOfPcfg)
}

// clearL4serviceEndpoints is a helper function to clear endpoints from the ingress configuration since they should be ignored when
// checking if the new configuration changes can be applied dynamically.
func clearL4serviceEndpoints(config *ingress.Configuration) {
	var clearedTCPL4Services []ingress.L4Service
	var clearedUDPL4Services []ingress.L4Service
	for _, service := range config.TCPEndpoints {
		copyofService := ingress.L4Service{
			Port:      service.Port,
			Backend:   service.Backend,
			Endpoints: []ingress.Endpoint{},
			Service:   nil,
		}
		clearedTCPL4Services = append(clearedTCPL4Services, copyofService)
	}
	for _, service := range config.UDPEndpoints {
		copyofService := ingress.L4Service{
			Port:      service.Port,
			Backend:   service.Backend,
			Endpoints: []ingress.Endpoint{},
			Service:   nil,
		}
		clearedUDPL4Services = append(clearedUDPL4Services, copyofService)
	}
	config.TCPEndpoints = clearedTCPL4Services
	config.UDPEndpoints = clearedUDPL4Services
}

// clearCertificates is a helper function to clear Certificates from the ingress configuration since they should be ignored when
// checking if the new configuration changes can be applied dynamically if dynamic certificates is on
func clearCertificates(config *ingress.Configuration) {
	var clearedServers []*ingress.Server
	for _, server := range config.Servers {
		copyOfServer := *server
		copyOfServer.SSLCert = nil
		clearedServers = append(clearedServers, &copyOfServer)
	}
	config.Servers = clearedServers
}

type redirect struct {
	From    string
	To      string
	SSLCert *ingress.SSLCert
}

// BuildRedirects build the redirects of servers based on configurations and certificates
func BuildRedirects(servers []*ingress.Server) []*redirect {
	names := sets.Set[string]{}
	redirectServers := make([]*redirect, 0)

	for _, srv := range servers {
		if !srv.RedirectFromToWWW {
			continue
		}

		to := srv.Hostname

		var from string
		if strings.HasPrefix(to, "www.") {
			from = strings.TrimPrefix(to, "www.")
		} else {
			from = fmt.Sprintf("www.%v", to)
		}

		if names.Has(to) {
			continue
		}

		klog.V(3).InfoS("Creating redirect", "from", from, "to", to)
		found := false
		for _, esrv := range servers {
			if esrv.Hostname == from {
				found = true
				break
			}
		}

		if found {
			klog.Warningf("Already exists an Ingress with %q hostname. Skipping creation of redirection from %q to %q.", from, from, to)
			continue
		}

		r := &redirect{
			From: from,
			To:   to,
		}

		if srv.SSLCert != nil {
			if ssl.IsValidHostname(from, srv.SSLCert.CN) {
				r.SSLCert = srv.SSLCert
			} else {
				klog.Warningf("the server %v has SSL configured but the SSL certificate does not contains a CN for %v. Redirects will not work for HTTPS to HTTPS", from, to)
			}
		}

		redirectServers = append(redirectServers, r)
		names.Insert(to)
	}

	return redirectServers
}
