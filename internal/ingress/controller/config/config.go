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

package config

import (
	apiv1 "k8s.io/api/core/v1"

	"k8s.io/ingress-nginx/internal/ingress"
)

/*
// Configuration represents the content of nginx.conf file
type Configuration struct {


}

// BuildLogFormatUpstream format the log_format upstream using
// proxy_protocol_addr as remote client address if UseProxyProtocol
// is enabled.
func (cfg Configuration) BuildLogFormatUpstream() string {
	if cfg.LogFormatUpstream == logFormatUpstream {
		return fmt.Sprintf(cfg.LogFormatUpstream, "$the_real_ip")
	}

	return cfg.LogFormatUpstream
}

*/
// TemplateConfig contains the nginx configuration to render the file nginx.conf
type TemplateConfig struct {
	ProxySetHeaders             map[string]string
	AddHeaders                  map[string]string
	MaxOpenFiles                int
	BacklogSize                 int
	Backends                    []*ingress.Backend
	PassthroughBackends         []*ingress.SSLPassthroughBackend
	Servers                     []*ingress.Server
	TCPBackends                 []ingress.L4Service
	UDPBackends                 []ingress.L4Service
	HealthzURI                  string
	CustomErrors                bool
	Cfg                         Configuration
	IsIPV6Enabled               bool
	IsSSLPassthroughEnabled     bool
	NginxStatusIpv4Whitelist    []string
	NginxStatusIpv6Whitelist    []string
	RedirectServers             map[string]string
	ListenPorts                 *ListenPorts
	PublishService              *apiv1.Service
	DynamicConfigurationEnabled bool
	DisableLua                  bool
}

// ListenPorts describe the ports required to run the
// NGINX Ingress controller
type ListenPorts struct {
	HTTP     int
	HTTPS    int
	Status   int
	Health   int
	Default  int
	SSLProxy int
}
