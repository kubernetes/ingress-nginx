# Prerequisites

Many of the examples in this directory have common prerequisites.

## Deploying a controller

Unless you're running on a cloudprovider that supports Ingress out of the box
(eg: GCE/GKE), you will need to deploy a controller. You can do so following
[these instructions](/examples/deployment).

## Firewall rules

If you're using a bare-metal controller (eg the nginx ingress controller), you
will need to create a firewall rule that targets port 80/443 on the specific VMs
the nginx controller is running on. On cloudproviders, the respective backend
will auto-create firewall rules for your Ingress.

If you'd like to auto-create firewall rules for an OSS Ingress controller,
you can put it behind a Service of `Type=Loadbalancer` as shown in
[this example](/examples/static-ip/nginx#acquiring-an-ip).

## TLS certificates

Unless otherwise mentioned, the TLS secret used in examples is a 2048 bit RSA
key/cert pair with an arbitrarily chosen hostname, created as follows

```console
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=nginxsvc/O=nginxsvc"
Generating a 2048 bit RSA private key
......................................................................................................................................+++
....................................................................+++
writing new private key to 'tls.key'
-----

$ kubectl create secret tls tls-secret --key tls.key --cert tls.crt
secret "tls-secret" created
```

## Test HTTP Service

All examples that require a test HTTP Service use the standard echoheaders pod,
which you can deploy as follows

```console
$ kubectl create -f http-svc.yaml
service "http-svc" created
replicationcontroller "http-svc" created

$ kubectl get po
NAME                READY     STATUS    RESTARTS   AGE
echoheaders-p1t3t   1/1       Running   0          1d

$ kubectl get svc
NAME          CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
echoheaders   10.0.122.116   <none>        80/TCP                       1d
```

You can test that the HTTP Service works by exposing it temporarily
```console
$ kubectl patch svc echoheaders -p '{"spec":{"type": "LoadBalancer"}}'
"echoheaders" patched

$ kubectl get svc echoheaders
NAME          CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
echoheaders   10.0.122.116   <pending>     80:32100/TCP   1d

$ kubectl describe svc echoheaders
Name:			echoheaders
Namespace:		default
Labels:			app=echoheaders
Selector:		app=echoheaders
Type:			LoadBalancer
IP:			10.0.122.116
LoadBalancer Ingress:	108.59.87.136
Port:			http	80/TCP
NodePort:		http	32100/TCP
Endpoints:		10.180.1.6:8080
Session Affinity:	None
Events:
  FirstSeen	LastSeen	Count	From			SubObjectPath	Type		Reason			Message
  ---------	--------	-----	----			-------------	--------	------			-------
  1m		1m		1	{service-controller }			Normal		Type			ClusterIP -> LoadBalancer
  1m		1m		1	{service-controller }			Normal		CreatingLoadBalancer	Creating load balancer
  16s		16s		1	{service-controller }			Normal		CreatedLoadBalancer	Created load balancer

$ curl 108.59.87.126
CLIENT VALUES:
client_address=10.240.0.3
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://108.59.87.136:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
host=108.59.87.136
user-agent=curl/7.46.0
BODY:
-no body in request-

$ kubectl patch svc echoheaders -p '{"spec":{"type": "NodePort"}}'
"echoheaders" patched
```

## Ingress Class

If you have multiple Ingress controllers in a single cluster, you can pick one
by specifying the `ingress.class` annotation, eg creating an Ingress with an
annotation like

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "gce"
```

will target the GCE controller, forcing the nginx controller to ignore it, while
an annotation like

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "nginx"
```

will target the nginx controller, forcing the GCE controller to ignore it.

__Note__: Deploying multiple ingress controller and not specifying the
annotation will result in both controllers fighting to satisfy the Ingress.
