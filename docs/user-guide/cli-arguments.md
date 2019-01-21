# Command line arguments

The following command line arguments are accepted by the Ingress controller executable.

They are set in the container spec of the `nginx-ingress-controller` Deployment manifest

| Argument | Description |
|----------|-------------|
| `--alsologtostderr`               | log to standard error as well as files |
| `--annotations-prefix string`     | Prefix of the Ingress annotations specific to the NGINX controller. (default "nginx.ingress.kubernetes.io") |
| `--apiserver-host string`         | Address of the Kubernetes API server. Takes the form "protocol://address:port". If not specified, it is assumed the program runs inside a Kubernetes cluster and local discovery is attempted. |
| `--configmap string`              | Name of the ConfigMap containing custom global configurations for the controller. |
| `--default-backend-service string` | Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form "namespace/name". The controller configures NGINX to forward requests to the first port of this Service. If not specified, a 404 page will be returned directly from NGINX.|
| `--default-server-port int`       | When `default-backend-service` is not specified or specified service does not have any endpoint, a local endpoint with this port will be used to serve 404 page from inside Nginx. |
| `--default-ssl-certificate string` | Secret containing a SSL certificate to be used by the default HTTPS server (catch-all). Takes the form "namespace/name". |
| `--election-id string`            | Election id to use for Ingress status updates. (default "ingress-controller-leader") |
| `--enable-dynamic-certificates`   | Dynamically serves certificates instead of reloading NGINX when certificates are created, updated, or deleted. Currently does not support OCSP stapling, so --enable-ssl-chain-completion must be turned off. Assuming the certificate is generated with a 2048 bit RSA key/cert pair, this feature can store roughly 5000 certificates. This is an experiemental feature that currently is not ready for production use. Feature backed by OpenResty Lua libraries. (disabled by default) |
| `--enable-ssl-chain-completion`   | Autocomplete SSL certificate chains with missing intermediate CA certificates. A valid certificate chain is required to enable OCSP stapling. Certificates uploaded to Kubernetes must have the "Authority Information Access" X.509 v3 extension for this to succeed. (default true) |
| `--enable-ssl-passthrough`        | Enable SSL Passthrough. |
| `--force-namespace-isolation`     | Force namespace isolation. Prevents Ingress objects from referencing Secrets and ConfigMaps located in a different namespace than their own. May be used together with watch-namespace. |
| `--health-check-path string`      | URL path of the health check endpoint. Configured inside the NGINX status server. All requests received on the port defined by the healthz-port parameter are forwarded internally to this path. (default "/healthz") |
| `--health-check-timeout duration` | Time limit, in seconds, for a probe to health-check-path to succeed. (default 10) |
| `--healthz-port int`              | Port to use for the healthz endpoint. (default 10254) |
| `--http-port int`                 | Port to use for servicing HTTP traffic. (default 80) |
| `--https-port int`                | Port to use for servicing HTTPS traffic. (default 443) |
| `--ingress-class string`          | Name of the ingress class this controller satisfies. The class of an Ingress object is set using the annotation "kubernetes.io/ingress.class". All ingress classes are satisfied if this parameter is left empty. |
| `--kubeconfig string`             | Path to a kubeconfig file containing authorization and API server information. |
| `--log_backtrace_at traceLocation` | when logging hits line file:N, emit a stack trace (default :0) |
| `--log_dir string`                | If non-empty, write log files in this directory |
| `--logtostderr`                   | log to standard error instead of files (default true) |
| `--profiling`                     | Enable profiling via web interface host:port/debug/pprof/ (default true) |
| `--publish-service string`        | Service fronting the Ingress controller. Takes the form "namespace/name". When used together with update-status, the controller mirrors the address of this service's endpoints to the load-balancer status of all Ingress objects it satisfies. |
| `--publish-status-address string` | Customized address to set as the load-balancer status of Ingress objects this controller satisfies. Requires the update-status parameter. |
| `--report-node-internal-ip-address` | Set the load-balancer status of Ingress objects to internal Node addresses instead of external. Requires the update-status parameter. |
| `--sort-backends`                 | Sort servers inside NGINX upstreams. |
| `--ssl-passthrough-proxy-port int` | Port to use internally for SSL Passthrough. (default 442) |
| `--stderrthreshold severity`      | logs at or above this threshold go to stderr (default 2) |
| `--sync-period duration`          | Period at which the controller forces the repopulation of its local object stores. Disabled by default. |
| `--sync-rate-limit float32`       | Define the sync frequency upper limit (default 0.3) |
| `--tcp-services-configmap string` | Name of the ConfigMap containing the definition of the TCP services to expose. The key in the map indicates the external port to be used. The value is a reference to a Service in the form "namespace/name:port", where "port" can either be a port number or name. TCP ports 80 and 443 are reserved by the controller for servicing HTTP traffic. |
| `--udp-services-configmap string` | Name of the ConfigMap containing the definition of the UDP services to expose. The key in the map indicates the external port to be used. The value is a reference to a Service in the form "namespace/name:port", where "port" can either be a port name or number. |
| `--update-status`                 | Update the load-balancer status of Ingress objects this controller satisfies. Requires setting the publish-service parameter to a valid Service reference. (default true) |
| `--update-status-on-shutdown`     | Update the load-balancer status of Ingress objects when the controller shuts down. Requires the update-status parameter. (default true) |
| `-v`, `--v Level`                 | log level for V logs |
| `--version`                       | Show release information about the NGINX Ingress controller and exit. |
| `--vmodule moduleSpec`            | comma-separated list of pattern=N settings for file-filtered logging |
| `--watch-namespace string`        | Namespace the controller watches for updates to Kubernetes objects. This includes Ingresses, Services and all configuration resources. All namespaces are watched if this parameter is left empty. |
| `--disable-catch-all`             | Disable support for catch-all Ingresses. |