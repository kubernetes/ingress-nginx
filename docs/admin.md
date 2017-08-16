# Ingress admin guide

This is a guide to the different deployment styles of an Ingress controller.

## Vanillla deployments

__GCP__: On GCE/GKE, the Ingress controller runs on the
master. If you wish to stop this controller and run another instance on your
nodes instead, you can do so by following this [example](/examples/deployment/gce).

__generic__: You can deploy a generic (nginx or haproxy) Ingress controller by simply
running it as a pod in your cluster, as shown in the [examples](/examples/deployment).
Please note that you must specify the `ingress.class`
[annotation](/examples/PREREQUISITES.md#ingress-class) if you're running on a
cloudprovider, or the cloudprovider controller will fight the nginx controller
for the Ingress.

__AWS__: Until we have an AWS ALB Ingress controller, you can deploy the nginx
Ingress controller behind an ELB on AWS, as shows in the [next section](#stacked-deployments).

## Stacked deployments

__Behind a LoadBalancer Service__: You can deploy an generic controller behind a
Service of `Type=LoadBalancer`, by following this [example](/examples/static-ip/nginx#acquiring-an-ip).
More specifically, first create a LoadBalancer Service that selects the generic
controller pods, then start the generic controller with the `--publish-service`
flag.


__Behind another Ingress__: Sometimes it is desirable to deploy a stack of
Ingresses, like the GCE Ingress -> nginx Ingress -> application. You might
want to do this because the GCE HTTP lb offers some features that the GCE
network LB does not, like a global static IP or CDN, but doesn't offer all the
features of nginx, like url rewriting or redirects.

TODO: Write an example

## Daemonset

Neither a single pod or bank of generic controllers scales with the cluster size.
If you create a daemonset of generic Ingress controllers, every new node
automatically gets an instance of the controller listening on the specified
ports.

TODO: Write an example

## Intra-cluster Ingress

Since generic Ingress controllers run in pods, you can deploy them as intra-cluster
proxies by just not exposing them on a `hostPort` and putting them behind a
Service of `Type=ClusterIP`.

TODO: Write an example


