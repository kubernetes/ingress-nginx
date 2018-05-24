# Installation Guide

## Contents

- [Mandatory command](#mandatory-command)
- [Custom Provider](#custom-provider)
  - [Docker for Mac](#docker-for-mac)
  - [minikube](#minikube)
  - [AWS](#aws)
  - [GCE - GKE](#gce---gke)
  - [Azure](#azure)
  - [Baremetal](#baremetal)
- [Using Helm](#using-helm)
- [Verify installation](#verify-installation)
- [Detect installed version](#detect-installed-version)

## Generic Deployment 

The following resources are required for a generic deployment.

### Mandatory command

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml
```

## Custom Service Provider Deployment

There are cloud provider specific yaml files.

### Docker for Mac

Kubernetes is available for Docker for Mac's Edge channel. Switch to the [Edge
channel][edge] and [enable Kubernetes][enable].

[edge]: https://docs.docker.com/docker-for-mac/install/
[enable]: https://docs.docker.com/docker-for-mac/#kubernetes

Create a service

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```

### minikube

For standard usage:

```console
minikube addons enable ingress
```

For development:

1. Disable the ingress addon:

```console
$ minikube addons disable ingress
```

2. Execute `make dev-env`
3. Confirm the `nginx-ingress-controller` deployment exists:

```console
$ kubectl get pods -n ingress-nginx 
NAME                                       READY     STATUS    RESTARTS   AGE
default-http-backend-66b447d9cf-rrlf9      1/1       Running   0          12s
nginx-ingress-controller-fdcdcd6dd-vvpgs   1/1       Running   0          11s
```

### AWS

In AWS we use an Elastic Load Balancer (ELB) to expose the NGINX Ingress controller behind a Service of `Type=LoadBalancer`.
Since Kubernetes v1.9.0 it is possible to use a classic load balancer (ELB) or network load balancer (NLB)
Please check the [elastic load balancing AWS details page](https://aws.amazon.com/es/elasticloadbalancing/details/)

#### Elastic Load Balancer - ELB

This setup requires to choose in which layer (L4 or L7) we want to configure the ELB:

- [Layer 4](https://en.wikipedia.org/wiki/OSI_model#Layer_4:_Transport_Layer): use TCP as the listener protocol for ports 80 and 443.
- [Layer 7](https://en.wikipedia.org/wiki/OSI_model#Layer_7:_Application_Layer): use HTTP as the listener protocol for port 80 and terminate TLS in the ELB

For L4:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-l4.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l4.yaml
```

For L7:

Change line of the file `provider/aws/service-l7.yaml` replacing the dummy id with a valid one `"arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"`
Then execute:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-l7.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l7.yaml
```

This example creates an ELB with just two listeners, one in port 80 and another in port 443

![Listeners](../images/elb-l7-listener.png)

#### Network Load Balancer (NLB)

This type of load balancer is supported since v1.10.0 as an ALPHA feature.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-nlb.yaml
```

### GCE - GKE

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```

**Important Note:** proxy protocol is not supported in GCE/GKE

### Azure


```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```

**Important Note:** proxy protocol is not supported in GCE/GKE

### Baremetal

Using [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport):

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/baremetal/service-nodeport.yaml
```

## Using Helm

NGINX Ingress controller can be installed via [Helm](https://helm.sh/) using the chart [stable/nginx](https://github.com/kubernetes/charts/tree/master/stable/nginx-ingress) from the official charts repository. 
To install the chart with the release name `my-nginx`:

```console
helm install stable/nginx-ingress --name my-nginx
```

If the kubernetes cluster has RBAC enabled, then run:

```console
helm install stable/nginx-ingress --name my-nginx --set rbac.create=true
```

## Verify installation

To check if the ingress controller pods have started, run the following command:

```console
kubectl get pods --all-namespaces -l app=ingress-nginx --watch
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.
Now, you are ready to create your first ingress.

## Detect installed version

To detect which version of the ingress controller is running, exec into the pod and run `nginx-ingress-controller version` command.

```console
POD_NAMESPACE=ingress-nginx
POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app=ingress-nginx -o jsonpath={.items[0].metadata.name})
kubectl exec -it $POD_NAME -n $POD_NAMESPACE -- /nginx-ingress-controller --version
```
