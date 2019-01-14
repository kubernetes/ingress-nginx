# Sticky Session

This example demonstrates how to achieve session affinity using cookies

## Deployment

Session stickiness is achieved through 3 annotations on the Ingress, as shown in the [example](ingress.yaml).

|Name|Description|Values|
| --- | --- | --- |
|nginx.ingress.kubernetes.io/affinity|Sets the affinity type|string (in NGINX only ``cookie`` is possible|
|nginx.ingress.kubernetes.io/session-cookie-name|Name of the cookie that will be used|string (default to INGRESSCOOKIE)|
|nginx.ingress.kubernetes.io/session-cookie-hash|Type of hash that will be used in cookie value|sha1/md5/index|
|nginx.ingress.kubernetes.io/session-cookie-expires|The value is a date as UNIX timestamp that the cookie will expire on, it corresponds to cookie Expires directive|number of seconds|
|nginx.ingress.kubernetes.io/session-cookie-max-age|Number of seconds until the cookie expires that will correspond to cookie `Max-Age` directive|number of seconds|

You can create the ingress to test this

```console
kubectl create -f ingress.yaml
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
  session-cookie-name:		INGRESSCOOKIE
  session-cookie-expires: 172800
  session-cookie-max-age: 172800
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
Set-Cookie: INGRESSCOOKIE=a9907b79b248140b56bb13723f72b67697baac3d; Expires=Sun, 12-Feb-17 14:11:12 GMT; Max-Age=172800; Path=/; HttpOnly
Last-Modified: Tue, 24 Jan 2017 14:02:19 GMT
ETag: "58875e6b-264"
Accept-Ranges: bytes
```
In the example above, you can see a line containing the 'Set-Cookie: INGRESSCOOKIE' setting the right defined stickiness cookie.
This cookie is created by NGINX, it contains the hash of the used upstream in that request and has an expires. 
If the user changes this cookie, NGINX creates a new one and redirect the user to another upstream.

If the backend pool grows up NGINX will keep sending the requests through the same server of the first request, even if it's overloaded.

When the backend server is removed, the requests are then re-routed to another upstream server and NGINX creates a new cookie, as the previous hash became invalid.

When you have more than one Ingress Object pointing to the same Service, but one containing affinity configuration and other don't, the first created Ingress will be used. 
This means that you can face the situation that you've configured Session Affinity in one Ingress and it doesn't reflects in NGINX configuration, because there is another Ingress Object pointing to the same service that doesn't configure this.

