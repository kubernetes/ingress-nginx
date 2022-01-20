# Installation Guide

There are multiple ways to install the NGINX ingress controller:

- with [Helm](https://helm.sh), using the project repository chart;
- with `kubectl apply`, using YAML manifests;
- with specific addons (e.g. for [minikube](#minikube) or [MicroK8s](#microk8s)).

On most Kubernetes clusters, the ingress controller will work without requiring any extra configuration. If you want to get started as fast as possible, you can check the [quick start](#quick-start) instructions. However, in many environments, you can improve the performance or get better logs by enabling extra features. we recommend that you check the [environment-specific instructions](#environment-specific-instructions) for details about optimizing the ingress controller for your particular environment or cloud provider.

## Contents

<!-- Quick tip: run `grep '^##' index.md` to check that the table of contents is up to date. -->

- [Quick start](#quick-start)

- [Environment-specific instructions](#environment-specific-instructions)
  - ... [Docker Desktop](#docker-desktop)
  - ... [minikube](#minikube)
  - ... [MicroK8s](#microk8s)
  - ... [AWS](#aws)
  - ... [GCE - GKE](#gce-gke)
  - ... [Azure](#azure)
  - ... [Digital Ocean](#digital-ocean)
  - ... [Scaleway](#scaleway)
  - ... [Exoscale](#exoscale)
  - ... [Oracle Cloud Infrastructure](#oracle-cloud-infrastructure)
  - ... [Bare-metal](#bare-metal-clusters)
- [Miscellaneous](#miscellaneous)

## Quick start

**If you have Helm,** you can deploy the ingress controller with the following command:

```console
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace
```

It will install the controller in the `ingress-nginx` namespace, creating that namespace if it doesn't already exist.

!!! info
    This command is *idempotent*:

    - if the ingress controller is not installed, it will install it,
    - if the ingress controller is already installed, it will upgrade it.

**If you don't have Helm** or if you prefer to use a YAML manifest, you can run the following command instead:

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/cloud/deploy.yaml
```

!!! info
    The YAML manifest in the command above was generated with `helm template`, so you will end up with almost the same resources as if you had used Helm to install the controller.

!!! attention
  If you are running an old version of Kubernetes (1.18 or earlier), please read
  [this paragraph](#running-on-Kubernetes-versions-older-than-1.19) for specific instructions.
  Because of api deprecations, the default manifest may not work on your cluster.
  Specific manifests for supported Kubernetes versions are available within a subfolder of each provider.

### Pre-flight check

A few pods should start in the `ingress-nginx` namespace:

```console
kubectl get pods --namespace=ingress-nginx
```

After a while, they should all be running. The following command will wait for the ingress controller pod to be up, running, and ready:

```console
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s
```

### Local testing

Let's create a simple web server and the associated service:

```console
kubectl create deployment demo --image=httpd --port=80
kubectl expose deployment demo
```

Then create an ingress resource. The following example uses an host that maps to `localhost`:

```console
kubectl create ingress demo-localhost --class=nginx \
  --rule=demo.localdev.me/*=demo:80
```

Now, forward a local port to the ingress controller:

```console
kubectl port-forward --namespace=ingress-nginx service/ingress-nginx-controller 8080:80
```

At this point, if you access http://demo.localdev.me:8080/, you should see an HTML page telling you "It works!".

### Online testing

If your Kubernetes cluster is a "real" cluster that supports services of type `LoadBalancer`, it will have allocated an external IP address or FQDN to the ingress controller.

You can see that IP address or FQDN with the following command:

```console
kubectl get service ingress-nginx-controller --namespace=ingress-nginx
```

It will be the `EXTERNAL-IP` field. If that field shows `<pending>`, this means that your Kubernetes cluster wasn't able to provision the load balancer (generally, this is because it doesn't support services of type `LoadBalancer`).

Once you have the external IP address (or FQDN), set up a DNS record pointing to it. Then you can create an ingress resource. The following example assumes that you have set up a DNS record for `www.demo.io`:

```console
kubectl create ingress demo --class=nginx \
  --rule="www.demo.io/*=demo:80"
```

Alternatively, the above command can be rewritten as follows for the ```--rule``` command and below.
```console
kubectl create ingress demo --class=nginx \
  --rule www.demo.io/=demo:80
```


You should then be able to see the "It works!" page when you connect to http://www.demo.io/. Congratulations, you are serving a public web site hosted on a Kubernetes cluster! 🎉

## Environment-specific instructions

### Local development clusters

#### minikube

The ingress controller can be installed through minikube's addons system:

```console
minikube addons enable ingress
```

#### MicroK8s

The ingress controller can be installed through MicroK8s's addons system:

```console
microk8s enable ingress
```

Please check the MicroK8s [documentation page](https://microk8s.io/docs/addon-ingress) for details.

#### Docker Desktop

Kubernetes is available in Docker Desktop:

- Mac, from [version 18.06.0-ce](https://docs.docker.com/docker-for-mac/release-notes/#stable-releases-of-2018)
- Windows, from [version 18.06.0-ce](https://docs.docker.com/docker-for-windows/release-notes/#docker-community-edition-18060-ce-win70-2018-07-25)

First, make sure that Kubernetes is enabled in the Docker settings. The command `kubectl get nodes` should show a single node called `docker-desktop`.

The ingress controller can be installed on Docker Desktop using the default [quick start](#quick-start) instructions.

On most systems, if you don't have any other service of type `LoadBalancer` bound to port 80, the ingress controller will be assigned the `EXTERNAL-IP` of `localhost`, which means that it will be reachable on localhost:80. If that doesn't work, you might have to fall back to the `kubectl port-forward` method described in the [local testing section](#local-testing).

### Cloud deployments

If the load balancers of your cloud provider do active healthchecks on their backends (most do), you can change the `externalTrafficPolicy` of the ingress controller Service to `Local` (instead of the default `Cluster`) to save an extra hop in some cases. If you're installing with Helm, this can be done by adding `--set controller.service.externalTrafficPolicy=Local` to the `helm install` or `helm upgrade` command.

Furthermore, if the load balancers of your cloud provider support the PROXY protocol, you can enable it, and it will let the ingress controller see the real IP address of the clients. Otherwise, it will generally see the IP address of the upstream load balancer. This must be done both in the ingress controller (with e.g. `--set controller.config.use-proxy-protocol=true`) and in the cloud provider's load balancer configuration to function correctly.

In the following sections, we provide YAML manifests that enable these options when possible, using the specific options of various cloud providers.

#### AWS

In AWS we use a Network load balancer (NLB) to expose the NGINX Ingress controller behind a Service of `Type=LoadBalancer`.

!!! info
    The provided templates illustrate the setup for legacy in-tree service load balancer for AWS NLB.
    AWS provides the documentation on how to use [Network load balancing on Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/network-load-balancing.html) with [AWS Load Balancer Controller](https://github.com/kubernetes-sigs/aws-load-balancer-controller).

##### Network Load Balancer (NLB)

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/aws/deploy.yaml
```

##### TLS termination in AWS Load Balancer (NLB)

By default, TLS is terminated in the ingress controller. But it is also possible to terminate TLS in the Load Balancer. This section explains how to do that on AWS using an NLB.

1. Download the [deploy.yaml](https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/aws/nlb-with-tls-termination/deploy.yaml) template

  ```console
  wget https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/aws/nlb-with-tls-termination/deploy.yaml
  ```

2. Edit the file and change the VPC CIDR in use for the Kubernetes cluster:
   ```
   proxy-real-ip-cidr: XXX.XXX.XXX/XX
   ```

3. Change the AWS Certificate Manager (ACM) ID as well:
   ```
   arn:aws:acm:us-west-2:XXXXXXXX:certificate/XXXXXX-XXXXXXX-XXXXXXX-XXXXXXXX
   ```

4. Deploy the manifest:
   ```console
   kubectl apply -f deploy.yaml
   ```

##### NLB Idle Timeouts

Idle timeout value for TCP flows is 350 seconds and [cannot be modified](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/network-load-balancers.html#connection-idle-timeout).

For this reason, you need to ensure the [keepalive_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout) value is configured less than 350 seconds to work as expected.

By default NGINX `keepalive_timeout` is set to `75s`.

More information with regards to timeouts can be found in the [official AWS documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/network-load-balancers.html#connection-idle-timeout)

#### GCE-GKE

First, your user needs to have `cluster-admin` permissions on the cluster. This can be done with the following command:

```console
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin \
  --user $(gcloud config get-value account)
```

Then, the ingress controller can be installed like this:


```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/cloud/deploy.yaml
```

!!! warning
    For private clusters, you will need to either add an additional firewall rule that allows master nodes access to port `8443/tcp` on worker nodes, or change the existing rule that allows access to ports `80/tcp`, `443/tcp` and `10254/tcp` to also allow access to port `8443/tcp`.

    See the [GKE documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters#add_firewall_rules) on adding rules and the [Kubernetes issue](https://github.com/kubernetes/kubernetes/issues/79739) for more detail.

!!! warning
    Proxy protocol is not supported in GCE/GKE.

#### Azure

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/cloud/deploy.yaml
```

More information with regards to Azure annotations for ingress controller can be found in the [official AKS documentation](https://docs.microsoft.com/en-us/azure/aks/ingress-internal-ip#create-an-ingress-controller).

#### Digital Ocean

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/do/deploy.yaml
```

#### Scaleway

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/scw/deploy.yaml
```

#### Exoscale

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/exoscale/deploy.yaml
```

The full list of annotations supported by Exoscale is available in the Exoscale Cloud Controller Manager [documentation](https://github.com/exoscale/exoscale-cloud-controller-manager/blob/master/docs/service-loadbalancer.md).

#### Oracle Cloud Infrastructure

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/cloud/deploy.yaml
```

A [complete list of available annotations for Oracle Cloud Infrastructure](https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/load-balancer-annotations.md) can be found in the [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager) documentation.

### Bare metal clusters

This section is applicable to Kubernetes clusters deployed on bare metal servers, as well as "raw" VMs where Kubernetes was installed manually, using generic Linux distros (like CentOS, Ubuntu...)

For quick testing, you can use a [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport). This should work on almost every cluster, but it will typically use a port in the range 30000-32767.

```console
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/baremetal/deploy.yaml
```

For more information about bare metal deployments (and how to use port 80 instead of a random port in the 30000-32767 range), see [bare-metal considerations](./baremetal.md).

## Miscellaneous

### Checking ingress controller version

Run `/nginx-ingress-controller --version` within the pod, for instance with `kubectl exec`:

```console
POD_NAMESPACE=ingress-nginx
POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/name=ingress-nginx --field-selector=status.phase=Running -o name)
kubectl exec $POD_NAME -n $POD_NAMESPACE -- /nginx-ingress-controller --version
```

### Scope

By default, the controller watches Ingress objects from all namespaces. If you want to change this behavior, use the flag `--watch-namespace` or check the Helm chart value `controller.scope` to limit the controller to a single namespace.

See also [“How to easily install multiple instances of the Ingress NGINX controller in the same cluster”](https://kubernetes.github.io/ingress-nginx/#how-to-easily-install-multiple-instances-of-the-ingress-nginx-controller-in-the-same-cluster) for more details.

### Webhook network access

!!! warning
    The controller uses an [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) to validate Ingress definitions. Make sure that you don't have [Network policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) or additional firewalls preventing connections from the API server to the `ingress-nginx-controller-admission` service.

### Certificate generation

!!! attention
    The first time the ingress controller starts, two [Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) create the SSL Certificate used by the admission webhook.

THis can cause an initial delay of up to two minutes until it is possible to create and validate Ingress definitions.

You can wait until it is ready to run the next command:

```yaml
 kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s
```

### Running on Kubernetes versions older than 1.19

Ingress resources evolved over time. They started with `apiVersion: extensions/v1beta1`, then moved to `apiVersion: networking.k8s.io/v1beta1` and more recently to `apiVersion: networking.k8s.io/v1`.

Here is how these Ingress versions are supported in Kubernetes:
- before Kubernetes 1.19, only `v1beta1` Ingress resources are supported
- from Kubernetes 1.19 to 1.21, both `v1beta1` and `v1` Ingress resources are supported
- in Kubernetes 1.22 and above, only `v1` Ingress resources are supported

And here is how these Ingress versions are supported in NGINX Ingress Controller:
- before version 1.0, only `v1beta1` Ingress resources are supported
- in version 1.0 and above, only `v1` Ingress resources are

As a result, if you're running Kubernetes 1.19 or later, you should be able to use the latest version of the NGINX Ingress Controller; but if you're using an old version of Kubernetes (1.18 or earlier) you will have to use version 0.X of the NGINX Ingress Controller (e.g. version 0.49).

The Helm chart of the NGINX Ingress Controller switched to version 1 in version 4 of the chart. In other words, if you're running Kubernetes 1.19 or earlier, you should use version 3.X of the chart (this can be done by adding `--version='<4'` to the `helm install` command).
