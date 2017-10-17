# Installation Guide

## Contents

- [Mandatory commands](#mandatory-commands)
- [Install without RBAC roles](#install-without-rbac-roles)
- [Install with RBAC roles](#install-with-rbac-roles)
- [Custom Provider](#custom-provider)
  - [minikube](#minikube)
  - [AWS](#aws)
  - [GCE - GKE](#gce-gke)
  - [Azure](#azure)
  - [Baremetal](#baremetal)
- [Using Helm](#using-helm)
- [Verify installation](#verify-installation)
- [Detect installed version](#detect-installed-version)

## Mandatory commands

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/namespace.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/default-backend.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/configmap.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/tcp-services-configmap.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/udp-services-configmap.yaml \
    | kubectl apply -f -
```

## Install without RBAC roles

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/without-rbac.yaml \
    | kubectl apply -f -
```

## Install with RBAC roles

Please check the [RBAC](rbac.md) document.

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/rbac.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/with-rbac.yaml \
    | kubectl apply -f -
```

## Custom Service provider

There are cloud provider specific yaml files

### minikube

```console
minikube addons enable ingress
```

### AWS

In AWS we use an Elastic Load Balancer (ELB) to expose the NGINX Ingress controller behind a Service of `Type=LoadBalancer`.
This setup requires to choose in which layer (L4 or L7) we want to configure the ELB:

- [Layer 4](https://en.wikipedia.org/wiki/OSI_model#Layer_4:_Transport_Layer): use TCP as the listener protocol for ports 80 and 443.
- [Layer 7](https://en.wikipedia.org/wiki/OSI_model#Layer_7:_Application_Layer): use HTTP as the listener protocol for port 80 and terminate TLS in the ELB

For L4:

```console
kubectl apply -f provider/aws/service-l4.yaml
kubectl apply -f provider/aws/patch-configmap-l4.yaml
```

For L7:

Change line of the file `provider/aws/service-l7.yaml` replacing the dummy id with a valid one `"arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"`
Then execute:

```console
kubectl apply -f provider/aws/service-l7.yaml
kubectl apply -f provider/aws/patch-configmap-l7.yaml
```

This example creates an ELB with just two listeners, one in port 80 and another in port 443

![Listeners](../docs/images/listener.png)

If the ingress controller uses RBAC run:

```console
kubectl apply -f provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f provider/patch-service-without-rbac.yaml
```

### GCE - GKE

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/gce-gke/service.yaml \
    | kubectl apply -f -
```

If the ingress controller uses RBAC run:

```console
kubectl apply -f provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f provider/patch-service-without-rbac.yaml
```

**Important Note:** proxy protocol is not supported in GCE/GKE

### Azure

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/azure/service.yaml \
    | kubectl apply -f -
```

If the ingress controller uses RBAC run:

```console
kubectl apply -f provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f provider/patch-service-without-rbac.yaml
```

**Important Note:** proxy protocol is not supported in GCE/GKE

### Baremetal

Using [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport):

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/baremetal/service-nodeport.yaml \
    | kubectl apply -f -
```

## Using Helm

NGINX Ingress controller can be installed via [Helm](https://helm.sh/) using the chart [stable/nginx](https://github.com/kubernetes/charts/tree/master/stable/nginx-ingress) from the official charts repository. 
To install the chart with the release name `my-nginx`:

```console
helm install stable/nginx-ingress --name my-nginx
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
kubectl exec -it $POD_NAME -n $POD_NAMESPACE /nginx-ingress-controller version
```
