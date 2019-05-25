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
    These commands depend on having kubectl version 1.14 or newer.

!!! attention
    The default configuration watches Ingress object from all the namespaces.
    To change this behavior use the flag `--watch-namespace` to limit the scope to a particular namespace.

!!! warning
    If multiple Ingresses define different paths for the same host, the ingress controller will merge the definitions.
    

```console
kubectl create namespace ingress-nginx
```

```console
cat << EOF > kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: ingress-nginx
bases:
- github.com/kubernetes/ingress-nginx/deploy/cluster-wide
- # provider-specific, see below
EOF
```

### Provider Specific Steps

There are cloud provider specific kustomize bases.

#### Docker for Mac

Kubernetes is available in Docker for Mac (from [version 18.06.0-ce](https://docs.docker.com/docker-for-mac/release-notes/#stable-releases-of-2018))

[enable]: https://docs.docker.com/docker-for-mac/#kubernetes

Add `github.com/kubernetes/ingress-nginx/deploy/cloud-generic` to the `bases` list in `kustomization.yaml` and run `kubectl apply --kustomize .`.

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


Check that no change is necessary with regards to the ELB idle timeout. In some scenarios, users may want to modify the ELB idle timeout, so please check the [ELB Idle Timeouts section](#elb-idle-timeouts) for additional information. If a change is required, users will need to override the value of the annotation `service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout` on the service object.

To do this, create a patch file which will replace the annotation.

```
cat << EOF > elb-timeout.yaml
kind: Service
apiVersion: v1
metadata:
  name: ingress-nginx
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout: "3600" # Recommended value for WebSockets
EOF
```

After creating the patch file, reference it in your `kustomization.yaml`:
```yaml
patchesStrategicMerge:
- elb-timeout.yaml
```

For L4:

To deploy the default example, add the base ` github.com/kubernetes/ingress-nginx/deploy/aws/l4` and then run `kubectl apply --kustomize .`

For L7:

Create a a patch that will annotate the ingress-controller's service with your ssl certificate id.

```console
cat << EOF > elb-ssl.yaml
kind: Service
apiVersion: v1
metadata:
  name: ingress-nginx
  annotations:
    # replace with the correct value of the generated certificate in the AWS console
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"
EOF
```

Reference this patch in your `kustomization.yaml`:

```yaml
patchesStrategicMerge:
- elb-ssl.yaml
```

Then add the l7 base, `github.com/kubernetes/ingress-nginx/deploy/aws/l7` and execute `kubectl apply --kustomize .`

This example creates an ELB with just two listeners, one in port 80 and another in port 443

![Listeners](../images/elb-l7-listener.png)

##### ELB Idle Timeouts
In some scenarios users will need to modify the value of the ELB idle timeout. Users need to ensure the idle timeout is less than the [keepalive_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout) that is configured for NGINX. By default NGINX `keepalive_timeout` is set to `75s`.

The default ELB idle timeout will work for most scenarios, unless the NGINX [keepalive_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout) has been modified, in which case `service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout` will need to be modified to ensure it is less than the `keepalive_timeout` the user has configured.

_Please Note: An idle timeout of `3600s` is recommended when using WebSockets._

More information with regards to idle timeouts for your Load Balancer can be found in the [official AWS documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

##### Network Load Balancer (NLB)

This type of load balancer is supported since v1.10.0 as an ALPHA feature.  Use the base `github.com/kubernetes/ingress-nginx/deploy/aws/nlb` and execute `kubectl apply --kustomize .`


#### GCE-GKE

!!! attention
    If you're using GKE you need to initialize your user as a cluster-admin with the following command: 
    ```kubectl create clusterrolebinding cluster-admin-binding   --clusterrole cluster-admin   --user $(gcloud config get-value account)```

Use the base `github.com/kubernetes/ingress-nginx/deploy/cloud-generic` and execute `kubectl apply --kustomize .`

**Important Note:** proxy protocol is not supported in GCE/GKE


#### Azure

Use the base `github.com/kubernetes/ingress-nginx/deploy/cloud-generic` and execute `kubectl apply --kustomize .`


#### Bare-metal

Using [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport):


Use the base `github.com/kubernetes/ingress-nginx/deploy/baremetal` and execute `kubectl apply --kustomize .`

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

