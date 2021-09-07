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

package controller

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/k8s"
)

// getEndpoints returns a list of Endpoint structs for a given service/target port combination.
func getEndpoints(s *corev1.Service, port *corev1.ServicePort, proto corev1.Protocol,
	getServiceEndpoints func(string) (*corev1.Endpoints, error)) []ingress.Endpoint {

	upsServers := []ingress.Endpoint{}

	if s == nil || port == nil {
		return upsServers
	}

	// using a map avoids duplicated upstream servers when the service
	// contains multiple port definitions sharing the same targetport
	processedUpstreamServers := make(map[string]struct{})

	svcKey := k8s.MetaNamespaceKey(s)

	// ExternalName services
	if s.Spec.Type == corev1.ServiceTypeExternalName {
		if ip := net.ParseIP(s.Spec.ExternalName); s.Spec.ExternalName == "localhost" ||
			(ip != nil && ip.IsLoopback()) {
			klog.Errorf("Invalid attempt to use localhost name %s in %q", s.Spec.ExternalName, svcKey)
			return upsServers
		}

		klog.V(3).Infof("Ingress using Service %q of type ExternalName.", svcKey)
		targetPort := port.TargetPort.IntValue()
		// if the externalName is not an IP address we need to validate is a valid FQDN
		if net.ParseIP(s.Spec.ExternalName) == nil {
			externalName := strings.TrimSuffix(s.Spec.ExternalName, ".")
			if errs := validation.IsDNS1123Subdomain(externalName); len(errs) > 0 {
				klog.Errorf("Invalid DNS name %s: %v", s.Spec.ExternalName, errs)
				return upsServers
			}
		}

		return append(upsServers, ingress.Endpoint{
			Address: s.Spec.ExternalName,
			Port:    fmt.Sprintf("%v", targetPort),
		})
	}

	klog.V(3).Infof("Getting Endpoints for Service %q and port %v", svcKey, port.String())
	ep, err := getServiceEndpoints(svcKey)
	if err != nil {
		klog.Warningf("Error obtaining Endpoints for Service %q: %v", svcKey, err)
		return upsServers
	}

	for _, ss := range ep.Subsets {
		matchedPortNameFound := false
		for i, epPort := range ss.Ports {

			if !reflect.DeepEqual(epPort.Protocol, proto) {
				continue
			}

			var targetPort int32

			if port.Name == "" {
				// port.Name is optional if there is only one port
				targetPort = epPort.Port
				matchedPortNameFound = true
			} else if port.Name == epPort.Name {
				targetPort = epPort.Port
				matchedPortNameFound = true
			}

			if i == len(ss.Ports)-1 && !matchedPortNameFound && port.TargetPort.Type == intstr.Int {
				// use service target port if it's a number and no port name matched
				// https://github.com/kubernetes/ingress-nginx/issues/7390
				targetPort = port.TargetPort.IntVal
			}

			if targetPort <= 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ep := net.JoinHostPort(epAddress.IP, strconv.Itoa(int(targetPort)))
				if _, exists := processedUpstreamServers[ep]; exists {
					continue
				}
				ups := ingress.Endpoint{
					Address: epAddress.IP,
					Port:    fmt.Sprintf("%v", targetPort),
					Target:  epAddress.TargetRef,
				}
				upsServers = append(upsServers, ups)
				processedUpstreamServers[ep] = struct{}{}
			}
		}
	}

	klog.V(3).Infof("Endpoints found for Service %q: %v", svcKey, upsServers)
	return upsServers
}
