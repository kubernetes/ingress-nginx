# Nginx Ingress Controller

This is a nginx Ingress controller that uses [ConfigMap](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/configmap.md) to store the nginx configuration. See [Ingress controller documentation](../README.md) for details on how it works.


## What it provides?

- Ingress controller
- nginx 1.9.x with [lua-nginx-module](https://github.com/openresty/lua-nginx-module)
- SSL support
- custom ssl_dhparam (optional). Just mount a secret with a file named `dhparam.pem`.
- support for TCP services (flag `--tcp-services`)
- custom nginx configuration using [ConfigMap](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/configmap.md)
- custom error pages. Using the flag `--custom-error-service` is possible to use a custom compatible [404-server](https://github.com/kubernetes/contrib/tree/master/404-server) image [nginx-error-server](https://github.com/aledbf/contrib/tree/nginx-debug-server/Ingress/images/nginx-error-server) that provides an additional `/errors` route that returns custom content for a particular error code. **This is completely optional**


## Requirements
- default backend [404-server](https://github.com/kubernetes/contrib/tree/master/404-server) (or a custom compatible image)
- DNS must be operational and able to resolve default-http-backend.default.svc.cluster.local

## SSL

Please follow [test.sh](https://github.com/bprashanth/Ingress/blob/master/examples/sni/nginx/test.sh) as a guide on how to generate secrets containing SSL certificates. The name of the secret can be different than the name of the certificate.

Currently Ingress does not support HTTPS. To bypass this the controller will check if there's a certificate for the the host in `Spec.Rules.Host` checking for a certificate in each of the mounted secrets. If exists it will create a nginx server listening in the port 443.

## Examples:

First we need to deploy some application to publish. To keep this simple we will use the [echoheaders app]() that just returns information about the http request as output
```
kubectl run echoheaders --image=gcr.io/google_containers/echoserver:1.0 --replicas=1 --port=8080
```

Now we expose the same application in two different services (so we can create different Ingress rules)
```
kubectl expose rc echoheaders --port=80 --target-port=8080 --name=echoheaders-x 
kubectl expose rc echoheaders --port=80 --target-port=8080 --name=echoheaders-y
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

Before the deploy of nginx we need a default backend [404-server](https://github.com/kubernetes/contrib/tree/master/404-server) (or a compatible custom image)
```
kubectl create -f examples/default-backend.yaml
kubectl expose rc default-http-backend --port=80 --target-port=8080 --name=default-http-backend
```

# Default configuration

The last step is the deploy of nginx Ingress rc (from the examples directory)
```
kubectl create -f examples/rc-default.yaml
```

To test if evertyhing is working correctly:

`curl -v http://<node IP address>:80/foo -H 'Host: foo.bar.com'`

You should see an output similar to
```
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET /foo HTTP/1.1
> Host: foo.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.9.8
< Date: Tue, 15 Dec 2015 13:45:13 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
< Vary: Accept-Encoding
<
CLIENT VALUES:
client_address=10.2.84.43
command=GET
real path=/foo
query=nil
request_version=1.1
request_uri=http://foo.bar.com:8080/foo

SERVER VALUES:
server_version=nginx: 1.9.7 - lua: 9019

HEADERS RECEIVED:
accept=*/*
connection=close
host=foo.bar.com
user-agent=curl/7.43.0
x-forwarded-for=172.17.4.1
x-forwarded-host=foo.bar.com
x-forwarded-server=foo.bar.com
x-real-ip=172.17.4.1
BODY:
* Connection #0 to host 172.17.4.99 left intact
```

If we try to get a non exising route like `/foobar` we should see
```
$ curl -v 172.17.4.99/foobar -H 'Host: foo.bar.com'
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET /foobar HTTP/1.1
> Host: foo.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.9.8
< Date: Tue, 15 Dec 2015 13:48:18 GMT
< Content-Type: text/html
< Transfer-Encoding: chunked
< Connection: keep-alive
< Vary: Accept-Encoding
<
default backend - 404
* Connection #0 to host 172.17.4.99 left intact
```

(this test checked that the default backend is properly working)

*Replacing the default backend with a custom one we can change the default error pages provided by nginx*

# Exposing TCP services

First we need to remove the running
```
kubectl delete rc nginx-ingress-3rdpartycfg
```

```
kubectl create -f examples/rc-tcp.yaml
```

Now we add the annotation to the replication controller that indicates with services should be exposed as TCP:
The annotation key is `nginx-ingress.kubernetes.io/tcpservices`. You can expose more than one service using comma as separator.
Each service must contain the namespace, service name and port to be use as public port

```
kubectl annotate rc nginx-ingress-3rdpartycfg "nginx-ingress.kubernetes.io/tcpservices=default/echoheaders-x:9000"
```

*Note:* the only reason to remove and create a new rc is that we cannot open new ports dynamically once the pod is running.


Once we run the `kubectl annotate` command nginx will reload.

Now we can test the new service:
```
$ (sleep 1; echo "GET / HTTP/1.1"; echo "Host: 172.17.4.99:9000"; echo;echo;sleep 2) | telnet 172.17.4.99 9000

Trying 172.17.4.99...
Connected to 172.17.4.99.
Escape character is '^]'.
HTTP/1.1 200 OK
Server: nginx/1.9.7
Date: Tue, 15 Dec 2015 14:46:28 GMT
Content-Type: text/plain
Transfer-Encoding: chunked
Connection: keep-alive

f
CLIENT VALUES:

1a
client_address=10.2.84.45

c
command=GET

c
real path=/

a
query=nil

14
request_version=1.1

25
request_uri=http://172.17.4.99:8080/

1


f
SERVER VALUES:

28
server_version=nginx: 1.9.7 - lua: 9019

1


12
HEADERS RECEIVED:

16
host=172.17.4.99:9000

6
BODY:

14
-no body in request-
0
```

## SSL

Currently Ingress rules does not contains SSL definitions. In order to support SSL in nginx this controller uses secrets mounted inside the directory `/etc/nginx-ssl` to detect if some Ingress rule contains a host for which it is possible the creation of an SSL server.

First create a secret containing the ssl certificate and key. This example creates the certificate and the secret (json):

`SECRET_NAME=secret-echoheaders-1 HOSTS=foo.bar.com ./examples/certs.sh`

Create the secret:
```
kubectl create -f secret-secret-echoheaders-1-foo.bar.com.json
```

Check if the secret was created:
```
$ kubectl get secrets
NAME                   TYPE                                  DATA      AGE
secret-echoheaders-1   Opaque                                2         9m
```


Like before we need to remove the running nginx rc
```
kubectl delete rc nginx-ingress-3rdpartycfg
```

Next create a new rc that uses the secret
```
kubectl create -f examples/rc-ssl.yaml
```

*Note:* this example uses a self signed certificate.

Example output:
```
$ curl -v https://172.17.4.99/foo -H 'Host: bar.baz.com' -k
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 4444 (#0)
* TLS 1.2 connection using TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
* Server certificate: foo.bar.com
> GET /foo HTTP/1.1
> Host: bar.baz.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.9.8
< Date: Thu, 17 Dec 2015 14:57:03 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
< Vary: Accept-Encoding
<
CLIENT VALUES:
client_address=10.2.84.34
command=GET
real path=/foo
query=nil
request_version=1.1
request_uri=http://bar.baz.com:8080/foo

SERVER VALUES:
server_version=nginx: 1.9.7 - lua: 9019

HEADERS RECEIVED:
accept=*/*
connection=close
host=bar.baz.com
user-agent=curl/7.43.0
x-forwarded-for=172.17.4.1
x-forwarded-host=bar.baz.com
x-forwarded-server=bar.baz.com
x-real-ip=172.17.4.1
BODY:
* Connection #0 to host 172.17.4.99 left intact
-no body in request-
```


## Custom errors

The default backend provides a way to customize the default 404 page. This helps but sometimes is not enough.
Using the flag `--custom-error-service` is possible to use an image that must be 404 compatible and provide the route /error 
[Here](https://github.com/aledbf/contrib/tree/nginx-debug-server/Ingress/images/nginx-error-server) there is an example of the the image

The route `/error` expects two arguments: code and format
* code defines the wich error code is expected to be returned (502,503,etc.)
* format the format that should be returned For instance /error?code=504&format=json or /error?code=502&format=html

Using a volume pointing to `/var/www/html` directory is possible to use a custom error

## Troubleshooting

Problems encountered during [1.2.0-alpha7 deployment](https://github.com/kubernetes/kubernetes/blob/master/docs/getting-started-guides/docker.md):
* make setup-files.sh file in hypercube does not provide 10.0.0.1 IP to make-ca-certs, resulting in CA certs that are issued to the external cluster IP address rather then 10.0.0.1 -> this results in nginx-third-party-lb appearing to get stuck at "Utils.go:177 - Waiting for default/default-http-backend" in the docker logs.  Kubernetes will eventually kill the container before nginx-third-party-lb times out with a message indicating that the CA certificate issuer is invalid (wrong ip), to verify this add zeros to the end of initialDelaySeconds and timeoutSeconds and reload the RC, and docker will log this error before kubernetes kills the container.
  * To fix the above, setup-files.sh must be patched before the cluster is inited (refer to https://github.com/kubernetes/kubernetes/pull/21504)
* if once the nginx-third-party-lb starts, its docker log spams this message continously "utils.go:(line #)] Requeuing default/echomap, err Post http://127.0.0.1:8080/update-ingress: dial tcp 127.0.0.1:8080: getsockopt: connection refused", it means that the container is unable to use DNS to resolve the service address, DNS autoconfigure is broken on 1.2.0-alpha7 (refer again to https://github.com/kubernetes/kubernetes/pull/21504 for fixes)

## TODO:
- multiple SSL certificates
- custom nginx configuration using [ConfigMap](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/configmap.md)

