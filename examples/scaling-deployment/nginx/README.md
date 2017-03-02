# Deploying multi Nginx Ingress Controllers

This example aims to demonstrate the Deployment of multi nginx ingress controllers.

## Default Backend

The default backend is a service of handling all url paths and hosts the nginx controller doesn't understand. Deploy the default-http-backend as follow:

```console
$ kubectl apply -f ../../deployment/nginx/default-backend.yaml 
deployment "default-http-backend" configured
service "default-http-backend" configured

$ kubectl -n kube-system get svc
NAME                   CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
default-http-backend   192.168.3.52   <none>        80/TCP    6m

$ kubectl -n kube-system get po
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-wz6o3   1/1       Running   0          6m
```

## Ingress Deployment

Deploy the Deployment of multi controllers as follows:

```console
$ kubectl apply -f nginx-ingress-deployment.yaml
deployment "nginx-ingress-controller" created

$ kubectl -n kube-system get deployment
NAME                       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
default-http-backend       1         1         1            1           16m
nginx-ingress-controller   2         2         2            2           24s

$ kubectl -n kube-system get po
NAME                                        READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-wz6o3       1/1       Running   0          16m
nginx-ingress-controller-3752011415-0qbi6   1/1       Running   0          39s
nginx-ingress-controller-3752011415-vi8fq   1/1       Running   0          39s
```
