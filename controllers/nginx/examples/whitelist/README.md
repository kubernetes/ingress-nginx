
This example shows how is possible to restrict access

```
echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: whitelist
  annotations:
    ingress.kubernetes.io/whitelist-source-range: "1.1.1.1/24"
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /
        backend:
          serviceName: echoheaders
          servicePort: 80
" | kubectl create -f -
```

Check the annotation is present in the Ingress rule:
```
$ kubectl get ingress whitelist -o yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress.kubernetes.io/whitelist-source-range: 1.1.1.1/24
  creationTimestamp: 2016-06-09T21:39:06Z
  generation: 2
  name: whitelist
  namespace: default
  resourceVersion: "419363"
  selfLink: /apis/extensions/v1beta1/namespaces/default/ingresses/whitelist
  uid: 97b74737-2e8a-11e6-90db-080027d2dc94
spec:
  rules:
  - host: whitelist.bar.com
    http:
      paths:
      - backend:
          serviceName: echoheaders
          servicePort: 80
        path: /
status:
  loadBalancer:
    ingress:
    - ip: 172.17.4.99
```

Finally test is not possible to access the URL

```
$ curl -v http://172.17.4.99/ -H 'Host: whitelist.bar.com'
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET / HTTP/1.1
> Host: whitelist.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 403 Forbidden
< Server: nginx/1.11.1
< Date: Thu, 09 Jun 2016 21:56:17 GMT
< Content-Type: text/html
< Content-Length: 169
< Connection: keep-alive
<
<html>
<head><title>403 Forbidden</title></head>
<body bgcolor="white">
<center><h1>403 Forbidden</h1></center>
<hr><center>nginx/1.11.1</center>
</body>
</html>
* Connection #0 to host 172.17.4.99 left intact
```

Removing the annotation removes the restriction

```
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET / HTTP/1.1
> Host: whitelist.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.1
< Date: Thu, 09 Jun 2016 21:57:44 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
<
CLIENT VALUES:
client_address=10.2.89.7
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://whitelist.bar.com:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=close
host=whitelist.bar.com
user-agent=curl/7.43.0
x-forwarded-for=10.2.89.1
x-forwarded-host=whitelist.bar.com
x-forwarded-port=80
x-forwarded-proto=http
x-real-ip=10.2.89.1
BODY:
* Connection #0 to host 172.17.4.99 left intact
```

