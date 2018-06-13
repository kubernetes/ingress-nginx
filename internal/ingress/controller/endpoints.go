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

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/ingress-nginx/internal/ingress"
	"k8s.io/ingress-nginx/internal/ingress/annotations/healthcheck"
)

// getEndpoints returns a list of Endpoint structs for a given service/target port combination.
func getEndpoints(s *corev1.Service, port *corev1.ServicePort, proto corev1.Protocol, hz *healthcheck.Config,
	getServiceEndpoints func(*corev1.Service) (*corev1.Endpoints, error)) []ingress.Endpoint {

	upsServers := []ingress.Endpoint{}

	if s == nil || port == nil {
		return upsServers
	}

	// using a map avoids duplicated upstream servers when the service
	// contains multiple port definitions sharing the same targetport
	processedUpstreamServers := make(map[string]struct{})

	// ExternalName services
	if s.Spec.Type == corev1.ServiceTypeExternalName {
		glog.V(3).Infof("Ingress using Service %q of type ExternalName.", s.Name)

		targetPort := port.TargetPort.IntValue()
		if targetPort <= 0 {
			glog.Errorf("ExternalName Service %q has an invalid port (%v)", s.Name, targetPort)
			return upsServers
		}

		if net.ParseIP(s.Spec.ExternalName) == nil {
			_, err := net.LookupHost(s.Spec.ExternalName)
			if err != nil {
				glog.Errorf("Error resolving host %q: %v", s.Spec.ExternalName, err)
				return upsServers
			}
		}

		return append(upsServers, ingress.Endpoint{
			Address:     s.Spec.ExternalName,
			Port:        fmt.Sprintf("%v", targetPort),
			MaxFails:    hz.MaxFails,
			FailTimeout: hz.FailTimeout,
		})
	}

	glog.V(3).Infof("Getting Endpoints for Service \"%v/%v\" and port %v", s.Namespace, s.Name, port.String())
	ep, err := getServiceEndpoints(s)
	if err != nil {
		glog.Warningf("Error obtaining Endpoints for Service \"%v/%v\": %v", s.Namespace, s.Name, err)
		return upsServers
	}

	for _, ss := range ep.Subsets {
		for _, epPort := range ss.Ports {

			if !reflect.DeepEqual(epPort.Protocol, proto) {
				continue
			}

			var targetPort int32

			if port.Name == "" {
				// port.Name is optional if there is only one port
				targetPort = epPort.Port
			} else if port.Name == epPort.Name {
				targetPort = epPort.Port
			}

			if targetPort <= 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ep := fmt.Sprintf("%v:%v", epAddress.IP, targetPort)
				if _, exists := processedUpstreamServers[ep]; exists {
					continue
				}
				ups := ingress.Endpoint{
					Address:     epAddress.IP,
					Port:        fmt.Sprintf("%v", targetPort),
					MaxFails:    hz.MaxFails,
					FailTimeout: hz.FailTimeout,
					Target:      epAddress.TargetRef,
				}
				upsServers = append(upsServers, ups)
				processedUpstreamServers[ep] = struct{}{}
			}
		}
	}

	glog.V(3).Infof("Endpoints found for Service \"%v/%v\": %v", s.Namespace, s.Name, upsServers)
	return upsServers
}
