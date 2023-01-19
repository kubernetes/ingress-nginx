# Command line arguments

The following command line arguments are accepted by the Ingress controller executable.

They are set in the container spec of the `ingress-nginx-controller` Deployment manifest

| Argument | Description |
|----------|-------------|
| `--annotations-prefix`             | Prefix of the Ingress annotations specific to the NGINX controller. (default "nginx.ingress.kubernetes.io") |
| `--apiserver-host`                 | Address of the Kubernetes API server. Takes the form "protocol://address:port". If not specified, it is assumed the program runs inside a Kubernetes cluster and local discovery is attempted. |
| `--certificate-authority`          | Path to a cert file for the certificate authority. This certificate is used only when the flag --apiserver-host is specified. |
| `--configmap`                      | Name of the ConfigMap containing custom global configurations for the controller. |
| `--controller-class`                      | Ingress Class Controller value this Ingress satisfies. The class of an Ingress object is set using the field IngressClassName in Kubernetes clusters version v1.19.0 or higher. The .spec.controller value of the IngressClass referenced in an Ingress Object should be the same value specified here to make this object be watched. |
| `--deep-inspect`                   | Enables ingress object security deep inspector. (default true) |
| `--default-backend-service`        | Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form "namespace/name". The controller configures NGINX to forward requests to the first port of this Service. |
| `--default-server-port`            | Port to use for exposing the default server (catch-all). (default 8181) |
| `--default-ssl-certificate`        | Secret containing a SSL certificate to be used by the default HTTPS server (catch-all). Takes the form "namespace/name". |
| `--disable-catch-all`              | Disable support for catch-all Ingresses. (default false) |
| `--disable-full-test` | Disable full test of all merged ingresses at the admission stage and tests the template of the ingress being created or updated  (full test of all ingresses is enabled by default). |
| `--disable-svc-external-name` | Disable support for Services of type ExternalName. (default false) |
| `--disable-sync-events` | Disables the creation of 'Sync' Event resources, but still logs them |
| `--dynamic-configuration-retries` | Number of times to retry failed dynamic configuration before failing to sync an ingress. (default 15) |
| `--election-id`                    | Election id to use for Ingress status updates. (default "ingress-controller-leader") |
| `--enable-metrics`                 | Enables the collection of NGINX metrics. (default true) |
| `--enable-ssl-chain-completion`    | Autocomplete SSL certificate chains with missing intermediate CA certificates. Certificates uploaded to Kubernetes must have the "Authority Information Access" X.509 v3 extension for this to succeed. (default false)|
| `--enable-ssl-passthrough`         | Enable SSL Passthrough. (default false) |
| `--enable-topology-aware-routing`  | Enable topology aware hints feature, needs service object annotation service.kubernetes.io/topology-aware-hints sets to auto. (default false) |
| `--health-check-path`              | URL path of the health check endpoint. Configured inside the NGINX status server. All requests received on the port defined by the healthz-port parameter are forwarded internally to this path. (default "/healthz") |
| `--health-check-timeout`           | Time limit, in seconds, for a probe to health-check-path to succeed. (default 10) |
| `--healthz-port`                   | Port to use for the healthz endpoint. (default 10254) |
| `--healthz-host`                   | Address to bind the healthz endpoint. |
| `--http-port`                      | Port to use for servicing HTTP traffic. (default 80) |
| `--https-port`                     | Port to use for servicing HTTPS traffic. (default 443) |
| `--ingress-class`                  | Name of the ingress class this controller satisfies. The class of an Ingress object is set using the field IngressClassName in Kubernetes clusters version v1.18.0 or higher or the annotation "kubernetes.io/ingress.class" (deprecated). If this parameter is not set, or set to the default value of "nginx", it will handle ingresses with either an empty or "nginx" class name. |
| `--ingress-class-by-name`          | Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class. (default false). |
| `--internal-logger-address`        | Address to be used when binding internal syslogger. (default 127.0.0.1:11514) |
| `--kubeconfig`                     | Path to a kubeconfig file containing authorization and API server information. |
| `--length-buckets`                     | Set of buckets which will be used for prometheus histogram metrics such as RequestLength, ResponseLength. (default `[10, 20, 30, 40, 50, 60, 70, 80, 90, 100]`) |
| `--maxmind-edition-ids`            | Maxmind edition ids to download GeoLite2 Databases. (default "GeoLite2-City,GeoLite2-ASN") |
| `--maxmind-retries-timeout`        | Maxmind downloading delay between 1st and 2nd attempt, 0s - do not retry to download if something went wrong. (default 0s) |
| `--maxmind-retries-count`          | Number of attempts to download the GeoIP DB. (default 1) |
| `--maxmind-license-key`            | Maxmind license key to download GeoLite2 Databases. https://blog.maxmind.com/2019/12/18/significant-changes-to-accessing-and-using-geolite2-databases . |
| `--maxmind-mirror`            | Maxmind mirror url (example: http://geoip.local/databases. |
| `--metrics-per-host`               | Export metrics per-host. (default true) |
| `--monitor-max-batch-size`               | Max batch size of NGINX metrics. (default 10000)|
| `--post-shutdown-grace-period`     | Additional delay in seconds before controller container exits. (default 10) |
| `--profiler-port`                  | Port to use for expose the ingress controller Go profiler when it is enabled. (default 10245) |
| `--profiling`                      | Enable profiling via web interface host:port/debug/pprof/ . (default true) |
| `--publish-service`                | Service fronting the Ingress controller. Takes the form "namespace/name". When used together with update-status, the controller mirrors the address of this service's endpoints to the load-balancer status of all Ingress objects it satisfies. |
| `--publish-status-address`         | Customized address (or addresses, separated by comma) to set as the load-balancer status of Ingress objects this controller satisfies. Requires the update-status parameter. |
| `--report-node-internal-ip-address`| Set the load-balancer status of Ingress objects to internal Node addresses instead of external. Requires the update-status parameter. (default false) |
| `--report-status-classes`          | If true, report status classes in metrics (2xx, 3xx, 4xx and 5xx) instead of full status codes. (default false) |
| `--ssl-passthrough-proxy-port`     | Port to use internally for SSL Passthrough. (default 442) |
| `--status-port`                    | Port to use for the lua HTTP endpoint configuration. (default 10246) |
| `--status-update-interval`         | Time interval in seconds in which the status should check if an update is required. Default is 60 seconds. (default 60) |
| `--stream-port`                    | Port to use for the lua TCP/UDP endpoint configuration. (default 10247) |
| `--sync-period`                    | Period at which the controller forces the repopulation of its local object stores. Disabled by default. |
| `--sync-rate-limit`                | Define the sync frequency upper limit. (default 0.3) |
| `--tcp-services-configmap`         | Name of the ConfigMap containing the definition of the TCP services to expose. The key in the map indicates the external port to be used. The value is a reference to a Service in the form "namespace/name:port", where "port" can either be a port number or name. TCP ports 80 and 443 are reserved by the controller for servicing HTTP traffic. |
| `--time-buckets`         | Set of buckets which will be used for prometheus histogram metrics such as RequestTime, ResponseTime. (default `[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]`) |
| `--udp-services-configmap`         | Name of the ConfigMap containing the definition of the UDP services to expose. The key in the map indicates the external port to be used. The value is a reference to a Service in the form "namespace/name:port", where "port" can either be a port name or number. |
| `--update-status`                  | Update the load-balancer status of Ingress objects this controller satisfies. Requires setting the publish-service parameter to a valid Service reference. (default true) |
| `--update-status-on-shutdown`      | Update the load-balancer status of Ingress objects when the controller shuts down. Requires the update-status parameter. (default true) |
| `--shutdown-grace-period`          | Seconds to wait after receiving the shutdown signal, before stopping the nginx process. (default 0) |
| `--size-buckets`          | Set of buckets which will be used for prometheus histogram metrics such as BytesSent. (default `[10, 100, 1000, 10000, 100000, 1e+06, 1e+07]`) |
| `-v, --v Level`                    | number for the log level verbosity |
| `--validating-webhook`             | The address to start an admission controller on to validate incoming ingresses. Takes the form "<host>:port". If not provided, no admission controller is started. |
| `--validating-webhook-certificate` | The path of the validating webhook certificate PEM. |
| `--validating-webhook-key`         | The path of the validating webhook key PEM. |
| `--version`                        | Show release information about the NGINX Ingress controller and exit. |
| `--watch-ingress-without-class`                        | Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified. (default false) |
| `--watch-namespace`                | Namespace the controller watches for updates to Kubernetes objects. This includes Ingresses, Services and all configuration resources. All namespaces are watched if this parameter is left empty. |
| `--watch-namespace-selector`       | The controller will watch namespaces whose labels match the given selector. This flag only takes effective when `--watch-namespace` is empty. |
