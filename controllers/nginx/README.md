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
kubectl run echoheaders --image=gcr.io/google_containers/echoserver:1.3 --replicas=1 --port=8080
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

Using this two headers is possible to use a custom backend service like [this one](https://github.com/aledbf/contrib/tree/nginx-debug-server/Ingress/images/nginx-error-server) that inspect each request and returns a custom error page with the format expected by the client. This images handles `html` and `json` responses.


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
- The IP address in the status of loadBalancer could contain old values

