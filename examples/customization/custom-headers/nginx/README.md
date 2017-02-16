# Deploying the Nginx Ingress controller

This example aims to demonstrate the deployment of an nginx ingress controller and
use a ConfigMap to configure a custom list of headers to be passed to the upstream
server

## Default Backend

The default backend is a Service capable of handling all url paths and hosts the
nginx controller doesn't understand. This most basic implementation just returns
a 404 page:

```console
$ kubectl apply -f default-backend.yaml
deployment "default-http-backend" created
service "default-http-backend" created

$ kubectl -n kube-system get po
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-qgwdd   1/1       Running   0          28s
```

## Custom configuration

```console
$ cat nginx-load-balancer-conf.yaml
apiVersion: v1
data:
  proxy-set-headers: "default/custom-headers"
kind: ConfigMap
metadata:
  name: nginx-load-balancer-conf
```

```console
$ kubectl create -f nginx-load-balancer-conf.yaml
```

## Custom headers

```console
$ cat custom-headers.yaml
apiVersion: v1
data:
  X-Different-Name: "true"
  X-Request-Start: t=${msec}
  X-Using-Nginx-Controller: "true"
kind: ConfigMap
metadata:
  name: proxy-headers
  namespace: default

```

```console
$ kubectl create -f custom-headers.yaml
```

## Controller

You can deploy the controller as follows:

```console
$ kubectl apply -f nginx-ingress-controller.yaml
deployment "nginx-ingress-controller" created

$ kubectl -n kube-system get po
NAME                                       READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-qgwdd      1/1       Running   0          2m
nginx-ingress-controller-873061567-4n3k2   1/1       Running   0          42s
```

## Test

Check the contents of the configmap is present in the nginx.conf file using:
`kubectl exec nginx-ingress-controller-873061567-4n3k2 -n kube-system cat /etc/nginx/nginx.conf`
