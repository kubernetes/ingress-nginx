# TLS termination

This example demonstrates how to terminate TLS through the GCE Ingress controller.

## Prerequisites

You need a [TLS cert](/examples/PREREQUISITES.md#tls-certificates) and a [test HTTP service](/examples/PREREQUISITES.md#test-http-service) for this example.
You will also need to make sure you Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](/examples/PREREQUISITES.md#ingress-class),
and that you have an ingress controller [running](/examples/deployment) in your cluster.

## Deployment

The following command instructs the controller to terminate traffic using
the provided TLS cert, and forward un-encrypted HTTP traffic to the test
HTTP service.

```console
$ kubectl create -f gce-tls-ingress.yaml
```

## Validation

You can confirm that the Ingress works.

```console
$ kubectl describe ing gce-test
Name:			gce-test
Namespace:		default
Address:		35.186.221.137
Default backend:	http-svc:80 (10.180.1.9:8080,10.180.3.6:8080)
TLS:
  tls-secret terminates
Rules:
  Host	Path	Backends
  ----	----	--------
  *	* 	http-svc:80 (10.180.1.9:8080,10.180.3.6:8080)
Annotations:
  target-proxy:			k8s-tp-default-gce-test--32658fa96c080068
  url-map:			k8s-um-default-gce-test--32658fa96c080068
  backends:			{"k8s-be-30301--32658fa96c080068":"Unknown"}
  forwarding-rule:		k8s-fw-default-gce-test--32658fa96c080068
  https-forwarding-rule:	k8s-fws-default-gce-test--32658fa96c080068
  https-target-proxy:		k8s-tps-default-gce-test--32658fa96c080068
  static-ip:			k8s-fw-default-gce-test--32658fa96c080068
Events:
  FirstSeen	LastSeen	Count	From				SubObjectPath	Type		Reason	Message
  ---------	--------	-----	----				-------------	--------	------	-------
  2m		2m		1	{loadbalancer-controller }			Normal		ADD	default/gce-test
  1m		1m		1	{loadbalancer-controller }			Normal		CREATE	ip: 35.186.221.137
  1m		1m		3	{loadbalancer-controller }			Normal		Service	default backend set to http-svc:30301

$ curl 35.186.221.137 -k
curl 35.186.221.137 -L
curl: (60) SSL certificate problem: self signed certificate
More details here: http://curl.haxx.se/docs/sslcerts.html

$ curl 35.186.221.137 -kl
CLIENT VALUES:
client_address=10.240.0.3
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://35.186.221.137:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=Keep-Alive
host=35.186.221.137
user-agent=curl/7.46.0
via=1.1 google
x-cloud-trace-context=bfa123130fd623989cca0192e43d9ba4/8610689379063045825
x-forwarded-for=104.132.0.80, 35.186.221.137
x-forwarded-proto=https
```
