
Create the Ingress controller
```
kubectl create -f rc-default.yaml
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
