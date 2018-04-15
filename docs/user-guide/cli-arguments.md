# Command line arguments

```console
Usage of :
      --alsologtostderr                   log to standard error as well as files
      --annotations-prefix string         Prefix of the ingress annotations. (default "nginx.ingress.kubernetes.io")
      --apiserver-host string             The address of the Kubernetes Apiserver to connect to in the format of protocol://address:port, e.g., http://localhost:8080. If not specified, the assumption is that the binary runs inside a Kubernetes cluster and local discovery is attempted.
      --configmap string                  Name of the ConfigMap that contains the custom configuration to use
      --default-backend-service string    Service used to serve a 404 page for the default backend. Takes the form
		namespace/name. The controller uses the first node port of this Service for
		the default backend.
      --default-server-port int           Default port to use for exposing the default server (catch all) (default 8181)
      --default-ssl-certificate string    Name of the secret
		that contains a SSL certificate to be used as default for a HTTPS catch-all server.
		Takes the form <namespace>/<secret name>.
      --election-id string                Election id to use for status update. (default "ingress-controller-leader")
      --enable-dynamic-configuration      When enabled controller will try to avoid Nginx reloads as much as possible by using Lua. Disabled by default.
      --enable-ssl-chain-completion       Defines if the nginx ingress controller should check the secrets for missing intermediate CA certificates.
		If the certificate contain issues chain issues is not possible to enable OCSP.
		Default is true. (default true)
      --enable-ssl-passthrough            Enable SSL passthrough feature. Default is disabled
      --force-namespace-isolation         Force namespace isolation. This flag is required to avoid the reference of secrets or
		configmaps located in a different namespace than the specified in the flag --watch-namespace.
      --health-check-path string          Defines
		the URL to be used as health check inside in the default server in NGINX. (default "/healthz")
      --healthz-port int                  port for healthz endpoint. (default 10254)
      --http-port int                     Indicates the port to use for HTTP traffic (default 80)
      --https-port int                    Indicates the port to use for HTTPS traffic (default 443)
      --ingress-class string              Name of the ingress class to route through this controller.
      --kubeconfig string                 Path to kubeconfig file with authorization and master location information.
      --log_backtrace_at traceLocation    when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                    If non-empty, write log files in this directory
      --logtostderr                       log to standard error instead of files (default true)
      --profiling                         Enable profiling via web interface host:port/debug/pprof/ (default true)
      --publish-service string            Service fronting the ingress controllers. Takes the form namespace/name.
		The controller will set the endpoint records on the ingress objects to reflect those on the service.
      --publish-status-address string     User customized address to be set in the status of ingress resources. The controller will set the
		endpoint records on the ingress using this address.
      --report-node-internal-ip-address   Defines if the nodes IP address to be returned in the ingress status should be the internal instead of the external IP address
      --sort-backends                     Defines if backends and its endpoints should be sorted
      --ssl-passtrough-proxy-port int     Default port to use internally for SSL when SSL Passthgough is enabled (default 442)
      --status-port int                   Indicates the TCP port to use for exposing the nginx status page (default 18080)
      --stderrthreshold severity          logs at or above this threshold go to stderr (default 2)
      --sync-period duration              Relist and confirm cloud resources this often. Default is 10 minutes (default 10m0s)
      --sync-rate-limit float32           Define the sync frequency upper limit (default 0.3)
      --tcp-services-configmap string     Name of the ConfigMap that contains the definition of the TCP services to expose.
		The key in the map indicates the external port to be used. The value is the name of the
		service with the format namespace/serviceName and the port of the service could be a
		number of the name of the port.
		The ports 80 and 443 are not allowed as external ports. This ports are reserved for the backend
      --udp-services-configmap string     Name of the ConfigMap that contains the definition of the UDP services to expose.
		The key in the map indicates the external port to be used. The value is the name of the
		service with the format namespace/serviceName and the port of the service could be a
		number of the name of the port.
      --update-status                     Indicates if the
		ingress controller should update the Ingress status IP/hostname. Default is true (default true)
      --update-status-on-shutdown         Indicates if the
		ingress controller should update the Ingress status IP/hostname when the controller
		is being stopped. Default is true (default true)
  -v, --v Level                           log level for V logs
      --version                           Shows release information about the NGINX Ingress controller
      --vmodule moduleSpec                comma-separated list of pattern=N settings for file-filtered logging
      --watch-namespace string            Namespace to watch for Ingress. Default is to watch all namespaces
```
