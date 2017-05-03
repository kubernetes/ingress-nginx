# Customize the HAProxy configuration

This example aims to demonstrate the deployment of an haproxy ingress controller and
use a ConfigMap to configure custom Diffie-Hellman parameters file to help with
"Perfect Forward Secrecy".

## Prerequisites

This document has the following prerequisites:

Deploy only the tls-secret and the default backend from the [deployment instructions](../../../deployment/haproxy/)

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Custom configuration

```console
$ cat haproxy-conf.yaml
apiVersion: v1
data:
  ssl-dh-param: "default/lb-dhparam"
kind: ConfigMap
metadata:
  name: haproxy-conf
```

```console
$ kubectl create -f haproxy-conf.yaml
```

## Custom DH parameters secret

```console
$> openssl dhparam 1024 2> /dev/null | base64
LS0tLS1CRUdJTiBESCBQQVJBTUVURVJ...
```

```console
$ cat ssl-dh-param.yaml
apiVersion: v1
data:
  dhparam.pem: "LS0tLS1CRUdJTiBESCBQQVJBTUVURVJ..."
kind: Secret
type: Opaque
metadata:
  name: lb-dhparam
```

```console
$ kubectl create -f ssl-dh-param.yaml
```

## Controller

You can deploy the controller as follows:

```console
$ kubectl apply -f haproxy-ingress-deployment.yaml
deployment "haproxy-ingress-deployment" created

$ kubectl get po
NAME                                       READY     STATUS    RESTARTS   AGE
default-http-backend-2198840601-0k6sv      1/1       Running   0          5m
haproxy-ingress-650604828-4vvwb            1/1       Running   0          57s
```

## Test

Check the contents of the configmap is present in the haproxy.cfg file using:
`kubectl exec -it haproxy-ingress-650604828-4vvwb cat /usr/local/etc/haproxy/haproxy.cfg`

Check all the config options in the [HAProxy Ingress docs](https://github.com/jcmoraisjr/haproxy-ingress#configmap)