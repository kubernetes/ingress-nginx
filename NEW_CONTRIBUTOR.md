## New Contributor Tips

Welcome to the Ingress Nginx new contributor tips.
This guide briefly outlines the necessary knowledge & tools, required to start working on Ingress-NGINX Issues.

### Prerequisites
- Basic understanding of linux
- Familiarity with the command line on linux
- OSI Model(Links below)

### Introduction
It all starts with the OSI model...
> The Open Systems Interconnection (OSI) model describes seven layers that computer systems use to communicate over a network. It was the first standard model for network communications, adopted by all major computer and telecommunication companies

![Describes the 7 Layers of the OSI Model](https://i.imgur.com/qF0KjBq.png)

#### Reading material for OSI Model
[OSI Model CertificationKits](https://www.certificationkits.com/cisco-certification/cisco-ccna-640-802-exam-certification-guide/cisco-ccna-the-osi-model/)

### Approaching the problem


Not everybody knows everything. But the factors that help are a love/passion for this to begin. But to move forward, its the approach and not the knowledge that sustains prolonged joy, while working on issues. If the approach is simple and powered by good-wishes-for-community, then info & tools are forthcoming and easy.

Here we take a bird's eye-view of the hops in the network plumbing, that a packet takes, from source to destination, when we run `curl`, from a laptop to a nginx webserver process, running in a container, inside a pod, inside a Kubernetes cluster, created using `kind` or `minikube` or any other cluster-management tool.

### [Kind](https://kind.sigs.k8s.io/) cluster example on a Linux Host

#### TL;DR
The destination of the packet from the curl command, is looked up, in the `routing table`. Based on the route, the the packet first travels to the virtual bridge `172.18.0.1` interface, created by docker, when we created the kind cluster on a laptop. Next the packet is forwarded to `172.18.0.2`(See below on how we got this IP address), within the kind cluster. The `kube-proxy` container creates iptables rules that make sure the packet goes to the correct pod ip in this case `10.244.0.5`

Command:
```
# docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED       STATUS          PORTS                       NAMES
230e7246a32c   kindest/node:v1.24.1   "/usr/local/bin/entr…"   2 weeks ago   Up 54 seconds   127.0.0.1:38143->6443/tcp   kind-control-plane

# docker inspect kind-control-plane -f '{{ .NetworkSettings.Networks.kind.IPAddress }}'
172.18.0.2

```



If this part is confusing, you would first need to understand what a [bridge](https://tldp.org/HOWTO/BRIDGE-STP-HOWTO/what-is-a-bridge.html) is and what [docker network](https://docs.docker.com/network/) is.



#### The journey of a curl packet.
Let's begin with creating a [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) Cluster on your laptop
```
# kind create cluster
```
This will create a cluster called `kind`, to view the clusters type
```
# kind get clusters                                                                                                                              
kind
```
Kind ships with `kubectl`, so we can use that to communicate with our clusters.
```
# kubectl get no -o wide                                                                                                                             
NAME                 STATUS   ROLES           AGE     VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE       KERNEL-VERSION    CONTAINER-RUNTIME
kind-control-plane   Ready    control-plane   5d23h   v1.24.1   172.18.0.2    <none>        Ubuntu 21.10   5.18.12-arch1-1   containerd://1.6.4
```
Kind creates  a cluster using docker container as nodes, it does this using [containerd](https://containerd.io/) within the docker container.
The concept of Docker in Docker is very important here.

To start with simply create a nginx deployment using `kubectl`.
```
# kubectl create deployment nginx --image nginx:alpine --port=80
deployment.apps/nginx created
```
Then we expose this as a NodePort Service.
```
# kubectl expose deployment/nginx --type=NodePort
service/nginx-new exposed
```
Command: Now we can see that the service has been exposed.
```
# kubectl get svc -o wide
NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE     SELECTOR
nginx          NodePort    10.96.176.241   <none>        80:32329/TCP   4d8h    app=nginx
```
Output Relevance: From the above output, we can see that our nginx pod is being exposed as the `NodePort` service type, and now we can curl the Node IP `172.18.0.2` with the exposed port `32329`

Command: The pod has an IP as shown below
```
# kubectl get po -o wide                                                                                                                            
NAME                               READY   STATUS    RESTARTS      AGE     IP            NODE                 NOMINATED NODE   READINESS GATES
nginx-6c8b449b8f-pdvdk             1/1     Running   1 (32h ago)   4d8h    10.244.0.5    kind-control-plane   <none>           <none>
```

Command: We can use `curl` on the laptop to view the nginx container that is running on port `32329`.

```
# curl  172.18.0.2:32329

<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```
Now, we can check the ip interfaces as well subnets for our system is connected to:

```
$ ifconfig
ethbr0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.31.9  netmask 255.255.255.0  broadcast 192.168.31.255
        inet6 fe80::7530:9ae5:3e8d:e45a  prefixlen 64  scopeid 0x20<link>
        ether 2e:90:b3:e8:52:5b  txqueuelen 1000  (Ethernet)
        RX packets 31220566  bytes 44930589084 (41.8 GiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 18104006  bytes 1757183680 (1.6 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

br-2fffe5cd5d9e: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 172.18.0.1  netmask 255.255.0.0  broadcast 172.18.255.255
        inet6 fc00:f853:ccd:e793::1  prefixlen 64  scopeid 0x0<global>
        inet6 fe80::42:12ff:fed3:8fb0  prefixlen 64  scopeid 0x20<link>
        inet6 fe80::1  prefixlen 64  scopeid 0x20<link>
        ether 02:42:12:d3:8f:b0  txqueuelen 0  (Ethernet)
        RX packets 3547  bytes 414792 (405.0 KiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 6267  bytes 8189931 (7.8 MiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
docker0: flags=4099<UP,BROADCAST,MULTICAST>  mtu 1500
        inet 172.17.0.1  netmask 255.255.0.0  broadcast 172.17.255.255
        inet6 fe80::42:a2ff:fe09:5edb  prefixlen 64  scopeid 0x20<link>
        ether 02:42:a2:09:5e:db  txqueuelen 0  (Ethernet)
        RX packets 14  bytes 2143 (2.0 KiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 40  bytes 6406 (6.2 KiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```
From the above output we can see that, there are two bridges connected to our systems network interface,one is the docker default bridge`docker0` and the other created by kind
`br-2fffe5cd5d9e`.

Since kind creates nodes as containers, this is easily accessible via `docker ps`.
```
$ docker ps

CONTAINER ID   IMAGE                  COMMAND                  CREATED      STATUS        PORTS                       NAMES
230e7246a32c   kindest/node:v1.24.1   "/usr/local/bin/entr…"   6 days ago   Up 33 hours   127.0.0.1:38143->6443/tcp   kind-control-plane
```
If we do a docker `exec` we can enter the container, we can also see the network interfaces within the container.
```
# docker exec -it 230e7246a32c bash

# root@kind-control-plane:/# ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: vethdb0d1da1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether a2:a1:ce:08:d2:39 brd ff:ff:ff:ff:ff:ff link-netns cni-ddc25710-030a-cc05-c600-5a183fae01f7
    inet 10.244.0.1/32 scope global vethdb0d1da1
       valid_lft forever preferred_lft forever
3: veth4d76603f@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 9a:9b:6b:3e:d1:53 brd ff:ff:ff:ff:ff:ff link-netns cni-f2270000-8fc8-6f89-e56b-4759ae10a084
    inet 10.244.0.1/32 scope global veth4d76603f
       valid_lft forever preferred_lft forever
4: vethcc2586d6@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 52:f9:20:63:62:a2 brd ff:ff:ff:ff:ff:ff link-netns cni-97e337cd-1322-c1fa-7523-789af94f397f
    inet 10.244.0.1/32 scope global vethcc2586d6
       valid_lft forever preferred_lft forever
5: veth783189a9@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether ba:e1:55:1f:6f:12 brd ff:ff:ff:ff:ff:ff link-netns cni-90849001-668a-03d2-7d9e-192de79ccc59
    inet 10.244.0.1/32 scope global veth783189a9
       valid_lft forever preferred_lft forever
6: veth79c98c12@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 22:05:55:c7:86:e9 brd ff:ff:ff:ff:ff:ff link-netns cni-734dfac9-9f70-ab33-265b-21569d90312a
    inet 10.244.0.1/32 scope global veth79c98c12
       valid_lft forever preferred_lft forever
7: veth5b221c83@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 92:3f:04:54:72:5a brd ff:ff:ff:ff:ff:ff link-netns cni-d8f6666b-1cfb-ef08-4bf8-237a7fc32da2
    inet 10.244.0.1/32 scope global veth5b221c83
       valid_lft forever preferred_lft forever
8: vethad630fb8@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 32:78:ec:f6:01:ea brd ff:ff:ff:ff:ff:ff link-netns cni-6cb3c179-cb17-3b81-2051-27231c44a3c4
    inet 10.244.0.1/32 scope global vethad630fb8
       valid_lft forever preferred_lft forever
9: veth573a629b@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether e2:57:f8:c9:bc:94 brd ff:ff:ff:ff:ff:ff link-netns cni-d2dbb903-8310-57b4-7ba4-9f353dbc79dc
    inet 10.244.0.1/32 scope global veth573a629b
       valid_lft forever preferred_lft forever
10: eth0@if11: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 02:42:ac:12:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.18.0.2/16 brd 172.18.255.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fc00:f853:ccd:e793::2/64 scope global nodad
       valid_lft forever preferred_lft forever
    inet6 fe80::42:acff:fe12:2/64 scope link
       valid_lft forever preferred_lft forever
11: vethd7368e27@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 8a:74:ec:f6:d6:c9 brd ff:ff:ff:ff:ff:ff link-netns cni-7c7eb9cd-bbb1-65b0-0480-b8f1265f2f36
    inet 10.244.0.1/32 scope global vethd7368e27
       valid_lft forever preferred_lft forever
12: veth7cadbf2b@if2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 12:48:10:b7:b8:f5 brd ff:ff:ff:ff:ff:ff link-netns cni-b39e37b5-1bc8-626a-a553-a0be2f94a117
    inet 10.244.0.1/32 scope global veth7cadbf2b
       valid_lft forever preferred_lft forever

```
When we run `curl  172.18.0.2:32329` on the laptop it first needs to figure out where `172.18.0.2`, to do this it refers to the host routing table.
```
sudo netstat -rn                                                                                                                           main 
Kernel IP routing table
Destination     Gateway         Genmask         Flags   MSS Window  irtt Iface
0.0.0.0         192.168.31.1    0.0.0.0         UG        0 0          0 ethbr0
172.17.0.0      0.0.0.0         255.255.0.0     U         0 0          0 docker0
172.18.0.0      0.0.0.0         255.255.0.0     U         0 0          0 br-2fffe5cd5d9e
172.19.0.0      0.0.0.0         255.255.0.0     U         0 0          0 br-be5b544733a3
192.168.31.0    0.0.0.0         255.255.255.0   U         0 0          0 ethbr0
192.168.31.0    0.0.0.0         255.255.255.0   U         0 0          0 ethbr0
192.168.39.0    0.0.0.0         255.255.255.0   U         0 0          0 virbr2
192.168.122.0   0.0.0.0         255.255.255.0   U         0 0          0 virbr0
```
Output Relevance: From the above output, you can see that the `iface`(Interface) for `172.18.0.0` is `br-2fffe5cd5d9e`, which means traffic that needs to go to `172.18.0.0` will go through `br-2fffe5cd5d9e` which is created by docker for the kind container (this is the node in case of kind cluster).

Now we need to understand how the packet travels from the container interface to the pod with IP `10.244.0.5`. The component that handles this is called kube-proxy

So what exactly is [kube-proxy](https://kubernetes.io/docs/concepts/overview/components/#kube-proxy):
> Kube-Proxy is a network proxy that runs on each node in your cluster, implementing part of the Kubernetes Service concept.
kube-proxy maintains network rules on nodes. These network rules allow network communication to your Pods from network sessions inside or outside of your cluster

So, as we can see that kube proxy handles the network rules required to aid the communication to the pods, we will look at the [iptables](https://linux.die.net/man/8/iptables)
> `iptables` is a command line interface used to set up and maintain tables for the Netfilter firewall for IPv4, included in the Linux kernel. The firewall matches packets with rules defined in these tables and then takes the specified action on a possible match. Tables is the name for a set of chains

Command:
```
# iptables -t nat -L PREROUTING -n
Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */
DOCKER_OUTPUT  all  --  0.0.0.0/0            172.18.0.1
CNI-HOSTPORT-DNAT  all  --  0.0.0.0/0            0.0.0.0/0            ADDRTYPE match dst-type LOCAL
```

```
# iptables-save | grep PREROUTING
-A PREROUTING -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
```
Output Relevance:
> -A: append new iptable rule
> -j: jump to the target
> KUBE-SERVICES: target

> The above output appends a new rule for PREROUTING which every network packet will go through first as they try to access any kubernetes service


What is `PREROUTING` in iptables?
>PREROUTING: This chain is used to make any routing related decisions before (PRE) sending any packets

To dig in further we need to go to the target, `KUBE-SERVICES` for our nginx service.
```
# iptables -t nat -L KUBE-SERVICES -n| grep nginx
KUBE-SVC-2CMXP7HKUVJN7L6M  tcp  --  0.0.0.0/0            10.96.176.241        /* default/nginx cluster IP */ tcp dpt:80
```
Command:
```
# iptables -t nat -L KUBE-SVC-2CMXP7HKUVJN7L6M -n
Chain KUBE-SVC-2CMXP7HKUVJN7L6M (2 references)
target     prot opt source               destination
KUBE-MARK-MASQ  tcp  -- !10.244.0.0/16        10.96.176.241        /* default/nginx cluster IP */ tcp dpt:80
KUBE-SEP-4IEO3WJHPKXV3AOH  all  --  0.0.0.0/0            0.0.0.0/0            /* default/nginx -> 10.244.0.5:80 */

# iptables -t nat -L KUBE-MARK-MASQ -n
Chain KUBE-MARK-MASQ (31 references)
target     prot opt source               destination
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x4000

# iptables -t nat -L KUBE-SEP-4IEO3WJHPKXV3AOH -n
Chain KUBE-SEP-4IEO3WJHPKXV3AOH (1 references)
target     prot opt source               destination
KUBE-MARK-MASQ  all  --  10.244.0.5           0.0.0.0/0            /* default/nginx */
DNAT       tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/nginx */ tcp to:10.244.0.5:80
```


```
iptables-save | grep 10.96.176.241

-A KUBE-SERVICES -d 10.96.176.241/32 -p tcp -m comment --comment "default/nginx cluster IP" -m tcp --dport 80 -j KUBE-SVC-2CMXP7HKUVJN7L6M
-A KUBE-SVC-2CMXP7HKUVJN7L6M ! -s 10.244.0.0/16 -d 10.96.176.241/32 -p tcp -m comment --comment "default/nginx cluster IP" -m tcp --dport 80 -j KUBE-MARK-MASQ
```

As you can see the rules added by `kube-proxy` helps the packet reach to the destination service.

### Minikube KVM VM Example on Linux

#### TL;DR
Now we look at the curl packet journey on minikube. The `routing table` is looked up to know the destination of the curl packet. The packet then first travels to the virtual bridge `192.168.39.1`, created by minikube kvm2 driver, when we created the minikube cluster, on a linux laptop. Then this packet is forwarded to `192.168.39.57`, within the minikube VM. We have docker containers running in the VM. Among them, the `kube-proxy` container creates iptables rules that make sure the packet goes to the correct pod ip, in this case `172.17.0.4`.


To begin with the minikube example, we first need to create a minikube cluster on a linux laptop. In this example I'll be using the `kvm2` driver option for `minikube start` command, as default.

```
minikube start
😄  minikube v1.26.0 on Arch "rolling"
🆕  Kubernetes 1.24.2 is now available. If you would like to upgrade, specify: --kubernetes-version=v1.24.2
✨  Using the kvm2 driver based on existing profile
👍  Starting control plane node minikube in cluster minikube
🏃  Updating the running kvm2 "minikube" VM ...
🐳  Preparing Kubernetes v1.23.3 on Docker 20.10.12 ...
    ▪ kubelet.housekeeping-interval=5m
🔎  Verifying Kubernetes components...
    ▪ Using image k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
    ▪ Using image k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
    ▪ Using image k8s.gcr.io/ingress-nginx/controller:v1.2.1
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🔎  Verifying ingress addon...
🌟  Enabled addons: ingress, storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```
**Note**: The KVM driver provides a lot of options on customizing the cluster, however that is currently beyond the scope of this guide.

Next we will get the Node IP.
```
$ kubectl get no -o wide                                                                                                                   
NAME       STATUS   ROLES                  AGE   VERSION   INTERNAL-IP     EXTERNAL-IP   OS-IMAGE              KERNEL-VERSION   CONTAINER-RUNTIME
minikube   Ready    control-plane,master   25d   v1.23.3   192.168.39.57   <none>        Buildroot 2021.02.4   4.19.202         docker://20.10.12
```
Minikube creates a Virtual Machine using the KVM2 driver(Other drivers such as Virtualbox do exist see `minikube start --help` for more information ), you should be able to see this with the following output(You may have to use sudo to get this output)

```
$ virsh --connect qemu:///system list
 Id   Name       State
--------------------------
 1    minikube   running

 or

 $ sudo virsh list
 Id   Name       State
--------------------------
 1    minikube   running

```

Moving on, simply create a nginx deployment using `kubectl`.
```
# kubectl create deployment nginx --image nginx:alpine --port=80
deployment.apps/nginx created
```
Then we expose this as a NodePort Service.
```
# kubectl expose deployment/nginx --type=NodePort
service/nginx-new exposed
```
Command: Now we can see that the service has been exposed.
```
# kubectl get svc -o wide                                                                                                                            main 
NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)        AGE   SELECTOR
kubernetes       ClusterIP   10.96.0.1    <none>        443/TCP        25d   <none>
nginx-minikube   NodePort    10.97.44.4   <none>        80:32007/TCP   45h   app=nginx-minikube
```
Output Relevance: From the above output, we can see that our nginx pod is being exposed as the `NodePort` service type, and now we can curl the Node IP `192.168.39.57` with the exposed port `32007`

Command: The pod has an IP as shown below
```
# kubectl get po -o wide
NAME                              READY   STATUS    RESTARTS      AGE   IP           NODE       NOMINATED NODE   READINESS GATES
nginx-minikube-7546f79bd8-x88bt   1/1     Running   3 (43m ago)   45h   172.17.0.4   minikube   <none>           <none>

```

Command: We can use `curl` on the laptop to view the nginx container that is running on port `32007`.
```
curl  192.168.39.57:32007
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

So, how does this packet travel, lets dive in.
We can check the ip interfaces as well subnets for our system is connected to:
```
$ ifconfig
virbr2: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.39.1  netmask 255.255.255.0  broadcast 192.168.39.255
        ether 52:54:00:19:29:93  txqueuelen 1000  (Ethernet)
        RX packets 5132  bytes 1777099 (1.6 MiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 6113  bytes 998530 (975.1 KiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

virbr0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.122.1  netmask 255.255.255.0  broadcast 192.168.122.255
        ether 52:54:00:48:ee:35  txqueuelen 1000  (Ethernet)
        RX packets 23648  bytes 1265196 (1.2 MiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 40751  bytes 60265308 (57.4 MiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```
Output Relevance: From the above output you can see there are two Virtual Bridges created by minikube when we created the cluster on the network. Here, `virbr0` is the default NAT network bridge while `virbr2` is a isolated network bridge on which the pods run.

Minikube creates a Virtual Machine, to enter the virtual machine we can simple do:
```
# minikube ssh
```

The interfaces within the Virtual Machine are as follows.
```
docker0   Link encap:Ethernet  HWaddr 02:42:03:24:26:78
          inet addr:172.17.0.1  Bcast:172.17.255.255  Mask:255.255.0.0
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:31478 errors:0 dropped:0 overruns:0 frame:0
          TX packets:36704 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:3264056 (3.1 MiB)  TX bytes:14061883 (13.4 MiB)

eth0      Link encap:Ethernet  HWaddr 52:54:00:C9:3A:73
          inet addr:192.168.39.57  Bcast:192.168.39.255  Mask:255.255.255.0
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:8245 errors:0 dropped:9 overruns:0 frame:0
          TX packets:3876 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:812006 (792.9 KiB)  TX bytes:1044724 (1020.2 KiB)

eth1      Link encap:Ethernet  HWaddr 52:54:00:7B:37:79
          inet addr:192.168.122.35  Bcast:192.168.122.255  Mask:255.255.255.0
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:4459 errors:0 dropped:9 overruns:0 frame:0
          TX packets:201 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:298528 (291.5 KiB)  TX bytes:25813 (25.2 KiB)

lo        Link encap:Local Loopback
          inet addr:127.0.0.1  Mask:255.0.0.0
          UP LOOPBACK RUNNING  MTU:65536  Metric:1
          RX packets:946772 errors:0 dropped:0 overruns:0 frame:0
          TX packets:946772 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:213465460 (203.5 MiB)  TX bytes:213465460 (203.5 MiB)

vetha4f1dc5 Link encap:Ethernet  HWaddr 3E:1C:FE:C9:75:86
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:10 errors:0 dropped:0 overruns:0 frame:0
          TX packets:16 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:1413 (1.3 KiB)  TX bytes:955 (955.0 B)

vethbf35613 Link encap:Ethernet  HWaddr BA:31:7D:AE:2A:BF
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:3526 errors:0 dropped:0 overruns:0 frame:0
          TX packets:3934 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:342408 (334.3 KiB)  TX bytes:380193 (371.2 KiB)

vethe092a51 Link encap:Ethernet  HWaddr 8A:37:D3:D9:D9:0E
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:9603 errors:0 dropped:0 overruns:0 frame:0
          TX packets:11151 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:1199235 (1.1 MiB)  TX bytes:5449408 (5.1 MiB)
```
Output Relevance: Here we have the Virtual Ethernet and we have docker bridges too since docker runs within the Virtual Machine.

When we do a `curl` to `192.168.39.57:32007` on the laptop the packet first goes to the route table
```
Destination     Gateway         Genmask         Flags   MSS Window  irtt Iface
0.0.0.0         192.168.31.1    0.0.0.0         UG        0 0          0 ethbr0
172.17.0.0      0.0.0.0         255.255.0.0     U         0 0          0 docker0
172.18.0.0      0.0.0.0         255.255.0.0     U         0 0          0 br-2fffe5cd5d9e
172.19.0.0      0.0.0.0         255.255.0.0     U         0 0          0 br-be5b544733a3
192.168.31.0    0.0.0.0         255.255.255.0   U         0 0          0 ethbr0
192.168.31.0    0.0.0.0         255.255.255.0   U         0 0          0 ethbr0
192.168.39.0    0.0.0.0         255.255.255.0   U         0 0          0 virbr2
192.168.122.0   0.0.0.0         255.255.255.0   U         0 0          0 virbr0
```
Output Relevance: As you can see multiple routes are defined here, of which our Virtual Machine Node IP(192.168.39.57) is also shown in the table, so the packet now knows where it has to go.

With that clear we now know how the packet goes from the laptop to the virtual bridge and then enters the Virtual Machine.

Inside the virtual machine, [kube-proxy](https://kubernetes.io/docs/concepts/overview/components/#kube-proxy) handles the routing using iptables.

So what exactly is [kube-proxy](https://kubernetes.io/docs/concepts/overview/components/#kube-proxy)(For those who skipped the kind example):
> Kube-Proxy is a network proxy that runs on each node in your cluster, implementing part of the Kubernetes Service concept.
kube-proxy maintains network rules on nodes. These network rules allow network communication to your Pods from network sessions inside or outside of your cluster

So, as we can see that kube proxy handles the network rules required to aid the communication to the pods, we will look at the [iptables](https://linux.die.net/man/8/iptables)
> `iptables` is a command line interface used to set up and maintain tables for the Netfilter firewall for IPv4, included in the Linux kernel. The firewall matches packets with rules defined in these tables and then takes the specified action on a possible match. Tables is the name for a set of chains

Command:

```
# minikube ssh                                                                                                                                      
                         _             _
            _         _ ( )           ( )
  ___ ___  (_)  ___  (_)| |/')  _   _ | |_      __
/' _ ` _ `\| |/' _ `\| || , <  ( ) ( )| '_`\  /'__`\
| ( ) ( ) || || ( ) || || |\`\ | (_) || |_) )(  ___/
(_) (_) (_)(_)(_) (_)(_)(_) (_)`\___/'(_,__/'`\____)

$ sudo iptables -t nat -L PREROUTING -n
Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
KUBE-SERVICES  all  --  0.0.0.0/0            0.0.0.0/0            /* kubernetes service portals */
DOCKER     all  --  0.0.0.0/0            0.0.0.0/0            ADDRTYPE match dst-type LOCAL

$ iptables-save | grep PREROUTING
-A PREROUTING -m comment --comment "kubernetes service portals" -j KUBE-SERVICES

```

Output Relevance:
> -A: append new iptable rule
> -j: jump to the target
> KUBE-SERVICES: target

> The above output appends a new rule for PREROUTING which every network packet will go through first as they try to access any kubernetes service


What is `PREROUTING` in iptables?
>PREROUTING: This chain is used to make any routing related decisions before (PRE) sending any packets

To dig in further we need to go to the target, `KUBE-SERVICES` for our nginx service.
```
# iptables -t nat -L KUBE-SERVICES -n| grep nginx
KUBE-SVC-NRDCJV6H42SDXARP  tcp  --  0.0.0.0/0            10.97.44.4           /* default/nginx-minikube cluster IP */ tcp dpt:80
```
Command:
```
$ sudo iptables -t nat -L| grep KUBE-SVC-NRDCJV6H42SDXARP
KUBE-SVC-NRDCJV6H42SDXARP  tcp  --  0.0.0.0/0            0.0.0.0/0            /* default/nginx-minikube */ tcp dpt:32007
KUBE-SVC-NRDCJV6H42SDXARP  tcp  --  0.0.0.0/0            10.97.44.4           /* default/nginx-minikube cluster IP */ tcp dpt:80

$ sudo iptables -t nat -L KUBE-MARK-MASQ -n
Chain KUBE-MARK-MASQ (19 references)
target     prot opt source               destination
MARK       all  --  0.0.0.0/0            0.0.0.0/0            MARK or 0x4000

sudo iptables-save | grep 172.17.0.4
-A KUBE-SEP-AHQQ7ZFXMEBNX76B -s 172.17.0.4/32 -m comment --comment "default/nginx-minikube" -j KUBE-MARK-MASQ
-A KUBE-SEP-AHQQ7ZFXMEBNX76B -p tcp -m comment --comment "default/nginx-minikube" -m tcp -j DNAT --to-destination 172.17.0.4:80
```
As you can see the rules added by kube-proxy helps the packet reach to the destination service.


### Connection termination
Connection termination is a type of event that occurs when there are load balancers present, the information for this is quite scarce, however I've found the following article, [IBM - Network Termination](https://www.ibm.com/docs/en/sva/9.0.4?topic=balancer-network-termination) that describes what it means by connection termination between clients(laptop) and server(load balancer) and the various other services.

### Different types of connection errors.
The following article on [TCP/IP errors](https://www.ibm.com/docs/en/db2/11.1?topic=message-tcpip-errors) has a list of the important tcp timeout errors that we need to know.


| Common TCP/IP errors | Meaning |
| -------- | -------- |
| Resource temporarily unavailable.| Self-explanatory.     |
| No space is left on a device or system table.|The disk partition is full|
|No route to the host is available.|The routing table doesn't know where to route the packet.|
|Connection was reset by the partner.|This usually means the packet was dropped as soon as it reached the server can be due to a firewall.|
|The connection was timed out.|This indicates the firewall blocking your connection or the connection took too long.|

## OSI Model Layer 7 (Application Layer)

[What is layer 7?](https://www.cloudflare.com/learning/ddos/what-is-layer-7/)
#### Summary
Layer 7 refers to the seventh and topmost layer of the Open Systems Interconnect (OSI) Model known as the application layer. This is the highest layer which supports end-user processes and applications. Layer 7 identifies the communicating parties and the quality of service between them, considers privacy and user authentication, as well as identifies any constraints on the data syntax. This layer is wholly application-specific.


## Setting up Ingress-Nginx Controller

Since we are doing this on our local laptop, we are going to use the following tools:
- [Minikube using KVM driver](https://minikube.sigs.k8s.io/docs/start/) - The host is linux-based in our example
- [Metallb](https://metallb.universe.tf/) - Baremetal load-balancer.
- [KVM](https://www.linux-kvm.org/page/Main_Page) / [Oracle VirtualBox](https://www.virtualbox.org/wiki/Downloads) / [VMWare](https://www.vmware.com/in/products/workstation-pro.html)


### So let's begin with Metallb and Ingress-Nginx setup.

For setting up metallb, we are going to follow the below steps:

 - To begin the installation, we will execute:
```
minikube start
```
- To install Metallb, one can install it using the [manifest](https://metallb.universe.tf/installation/#installation-by-manifest) or by using [helm](https://metallb.universe.tf/installation/#installation-with-helm), for now we will use the Manifest method:
```
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.4/config/manifests/metallb-native.yaml
```

- We need to now configure Metallb, we are using [Layer 2 configuration](https://metallb.universe.tf/configuration/#announce-the-service-ips), let's head over to the [Metallb Configuration](https://metallb.universe.tf/configuration/) website, here you will see how to setup metallb.
>Layer 2 mode does not require the IPs to be bound to the network interfaces of your worker nodes. It works by responding to ARP requests on your local network directly, to give the machine’s MAC address to clients.
In order to advertise the IP coming from an IPAddressPool, an L2Advertisement instance must be associated to the IPAddressPool.
-  We have modified the IP address pool so that our loadbalancer knows which subnet to choose an IP from.Since we have only one minikube IP we need to modify the code given in the documentation.
Save this as `metallb-config.yaml`:
```
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: first-pool
  namespace: metallb-system
spec:
  addresses:
  # The configuration website show's you this

  #- 192.168.10.0/24
  #- 192.168.9.1-192.168.9.5
  #- fc00:f853:0ccd:e799::/124

  # We are going to change this to `minikube ip` as such
  - 192.168.39.57/32
```
Now deploy it using `kubectl`
```
kubectl apply -f metallb-config.yaml
```
- Now that metallb is setup, let's install [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/#quick-start) on the laptop.
Note: We are using the install by manifest option from the Installation manual
```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.3.0/deploy/static/provider/cloud/deploy.yaml
```
or one can also install it using the minikube addons:
```
minikube addons enable ingress
```
 - Once your Ingress-Nginx controller is created you can run the following commands to see the output of the setup done.
```
kubectl get pods -n ingress-nginx
NAME                                        READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-65bld        0/1     Completed   0          14m
ingress-nginx-admission-patch-rwq4x         0/1     Completed   0          14m
ingress-nginx-controller-6dc865cd86-7c5zd   1/1     Running     0          14m
```
The Ingress controller creates a Service with the type LoadBalancer and metallb provides the IP address.

```
kubectl -n ingress-nginx get svc

NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                      AGE
ingress-nginx-controller             LoadBalancer   10.108.154.53   192.168.39.223   80:30367/TCP,443:31491/TCP   4d15h
ingress-nginx-controller-admission   ClusterIP      10.98.54.3      <none>           443/TCP                      4d15h
```

#### Creating an Ingress

We will deploy a `httpd` service in a `httpd` namespace and create a ingress for it.

First, let's create a namespace.
```
kubectl create namespace httpd
```

Next we will create a deployment
```
kubectl create deployment httpd -n httpd --image=httpd:alpine
```

Now, In order to create a service, let's expose this deployment
```
kubectl expose deployment -n httpd httpd --port 80
```
Let's check the `pod` that is created

```
kubectl get po -n httpd
NAME                    READY   STATUS    RESTARTS   AGE
httpd-fb7fcdc77-w287c   1/1     Running   0          64s
```

Let's list the services in the `httpd` namespace
```
kubectl get svc -n httpd
NAME    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
httpd   ClusterIP   10.104.111.0   <none>        80/TCP    13s
```

Once we have this we can now create a n ingress using the following
```
kubectl -n httpd create ingress httpd --class nginx --rule httpd.dev.leonnunes.com/"*"=httpd:80
```
The above output, creates an ingress, for us with the rule to match the service if the host is `httpd.dev.leonnunes.com`. The class here is retrieved from the below command.

To list the `ingressclasses` use
```
kubectl get ingressclasses
NAME    CONTROLLER             PARAMETERS   AGE
nginx   k8s.io/ingress-nginx   <none>       6h49m
```

The following command shows the ingress created
```
$ kubectl get ingress -A -o wide

NAMESPACE   NAME    CLASS   HOSTS                     ADDRESS          PORTS   AGE
httpd       httpd   nginx   httpd.dev.leonnunes.com   192.168.39.223   80      11d
```

To test if the rule works we can now do
```
$ minikube ip
192.168.39.223

$ curl --resolve httpd.dev.leonnunes.com:80:192.168.39.223 httpd.dev.leonnunes.com
<html><body><h1>It works!</h1></body></html>

or

curl -H "Host: httpd.dev.leonnunes.com" 192.168.39.223
```

#### Example of Information found on layer 7
We have setup `Ingress-Nginx`, using `nginx` as a class and `httpd` for this example.

In order to display the info on Layer - 7, we have extracted the Layer 7 information from a simple `curl` request, and then using `tcpdump` command within the `httpd` pod we extracted the network packets and opened it using the `Wireshark` utility.

Below given is the output that is important:
```bash
Frame 4: 391 bytes on wire (3128 bits), 391 bytes captured (3128 bits)
Linux cooked capture v2
Internet Protocol Version 4, Src: 172.17.0.4, Dst: 172.17.0.3
Transmission Control Protocol, Src Port: 49074, Dst Port: 80, Seq: 1, Ack: 1, Len: 319
Hypertext Transfer Protocol
    GET / HTTP/1.1\r\n
    Host: httpd.dev.leonnunes.com\r\n
    X-Request-ID: 6e1a790412a0d1615dc0231358dc9c8b\r\n
    X-Real-IP: 172.17.0.1\r\n
    X-Forwarded-For: 172.17.0.1\r\n
    X-Forwarded-Host: httpd.dev.leonnunes.com\r\n
    X-Forwarded-Port: 80\r\n
    X-Forwarded-Proto: http\r\n
    X-Forwarded-Scheme: http\r\n
    X-Scheme: http\r\n
    User-Agent: curl/7.84.0\r\n
    Accept: */*\r\n
    \r\n
    [Full request URI: http://httpd.dev.leonnunes.com/]
    [HTTP request 1/1]
    [Response in frame: 6]

```
The above output shows the information that the `httpd` pod recieves. The `curl` command sends the host header, `Host: httpd.dev.leonnunes.com`, to the nginx controller, that then matches the rule and sends the information to the right controller

The following output shows what is sent via the laptop.
```
curl --resolve httpd.dev.leonnunes.com:80:192.168.39.57 -H "Host: httpd.dev.leonnunes.com" 192.168.39.57 -vL
* Added httpd.dev.leonnunes.com:80:192.168.39.57 to DNS cache
*   Trying 192.168.39.57:80...
* Connected to 192.168.39.57 (192.168.39.57) port 80 (#0)
> GET / HTTP/1.1
> Host: httpd.dev.leonnunes.com
> User-Agent: curl/7.84.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Mon, 22 Aug 2022 16:05:27 GMT
< Content-Type: text/html
< Content-Length: 45
< Connection: keep-alive
< Last-Modified: Mon, 11 Jun 2007 18:53:14 GMT
< ETag: "2d-432a5e4a73a80"
< Accept-Ranges: bytes
<
<html><body><h1>It works!</h1></body></html>
* Connection #0 to host 192.168.39.57 left intact
```
As you can see from the above output there are several headers added to the curl output after it reaches the `httpd` pod, these headers are added by the Ingress Nginx Controller.


### References
#### Basics of Networking
   - https://www.cisco.com/en/US/docs/security/vpn5000/manager/reference/guide/appA.html
   - http://web.stanford.edu/class/cs101/
   - https://www.geeksforgeeks.org/basics-computer-networking/
   - Subnetting
     - https://www.computernetworkingnotes.com/ccna-study-guide/subnetting-tutorial-subnetting-explained-with-examples.html

#### Video Links
   - https://www.youtube.com/playlist?list=PLhfrWIlLOoKPc2RecyiM_A9nf3fUU3e6g
   - https://www.youtube.com/watch?v=S7MNX_UD7vY&list=PLIhvC56v63IJVXv0GJcl9vO5Z6znCVb1P

### Topics to read about
   - Docker in Docker
   - [Docker/Containers](https://www.oreilly.com/library/view/docker-deep-dive/9781800565135/)
   - Containers

### Basics of Kubernetes
#### Reading Material
- https://nubenetes.com/kubernetes-tutorials/
- https://kubernetes.io/docs/concepts/
#### Video Material
- [Techworld with Nana 101](https://www.youtube.com/playlist?list=PLy7NrYWoggjziYQIDorlXjTvvwweTYoNC)
- [Jeff Geerling Kubernetes 101](https://www.youtube.com/watch?v=IcslsH7OoYo&list=PL2_OBreMn7FoYmfx27iSwocotjiikS5BD)

#### Hands-On Kubernetes
- https://kube.academy/
- https://www.civo.com/academy

### Networking in Kubernetes
- [Kubernetes Networking 101](https://youtu.be/CYnwBIpvSlM?t=284)
- [CNCF Kubernetes 101](https://www.youtube.com/watch?v=cUGXu2tiZMc)

### Tools/Commands to help with troubleshooting.
- [mtr](https://www.redhat.com/sysadmin/linux-mtr-command) - Tracing the packet from the source to destination
- [tcpdump](https://linuxconfig.org/how-to-use-tcpdump-command-on-linux) - Monitor packets
- [wireshark](https://www.lifewire.com/wireshark-tutorial-4143298) - Read/Sniff packets
- [nslookup](https://phoenixnap.com/kb/nslookup-command) - Lookup Nameservers
- [netstat](https://www.lifewire.com/netstat-command-2618098) - List network details
- [curl](https://linuxhandbook.com/curl-command-examples/) - Curl a website from the command line
- [ifconfig](https://www.tecmint.com/ifconfig-command-examples/)/[ip](https://www.geeksforgeeks.org/ip-command-in-linux-with-examples/) - Show ip address configuration
- [dig](https://www.geeksforgeeks.org/dig-command-in-linux-with-examples/) - Query Nameservers
- [ipcalc](https://www.linux.com/topic/networking/how-calculate-network-addresses-ipcalc/) - Calculate IP addresses
- Advanced Tools for troubleshooting
    - [Netshoot](https://github.com/nicolaka/netshoot) - Troubleshoot Networks
- Cluster Creation tools
    - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
    - [minikube](https://minikube.sigs.k8s.io/docs/start/)
- MacOS users
	- [docker-mac-net-connect](https://github.com/chipmk/docker-mac-net-connect) - See this [issue](https://github.com/kubernetes/minikube/issues/7332)
