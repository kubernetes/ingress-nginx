# Nginx Ingress DaemonSet

In some cases, the Ingress controller will be required to be run at all the nodes in cluster. Using [DaemonSet](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/daemon.md) can achieve this requirement.

## Default Backend

The default backend is a service of handling all url paths and hosts the nginx controller doesn't understand. Deploy the default-http-backend as follow:

```console
$ kubectl apply -f ../../deployment/nginx/default-backend.yaml 
deployment "default-http-backend" configured
service "default-http-backend" configured

$ kubectl -n kube-system get svc
NAME                   CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
default-http-backend   192.168.3.6   <none>        80/TCP    1h

$ kubectl -n kube-system get po
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-6b47n   1/1       Running   0          1h
```

## Ingress DaemonSet

Deploy the daemonset as follows:

```console
$ kubectl apply -f nginx-ingress-daemonset.yaml
daemonset "nginx-ingress-lb" created

$ kubectl -n kube-system get ds
NAME               DESIRED   CURRENT   READY     NODE-SELECTOR   AGE
nginx-ingress-lb   2         2         2         <none>          21s

$ kubectl -n kube-system get po
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-6b47n   1/1       Running   0          2h
nginx-ingress-lb-8381i                  1/1       Running   0          56s
nginx-ingress-lb-h54gf                  1/1       Running   0          56s
```
