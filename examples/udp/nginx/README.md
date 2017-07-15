# UDP loadbalancing

This example show how to implement UDP loadbalancing throught the Nginx Controller

## Prerequisites

You need a [Default Backend service](/examples/deployment/nginx/README.md#default-backend) and a [kube-dns service](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns#kube-dns) for this example
```
$ kubectl -n kube-system get svc 
NAMESPACE     NAME                   CLUSTER-IP      EXTERNAL-IP   PORT(S)         AGE
kube-system   default-http-backend   192.168.3.204   <none>        80/TCP          1d
kube-system   kube-dns               192.168.3.10    <none>        53/UDP,53/TCP   23h
```

## Config UDP Service

To configure which services and ports will be exposed:
```
$ kubectl create -f nginx-udp-ingress-configmap.yaml
configmap "nginx-udp-ingress-configmap" created

$ kubectl -n kube-system get configmap 
NAME                                 DATA      AGE
extension-apiserver-authentication   1         1d
kube-dns                             0         1d
nginx-udp-ingress-configmap          1         15m

$ kubectl -n kube-system describe configmap nginx-udp-ingress-configmap
Name:           nginx-udp-ingress-configmap
Namespace:      kube-system
Labels:         <none>
Annotations:    <none>

Data
====
9001:
----
kube-system/kube-dns:53
```

The file `nginx-udp-ingress-configmap.yaml` uses a ConfigMap where the key is the external port to use and the value is
`<namespace/service name>:<service port>`

## Deploy
```
$ kubectl create -f nginx-udp-ingress-controller.yaml
replicationcontroller "nginx-udp-ingress-controller" created

$ kubectl -n kube-system get rc
NAME                           DESIRED   CURRENT   READY     AGE
nginx-udp-ingress-controller   1         1         1         13m

$ kubectl -n kube-system describe rc nginx-udp-ingress-controller
Name:           nginx-udp-ingress-controller
Namespace:      kube-system
Image(s):       gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.11
Selector:       k8s-app=nginx-udp-ingress-lb
Labels:         k8s-app=nginx-udp-ingress-lb
Annotations:    <none>
Replicas:       1 current / 1 desired
Pods Status:    1 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen     LastSeen        Count   From                    SubObjectPath   Type            Reason                  Message
  ---------     --------        -----   ----                    -------------   --------        ------                  -------
  46s           46s             1       replication-controller                  Normal          SuccessfulCreate        Created pod: nginx-udp-ingress-controller-m0pjl
  
$ kubectl -n kube-system get po -o wide
NAME                                    READY     STATUS    RESTARTS   AGE       IP           
NAME                                    READY     STATUS    RESTARTS   AGE       IP            NODE
default-http-backend-2198840601-5j1zc   1/1       Running   0          1d        172.16.45.3   10.114.51.28
kube-dns-1874783228-nvs9f               3/3       Running   0          23h       172.16.10.3   10.114.51.217
nginx-udp-ingress-controller-m0pjl      1/1       Running   0          1m        172.16.10.2   10.114.51.217
```

## Test
```
$ nc -uzv 172.16.10.2 9001
Connection to 172.16.10.2 9001 port [udp/*] succeeded!
```
