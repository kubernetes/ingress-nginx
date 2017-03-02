# Haproxy Ingress DaemonSet

In some cases, the Ingress controller will be required to be run at all the nodes in cluster. Using [DaemonSet](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/daemon.md) can achieve this requirement.

## Prerequisites

This ingress controller doesn't yet have support for
[ingress classes](/examples/PREREQUISITES.md#ingress-class). You MUST turn
down any existing ingress controllers before running HAProxy Ingress controller or
they will fight for Ingresses. This includes any cloudprovider controller.

This document has also the following prerequisites:

* Create a [TLS secret](/examples/PREREQUISITES.md#tls-certificates) named `tls-secret` to be used as default TLS certificate

Creating the TLS secret:

```console
$ openssl req \
  -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout tls.key -out tls.crt -subj '/CN=localhost'
$ kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key
$ rm -v tls.crt tls.key
```

## Default Backend

The default backend is a service of handling all url paths and hosts the haproxy controller doesn't understand. Deploy the default-http-backend as follow:

```console
$ kubectl apply -f ../../deployment/nginx/default-backend.yaml 
deployment "default-http-backend" configured
service "default-http-backend" configured

$ kubectl -n kube-system get svc
NAME                   CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
default-http-backend   192.168.3.4   <none>        80/TCP    30m

$ kubectl -n kube-system get pods
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-q5sb6              1/1       Running   0          30m
```

## Ingress DaemonSet

Deploy the daemonset as follows:

```console
$ kubectl apply -f haproxy-ingress-daemonset.yaml
```

Check if the controller was successfully deployed:
```console
$ kubectl -n kube-system get ds
NAME               DESIRED   CURRENT   READY     NODE-SELECTOR   AGE
haproxy-ingress   2         2         2         <none>          45s

$ kubectl -n kube-system get pods
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-q5sb6              1/1       Running   0          45m
haproxy-ingress-km32x                   1/1       Running   0          1m
```
