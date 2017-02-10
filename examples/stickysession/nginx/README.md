# Sticky Session

This example demonstrates how to Stickness in a Ingress.

## Prerequisites

You will need to make sure you Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](/examples/PREREQUISITES.md#ingress-class),
and that you have an ingress controller [running](/examples/deployment) in your cluster.

Also, you need to have a deployment with replica > 1. Using a deployment with only one replica doesn't set the 'sticky' cookie.

## Deployment

The following command instructs the controller to set Stickness in all Upstreams of an Ingress

```console
$ kubectl create -f sticky-ingress.yaml
```

## Validation

You can confirm that the Ingress works.

```console
$ kubectl describe ing nginx-test
Name:			nginx-test
Namespace:		default
Address:		
Default backend:	default-http-backend:80 (10.180.0.4:8080,10.240.0.2:8080)
Rules:
  Host	                        Path	Backends
  ----	                        ----	--------
  stickyingress.example.com     
                                /   	 nginx-service:80 (<none>)
Annotations:
  sticky-enabled:	true
  sticky-hash:		sha1
  sticky-name:		route
Events:
  FirstSeen	LastSeen	Count	From				SubObjectPath	Type		Reason	Message
  ---------	--------	-----	----				-------------	--------	------	-------
  7s		7s		1	{nginx-ingress-controller }			Normal		CREATE	default/nginx-test
  

$ curl -I http://stickyingress.example.com
HTTP/1.1 200 OK
Server: nginx/1.11.9
Date: Fri, 10 Feb 2017 14:11:12 GMT
Content-Type: text/html
Content-Length: 612
Connection: keep-alive
Set-Cookie: route=a9907b79b248140b56bb13723f72b67697baac3d; Path=/; HttpOnly
Last-Modified: Tue, 24 Jan 2017 14:02:19 GMT
ETag: "58875e6b-264"
Accept-Ranges: bytes
```
In the example avove, you can see a line containing the 'Set-Cookie: route' setting the right defined stickness cookie.