# Deploying the Nginx Ingress controller

This example aims to demonstrate the deployment of an nginx ingress controller and
use a ConfigMap to enable nginx vts module and export metrics for prometheus,to enable 
vts metric,you can simply run `kubectl apply -f nginx`,a deployment and service will be
created which already has a `prometheus.io/scrap: 'true'` annotation and if you added
the recommended Prometheus service-endpoint scraping [configuration](https://raw.githubusercontent.com/prometheus/prometheus/master/documentation/examples/prometheus-kubernetes.yml),
Prometheus will scrape it automatically and you start using the generated metrics right away.

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
$ cat nginx-vts-metrics-conf.yaml
apiVersion: v1
data:
  enable-vts-status: "true"
kind: ConfigMap
metadata:
  name: nginx-vts-metrics-conf
  namespace: kube-system
```

```console
$ kubectl create -f nginx-vts-metrics-conf.yaml
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

## Result
Check  wether to open the vts status:
```console
$ kubectl exec nginx-ingress-controller-873061567-4n3k2 -n kube-system cat /etc/nginx/nginx.conf|grep vhost_traffic_status_display
 vhost_traffic_status_display;
 vhost_traffic_status_display_format html;
```
