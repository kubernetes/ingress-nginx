
To configure which services and ports will be exposed
```
kubectl create -f tcp-configmap-example.yaml
```

The file `tcp-configmap-example.yaml` uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`
It is possible to use a number or the name of the port.

```
kubectl create -f rc-tcp.yaml
```

Now we can test the new service:
```
$ (sleep 1; echo "GET / HTTP/1.1"; echo "Host: 172.17.4.99:9000"; echo;echo;sleep 2) | telnet 172.17.4.99 9000

Trying 172.17.4.99...
Connected to 172.17.4.99.
Escape character is '^]'.
HTTP/1.1 200 OK
Server: nginx/1.9.7
Date: Tue, 15 Dec 2015 14:46:28 GMT
Content-Type: text/plain
Transfer-Encoding: chunked
Connection: keep-alive

f
CLIENT VALUES:

1a
client_address=10.2.84.45

c
command=GET

c
real path=/

a
query=nil

14
request_version=1.1

25
request_uri=http://172.17.4.99:8080/

1


f
SERVER VALUES:

28
server_version=nginx: 1.9.7 - lua: 9019

1


12
HEADERS RECEIVED:

16
host=172.17.4.99:9000

6
BODY:

14
-no body in request-
0
```
