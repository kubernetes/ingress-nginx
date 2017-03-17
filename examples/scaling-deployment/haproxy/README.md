# Deploying multi Haproxy Ingress Controllers

This example aims to demonstrate the Deployment of multi haproxy ingress controllers.

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
$ kubectl create -f default-backend.yaml 
deployment "default-http-backend" created
service "default-http-backend" created

$ kubectl get svc
NAME                   CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
default-http-backend   192.168.3.4   <none>        80/TCP    30m

$ kubectl get pods
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-q5sb6              1/1       Running   0          30m
```

## Ingress Deployment

Deploy the Deployment of multi controllers as follows:

```console
$ kubectl apply -f haproxy-ingress-deployment.yaml
deployment "haproxy-ingress" created
```

Check if the controller was successfully deployed:
```console
$ kubectl get deployment
NAME                   DESIRED   CURRENT   UP-TO-DATE     AVAILABLE   AGE
default-http-backend   1         1         1              1           30m
haproxy-ingress        2         2         2              2           45s

$ kubectl get pods
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-q5sb6              1/1       Running   0          35m
haproxy-ingress-1779899633-k045t        1/1       Running   0          1m
haproxy-ingress-1779899633-mhthv        1/1       Running   0          1m
```
