package defaults

import "net"

// Backend defines the mandatory configuration that an Ingress controller must provide
// The reason of this requirements is the annotations are generic. If some implementation do not supports
// one or more annotations it just can provides defaults
type Backend struct {
	// AppRoot contains the AppRoot for apps that doesn't exposes its content in the 'root' context
	AppRoot string `json:"app-root"`

	// enables which HTTP codes should be passed for processing with the error_page directive
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page
	// By default this is disabled
	CustomHTTPErrors []int `json:"custom-http-errors,-"`

	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	ProxyBodySize string `json:"proxy-body-size"`

	// Defines a timeout for establishing a connection with a proxied server.
	// It should be noted that this timeout cannot usually exceed 75 seconds.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout
	ProxyConnectTimeout int `json:"proxy-connect-timeout"`

	// Timeout in seconds for reading a response from the proxied server. The timeout is set only between
	// two successive read operations, not for the transmission of the whole response
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout
	ProxyReadTimeout int `json:"proxy-read-timeout"`

	// Timeout in seconds for transmitting a request to the proxied server. The timeout is set only between
	// two successive write operations, not for the transmission of the whole request.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout
	ProxySendTimeout int `json:"proxy-send-timeout"`

	// Sets the size of the buffer used for reading the first part of the response received from the
	// proxied server. This part usually contains a small response header.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size)
	ProxyBufferSize string `json:"proxy-buffer-size"`

	// Sets a text that should be changed in the path attribute of the “Set-Cookie” header fields of
	// a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_path
	ProxyCookiePath string `json:"proxy-cookie-path"`

	// Sets a text that should be changed in the domain attribute of the “Set-Cookie” header fields
	// of a proxied server response.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cookie_domain
	ProxyCookieDomain string `json:"proxy-cookie-domain"`

	// Specifies in which cases a request should be passed to the next server.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream
	ProxyNextUpstream string `json:"proxy-next-upstream"`

	// Name server/s used to resolve names of upstream servers into IP addresses.
	// The file /etc/resolv.conf is used as DNS resolution configuration.
	Resolver []net.IP

	// SkipAccessLogURLs sets a list of URLs that should not appear in the NGINX access log
	// This is useful with urls like `/health` or `health-check` that make "complex" reading the logs
	// By default this list is empty
	SkipAccessLogURLs []string `json:"skip-access-log-urls,-"`

	// Enables or disables the redirect (301) to the HTTPS port
	SSLRedirect bool `json:"ssl-redirect"`

	// Enables or disables the redirect (301) to the HTTPS port even without TLS cert
	// This is useful if doing SSL offloading outside of cluster eg AWS ELB
	ForceSSLRedirect bool `json:"force-ssl-redirect"`

	// Enables or disables the specification of port in redirects
	// Default: false
	UsePortInRedirects bool `json:"use-port-in-redirects"`

	// Number of unsuccessful attempts to communicate with the server that should happen in the
	// duration set by the fail_timeout parameter to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	UpstreamMaxFails int `json:"upstream-max-fails"`

	// Time during which the specified number of unsuccessful attempts to communicate with
	// the server should happen to consider the server unavailable
	// http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream
	// Default: 0, ie use platform liveness probe
	UpstreamFailTimeout int `json:"upstream-fail-timeout"`

	// WhitelistSourceRange allows limiting access to certain client addresses
	// http://nginx.org/en/docs/http/ngx_http_access_module.html
	WhitelistSourceRange []string `json:"whitelist-source-range,-"`

	// Limits the rate of response transmission to a client.
	// The rate is specified in bytes per second. The zero value disables rate limiting.
	// The limit is set per a request, and so if a client simultaneously opens two connections,
	// the overall rate will be twice as much as the specified limit.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#limit_rate
	LimitRate int `json:"limit-rate"`

	// Sets the initial amount after which the further transmission of a response to a client will be rate limited.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#limit_rate_after
	LimitRateAfter int `json:"limit-rate-after"`
}
