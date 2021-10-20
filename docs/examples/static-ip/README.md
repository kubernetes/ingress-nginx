# Static IPs

This example demonstrates how to assign a static-ip to an Ingress on through the Ingress-NGINX controller.

## Prerequisites

You need a [TLS cert](../PREREQUISITES.md#tls-certificates) and a [test HTTP service](../PREREQUISITES.md#test-http-service) for this example.
You will also need to make sure your Ingress targets exactly one Ingress
controller by specifying the [ingress.class annotation](../../user-guide/multiple-ingress.md),
and that you have an ingress controller [running](../../deploy/) in your cluster.

## Acquiring an IP

Since instances of the ingress nginx controller actually run on nodes in your cluster,
by default nginx Ingresses will only get static IPs if your cloudprovider
supports static IP assignments to nodes. On GKE/GCE for example, even though
nodes get static IPs, the IPs are not retained across upgrades.

To acquire a static IP for the ingress-nginx-controller, simply put it
behind a Service of `Type=LoadBalancer`.

First, create a loadbalancer Service and wait for it to acquire an IP:

```console
$ kubectl create -f static-ip-svc.yaml
service "ingress-nginx-lb" created

$ kubectl get svc ingress-nginx-lb
NAME               CLUSTER-IP     EXTERNAL-IP       PORT(S)                      AGE
ingress-nginx-lb   10.0.138.113   104.154.109.191   80:31457/TCP,443:32240/TCP   15m
```

Then, update the ingress controller so it adopts the static IP of the Service
by passing the `--publish-service` flag (the example yaml used in the next step
already has it set to "ingress-nginx-lb").

```console
$ kubectl create -f ingress-nginx-controller.yaml
deployment "ingress-nginx-controller" created
```

## Assigning the IP to an Ingress

From here on every Ingress created with the `ingress.class` annotation set to
`nginx` will get the IP allocated in the previous step.

```console
$ kubectl create -f ingress-nginx.yaml
ingress "ingress-nginx" created

$ kubectl get ing ingress-nginx
NAME            HOSTS     ADDRESS           PORTS     AGE
ingress-nginx   *         104.154.109.191   80, 443   13m

$ curl 104.154.109.191 -kL
CLIENT VALUES:
client_address=10.180.1.25
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://104.154.109.191:8080/
...
```

## Retaining the IP

You can test retention by deleting the Ingress:

```console
$ kubectl delete ing ingress-nginx
ingress "ingress-nginx" deleted

$ kubectl create -f ingress-nginx.yaml
ingress "ingress-nginx" created

$ kubectl get ing ingress-nginx
NAME            HOSTS     ADDRESS           PORTS     AGE
ingress-nginx   *         104.154.109.191   80, 443   13m
```

> Note that unlike the GCE Ingress, the same loadbalancer IP is shared amongst all
> Ingresses, because all requests are proxied through the same set of nginx
> controllers.

## Promote ephemeral to static IP

To promote the allocated IP to static, you can update the Service manifest:

```console
$ kubectl patch svc ingress-nginx-lb -p '{"spec": {"loadBalancerIP": "104.154.109.191"}}'
"ingress-nginx-lb" patched
```

... and promote the IP to static (promotion works differently for cloudproviders,
provided example is for GKE/GCE):

```console
$ gcloud compute addresses create ingress-nginx-lb --addresses 104.154.109.191 --region us-central1
Created [https://www.googleapis.com/compute/v1/projects/kubernetesdev/regions/us-central1/addresses/ingress-nginx-lb].
---
address: 104.154.109.191
creationTimestamp: '2017-01-31T16:34:50.089-08:00'
description: ''
id: '5208037144487826373'
kind: compute#address
name: ingress-nginx-lb
region: us-central1
selfLink: https://www.googleapis.com/compute/v1/projects/kubernetesdev/regions/us-central1/addresses/ingress-nginx-lb
status: IN_USE
users:
- us-central1/forwardingRules/a09f6913ae80e11e6a8c542010af0000
```

Now even if the Service is deleted, the IP will persist, so you can recreate the
Service with `spec.loadBalancerIP` set to `104.154.109.191`.
