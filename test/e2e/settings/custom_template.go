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

package settings

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Custom Template", func() {
	f := framework.NewDefaultFramework("custom-template")

	BeforeEach(func() {
		configmap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: "custom-template",
			},
			Data: map[string]string{
				"nginx.tmpl": `
			# custom-template-test
			{{ $all := . }}
			{{ $servers := .Servers }}
			{{ $cfg := .Cfg }}
			{{ $IsIPV6Enabled := .IsIPV6Enabled }}
			{{ $healthzURI := .HealthzURI }}
			{{ $backends := .Backends }}
			{{ $proxyHeaders := .ProxySetHeaders }}
			{{ $addHeaders := .AddHeaders }}
			
			# Configuration checksum: {{ $all.Cfg.Checksum }}
			
			# setup custom paths that do not require root access
			pid /tmp/nginx.pid;
			
			daemon off;
			
			events {
				multi_accept        on;
				worker_connections  {{ $cfg.MaxWorkerConnections }};
				use                 epoll;
			}
			
			http {
				{{ if not $all.DisableLua }}
				lua_package_cpath "/usr/local/lib/lua/?.so;/usr/lib/lua-platform-path/lua/5.1/?.so;;";
				lua_package_path "/etc/nginx/lua/?.lua;/etc/nginx/lua/vendor/?.lua;/usr/local/lib/lua/?.lua;;";
			
				{{ buildLuaSharedDictionaries $servers $all.DynamicConfigurationEnabled $all.Cfg.DisableLuaRestyWAF }}
			
				init_by_lua_block {
					require("resty.core")
					collectgarbage("collect")
			
					local lua_resty_waf = require("resty.waf")
					lua_resty_waf.init()
			
					{{ if $all.DynamicConfigurationEnabled }}
					-- init modules
					local ok, res
			
					ok, res = pcall(require, "configuration")
					if not ok then
						error("require failed: " .. tostring(res))
					else
						configuration = res
						configuration.nameservers = { {{ buildResolversForLua $cfg.Resolver $cfg.DisableIpv6DNS }} }
					end
			
					ok, res = pcall(require, "balancer")
					if not ok then
						error("require failed: " .. tostring(res))
					else
						balancer = res
					end
					{{ end }}
			
					ok, res = pcall(require, "monitor")
					if not ok then
						error("require failed: " .. tostring(res))
					else
						monitor = res
					end
				}
			
				{{ if $all.DynamicConfigurationEnabled }}
				init_worker_by_lua_block {
					balancer.init_worker()
				}
				{{ end }}
				{{ end }}
				{{/* we use the value of the header X-Forwarded-For to be able to use the geo_ip module */}}
				{{ if $cfg.UseProxyProtocol }}
				real_ip_header      proxy_protocol;
				{{ else }}
				real_ip_header      {{ $cfg.ForwardedForHeader }};
				{{ end }}
			
				real_ip_recursive   on;
				{{ range $trusted_ip := $cfg.ProxyRealIPCIDR }}
				set_real_ip_from    {{ $trusted_ip }};
				{{ end }}
			
				{{ if $cfg.UseGeoIP }}
				{{/* databases used to determine the country depending on the client IP address */}}
				{{/* http://nginx.org/en/docs/http/ngx_http_geoip_module.html */}}
				{{/* this is require to calculate traffic for individual country using GeoIP in the status page */}}
				geoip_country       /etc/nginx/geoip/GeoIP.dat;
				geoip_city          /etc/nginx/geoip/GeoLiteCity.dat;
				geoip_org           /etc/nginx/geoip/GeoIPASNum.dat;
				geoip_proxy_recursive on;
				{{ end }}
		
				client_body_temp_path           /tmp/client-body;
				fastcgi_temp_path               /tmp/fastcgi-temp;
				proxy_temp_path                 /tmp/proxy-temp;
			
				include /etc/nginx/mime.types;
				default_type text/html;
			
				# Additional available variables:
				# $namespace
				# $ingress_name
				# $service_name
				# $service_port
				log_format upstreaminfo {{ if $cfg.LogFormatEscapeJSON }}escape=json {{ end }}'{{ buildLogFormatUpstream $cfg }}';
			
				{{/* map urls that should not appear in access.log */}}
				{{/* http://nginx.org/en/docs/http/ngx_http_log_module.html#access_log */}}
				map $request_uri $loggable {
					{{ range $reqUri := $cfg.SkipAccessLogURLs }}
					{{ $reqUri }} 0;{{ end }}
					default 1;
				}
			
				{{ if $cfg.DisableAccessLog }}
				access_log off;
				{{ else }}
				{{ if $cfg.EnableSyslog }}
				access_log syslog:server={{ $cfg.SyslogHost }}:{{ $cfg.SyslogPort }} upstreaminfo if=$loggable;
				{{ else }}
				access_log {{ $cfg.AccessLogPath }} upstreaminfo if=$loggable;
				{{ end }}
				{{ end }}
			
				{{ buildResolvers $cfg.Resolver $cfg.DisableIpv6DNS }}
			
				{{/* Whenever nginx proxies a request without a "Connection" header, the "Connection" header is set to "close" */}}
				{{/* when making the target request.  This means that you cannot simply use */}}
				{{/* "proxy_set_header Connection $http_connection" for WebSocket support because in this case, the */}}
				{{/* "Connection" header would be set to "" whenever the original request did not have a "Connection" header, */}}
				{{/* which would mean no "Connection" header would be in the target request.  Since this would deviate from */}}
				{{/* normal nginx behavior we have to use this approach. */}}
				# Retain the default nginx handling of requests without a "Connection" header
				map $http_upgrade $connection_upgrade {
					default          upgrade;
					''               close;
				}
			
				map {{ buildForwardedFor $cfg.ForwardedForHeader }} $the_real_ip {
				{{ if $cfg.UseProxyProtocol }}
					# Get IP address from Proxy Protocol
					default          $proxy_protocol_addr;
				{{ else }}
					default          $remote_addr;
				{{ end }}
				}
			
				# trust http_x_forwarded_proto headers correctly indicate ssl offloading
				map $http_x_forwarded_proto $pass_access_scheme {
					default          $http_x_forwarded_proto;
					''               $scheme;
				}
			
				# validate $pass_access_scheme and $scheme are http to force a redirect
				map "$scheme:$pass_access_scheme" $redirect_to_https {
					default          0;
					"http:http"      1;
					"https:http"     1;
				}
			
				map $http_x_forwarded_port $pass_server_port {
					default           $http_x_forwarded_port;
					''                $server_port;
				}
			
				{{ if $all.IsSSLPassthroughEnabled }}
				# map port {{ $all.ListenPorts.SSLProxy }} to 443 for header X-Forwarded-Port
				map $pass_server_port $pass_port {
					{{ $all.ListenPorts.SSLProxy }}              443;
					default          $pass_server_port;
				}
				{{ else }}
				map $pass_server_port $pass_port {
					{{ $all.ListenPorts.HTTPS }}              443;
					default          $pass_server_port;
				}
				{{ end }}
			
				# Obtain best http host
				map $http_host $this_host {
					default          $http_host;
					''               $host;
				}
			
				map $http_x_forwarded_host $best_http_host {
					default          $http_x_forwarded_host;
					''               $this_host;
				}
			
				# Reverse proxies can detect if a client provides a X-Request-ID header, and pass it on to the backend server.
				# If no such header is provided, it can provide a random value.
				map $http_x_request_id $req_id {
					default   $http_x_request_id;
					{{ if $cfg.GenerateRequestId }}
					""        $request_id;
					{{ end }}
				}
			
				{{ if $cfg.ComputeFullForwardedFor }}
				# We can't use $proxy_add_x_forwarded_for because the realip module
				# replaces the remote_addr too soon
				map $http_x_forwarded_for $full_x_forwarded_for {
					{{ if $all.Cfg.UseProxyProtocol }}
					default          "$http_x_forwarded_for, $proxy_protocol_addr";
					''               "$proxy_protocol_addr";
					{{ else }}
					default          "$http_x_forwarded_for, $realip_remote_addr";
					''               "$realip_remote_addr";
					{{ end}}
				}
				{{ end }}
			
				server_name_in_redirect off;
				port_in_redirect        off;
			
				ssl_protocols {{ $cfg.SSLProtocols }};
			
				# turn on session caching to drastically improve performance
				{{ if $cfg.SSLSessionCache }}
				ssl_session_cache builtin:1000 shared:SSL:{{ $cfg.SSLSessionCacheSize }};
				ssl_session_timeout {{ $cfg.SSLSessionTimeout }};
				{{ end }}
			
				# allow configuring ssl session tickets
				ssl_session_tickets {{ if $cfg.SSLSessionTickets }}on{{ else }}off{{ end }};
			
				{{ if not (empty $cfg.SSLSessionTicketKey ) }}
				ssl_session_ticket_key /etc/nginx/tickets.key;
				{{ end }}
			
				# slightly reduce the time-to-first-byte
				ssl_buffer_size {{ $cfg.SSLBufferSize }};
			
				{{ if not (empty $cfg.SSLCiphers) }}
				# allow configuring custom ssl ciphers
				ssl_ciphers '{{ $cfg.SSLCiphers }}';
				ssl_prefer_server_ciphers on;
				{{ end }}
			
				{{ if not (empty $cfg.SSLDHParam) }}
				# allow custom DH file http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam
				ssl_dhparam {{ $cfg.SSLDHParam }};
				{{ end }}
			
				{{ if not $cfg.EnableDynamicTLSRecords }}
				ssl_dyn_rec_size_lo 0;
				{{ end }}
			
				ssl_ecdh_curve {{ $cfg.SSLECDHCurve }};
			
				{{ if .CustomErrors }}
				# Custom error pages
				proxy_intercept_errors on;
				{{ end }}
			
				{{ range $errCode := $cfg.CustomHTTPErrors }}
				error_page {{ $errCode }} = @custom_{{ $errCode }};{{ end }}
			
				proxy_ssl_session_reuse on;
			
				{{ if $cfg.AllowBackendServerHeader }}
				proxy_pass_header Server;
				{{ end }}
			
				{{ range $header := $cfg.HideHeaders }}proxy_hide_header {{ $header }};
				{{ end }}
			
				{{ if not (empty $cfg.HTTPSnippet) }}
				# Custom code snippet configured in the configuration configmap
				{{ $cfg.HTTPSnippet }}
				{{ end }}
			
				{{ if not $all.DynamicConfigurationEnabled }}
				{{ range  $upstream := $backends }}
				{{ if eq $upstream.SessionAffinity.AffinityType "cookie" }}
				upstream sticky-{{ $upstream.Name }} {
					sticky hash={{ $upstream.SessionAffinity.CookieSessionAffinity.Hash }} name={{ $upstream.SessionAffinity.CookieSessionAffinity.Name }}{{if eq (len $upstream.SessionAffinity.CookieSessionAffinity.Locations) 1 }}{{ range $locationName, $locationPaths := $upstream.SessionAffinity.CookieSessionAffinity.Locations }}{{ if eq (len $locationPaths) 1 }} path={{ index $locationPaths 0 }}{{ end }}{{ end }}{{ end }} httponly;
			
					{{ if (gt $cfg.UpstreamKeepaliveConnections 0) }}
					keepalive {{ $cfg.UpstreamKeepaliveConnections }};
					{{ end }}
			
					{{ range $server := $upstream.Endpoints }}server {{ $server.Address | formatIP }}:{{ $server.Port }} max_fails={{ $server.MaxFails }} fail_timeout={{ $server.FailTimeout }};
					{{ end }}
				}
				{{ end }}
			
				upstream {{ $upstream.Name }} {
					{{ buildLoadBalancingConfig $upstream $cfg.LoadBalanceAlgorithm }}
			
					{{ if (gt $cfg.UpstreamKeepaliveConnections 0) }}
					keepalive {{ $cfg.UpstreamKeepaliveConnections }};
					{{ end }}
			
					{{ range $server := $upstream.Endpoints }}server {{ $server.Address | formatIP }}:{{ $server.Port }} max_fails={{ $server.MaxFails }} fail_timeout={{ $server.FailTimeout }};
					{{ end }}
				}
				{{ end }}
				{{ end }}
			
				{{ if $all.DynamicConfigurationEnabled }}
				upstream upstream_balancer {
					server 0.0.0.1; # placeholder
			
					balancer_by_lua_block {
						balancer.balance()
					}
			
					{{ if (gt $cfg.UpstreamKeepaliveConnections 0) }}
					keepalive {{ $cfg.UpstreamKeepaliveConnections }};
					{{ end }}
				}
				{{ end }}
			
				{{/* build the maps that will be use to validate the Whitelist */}}
				{{ range $server := $servers }}
				{{ range $location := $server.Locations }}
				{{ $path := buildLocation $location }}
			
				{{ if isLocationAllowed $location }}
				{{ if gt (len $location.Whitelist.CIDR) 0 }}
			
				# Deny for {{ print $server.Hostname  $path }}
				geo $the_real_ip {{ buildDenyVariable (print $server.Hostname "_"  $path) }} {
					default 1;
			
					{{ range $ip := $location.Whitelist.CIDR }}
					{{ $ip }} 0;{{ end }}
				}
				{{ end }}
				{{ end }}
				{{ end }}
				{{ end }}
			
				{{ range $rl := (filterRateLimits $servers ) }}
				# Ratelimit {{ $rl.Name }}
				geo $the_real_ip $whitelist_{{ $rl.ID }} {
					default 0;
					{{ range $ip := $rl.Whitelist }}
					{{ $ip }} 1;{{ end }}
				}
			
				# Ratelimit {{ $rl.Name }}
				map $whitelist_{{ $rl.ID }} $limit_{{ $rl.ID }} {
					0 {{ $cfg.LimitConnZoneVariable }};
					1 "";
				}
				{{ end }}
			
				{{/* build all the required rate limit zones. Each annotation requires a dedicated zone */}}
				{{/* 1MB -> 16 thousand 64-byte states or about 8 thousand 128-byte states */}}
				{{ range $zone := (buildRateLimitZones $servers) }}
				{{ $zone }}
				{{ end }}
			
				{{/* Build server redirects (from/to www) */}}
				{{ range $hostname, $to := .RedirectServers }}
				server {
					{{ range $address := $all.Cfg.BindAddressIpv4 }}
					listen {{ $address }}:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }};
					listen {{ $address }}:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }} ssl;
					{{ else }}
					listen {{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }};
					listen {{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }} ssl;
					{{ end }}
					{{ if $IsIPV6Enabled }}
					{{ range $address := $all.Cfg.BindAddressIpv6 }}
					listen {{ $address }}:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }};
					listen {{ $address }}:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }};
					{{ else }}
					listen [::]:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }};
					listen [::]:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }};
					{{ end }}
					{{ end }}
					server_name {{ $hostname }};
			
					{{ if ne $all.ListenPorts.HTTPS 443 }}
					{{ $redirect_port := (printf ":%v" $all.ListenPorts.HTTPS) }}
					return {{ $all.Cfg.HTTPRedirectCode }} $scheme://{{ $to }}{{ $redirect_port }}$request_uri;
					{{ else }}
					return {{ $all.Cfg.HTTPRedirectCode }} $scheme://{{ $to }}$request_uri;
					{{ end }}
				}
				{{ end }}
			
				{{ range $server := $servers }}
			
				## start server {{ $server.Hostname }}
				server {
					server_name {{ $server.Hostname }} {{ $server.Alias }};
					{{ template "SERVER" serverConfig $all $server }}
			
					{{ if not (empty $cfg.ServerSnippet) }}
					# Custom code snippet configured in the configuration configmap
					{{ $cfg.ServerSnippet }}
					{{ end }}
			
					{{ template "CUSTOM_ERRORS" $all }}
				}
				## end server {{ $server.Hostname }}
			
				{{ end }}
			
				# default server, used for NGINX healthcheck and access to nginx stats
				server {
					# Use the port {{ $all.ListenPorts.Status }} (random value just to avoid known ports) as default port for nginx.
					# Changing this value requires a change in:
					# https://github.com/kubernetes/ingress-nginx/blob/master/controllers/nginx/pkg/cmd/controller/nginx.go
					listen {{ $all.ListenPorts.Status }} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }};
					{{ if $IsIPV6Enabled }}listen [::]:{{ $all.ListenPorts.Status }} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }};{{ end }}
					set $proxy_upstream_name "-";
			
					location {{ $healthzURI }} {
						{{ if $cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
						access_log off;
						return 200;
					}
					{{ if not $all.DisableLua }}
					location /is-dynamic-lb-initialized {
						{{ if $cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
						access_log off;
			
						content_by_lua_block {
							local configuration = require("configuration")
							local backend_data = configuration.get_backends_data()
							if not backend_data then
								ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
								return
							end
			
							ngx.say("OK")
							ngx.exit(ngx.HTTP_OK)
						}
					}
					{{ end }}
					location /nginx_status {
						set $proxy_upstream_name "internal";
						{{ if $cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
			
						access_log off;
						stub_status on;
					}
			
					{{ if $all.DynamicConfigurationEnabled }}
					location /configuration {
						access_log off;
						{{ if $cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
			
						allow 127.0.0.1;
						{{ if $IsIPV6Enabled }}
						allow ::1;
						{{ end }}
						deny all;
			
						# this should be equals to configuration_data dict
						client_max_body_size                    "10m";
						proxy_buffering                         off;
			
						content_by_lua_block {
							configuration.call()
						}
					}
					{{ end }}
			
					location / {
						{{ if .CustomErrors }}
						proxy_set_header    X-Code 404;
						{{ end }}
						set $proxy_upstream_name "upstream-default-backend";
						{{ if $all.DynamicConfigurationEnabled }}
						proxy_pass          http://upstream_balancer;
						{{ else }}
						proxy_pass          http://upstream-default-backend;
						{{ end }}
					}
			
					{{ template "CUSTOM_ERRORS" $all }}
				}
			}
			
			{{/* definition of templates to avoid repetitions */}}
			{{ define "CUSTOM_ERRORS" }}
					{{ $dynamicConfig := .DynamicConfigurationEnabled}}
					{{ $proxySetHeaders := .ProxySetHeaders }}
					{{ range $errCode := .Cfg.CustomHTTPErrors }}
					location @custom_{{ $errCode }} {
						internal;
			
						proxy_intercept_errors off;
			
						proxy_set_header       X-Code             {{ $errCode }};
						proxy_set_header       X-Format           $http_accept;
						proxy_set_header       X-Original-URI     $request_uri;
						proxy_set_header       X-Namespace        $namespace;
						proxy_set_header       X-Ingress-Name     $ingress_name;
						proxy_set_header       X-Service-Name     $service_name;
						proxy_set_header       X-Service-Port     $service_port;
			
						set $proxy_upstream_name "upstream-default-backend";
			
						rewrite                (.*) / break;
			
						{{ if $dynamicConfig }}
						proxy_pass            http://upstream_balancer;
						{{ else }}
						proxy_pass            http://upstream-default-backend;
						{{ end }}
					}
					{{ end }}
			{{ end }}
			
			{{/* CORS support from https://michielkalkman.com/snippets/nginx-cors-open-configuration.html */}}
			{{ define "CORS" }}
					{{ $cors := .CorsConfig }}
					# Cors Preflight methods needs additional options and different Return Code
					if ($request_method = 'OPTIONS') {
					more_set_headers 'Access-Control-Allow-Origin: {{ $cors.CorsAllowOrigin }}';
					{{ if $cors.CorsAllowCredentials }} more_set_headers 'Access-Control-Allow-Credentials: {{ $cors.CorsAllowCredentials }}'; {{ end }}
					more_set_headers 'Access-Control-Allow-Methods: {{ $cors.CorsAllowMethods }}';
					more_set_headers 'Access-Control-Allow-Headers: {{ $cors.CorsAllowHeaders }}';
					more_set_headers 'Access-Control-Max-Age: {{ $cors.CorsMaxAge }}';
					more_set_headers 'Content-Type: text/plain charset=UTF-8';
					more_set_headers 'Content-Length: 0';
					return 204;
					}
			
					more_set_headers 'Access-Control-Allow-Origin: {{ $cors.CorsAllowOrigin }}';
					{{ if $cors.CorsAllowCredentials }} more_set_headers 'Access-Control-Allow-Credentials: {{ $cors.CorsAllowCredentials }}'; {{ end }}
					more_set_headers 'Access-Control-Allow-Methods: {{ $cors.CorsAllowMethods }}';
					more_set_headers 'Access-Control-Allow-Headers: {{ $cors.CorsAllowHeaders }}';
			
			{{ end }}
			
			{{/* definition of server-template to avoid repetitions with server-alias */}}
			{{ define "SERVER" }}
					{{ $all := .First }}
					{{ $server := .Second }}
					{{ range $address := $all.Cfg.BindAddressIpv4 }}
					listen {{ $address }}:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}};
					{{ else }}
					listen {{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}};
					{{ end }}
					{{ if $all.IsIPV6Enabled }}
					{{ range $address := $all.Cfg.BindAddressIpv6 }}
					listen {{ $address }}:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{ end }};
					{{ else }}
					listen [::]:{{ $all.ListenPorts.HTTP }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{ end }};
					{{ end }}
					{{ end }}
					set $proxy_upstream_name "-";
			
					{{/* Listen on {{ $all.ListenPorts.SSLProxy }} because port {{ $all.ListenPorts.HTTPS }} is used in the TLS sni server */}}
					{{/* This listener must always have proxy_protocol enabled, because the SNI listener forwards on source IP info in it. */}}
					{{ if not (empty $server.SSLCert.PemFileName) }}
					{{ range $address := $all.Cfg.BindAddressIpv4 }}
					listen {{ $address }}:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol {{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }} {{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}} ssl {{ if $all.Cfg.UseHTTP2 }}http2{{ end }};
					{{ else }}
					listen {{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol {{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }} {{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}} ssl {{ if $all.Cfg.UseHTTP2 }}http2{{ end }};
					{{ end }}
					{{ if $all.IsIPV6Enabled }}
					{{ range $address := $all.Cfg.BindAddressIpv6 }}
					{{ if not (empty $server.SSLCert.PemFileName) }}listen {{ $address }}:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }}{{ end }} {{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}} ssl {{ if $all.Cfg.UseHTTP2 }}http2{{ end }};
					{{ else }}
					{{ if not (empty $server.SSLCert.PemFileName) }}listen [::]:{{ if $all.IsSSLPassthroughEnabled }}{{ $all.ListenPorts.SSLProxy }} proxy_protocol{{ else }}{{ $all.ListenPorts.HTTPS }}{{ if $all.Cfg.UseProxyProtocol }} proxy_protocol{{ end }}{{ end }}{{ end }} {{ if eq $server.Hostname "_"}} default_server {{ if $all.Cfg.ReusePort }}reuseport{{ end }} backlog={{ $all.BacklogSize }}{{end}} ssl {{ if $all.Cfg.UseHTTP2 }}http2{{ end }};
					{{ end }}
					{{ end }}
					{{/* comment PEM sha is required to detect changes in the generated configuration and force a reload */}}
					# PEM sha: {{ $server.SSLCert.PemSHA }}
					ssl_certificate                         {{ $server.SSLCert.PemFileName }};
					ssl_certificate_key                     {{ $server.SSLCert.PemFileName }};
					{{ if not (empty $server.SSLCert.FullChainPemFileName)}}
					ssl_trusted_certificate                 {{ $server.SSLCert.FullChainPemFileName }};
					ssl_stapling                            on;
					ssl_stapling_verify                     on;
					{{ end }}
					{{ end }}
			
					{{ if not (empty $server.AuthTLSError) }}
					# {{ $server.AuthTLSError }}
					return 403;
					{{ else }}
			
					{{ if not (empty $server.CertificateAuth.CAFileName) }}
					# PEM sha: {{ $server.CertificateAuth.PemSHA }}
					ssl_client_certificate                  {{ $server.CertificateAuth.CAFileName }};
					ssl_verify_client                       {{ $server.CertificateAuth.VerifyClient }};
					ssl_verify_depth                        {{ $server.CertificateAuth.ValidationDepth }};
					{{ if not (empty $server.CertificateAuth.ErrorPage)}}
					error_page 495 496 = {{ $server.CertificateAuth.ErrorPage }};
					{{ end }}
					{{ end }}
			
					{{ if not (empty $server.SSLCiphers) }}
					ssl_ciphers                             {{ $server.SSLCiphers }};
					{{ end }}
			
					{{ if not (empty $server.ServerSnippet) }}
					{{ $server.ServerSnippet }}
					{{ end }}
			
					{{ range $location := $server.Locations }}
					{{ $path := buildLocation $location }}
					{{ $proxySetHeader := proxySetHeader $location }}
					{{ $authPath := buildAuthLocation $location }}
			
					{{ if not (empty $location.Rewrite.AppRoot)}}
					if ($uri = /) {
						return 302 {{ $location.Rewrite.AppRoot }};
					}
					{{ end }}
			
					location {{ $path }} {
					}
					{{ end }}
					{{ end }}
			
					{{ if eq $server.Hostname "_" }}
					# health checks in cloud providers require the use of port {{ $all.ListenPorts.HTTP }}
					location {{ $all.HealthzURI }} {
						{{ if $all.Cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
			
						access_log off;
						return 200;
					}
			
					# this is required to avoid error if nginx is being monitored
					# with an external software (like sysdig)
					location /nginx_status {
						{{ if $all.Cfg.EnableOpentracing }}
						opentracing off;
						{{ end }}
			
						{{ range $v := $all.NginxStatusIpv4Whitelist }}
						allow {{ $v }};
						{{ end }}
						{{ if $all.IsIPV6Enabled -}}
						{{ range $v := $all.NginxStatusIpv6Whitelist }}
						allow {{ $v }};
						{{ end }}
						{{ end -}}
						deny all;
			
						access_log off;
						stub_status on;
					}
			
					{{ end }}
			
			{{ end }}				
				`,
			},
		}

		_, err := f.KubeClientSet.CoreV1().ConfigMaps(f.IngressController.Namespace).Create(configmap)
		Expect(err).NotTo(HaveOccurred())

		err = framework.UpdateDeployment(f.KubeClientSet, f.IngressController.Namespace, "nginx-ingress-controller", 1,
			func(deployment *appsv1beta1.Deployment) error {
				deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
					{
						Name:      "custom-template",
						ReadOnly:  true,
						MountPath: "/etc/nginx/template",
					},
				}

				deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
					{
						Name: "custom-template",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "custom-template",
								},
								Items: []corev1.KeyToPath{
									{
										Key:  "nginx.tmpl",
										Path: "nginx.tmpl",
									},
								},
							},
						},
					},
				}

				_, err := f.KubeClientSet.AppsV1beta1().Deployments(f.IngressController.Namespace).Update(deployment)
				return err
			})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
	})

	It("should return a custom tempĺate", func() {
		resp, _, errs := gorequest.New().
			Get(f.IngressController.HTTPURL).
			End()
		Expect(len(errs)).Should(BeNumerically("==", 0))
		Expect(resp.StatusCode).Should(Equal(404))
	})
})
