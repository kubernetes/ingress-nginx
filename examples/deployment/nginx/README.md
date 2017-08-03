# Deploying the Nginx Ingress controller

This example aims to demonstrate the deployment of an nginx ingress controller.

## Default Backend

The default backend is a Service capable of handling all url paths and hosts the
nginx controller doesn't understand. This most basic implementation just returns
a 404 page:

```console
$ kubectl apply -f default-backend.yaml
deployment "default-http-backend" created
service "default-http-backend" created

$ kubectl -n kube-system get pods
NAME                                    READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-qgwdd   1/1       Running   0          28s
```

## Controller

You can deploy the controller as follows:

1. Disable the ingress addon:
```console
$ minikube addons disable ingress
```
2. Use the [docker daemon](https://github.com/kubernetes/minikube/blob/master/docs/reusing_the_docker_daemon.md)
3. [Build the image](../../../docs/dev/getting-started.md)
4. Change [nginx-ingress-controller.yaml](nginx-ingress-controller.yaml) to use the appropriate image. Local images can be
seen by performing `docker images`.
```yaml
image: <IMAGE-NAME>:<TAG>
```
5. Create the nginx-ingress-controller deployment:
```console
$ kubectl apply -f nginx-ingress-controller.yaml
deployment "nginx-ingress-controller" created

$ kubectl -n kube-system get pods
NAME                                       READY     STATUS    RESTARTS   AGE
default-http-backend-2657704409-qgwdd      1/1       Running   0          2m
nginx-ingress-controller-873061567-4n3k2   1/1       Running   0          42s
```

Note the default settings of this controller:
* serves a `/healthz` url on port 10254, as a status probe
* takes a `--default-backend-service` argument pointing to the Service created above

## Running on a cloud provider

If you're running this ingress controller on a cloud-provider, you should assume
the provider also has a native Ingress controller and set the annotation
`kubernetes.io/ingress.class: nginx` in all Ingresses meant for this controller.
You might also need to open a firewall-rule for ports 80/443 of the nodes the
controller is running on.
