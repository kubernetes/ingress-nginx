# Role Based Access Control (RBAC)

## Overview

This example applies to ingress-nginx-controllers being deployed in an environment with RBAC enabled.

Role Based Access Control is comprised of four layers:

1. `ClusterRole` - permissions assigned to a role that apply to an entire cluster
2. `ClusterRoleBinding` - binding a ClusterRole to a specific account
3. `Role` - permissions assigned to a role that apply to a specific namespace
4. `RoleBinding` - binding a Role to a specific account

In order for RBAC to be applied to an ingress-nginx-controller, that controller
should be assigned to a `ServiceAccount`.  That `ServiceAccount` should be
bound to the `Role`s and `ClusterRole`s defined for the ingress-nginx-controller.

## Service Accounts created in this example

One ServiceAccount is created in this example, `ingress-nginx`.

## Permissions Granted in this example

There are two sets of permissions defined in this example.  Cluster-wide
permissions defined by the `ClusterRole` named `ingress-nginx`, and
namespace specific permissions defined by the `Role` named `ingress-nginx`.

### Cluster Permissions

These permissions are granted in order for the ingress-nginx-controller to be
able to function as an ingress across the cluster.  These permissions are
granted to the ClusterRole named `ingress-nginx`

* `configmaps`, `endpoints`, `nodes`, `pods`, `secrets`: list, watch
* `nodes`: get
* `services`, `ingresses`: get, list, watch
* `events`: create, patch
* `ingresses/status`: update

### Namespace Permissions

These permissions are granted specific to the ingress-nginx namespace.  These
permissions are granted to the Role named `ingress-nginx`

* `configmaps`, `pods`, `secrets`: get
* `endpoints`: get

Furthermore to support leader-election, the ingress-nginx-controller needs to
have access to a `configmap` using the resourceName `ingress-controller-leader-nginx`

> Note that resourceNames can NOT be used to limit requests using the “create”
> verb because authorizers only have access to information that can be obtained
> from the request URL, method, and headers (resource names in a “create” request
> are part of the request body).

* `configmaps`: get, update (for resourceName `ingress-controller-leader-nginx`)
* `configmaps`: create

This resourceName is the concatenation of the `election-id` and the
`ingress-class` as defined by the ingress-controller, which defaults to:

* `election-id`: `ingress-controller-leader`
* `ingress-class`: `nginx`
* `resourceName` : `<election-id>-<ingress-class>`

Please adapt accordingly if you overwrite either parameter when launching the
ingress-nginx-controller.

### Bindings

The ServiceAccount `ingress-nginx` is bound to the Role
`ingress-nginx` and the ClusterRole `ingress-nginx`.

The serviceAccountName associated with the containers in the deployment must
match the serviceAccount. The namespace references in the Deployment metadata, 
container arguments, and POD_NAMESPACE should be in the ingress-nginx namespace.
