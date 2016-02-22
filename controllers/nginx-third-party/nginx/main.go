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

package nginx

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"k8s.io/contrib/ingress/controllers/nginx-third-party/ssl"

	"k8s.io/kubernetes/pkg/client/record"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	k8sruntime "k8s.io/kubernetes/pkg/runtime"
)

const (
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	bodySize = "1m"

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	errorLevel = "info"

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.
	htsMaxAge = "15724800"

	// If UseProxyProtocol is enabled defIPCIDR defines the default the IP/network address of your external load balancer
	defIPCIDR = "0.0.0.0/0"

	gzipTypes = "application/atom+xml application/javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/plain text/x-component"

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA"

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1 TLSv1.1 TLSv1.2"

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"

	// Base directory that contains the mounted secrets with SSL certificates, keys and
	sslDirectory = "/etc/nginx-ssl"
)

type nginxConfiguration struct {
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	BodySize string `json:"bodySize,omitempty" structs:",omitempty"`

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `json:"errorLogLevel,omitempty" structs:",omitempty"`

	// Enables or disables the header HTS in servers running SSL
	UseHTS bool `json:"useHTS,omitempty" structs:",omitempty"`

	// Enables or disables the use of HTS in all the subdomains of the servername
	HTSIncludeSubdomains bool `json:"htsIncludeSubdomains,omitempty" structs:",omitempty"`

	// HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header)
	// that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
	// https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
	// max-age is the time, in seconds, that the browser should remember that this site is only to be
	// accessed using HTTPS.
	HTSMaxAge string `json:"htsMaxAge,omitempty" structs:",omitempty"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `json:"keepAlive,omitempty" structs:",omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// http://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `json:"maxWorkerConnections,omitempty" structs:",omitempty"`

	// Defines a timeout for establishing a connection with a proxied server.
	// It should be noted that this timeout cannot usually exceed 75 seconds.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout
	ProxyConnectTimeout int `json:"proxyConnectTimeout,omitempty" structs:",omitempty"`

	// If UseProxyProtocol is enabled ProxyRealIPCIDR defines the default the IP/network address
	// of your external load balancer
	ProxyRealIPCIDR string `json:"proxyRealIPCIDR,omitempty" structs:",omitempty"`

	// Timeout in seconds for reading a response from the proxied server. The timeout is set only between
	// two successive read operations, not for the transmission of the whole response
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout
	ProxyReadTimeout int `json:"proxyReadTimeout,omitempty" structs:",omitempty"`

	// Timeout in seconds for transmitting a request to the proxied server. The timeout is set only between
	// two successive write operations, not for the transmission of the whole request.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout
	ProxySendTimeout int `json:"proxySendTimeout,omitempty" structs:",omitempty"`

	// Configures name servers used to resolve names of upstream servers into addresses
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#resolver
	Resolver string `json:"resolver,omitempty" structs:",omitempty"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `json:"serverNameHashMaxSize,omitempty" structs:",omitempty"`

	// Size of the bucker for the server names hash tables
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `json:"serverNameHashBucketSize,omitempty" structs:",omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	SSLBufferSize string `json:"sslBufferSize,omitempty" structs:",omitempty"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	SSLCiphers string `json:"sslCiphers,omitempty" structs:",omitempty"`

	// Base64 string that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
	// https://www.openssl.org/docs/manmaster/apps/dhparam.html
	// https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
	SSLDHParam string `json:"sslDHParam,omitempty" structs:",omitempty"`

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	SSLProtocols string `json:"sslProtocols,omitempty" structs:",omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCache bool `json:"sslSessionCache,omitempty" structs:",omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCacheSize string `json:"sslSessionCacheSize,omitempty" structs:",omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SSLSessionTickets bool `json:"sslSessionTickets,omitempty" structs:",omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SSLSessionTimeout string `json:"sslSessionTimeout,omitempty" structs:",omitempty"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	UseProxyProtocol bool `json:"useProxyProtocol,omitempty" structs:",omitempty"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// http://nginx.org/en/docs/http/ngx_http_gzip_module.html
	UseGzip bool `json:"useGzip,omitempty" structs:",omitempty"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `json:"gzipTypes,omitempty" structs:",omitempty"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// http://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses string `json:"workerProcesses,omitempty" structs:",omitempty"`
}

// Service service definition to use in nginx template
type Service struct {
	ServiceName string
	ServicePort string
	Namespace   string
	// ExposedPort port used by nginx to listen for the stream upstream
	ExposedPort string
}

// NginxManager ...
type NginxManager struct {
	defBackend  Service
	defCfg      *nginxConfiguration
	defError    Service
	defResolver string

	// path to the configuration file to be used by nginx
	ConfigFile string

	sslCertificates []ssl.Certificate
	sslDHParam      string
	servicesL4      []Service

	client *client.Client
	// template loaded ready to be used to generate the nginx configuration file
	template *template.Template

	// obj runtime object to be used in events
	obj k8sruntime.Object

	recorder record.EventRecorder

	reloadLock *sync.Mutex
}

// defaultConfiguration returns the default configuration contained
// in the file default-conf.json
func newDefaultNginxCfg() *nginxConfiguration {
	return &nginxConfiguration{
		BodySize:                 bodySize,
		ErrorLogLevel:            errorLevel,
		UseHTS:                   true,
		HTSIncludeSubdomains:     true,
		HTSMaxAge:                htsMaxAge,
		GzipTypes:                gzipTypes,
		KeepAlive:                75,
		MaxWorkerConnections:     16384,
		ProxyConnectTimeout:      30,
		ProxyRealIPCIDR:          defIPCIDR,
		ProxyReadTimeout:         30,
		ProxySendTimeout:         30,
		ServerNameHashMaxSize:    512,
		ServerNameHashBucketSize: 64,
		SSLBufferSize:            sslBufferSize,
		SSLCiphers:               sslCiphers,
		SSLProtocols:             sslProtocols,
		SSLSessionCache:          true,
		SSLSessionCacheSize:      sslSessionCacheSize,
		SSLSessionTickets:        true,
		SSLSessionTimeout:        sslSessionTimeout,
		UseProxyProtocol:         false,
		UseGzip:                  true,
		WorkerProcesses:          strconv.Itoa(runtime.NumCPU()),
	}
}

// NewManager ...
func NewManager(kubeClient *client.Client, defaultSvc, customErrorSvc Service) *NginxManager {
	ngx := &NginxManager{
		ConfigFile:      "/etc/nginx/nginx.conf",
		defBackend:      defaultSvc,
		defCfg:          newDefaultNginxCfg(),
		defError:        customErrorSvc,
		defResolver:     strings.Join(getDnsServers(), " "),
		reloadLock:      &sync.Mutex{},
		sslDHParam:      ssl.SearchDHParamFile(sslDirectory),
		sslCertificates: ssl.CreateSSLCerts(sslDirectory),
	}

	ngx.loadTemplate()

	return ngx
}
