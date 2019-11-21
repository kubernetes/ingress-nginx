# External authentication, authentication service response headers propagation

This example demonstrates propagation of selected authentication service response headers
to backend service.

Sample configuration includes:

* Sample authentication service producing several response headers
  * Authentication logic is based on HTTP header: requests with header `User` containing string `internal` are considered authenticated
  * After successful authentication service generates response headers `UserID` and `UserRole`
* Sample echo service displaying header information
* Two ingress objects pointing to echo service
  * Public, which allows access from unauthenticated users
  * Private, which allows access from authenticated users only

You can deploy the controller as
follows:

```console
$ kubectl create -f deploy/
deployment "demo-auth-service" created
service "demo-auth-service" created
ingress "demo-auth-service" created
deployment "demo-echo-service" created
service "demo-echo-service" created
ingress "public-demo-echo-service" created
ingress "secure-demo-echo-service" created

$ kubectl get po
NAME                                        READY     STATUS    RESTARTS   AGE
demo-auth-service-2769076528-7g9mh          1/1       Running            0          30s
demo-echo-service-3636052215-3vw8c          1/1       Running            0          29s

kubectl get ing
NAME                       HOSTS                                 ADDRESS   PORTS     AGE
public-demo-echo-service   public-demo-echo-service.kube.local             80        1m
secure-demo-echo-service   secure-demo-echo-service.kube.local             80        1m
```

Test 1: public service with no auth header

```console
$ curl -H 'Host: public-demo-echo-service.kube.local' -v 192.168.99.100
* Rebuilt URL to: 192.168.99.100/
*   Trying 192.168.99.100...
* Connected to 192.168.99.100 (192.168.99.100) port 80 (#0)
> GET / HTTP/1.1
> Host: public-demo-echo-service.kube.local
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.10
< Date: Mon, 13 Mar 2017 20:19:21 GMT
< Content-Type: text/plain; charset=utf-8
< Content-Length: 20
< Connection: keep-alive
<
* Connection #0 to host 192.168.99.100 left intact
UserID: , UserRole:
```

Test 2: secure service with no auth header

```console
$ curl -H 'Host: secure-demo-echo-service.kube.local' -v 192.168.99.100
* Rebuilt URL to: 192.168.99.100/
*   Trying 192.168.99.100...
* Connected to 192.168.99.100 (192.168.99.100) port 80 (#0)
> GET / HTTP/1.1
> Host: secure-demo-echo-service.kube.local
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 403 Forbidden
< Server: nginx/1.11.10
< Date: Mon, 13 Mar 2017 20:18:48 GMT
< Content-Type: text/html
< Content-Length: 170
< Connection: keep-alive
<
<html>
<head><title>403 Forbidden</title></head>
<body bgcolor="white">
<center><h1>403 Forbidden</h1></center>
<hr><center>nginx/1.11.10</center>
</body>
</html>
* Connection #0 to host 192.168.99.100 left intact
```

Test 3: public service with valid auth header

```console
$ curl -H 'Host: public-demo-echo-service.kube.local' -H 'User:internal' -v 192.168.99.100
* Rebuilt URL to: 192.168.99.100/
*   Trying 192.168.99.100...
* Connected to 192.168.99.100 (192.168.99.100) port 80 (#0)
> GET / HTTP/1.1
> Host: public-demo-echo-service.kube.local
> User-Agent: curl/7.43.0
> Accept: */*
> User:internal
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.10
< Date: Mon, 13 Mar 2017 20:19:59 GMT
< Content-Type: text/plain; charset=utf-8
< Content-Length: 44
< Connection: keep-alive
<
* Connection #0 to host 192.168.99.100 left intact
UserID: 1443635317331776148, UserRole: admin
```

Test 4: secure service with valid auth header

```console
$ curl -H 'Host: secure-demo-echo-service.kube.local' -H 'User:internal' -v 192.168.99.100
* Rebuilt URL to: 192.168.99.100/
*   Trying 192.168.99.100...
* Connected to 192.168.99.100 (192.168.99.100) port 80 (#0)
> GET / HTTP/1.1
> Host: secure-demo-echo-service.kube.local
> User-Agent: curl/7.43.0
> Accept: */*
> User:internal
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.10
< Date: Mon, 13 Mar 2017 20:17:23 GMT
< Content-Type: text/plain; charset=utf-8
< Content-Length: 43
< Connection: keep-alive
<
* Connection #0 to host 192.168.99.100 left intact
UserID: 605394647632969758, UserRole: admin
```
