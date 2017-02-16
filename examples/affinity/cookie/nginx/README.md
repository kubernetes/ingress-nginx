# Sticky Session

This example demonstrates how to achieve session affinity using cookies 

## Prerequisites

You will need to make sure you Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](/examples/PREREQUISITES.md#ingress-class),
and that you have an ingress controller [running](/examples/deployment) in your cluster.

You will also need to deploy multiple replicas of your application that show up as endpoints for the Service referenced in the Ingress object, to test session stickyness.
Using a deployment with only one replica doesn't set the 'sticky' cookie.

## Deployment

Session stickyness is achieved through 3 annotations on the Ingress, as shown in the [example](sticky-ingress.yaml).

|Name|Description|Values|
| --- | --- | --- |
|ingress.kubernetes.io/affinity|Sets the affinity type|string (in NGINX only ``cookie`` is possible|
|ingress.kubernetes.io/session-cookie-name|Name of the cookie that will be used|string (default to route)|
|ingress.kubernetes.io/session-cookie-hash|Type of hash that will be used in cookie value|sha1/md5/index|

You can create the ingress to test this

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
  affinity:	cookie
  session-cookie-hash:		sha1
  session-cookie-name:		route
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
In the example above, you can see a line containing the 'Set-Cookie: route' setting the right defined stickness cookie.
This cookie is created by NGINX containing the hash of the used upstream in that request. 
If the user changes this cookie, NGINX creates a new one and redirect the user to another upstream.

If the backend pool grows up NGINX will keep sending the requests through the same server of the first request, even if it's overloaded.

When the backend server is removed, the requests are then re-routed to another upstream server and NGINX creates a new cookie, as the previous hash became invalid.

When you have more than one Ingress Object pointing to the same Service, but one containing affinity configuration and other don't, the first created Ingress will be used. 
This means that you can face the situation that you've configured Session Affinity in one Ingress and it doesn't reflects in NGINX configuration, because there is another Ingress Object pointing to the same service that doesn't configure this.

