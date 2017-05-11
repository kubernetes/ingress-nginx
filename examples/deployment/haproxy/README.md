# Deploying HAProxy Ingress Controller

If you don't have a Kubernetes cluster, please refer to [setup](/docs/dev/setup.md)
for instructions on how to create a new one.

## Prerequisites

This ingress controller doesn't yet have support for
[ingress classes](/examples/PREREQUISITES.md#ingress-class). You MUST turn
down any existing ingress controllers before running HAProxy Ingress controller or
they will fight for Ingresses. This includes any cloudprovider controller.

This document has also the following prerequisites:

* Create a [TLS secret](/examples/PREREQUISITES.md#tls-certificates) named `tls-secret` to be used as default TLS certificate
* Optional: deploy a [web app](/examples/PREREQUISITES.md#test-http-service) for testing

Creating the TLS secret:

```console
$ openssl req \
  -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout tls.key -out tls.crt -subj '/CN=localhost'
$ kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key
$ rm -v tls.crt tls.key
```

The optional web app can be created as follow:

```console
$ kubectl run http-svc \
  --image=gcr.io/google_containers/echoserver:1.3 \
  --port=8080 \
  --replicas=1 \
  --expose
```

## Default backend

Deploy a default backend used to serve `404 Not Found` pages:

```console
$ kubectl run ingress-default-backend \
  --image=gcr.io/google_containers/defaultbackend:1.0 \
  --port=8080 \
  --limits=cpu=10m,memory=20Mi \
  --expose
```

Check if the default backend is up and running:

```console
$ kubectl get pod
NAME                                       READY     STATUS    RESTARTS   AGE
ingress-default-backend-1110790216-gqr61   1/1       Running   0          10s
```

## Configmap

Create a configmap named `haproxy-ingress`:

```console
$ kubectl create configmap haproxy-ingress
configmap "haproxy-ingress" created
```

A configmap is used to provide global or default configuration like
timeouts, SSL/TLS settings, a syslog service endpoint and so on. The
configmap can be edited or replaced later in order to apply new
configuration on a running ingress controller. All supported options
are [here](https://github.com/jcmoraisjr/haproxy-ingress#configmap).

## Controller

Deploy HAProxy Ingress:

```console
$ kubectl create -f haproxy-ingress.yaml
```

Check if the controller was successfully deployed:

```console
$ kubectl get pod -w
NAME                                       READY     STATUS    RESTARTS   AGE
haproxy-ingress-2556761959-tv20k           1/1       Running   0          12s
ingress-default-backend-1110790216-gqr61   1/1       Running   0          3m
^C
```

## Testing

From now the optional web app should be deployed. Deploy an ingress resource to expose this app:

```console
$ kubectl create -f - <<EOF
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: app
spec:
  rules:
  - host: foo.bar
    http:
      paths:
      - path: /
        backend:
          serviceName: http-svc
          servicePort: 8080
EOF
```

Expose the Ingress controller as a `type=NodePort` service:

```console
$ kubectl expose deploy/haproxy-ingress --type=NodePort
$ kubectl get svc/haproxy-ingress -oyaml
```

Look for `nodePort` field next to `port: 80`.

Change below `172.17.4.99` to the host's IP and `30876` to the `nodePort`:

```console
$ curl -i 172.17.4.99:30876
HTTP/1.1 404 Not Found
Date: Mon, 05 Feb 2017 22:59:36 GMT
Content-Length: 21
Content-Type: text/plain; charset=utf-8

default backend - 404
```

Using default backend because host was not found.

Now try to send a header:

```console
$ curl -i 172.17.4.99:30876 -H 'Host: foo.bar'
HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Mon, 05 Feb 2017 23:00:33 GMT
Content-Type: text/plain
Transfer-Encoding: chunked

CLIENT VALUES:
client_address=10.2.18.5
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://foo.bar:8080/
...
```

## Troubleshooting

If you have any problem, check logs and events of HAProxy Ingress POD:

```console
$ kubectl get pod
NAME                                       READY     STATUS    RESTARTS   AGE
haproxy-ingress-2556761959-tv20k           1/1       Running   0          9m
...

$ kubectl logs haproxy-ingress-2556761959-tv20k
$ kubectl describe pod/haproxy-ingress-2556761959-tv20k
```
