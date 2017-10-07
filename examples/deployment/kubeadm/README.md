# Deploying the Nginx Ingress controller on kubeadm clusters

This example aims to demonstrate the deployment of an nginx ingress controller with kubeadm, 
and is nearly the same as the example above, but here the Ingress Controller is using 
`hostNetwork: true` until the CNI kubelet networking plugin is compatible with `hostPort`
(see issue: [kubernetes/kubernetes#31307](https://github.com/kubernetes/kubernetes/issues/31307))

## Default Backend

The default backend is a Service capable of handling all url paths and hosts the
nginx controller doesn't understand. This most basic implementation just returns
a 404 page.

## Controller

The Nginx Ingress Controller uses nginx (surprisingly!) to loadbalance requests that are coming to
ports 80 and 443 to Services in the cluster.

```console
$ kubectl apply -f https://rawgit.com/kubernetes/ingress/master/examples/deployment/nginx/kubeadm/nginx-ingress-controller.yaml
deployment "default-http-backend" created
service "default-http-backend" created
deployment "nginx-ingress-controller" created
```

Note the default settings of this controller:
* serves a `/healthz` url on port 10254, as both a liveness and readiness probe
* automatically deploys the `gcr.io/google_containers/defaultbackend:1.0` image for serving 404 requests.

At its current state, it only supports running on `amd64` nodes.

## Running on a cloud provider

If you're running this ingress controller on a cloudprovider, you should assume
the provider also has a native Ingress controller and set the annotation
`kubernetes.io/ingress.class: nginx` in all Ingresses meant for this controller.
You might also need to open a firewall-rule for ports 80/443 of the nodes the
controller is running on.
