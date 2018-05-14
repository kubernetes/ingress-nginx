# Role Based Access Control (RBAC)

## Overview

This example applies to nginx-ingress-controllers being deployed in an environment with RBAC enabled.

Role Based Access Control is comprised of four layers:

1. `ClusterRole` - permissions assigned to a role that apply to an entire cluster
2. `ClusterRoleBinding` - binding a ClusterRole to a specific account
3. `Role` - permissions assigned to a role that apply to a specific namespace
4. `RoleBinding` - binding a Role to a specific account

In order for RBAC to be applied to an nginx-ingress-controller, that controller
should be assigned to a `ServiceAccount`.  That `ServiceAccount` should be
bound to the `Role`s and `ClusterRole`s defined for the nginx-ingress-controller.

## Service Accounts created in this example

One ServiceAccount is created in this example, `nginx-ingress-serviceaccount`.

## Permissions Granted in this example

There are two sets of permissions defined in this example.  Cluster-wide
permissions defined by the `ClusterRole` named `nginx-ingress-clusterrole`, and
namespace specific permissions defined by the `Role` named `nginx-ingress-role`.

### Cluster Permissions

These permissions are granted in order for the nginx-ingress-controller to be
able to function as an ingress across the cluster.  These permissions are
granted to the ClusterRole named `nginx-ingress-clusterrole`

* `configmaps`, `endpoints`, `nodes`, `pods`, `secrets`: list, watch
* `nodes`: get
* `services`, `ingresses`: get, list, watch
* `events`: create, patch
* `ingresses/status`: update

### Namespace Permissions

These permissions are granted specific to the nginx-ingress namespace.  These
permissions are granted to the Role named `nginx-ingress-role`

* `configmaps`, `pods`, `secrets`: get
* `endpoints`: get

Furthermore to support leader-election, the nginx-ingress-controller needs to
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
nginx-ingress-controller.

### Bindings

The ServiceAccount `nginx-ingress-serviceaccount` is bound to the Role
`nginx-ingress-role` and the ClusterRole `nginx-ingress-clusterrole`.

The serviceAccountName associated with the containers in the deployment must
match the serviceAccount. The namespace references in the Deployment metadata, 
container arguments, and POD_NAMESPACE should be in the nginx-ingress namespace.
