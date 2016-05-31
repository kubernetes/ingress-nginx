# Nginx Ingress Controller

This is a nginx Ingress controller that uses [ConfigMap](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/configmap.md) to store the nginx configuration. See [Ingress controller documentation](../README.md) for details on how it works.


## What it provides?

- Ingress controller
- nginx 1.9.x with
- SSL support
- custom ssl_dhparam (optional). Just mount a secret with a file named `dhparam.pem`.
- support for TCP services (flag `--tcp-services-configmap`)
- custom nginx configuration using [ConfigMap](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/configmap.md)


## Requirements
- default backend [404-server](https://github.com/kubernetes/contrib/tree/master/404-server)


## Dry running the Ingress controller

Before deploying the controller to production you might want to run it outside the cluster and observe it.

```console
$ make controller
$ mkdir /etc/nginx-ssl
$ ./nginx-ingress-controller --running-in-cluster=false --default-backend-service=kube-system/default-http-backend
```


## Deploy the Ingress controller

First create a default backend:
```
$ kubectl create -f examples/default-backend.yaml
$ kubectl expose rc default-http-backend --port=80 --target-port=8080 --name=default-http-backend
```

Loadbalancers are created via a ReplicationController or Daemonset:

```
$ kubectl create -f examples/default/rc-default.yaml
```

## HTTP

First we need to deploy some application to publish. To keep this simple we will use the [echoheaders app](https://github.com/kubernetes/contrib/blob/master/ingress/echoheaders/echo-app.yaml) that just returns information about the http request as output
```
kubectl run echoheaders --image=gcr.io/google_containers/echoserver:1.4 --replicas=1 --port=8080
```

Now we expose the same application in two different services (so we can create different Ingress rules)
```
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-x
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-y
```

Next we create a couple of Ingress rules
```
kubectl create -f examples/ingress.yaml
```

we check that ingress rules are defined:
```
$ kubectl get ing
NAME      RULE          BACKEND   ADDRESS
echomap   -
          foo.bar.com
          /foo          echoheaders-x:80
          bar.baz.com
          /bar          echoheaders-y:80
          /foo          echoheaders-x:80
```

Before the deploy of the Ingress controller we need a default backend [404-server](https://github.com/kubernetes/contrib/tree/master/404-server)
```
kubectl create -f examples/default-backend.yaml
kubectl expose rc default-http-backend --port=80 --target-port=8080 --name=default-http-backend
```

Check NGINX it is running with the defined Ingress rules:

```
$ LBIP=$(kubectl get node `kubectl get po -l name=nginx-ingress-lb --template '{{range .items}}{{.spec.nodeName}}{{end}}'` --template '{{range $i, $n := .status.addresses}}{{if eq $n.type "ExternalIP"}}{{$n.address}}{{end}}{{end}}')
$ curl $LBIP/foo -H 'Host: foo.bar.com'
```

## TLS

You can secure an Ingress by specifying a secret that contains a TLS private key and certificate. Currently the Ingress only supports a single TLS port, 443, and assumes TLS termination. This controller supports SNI. The TLS secret must contain keys named tls.crt and tls.key that contain the certificate and private key to use for TLS, eg:

```
apiVersion: v1
data:
  tls.crt: base64 encoded cert
  tls.key: base64 encoded key
kind: Secret
metadata:
  name: testsecret
  namespace: default
type: Opaque
```

Referencing this secret in an Ingress will tell the Ingress controller to secure the channel from the client to the loadbalancer using TLS:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: no-rules-map
spec:
  tls:
    secretName: testsecret
  backend:
    serviceName: s1
    servicePort: 80
```
Please follow [test.sh](https://github.com/bprashanth/Ingress/blob/master/examples/sni/nginx/test.sh) as a guide on how to generate secrets containing SSL certificates. The name of the secret can be different than the name of the certificate.

Check the [example](examples/tls/README.md)

### HTTP Strict Transport Security

HTTP Strict Transport Security (HSTS) is an opt-in security enhancement specified through the use of a special response header. Once a supported browser receives this header that browser will prevent any communications from being sent over HTTP to the specified domain and will instead send all communications over HTTPS.

By default the controller redirects (301) to HTTPS if there is a TLS Ingress rule. 

To disable this behavior use `hsts=false` in the NGINX ConfigMap.

#### Optimizing TLS Time To First Byte (TTTFB)

NGINX provides the configuration option [ssl_buffer_size](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) to allow the optimization of the TLS record size. This improves the [Time To First Byte](https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/) (TTTFB). The default value in the Ingress controller is `4k` (nginx default is `16k`);


## Proxy Protocol

If you are using a L4 proxy to forward the traffic to the NGINX pods and terminate HTTP/HTTPS there, you will lose the remote endpoint's IP addresses. To prevent this you could use the [Proxy Protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) for forwarding traffic, this will send the connection details before forwarding the acutal TCP connection itself.

Amongst others [ELBs in AWS](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) and [HAProxy](http://www.haproxy.org/) support Proxy Protocol.

Please check the [proxy-protocol](examples/proxy-protocol/) example

## Exposing TCP services

Ingress does not support TCP services (yet). For this reason this Ingress controller uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`
It is possible to use a number or the name of the port.

The next example shows how to expose the service `example-go` running in the namespace `default` in the port `8080` using the port `9000`
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-configmap-example
data:
  9000: "default/example-go:8080"
```


Please check the [tcp services](examples/tcp/README.md) example

## Exposing UDP services

Since 1.9.13 NGINX provides [UDP Load Balancing](https://www.nginx.com/blog/announcing-udp-load-balancing/).

Ingress does not support UDP services (yet). For this reason this Ingress controller uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`
It is possible to use a number or the name of the port.

The next example shows how to expose the service `kube-dns` running in the namespace `kube-system` in the port `53` using the port `53`
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: udp-configmap-example
data:
  53: "kube-system/kube-dns:53"
```


Please check the [udp services](examples/udp/README.md) example


## Custom NGINX configuration

Using a ConfigMap it is possible to customize the defaults in nginx.

Please check the [tcp services](examples/custom-configuration/README.md) example


## Custom NGINX template

The NGINX template is located in the file `/etc/nginx/template/nginx.tmpl`. Mounting a volume is possible to use a custom version.
Use the [custom-template](examples/custom-template/README.md) example as a guide

**Please note the template is tied to the go code. Be sure to no change names in the variable `$cfg`**


### Custom NGINX upstream checks

NGINX exposes some flags in the [upstream configuration](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) that enables the configuration of each server in the upstream. The ingress controller allows custom `max_fails` and `fail_timeout` parameters in a global context using `upstream-max-fails` or `upstream-fail-timeout` in the NGINX Configmap or in a particular Ingress rule. By default this values are 0. This means NGINX will respect the `readinessProbe`, if is defined. If there is no probe, NGINX will not mark a server inside an upstream down.

**With the default values NGINX will not health check your backends, and whenever the endpoints controller notices a readiness probe failure that pod's ip will be removed from the list of endpoints, causing nginx to also remove it from the upstreams.**

To use custom values in an Ingress rule define this annotations:

`ingress-nginx.kubernetes.io/upstream-max-fails`: number of unsuccessful attempts to communicate with the server that should happen in the duration set by the fail_timeout parameter to consider the server unavailable

`ingress-nginx.kubernetes.io/upstream-fail-timeout`: time in seconds during which the specified number of unsuccessful attempts to communicate with the server should happen to consider the server unavailable. Also the period of time the server will be considered unavailable.

**Important:** 
The upstreams are shared. i.e. Ingress rule using the same service will use the same upstream. 
This means only one of the rules should define annotations to configure the upstream servers


Please check the [auth](examples/custom-upstream-check/README.md) example

 ### Authentication
 
 Is possible to add authentication adding additional annotations in the Ingress rule. The source of the authentication is a secret that contains usernames and passwords inside the the key `auth`
 
 The annotations are:
 
 ```
 ingress-nginx.kubernetes.io/auth-type:[basic|digest]
 ```
 
 Indicates the [HTTP Authentication Type: Basic or Digest Access Authentication](https://tools.ietf.org/html/rfc2617).
 
 ```
 ingress-nginx.kubernetes.io/auth-secret:secretName
 ```
 
 Name of the secret that contains the usernames and passwords with access to the `path/s` defined in the Ingress Rule.
 The secret must be created in the same namespace than the Ingress rule
 
 ```
 ingress-nginx.kubernetes.io/auth-realm:"realm string"
 ```


### NGINX status page

The ngx_http_stub_status_module module provides access to basic status information. This is the default module active in the url `/nginx_status`.
This controller provides an alternitive to this module using [nginx-module-vts](https://github.com/vozlt/nginx-module-vts) third party module.
To use this module just provide a ConfigMap with the key `enable-vts-status=true`. The URL is exposed in the port 8080.
Please check the example `example/rc-default.yaml`

![nginx-module-vts screenshot](https://cloud.githubusercontent.com/assets/3648408/10876811/77a67b70-8183-11e5-9924-6a6d0c5dc73a.png "screenshot with filter")

To extract the information in JSON format the module provides a custom URL: `/nginx_status/format/json`


### Custom errors

In case of an error in a request the body of the response is obtained from the `default backend`. Each request to the default backend includes two headers: 
- `X-Code` indicates the HTTP code
- `X-Format` the value of the `Accept` header

Using this two headers is possible to use a custom backend service like [this one](https://github.com/aledbf/contrib/tree/nginx-debug-server/Ingress/images/nginx-error-server) that inspect each request and returns a custom error page with the format expected by the client. Please check the example [custom-errors](examples/custom-errors/README.md)


### Annotations

|Annotation                 |Values|Description|
|---------------------------|------|-----------|
|ingress.kubernetes.io/rewrite-target|URI|   |
|ingress.kubernetes.io/add-base-url|true\|false| |
|ingress.kubernetes.io/limit-connections| ||
|ingress.kubernetes.io/limit-rps|||
|ingress.kubernetes.io/auth-type|basic\|digest|Indicates the [HTTP Authentication Type: Basic or Digest Access Authentication](https://tools.ietf.org/html/rfc2617)||
|ingress.kubernetes.io/auth-secret|string|Name of the secret that contains the usernames and passwords.
| | |The secret must be created in the same namespace than the Ingress rule||
|ingress.kubernetes.io/auth-realm|string| |


### Custom configuration options

Running `/nginx-ingress-controller --dump-nginx-configuration` is possible to get the value of the options that can be changed.
The next table shows the options, the default value and a description

|name                 |default|
|---------------------------|------|
|body-size|1m|
|custom-http-errors|" "|
|enable-sticky-sessions|"false"|
|enable-vts-status|"false"|
|error-log-level|notice|
|gzip-types||
|hsts|"true"|
|hsts-include-subdomains|"true"|
|hsts-max-age|"15724800"|
|keep-alive|"75"|
|max-worker-connections|"16384"|
|proxy-connect-timeout|"5"|
|proxy-read-timeout|"60"|
|proxy-real-ip-cidr|0.0.0.0/0|
|proxy-send-timeout|"60"|
|retry-non-idempotent|"false"|
|server-name-hash-bucket-size|"64"|
|server-name-hash-max-size|"512"|
|ssl-buffer-size|4k|
|ssl-ciphers||
|ssl-protocols|TLSv1 TLSv1.1 TLSv1.2|
|ssl-session-cache|"true"|
|ssl-session-cache-size|10m|
|ssl-session-tickets|"true"|
|ssl-session-timeout|10m|
|use-gzip|"true"|
|use-http2|"true"|
|vts-status-zone-size|10m|
|worker-processes|<number of CPUs>|


**Description:**

*body-size:*

Sets the maximum allowed size of the client request body. See NGINX [client_max_body_size](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size)


*custom-http-errors:*

Enables which HTTP codes should be passed for processing with the [error_page directive](http://nginx.org/en/docs/http/ngx_http_core_module.html#error_page)
Setting at least one code this also enables [proxy_intercept_errors](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_intercept_errors) (required to process error_page)


*enable-sticky-sessions:* 

Enables sticky sessions using cookies. This is provided by [nginx-sticky-module-ng](https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng) module


*enable-vts-status:*

Allows the replacement of the default status page with a third party module named [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)


*error-log-level:*

Configures the logging level of errors. Log levels above are listed in the order of increasing severity
http://nginx.org/en/docs/ngx_core_module.html#error_log


*retry-non-idempotent:*

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error in the upstream server. 
The previous behavior can be restored using the value "true"


*hsts:*

Enables or disables the header HSTS in servers running SSL.
HTTP Strict Transport Security (often abbreviated as HSTS) is a security feature (HTTP header) that tell browsers that it should only be communicated with using HTTPS, instead of using HTTP.
https://developer.mozilla.org/en-US/docs/Web/Security/HTTP_strict_transport_security
Why HSTS is important?


*hsts-include-subdomains:*

Enables or disables the use of HSTS in all the subdomains of the servername


*hsts-max-age:*

Sets the time, in seconds, that the browser should remember that this site is only to be accessed using HTTPS.


*keep-alive:*

Sets the time during which a keep-alive client connection will stay open on the server side.
The zero value disables keep-alive client connections
http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout


*max-worker-connections:*

Sets the maximum number of simultaneous connections that can be opened by each [worker process](http://nginx.org/en/docs/ngx_core_module.html#worker_connections)


*proxy-connect-timeout*:
  
Sets the timeout for [establishing a connection with a proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_connect_timeout). It should be noted that this timeout cannot usually exceed 75 seconds.


*proxy-read-timeout:*

Sets the timeout in seconds for [reading a response from the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_read_timeout). The timeout is set only between two successive read operations, not for the transmission of the whole response 


*proxy-send-timeout:*

Sets the timeout in seconds for [transmitting a request to the proxied server](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_send_timeout). The timeout is set only between two successive write operations, not for the transmission of the whole request.


*resolver:*

Configures name servers used to [resolve](http://nginx.org/en/docs/http/ngx_http_core_module.html#resolver) names of upstream servers into addresses


*server-name-hash-max-size:*

Sets the maximum size of the [server names hash tables](http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size) used in server names, map directive’s values, MIME types, names of request header strings, etc.
http://nginx.org/en/docs/hash.html


*server-name-hash-bucket-size:*

Sets the size of the bucker for the server names hash tables
http://nginx.org/en/docs/hash.html
http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size

*ssl-buffer-size:*

Sets the size of the [SSL buffer](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) used for sending data.
4k helps NGINX to improve TLS Time To First Byte (TTTFB)
https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/

*ssl-ciphers:*

Sets the [ciphers](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers) list to enable. The ciphers are specified in the format understood by the OpenSSL library


*ssl-dh-param:*

Base64 string that contains Diffie-Hellman key to help with "Perfect Forward Secrecy"
https://www.openssl.org/docs/manmaster/apps/dhparam.html
https://wiki.mozilla.org/Security/Server_Side_TLS#DHE_handshake_and_dhparam
http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_dhparam

*ssl-protocols:*

Sets the [SSL protocols](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols) to use


*ssl-session-cache:*

Enables or disables the use of shared [SSL cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) among worker processes.


*ssl-session-cache-size:*

Sets the size of the [SSL shared session cache](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache) between all worker processes.


*ssl-session-tickets:*

Enables or disables session resumption through [TLS session tickets](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets)


*ssl-session-timeout:*

Sets the time during which a client may [reuse the session](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout) parameters stored in a cache.


*upstream-max-fails:*

Sets the number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) that should happen in the duration set by the fail_timeout parameter to consider the server unavailable


*upstream-fail-timeout:*

Sets the time during which the specified number of unsuccessful attempts to communicate with the [server](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) should happen to consider the server unavailable


*use-proxy-protocol:*

Enables or disables the use of the [PROXY protocol](https://www.nginx.com/resources/admin-guide/proxy-protocol/) to receive client connection (real IP address) information passed through proxy servers and load balancers such as HAproxy and Amazon Elastic Load Balancer (ELB).


*use-gzip:*

Enables or disables the use of the nginx module that compresses responses using the ["gzip" module](http://nginx.org/en/docs/http/ngx_http_gzip_module.html)


*use-http2:*

Enables or disables the [HTTP/2](http://nginx.org/en/docs/http/ngx_http_v2_module.html) support in secure connections 


*gzip-types:*

MIME types in addition to "text/html" to compress. The special value "*"" matches any MIME type.
Responses with the "text/html" type are always compressed if `use-gzip` is enabled


*worker-processes:*

Sets the number of [worker processes](http://nginx.org/en/docs/ngx_core_module.html#worker_processes). By default "auto" means number of available CPU cores



## Troubleshooting

Problems encountered during [1.2.0-alpha7 deployment](https://github.com/kubernetes/kubernetes/blob/master/docs/getting-started-guides/docker.md):
* make setup-files.sh file in hypercube does not provide 10.0.0.1 IP to make-ca-certs, resulting in CA certs that are issued to the external cluster IP address rather then 10.0.0.1 -> this results in nginx-third-party-lb appearing to get stuck at "Utils.go:177 - Waiting for default/default-http-backend" in the docker logs.  Kubernetes will eventually kill the container before nginx-third-party-lb times out with a message indicating that the CA certificate issuer is invalid (wrong ip), to verify this add zeros to the end of initialDelaySeconds and timeoutSeconds and reload the RC, and docker will log this error before kubernetes kills the container.
  * To fix the above, setup-files.sh must be patched before the cluster is inited (refer to https://github.com/kubernetes/kubernetes/pull/21504)

### Debug

Using the flag `--v=XX` it is possible to increase the level of logging.
In particular:
- `--v=2` shows details using `diff` about the changes in the configuration in nginx

```
I0316 12:24:37.581267       1 utils.go:148] NGINX configuration diff a//etc/nginx/nginx.conf b//etc/nginx/nginx.conf
I0316 12:24:37.581356       1 utils.go:149] --- /tmp/922554809  2016-03-16 12:24:37.000000000 +0000
+++ /tmp/079811012  2016-03-16 12:24:37.000000000 +0000
@@ -235,7 +235,6 @@

     upstream default-echoheadersx {
         least_conn;
-        server 10.2.112.124:5000;
         server 10.2.208.50:5000;

     }
I0316 12:24:37.610073       1 command.go:69] change in configuration detected. Reloading...
```

- `--v=3` shows details about the service, Ingress rule, endpoint changes and it dumps the nginx configuration in JSON format
- `--v=5` configures NGINX in [debug mode](http://nginx.org/en/docs/debugging_log.html)



### Retries in no idempotent methods

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error.
The previous behavior can be restored using `retry-non-idempotent=true` in the configuration ConfigMap


## Limitations

- Ingress rules for TLS require the definition of the field `host`


