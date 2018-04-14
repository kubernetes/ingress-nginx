# Installation Guide

## Contents

- [Mandatory commands](#mandatory-commands)
- [Install without RBAC roles](#install-without-rbac-roles)
- [Install with RBAC roles](#install-with-rbac-roles)
- [Custom Provider](#custom-provider)
  - [minikube](#minikube)
  - [AWS](#aws)
  - [GCE - GKE](#gce---gke)
  - [Azure](#azure)
  - [Baremetal](#baremetal)
- [Using Helm](#using-helm)
- [Verify installation](#verify-installation)
- [Detect installed version](#detect-installed-version)
- [Deploying the config-map](#deploying-the-config-map)

## Generic Deployment 

The following resources are required for a generic deployment.

### Mandatory commands

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

### Install without RBAC roles

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/without-rbac.yaml \
    | kubectl apply -f -
```

### Install with RBAC roles

Please check the [RBAC](rbac.md) document.

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/rbac.yaml \
    | kubectl apply -f -

curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/with-rbac.yaml \
    | kubectl apply -f -
```

## Custom Service Provider Deployment

There are cloud provider specific yaml files.

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

2. Use the [docker daemon](https://github.com/kubernetes/minikube/blob/master/docs/reusing_the_docker_daemon.md)
3. [Build the image](../docs/development.md)
4. Perform [Mandatory commands](#mandatory-commands)
5. Install the `nginx-ingress-controller` deployment [without RBAC roles](#install-without-rbac-roles) or [with RBAC roles](#install-with-rbac-roles)
6. Edit the `nginx-ingress-controller` deployment to use your custom image. Local images can be seen by performing `docker images`.

```console
$ kubectl edit deployment nginx-ingress-controller -n ingress-nginx
```

edit the following section:

```yaml
image: <IMAGE-NAME>:<TAG>
imagePullPolicy: IfNotPresent
name: nginx-ingress-controller
```

7. Confirm the `nginx-ingress-controller` deployment exists:

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

Patch the nginx ingress controller deployment to add the flag `--publish-service`

```console
kubectl patch deployment -n ingress-nginx nginx-ingress-controller --type='json' \
  --patch="$(curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/publish-service-patch.yaml)"
```

For L4:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-l4.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l4.yaml
```

For L7:

Change line of the file `provider/aws/service-l7.yaml` replacing the dummy id with a valid one `"arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX"`
Then execute:

```console
kubectl apply -f provider/aws/service-l7.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/patch-configmap-l7.yaml
```

This example creates an ELB with just two listeners, one in port 80 and another in port 443

![Listeners](../docs/images/elb-l7-listener.png)

If the ingress controller uses RBAC run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-without-rbac.yaml
```

#### Network Load Balancer (NLB)

This type of load balancer is supported since v1.10.0 as an ALPHA feature.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/aws/service-nlb.yaml
```

If the ingress controller uses RBAC run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-without-rbac.yaml
```

### GCE - GKE

Patch the nginx ingress controller deployment to add the flag `--publish-service`

```console
kubectl patch deployment -n ingress-nginx nginx-ingress-controller --type='json' \
  --patch="$(curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/publish-service-patch.yaml)"
```

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/gce-gke/service.yaml \
    | kubectl apply -f -
```

If the ingress controller uses RBAC run:

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-with-rbac.yaml | kubectl apply -f -
```

If not run:

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-without-rbac.yaml | kubectl apply -f -
```

**Important Note:** proxy protocol is not supported in GCE/GKE

### Azure

Patch the nginx ingress controller deployment to add the flag `--publish-service`

```console
kubectl patch deployment -n ingress-nginx nginx-ingress-controller --type='json' \
  --patch="$(curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/publish-service-patch.yaml)"
```

```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/azure/service.yaml \
    | kubectl apply -f -
```

If the ingress controller uses RBAC run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-with-rbac.yaml
```

If not run:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/provider/patch-service-without-rbac.yaml
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

## Deploying the config-map

A config map can be used to configure system components for the nginx-controller. In order to begin using a config-map
make sure it has been created and is being used in the deployment.

It is created as seen in the [Mandatory Commands](#mandatory-commands) section above.
```console
curl https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/configmap.yaml \
    | kubectl apply -f -
```

and is setup to be used in the deployment [without-rbac](without-rbac.yaml) or [with-rbac](with-rbac.yaml) with the following line:
```yaml
- --configmap=$(POD_NAMESPACE)/nginx-configuration
```

For information on using the config-map, see its [user-guide](../docs/user-guide/configmap.md).
