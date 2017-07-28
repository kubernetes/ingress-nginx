# TCP loadbalancing

This example show how to implement TCP loadbalancing through the Nginx Controller

## Prerequisites

You need a [Default Backend service](/examples/deployment/nginx/README.md#default-backend) and a [test HTTP service](/examples/PREREQUISITES.md#test-http-service) for this example

## Config TCP Service

To configure which services and ports will be exposed:
```
$ kubectl create -f nginx-tcp-ingress-configmap.yaml
configmap "nginx-tcp-ingress-configmap" created

$ kubectl -n kube-system get configmap 
NAME                                 DATA      AGE
nginx-tcp-ingress-configmap          1         10m

$ kubectl -n kube-system describe configmap nginx-tcp-ingress-configmap
Name:           nginx-tcp-ingress-configmap
Namespace:      kube-system
Labels:         <none>
Annotations:    <none>

Data
====
9000:
----
default/http-svc:80
```

The file `nginx-tcp-ingress-configmap.yaml` uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`

It is possible to use a number or the name of the port

## Deploy
```
$ kubectl create -f nginx-tcp-ingress-controller.yaml
replicationcontroller "nginx-ingress-controller" created

$ kubectl -n kube-system get rc
NAME                       DESIRED   CURRENT   READY     AGE
nginx-ingress-controller   1         1         1         3m

$ kubectl -n kube-system describe rc nginx-ingress-controller
Name:           nginx-ingress-controller
Namespace:      kube-system
Image(s):       gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.11
Selector:       k8s-app=nginx-tcp-ingress-lb
Labels:         k8s-app=nginx-ingress-lb
Annotations:    <none>
Replicas:       1 current / 1 desired
Pods Status:    1 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                    -------------   --------        ------                  -------
  1m            1m              1       replication-controller                  Normal          SuccessfulCreate        Created pod: nginx-ingress-controller-mv92m
  
$ kubectl -n kube-system get po -o wide
NAME                                    READY     STATUS    RESTARTS   AGE       IP           
default-http-backend-2198840601-fxxjg   1/1       Running   0          2h        172.16.22.4   10.114.51.137
nginx-ingress-controller-mv92m          1/1       Running   0          2m        172.16.63.6   10.114.51.207
```

## Test
```
$ (sleep 1; echo "GET / HTTP/1.1"; echo "Host: 172.16.63.6:9000"; echo;echo;sleep 2) | telnet 172.16.63.6 9000
Trying 172.16.63.6...
Connected to 172.16.63.6.
Escape character is '^]'.
HTTP/1.1 200 OK
Server: nginx/1.9.11
Date: Thu, 20 Apr 2017 07:53:30 GMT
Content-Type: text/plain
Transfer-Encoding: chunked
Connection: keep-alive

f
CLIENT VALUES:

1b
client_address=172.16.63.6

c
command=GET

c
real path=/

a
query=nil

14
request_version=1.1

25
request_uri=http://172.16.63.6:8080/

1


f
SERVER VALUES:

2a
server_version=nginx: 1.9.11 - lua: 10001

1


12
HEADERS RECEIVED:

16
host=172.16.63.6:9000

6
BODY:

14
-no body in request-
0

Connection closed by foreign host.
```
