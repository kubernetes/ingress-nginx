# Installation Guide

## Contents

- [Prerequisite Generic Deployment Command](#prerequisite-generic-deployment-command)
  - [Provider Specific Steps](#provider-specific-steps)
    - [Docker for Mac](#docker-for-mac)
    - [minikube](#minikube)
    - [AWS](#aws)
    - [GCE - GKE](#gce-gke)
    - [Azure](#azure)
    - [Bare-metal](#bare-metal)
  - [Verify installation](#verify-installation)
  - [Detect installed version](#detect-installed-version)
- [Using Helm](#using-helm)

## Prerequisite Generic Deployment Command

The following **Mandatory Command** is required for all deployments.

!!! attention
    The default configuration watches Ingress object from all the namespaces.
    To change this behavior use the flag `--watch-namespace` to limit the scope to a particular namespace.

!!! warning
    If multiple Ingresses define different paths for the same host, the ingress controller will merge the definitions.
    
!!! attention
    If you're using GKE you need to initialize your user as a cluster-admin with the following command: 
    ```kubectl create clusterrolebinding cluster-admin-binding   --clusterrole cluster-admin   --user $(gcloud config get-value account)```

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/mandatory.yaml
```

### Provider Specific Steps

There are cloud provider specific yaml files.

#### Docker for Mac

Kubernetes is available in Docker for Mac (from [version 18.06.0-ce](https://docs.docker.com/docker-for-mac/release-notes/#stable-releases-of-2018))

[enable]: https://docs.docker.com/docker-for-mac/#kubernetes

Create a service

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```

#### minikube

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

#### AWS

In AWS we use an Elastic Load Balancer (ELB) to expose the NGINX Ingress controller behind a Service of `Type=LoadBalancer`.
Since Kubernetes v1.9.0 it is possible to use a classic load balancer (ELB) or network load balancer (NLB)
Please check the [elastic load balancing AWS details page](https://aws.amazon.com/elasticloadbalancing/details/)

##### Elastic Load Balancer - ELB

This setup requires to choose in which layer (L4 or L7) we want to configure the ELB:

- [Layer 4](https://en.wikipedia.org/wiki/OSI_model#Layer_4:_Transport_Layer): use TCP as the listener protocol for ports 80 and 443.
- [Layer 7](https://en.wikipedia.org/wiki/OSI_model#Layer_7:_Application_Layer): use HTTP as the listener protocol for port 80 and terminate TLS in the ELB

For L4:

Check that no change is necessary with regards to the ELB idle timeout. In some scenarios, users may want to modify the ELB idle timeout, so please check the [ELB Idle Timeouts section](#elb-idle-timeouts) for additional information. If a change is required, users will need to update the value of `service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout` in `provider/aws/service-l4.yaml`

Then execute:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-l4.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l4.yaml
```

For L7:

Change line of the file `provider/aws/service-l7.yaml` replacing the dummy id with a valid one `"arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"`

Check that no change is necessary with regards to the ELB idle timeout. In some scenarios, users may want to modify the ELB idle timeout, so please check the [ELB Idle Timeouts section](#elb-idle-timeouts) for additional information. If a change is required, users will need to update the value of `service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout` in `provider/aws/service-l7.yaml`

Then execute:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-l7.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l7.yaml
```

This example creates an ELB with just two listeners, one in port 80 and another in port 443

![Listeners](../images/elb-l7-listener.png)

##### ELB Idle Timeouts
In some scenarios users will need to modify the value of the ELB idle timeout. Users need to ensure the idle timeout is less than the [keepalive_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout) that is configured for NGINX. By default NGINX `keepalive_timeout` is set to `75s`.

The default ELB idle timeout will work for most scenarios, unless the NGINX [keepalive_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout) has been modified, in which case `service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout` will need to be modified to ensure it is less than the `keepalive_timeout` the user has configured.

_Please Note: An idle timeout of `3600s` is recommended when using WebSockets._

More information with regards to idle timeouts for your Load Balancer can be found in the [official AWS documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

##### Network Load Balancer (NLB)

This type of load balancer is supported since v1.10.0 as an ALPHA feature.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-nlb.yaml
```

#### GCE - GKE

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```

**Important Note:** proxy protocol is not supported in GCE/GKE

#### Azure


```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/cloud-generic.yaml
```


#### Bare-metal

Using [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport):

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/baremetal/service-nodeport.yaml
```

!!! tip
    For extended notes regarding deployments on bare-metal, see [Bare-metal considerations](./baremetal.md).

### Verify installation

To check if the ingress controller pods have started, run the following command:

```console
kubectl get pods --all-namespaces -l app.kubernetes.io/name=ingress-nginx --watch
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.
Now, you are ready to create your first ingress.

### Detect installed version

To detect which version of the ingress controller is running, exec into the pod and run `nginx-ingress-controller version` command.

```console
POD_NAMESPACE=ingress-nginx
POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/name=ingress-nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD_NAME -n $POD_NAMESPACE -- /nginx-ingress-controller --version
```

## Using Helm

NGINX Ingress controller can be installed via [Helm](https://helm.sh/) using the chart [stable/nginx-ingress](https://github.com/kubernetes/charts/tree/master/stable/nginx-ingress) from the official charts repository. 
To install the chart with the release name `my-nginx`:

```console
helm install stable/nginx-ingress --name my-nginx
```

If the kubernetes cluster has RBAC enabled, then run:

```console
helm install stable/nginx-ingress --name my-nginx --set rbac.create=true
```

Detect installed version:

```console
POD_NAME=$(kubectl get pods -l app.kubernetes.io/name=ingress-nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD_NAME -- /nginx-ingress-controller --version
```

