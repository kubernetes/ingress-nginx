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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/ingress-nginx/internal/net/ssl"
	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/ingress-nginx/pkg/apis/ingress"
)

const (
	emptyUID = "-1"
)

// configureDynamically encodes new Backends in JSON format and POSTs the
// payload to an internal HTTP endpoint handled by Lua.
func ConfigureDynamically(newconfig *ingress.Configuration, oldconfig *ingress.Configuration) error {
	backendsChanged := !reflect.DeepEqual(oldconfig.Backends, newconfig.Backends)
	if backendsChanged {
		err := configureBackends(newconfig.Backends)
		if err != nil {
			return err
		}
	}

	streamConfigurationChanged := !reflect.DeepEqual(oldconfig.TCPEndpoints, newconfig.TCPEndpoints) || !reflect.DeepEqual(oldconfig.UDPEndpoints, newconfig.UDPEndpoints)
	if streamConfigurationChanged {
		err := updateStreamConfiguration(newconfig.TCPEndpoints, newconfig.UDPEndpoints)
		if err != nil {
			return err
		}
	}

	serversChanged := !reflect.DeepEqual(oldconfig.Servers, newconfig.Servers)
	if serversChanged {
		err := configureCertificates(newconfig.Servers)
		if err != nil {
			return err
		}
	}

	return nil
}

func configureBackends(rawBackends []*ingress.Backend) error {
	backends := make([]*ingress.Backend, len(rawBackends))

	for i, backend := range rawBackends {
		var service *apiv1.Service
		if backend.Service != nil {
			service = &apiv1.Service{Spec: backend.Service.Spec}
		}
		luaBackend := &ingress.Backend{
			Name:                 backend.Name,
			Port:                 backend.Port,
			SSLPassthrough:       backend.SSLPassthrough,
			SessionAffinity:      backend.SessionAffinity,
			UpstreamHashBy:       backend.UpstreamHashBy,
			LoadBalancing:        backend.LoadBalancing,
			Service:              service,
			NoServer:             backend.NoServer,
			TrafficShapingPolicy: backend.TrafficShapingPolicy,
			AlternativeBackends:  backend.AlternativeBackends,
		}

		var endpoints []ingress.Endpoint
		for _, endpoint := range backend.Endpoints {
			endpoints = append(endpoints, ingress.Endpoint{
				Address: endpoint.Address,
				Port:    endpoint.Port,
			})
		}

		luaBackend.Endpoints = endpoints
		backends[i] = luaBackend
	}

	statusCode, _, err := nginx.NewPostStatusRequest("/configuration/backends", "application/json", backends)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("unexpected error code: %d", statusCode)
	}

	return nil
}

type sslConfiguration struct {
	Certificates map[string]string `json:"certificates"`
	Servers      map[string]string `json:"servers"`
}

// configureCertificates JSON encodes certificates and POSTs it to an internal HTTP endpoint
// that is handled by Lua
func configureCertificates(rawServers []*ingress.Server) error {
	configuration := &sslConfiguration{
		Certificates: map[string]string{},
		Servers:      map[string]string{},
	}

	configure := func(hostname string, sslCert *ingress.SSLCert) {
		uid := emptyUID

		if sslCert != nil {
			uid = sslCert.UID

			if _, ok := configuration.Certificates[uid]; !ok {
				configuration.Certificates[uid] = sslCert.PemCertKey
			}
		}

		configuration.Servers[hostname] = uid
	}

	for _, rawServer := range rawServers {
		configure(rawServer.Hostname, rawServer.SSLCert)

		for _, alias := range rawServer.Aliases {
			if rawServer.SSLCert != nil && ssl.IsValidHostname(alias, rawServer.SSLCert.CN) {
				configuration.Servers[alias] = rawServer.SSLCert.UID
			} else {
				configuration.Servers[alias] = emptyUID
			}
		}
	}

	redirects := BuildRedirects(rawServers)
	for _, redirect := range redirects {
		configure(redirect.From, redirect.SSLCert)
	}

	statusCode, _, err := nginx.NewPostStatusRequest("/configuration/servers", "application/json", configuration)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated {
		return fmt.Errorf("unexpected error code: %d", statusCode)
	}

	return nil
}

func updateStreamConfiguration(TCPEndpoints []ingress.L4Service, UDPEndpoints []ingress.L4Service) error {
	streams := make([]ingress.Backend, 0)
	for _, ep := range TCPEndpoints {
		var service *apiv1.Service
		if ep.Service != nil {
			service = &apiv1.Service{Spec: ep.Service.Spec}
		}

		key := fmt.Sprintf("tcp-%v-%v-%v", ep.Backend.Namespace, ep.Backend.Name, ep.Backend.Port.String())
		streams = append(streams, ingress.Backend{
			Name:      key,
			Endpoints: ep.Endpoints,
			Port:      intstr.FromInt(ep.Port),
			Service:   service,
		})
	}
	for _, ep := range UDPEndpoints {
		var service *apiv1.Service
		if ep.Service != nil {
			service = &apiv1.Service{Spec: ep.Service.Spec}
		}

		key := fmt.Sprintf("udp-%v-%v-%v", ep.Backend.Namespace, ep.Backend.Name, ep.Backend.Port.String())
		streams = append(streams, ingress.Backend{
			Name:      key,
			Endpoints: ep.Endpoints,
			Port:      intstr.FromInt(ep.Port),
			Service:   service,
		})
	}

	buf, err := json.Marshal(streams)
	if err != nil {
		return err
	}

	hostPort := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%v", nginx.StreamPort))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(buf)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(conn, "\r\n")
	if err != nil {
		return err
	}

	return nil
}
