# Deploying the Nginx Ingress controller

This example aims to demonstrate the deployment of an nginx ingress controller and
with the use of an annotation in the Ingress rule be able to customize the nginx 
configuration.

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

## Ingress
The Ingress in this example adds a custom header to Nginx configuration that only applies to that specific Ingress. If you want to add headers that apply globally to all Ingresses, please have a look at [this example](/examples/customization/custom-headers/nginx).

```console
$ kubectl apply -f ingress.yaml
deployment "nginx-ingress-controller" created
```

## Test

Check if the contents of the annotation are present in the nginx.conf file using:
`kubectl exec nginx-ingress-controller-873061567-4n3k2 -n kube-system cat /etc/nginx/nginx.conf`
