package main

type EnvoyCluster struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	ConnectTimeoutMs int               `json:"connect_timeout_ms"`
	LbType           string            `json:"lb_type"`
	ServiceName      string            `json:"service_name"`
	HealthCheck      *EnvoyHealthCheck `json:"health_check"`
	// TODO(cmaloney): max_requests_per_connection, circuit_breakers, ssl_context
	// features, http_codec_options, http2_settings, dns*, outlier_detection

}

type RdsConfig struct {
	Cluster         string `json:"cluster"`
	RouteConfigName string `json:"route_config_name"`
	RefreshDelayMs  int    `json:"refresh_delay_ms"`
}

type HttpConnectionManagerConfig struct {
	CodecType  string    `json:"codec_type"`
	StatPrefix string    `json:"stat_prefix"`
	Rds        RdsConfig `json:"rds"`
	// TODO(cmaloney): route_config
	Filters []EnvoyFilter `json:"filters"`
	// TODO(cmaloney): add_user_agent, tracing, http2_settings, server_name,
	// idle_timeout_s, drain_timeout_ms, access_log, use_remote_address,
	// forward_client_cert, set_current_client_cert_details, generate_request_id
}

type EnvoyFilter struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type EnvoyHealthCheck struct {
	Type               string `json:"type"`
	TimeoutMs          int    `json:"timeout_ms"`
	IntervalMs         int    `json:"interval_ms"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	Path               string `json:"path"`
	// TODO(cmaloney): send/recieve for TCP healthchecks
}

type EnvoyHost struct {
	IpAddress string `json:"ip_address"`
	Port      int    `json:"port"`
	/* TODO(cmaloney): Do based on labels
	Tags      *struct {
		Az                  string `json:az`
		Canary              string `json:canary`
		LoadBalancingWeight string `json:load_balancing_weight`
	} `json:tags`
	*/
}

type EnvoyListener struct {
	Name    string        `json:"name"`
	Address string        `json:"address"`
	Filters []EnvoyFilter `json:"filters"`
	// TODO(cmaloney): ssl_context, bind_to_port
	UseProxyProto bool `json:"use_proxy_proto"`
	// TODO(cmaloney): use_original_dst, per_connection_buffer_limit_bytes
}

type EnvoyRetryPolicy struct {
	RetryOn         string `json:"retry_on"`
	NumRetries      int    `json:"num_retries"`
	PerTryTimeoutMs int    `json:"per_try_timeout_ms"`
}

type EnvoyRoute struct {
	Prefix string `json:"prefix"`
	// Path    string `json:path`
	Cluster string `json:"cluster"`
	// TODO(cmaloney): cluster_header, weighted_clusters, host_redirect, path_redirect
	// prefix_rewrite, host_rewrite, auto_host_rewrite, case_sensitive, runtime
	TimeoutMs   int              `json:"timeout_ms"`
	RetryPolicy EnvoyRetryPolicy `json:"retry_policy"`
	// TODO(cmaloney): priority, headers, request_headers_to_add, opaque_config, rate_limits
	// include_vh_rate_limits, hash_policy
}

type EnvoyVirtualHost struct {
	Name    string       `json:"name"`
	Domains []string     `json:"domains"`
	Routes  []EnvoyRoute `json:"routes"`
	// TODO(cmaloney): requre_ssl, virtual_clusters, rate_limits, request_headers_to_add
}

type CdsResponse struct {
	Clusters []EnvoyCluster `json:"clusters"`
}

type LdsResponse struct {
	Listeners []EnvoyListener `json:"listeners"`
}

type RdsResponse struct {
	ValidateClusters bool               `json:"validate_clusters"`
	VirtualHosts     []EnvoyVirtualHost `json:"virtual_hosts"`
	// TODO(cmaloney): internal_only_headers, response_headers_to_add,
	// response_headers_to_remove, request_headers_to_add
}

type SdsResponse struct {
	Hosts []EnvoyHost `json:"hosts"`
}

type Responses struct {
	Cds CdsResponse
	Lds LdsResponse
	Rds RdsResponse
	Sds map[string]SdsResponse
}
