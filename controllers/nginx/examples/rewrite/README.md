
Create an Ingress rule with a rewrite annotation:
```
$ echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress.kubernetes.io/rewrite-target: /
  name: rewrite
  namespace: default
spec:
  rules:
  - host: rewrite.bar.com
    http:
      paths:
      - backend:
          serviceName: echoheaders
          servicePort: 80
        path: /something
" | kubectl create -f -
```

Check the rewrite is working

```
$ curl -v http://172.17.4.99/something -H 'Host: rewrite.bar.com'
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET /something HTTP/1.1
> Host: rewrite.bar.com
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.0
< Date: Tue, 31 May 2016 16:07:31 GMT
< Content-Type: text/plain
< Transfer-Encoding: chunked
< Connection: keep-alive
<
CLIENT VALUES:
client_address=10.2.56.9
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://rewrite.bar.com:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=close
host=rewrite.bar.com
user-agent=curl/7.43.0
x-forwarded-for=10.2.56.1
x-forwarded-host=rewrite.bar.com
x-forwarded-port=80
x-forwarded-proto=http
x-real-ip=10.2.56.1
BODY:
* Connection #0 to host 172.17.4.99 left intact
-no body in request-
```

