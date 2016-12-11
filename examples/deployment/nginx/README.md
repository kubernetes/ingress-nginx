# Deploying an Nginx Ingress controller

This example aims to demonstrate the deployment of an nginx ingress controller.

## Default Backend

The default backend is a Service capable of handling all url paths and hosts the
nginx controller doesn't understand. This most basic implementation just returns
a 404 page:
```console
$ kubectl create -f default-backend.yaml
replicationcontroller "default-http-backend" created

$ kubectl expose rc default-http-backend --port=80 --target-port=8080 --name=default-http-backend
service "default-http-backend" exposed

$ kubectl get po
NAME                             READY     STATUS              RESTARTS   AGE
default-http-backend-ppqdj       1/1       Running             0          1m
```

## Controller

You can deploy the controller as follows:

```console
$ kubectl create -f rc.yaml
replicationcontroller "nginx-ingress-controller" created

$ kubectl get po
NAME                             READY     STATUS              RESTARTS   AGE
default-http-backend-ppqdj       1/1       Running             0          1m
nginx-ingress-controller-vbgf9   0/1       ContainerCreating   0          2s
```

Note the default settings of this controller:
* serves a `/healthz` url on port 10254, as both a liveness and readiness probe
* takes a `--default-backend-service` arg pointing to a Service, created above

## Running on a cloud provider

If you're running this ingress controller on a cloudprovider, you should assume
the provider also has a native Ingress controller and set the annotation
`kubernetes.io/ingress.class: nginx` in all Ingresses meant for this controller.
You might also need to open a firewall-rule for ports 80/443 of the nodes the
controller is running on.



