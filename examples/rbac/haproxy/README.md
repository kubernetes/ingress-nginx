# Role Based Access Control

This example demonstrates how to authorize an ingress controller on a cluster
with role based access control.

## Overview

This example applies to ingress controllers being deployed in an environment with
[RBAC](https://kubernetes.io/docs/admin/authorization/rbac/) enabled.

## Service Account created in this example

One ServiceAccount is created in this example, `ingress-controller`. See
[Using cert based authentication](#using-cert-based-authentication)
below if using client cert authentication.

## Permissions Granted in this example

There are two sets of permissions defined in this example.  Cluster-wide
permissions defined by a `ClusterRole` and namespace specific permissions
defined by a `Role`, both named `ingress-controller`.

### Cluster Permissions

These permissions are granted in order for the ingress-controller to be
able to function as an ingress across the cluster. These permissions are
granted to the ClusterRole:

* `configmaps`, `endpoints`, `nodes`, `pods`, `secrets`: list, watch
* `nodes`: get
* `services`, `ingresses`: get, list, watch
* `events`: create, patch
* `ingresses/status`: update

### Namespace Permissions

These permissions are granted specific to the `ingress-controller` namespace.
The Role permissions are:

* `configmaps`, `pods`, `secrets`: get
* `endpoints`: create, get, update

Furthermore to support leader-election, the ingress controller needs to
have access to a `configmap` in the `ingress-controller` namespace:

* `configmaps`: get, update, create

## Namespace created in this example

The `Namespace` named `ingress-controller` is defined in this example. The
namespace name can be changed arbitrarily as long as all of the references
change as well.

## Usage

1. Create the `Namespace`, `Service Account`, `ClusterRole`, `Role`,
`ClusterRoleBinding`, and `RoleBinding`:

```console
$ kubectl create -f ingress-controller-rbac.yml
```

2. Deploy the ingress controller. The deployment should be configured to use
the `ingress-controller` service account name if not using kubeconfig and
client cert based authentication. Add the `serviceAccountName` to the pod
template spec:

```yaml
spec:
  template:
    spec:
      serviceAccountName: ingress-controller
```

## Using cert based authentication

A client certificate based authentication can also be used with the following changes:

1. No need to add the `serviceAccountName` to the pod template spec.
2. Sign a client certificate using `ingress-controller` as it's common name.
