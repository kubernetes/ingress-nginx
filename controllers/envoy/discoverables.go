package main

// k8s endpoint
// envoy service
type Host struct {
	IpAddress string
	Port      string
	Canary    bool
}

// k8s services
// envoy cluster
type Cluster struct {
	HealthCheckPath string
	Hosts           []Host
}

type Path struct {
	// TODO(cmaloney): Formally in k8s path can be regex. We only allow strings.
	Prefix  string
	Cluster string
}

type VirtualHost struct {
	Domain string
	// Matched in order
	Paths []Path
}

// K8s Servers
type DiscoveryItems struct {
	// Name -> Cluster
	Clusters map[string]Cluster
	Routes   map[string]VirtualHost
}
