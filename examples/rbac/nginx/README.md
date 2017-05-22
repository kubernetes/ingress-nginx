# Role Based Access Control

This example demontrates how to apply role based access control

## Overview

This example applies to nginx-ingress-controllers being deployed in an
environment with RBAC enabled.

Role Based Access Control is comprised of four layers:

1.  `ClusterRole` - permissions assigned to a role that apply to an entire cluster
2.  `ClusterRoleBinding` - binding a ClusterRole to a specific account
3.  `Role` - permissions assigned to a role that apply to a specific namespace
4.  `RoleBinding` - binding a Role to a specific account

In order for RBAC to be applied to an nginx-ingress-controller, that controller
should be assigned to a `ServiceAccount`.  That `ServiceAccount` should be
bound to the `Role`s and `ClusterRole`s defined for the
nginx-ingress-controller.

## Service Accounts created in this example

One ServiceAccount is created in this example, `nginx-ingress-serviceaccount`.

## Permissions Granted in this example

There are two sets of permissions defined in this example.  Cluster-wide
permissions defined by the `ClusterRole` named `nginx-ingress-clusterrole`, and
namespace specific permissions defined by the `Role` named
`nginx-ingress-role`.

### Cluster Permissions

These permissions are granted in order for the nginx-ingress-controller to be
able to function as an ingress across the cluster.  These permissions are
granted to the ClusterRole named `nginx-ingress-clusterrole`

* `configmaps`, `endpoints`, `nodes`, `pods`, `secrets`: list, watch
* `services`, `ingresses`: get, list, watch
* `events`: create, patch
* `ingresses/status`: update

### Namespace Permissions

These permissions are granted specific to the nginx-ingress namespace.  These
permissions are granted to the Role named `nginx-ingress-role`

* `configmaps`, `pods`, `secrets`: get
* `endpoints`: create, get, update

### Bindings

The ServiceAccount `nginx-ingress-serviceaccount` is bound to the Role
`nginx-ingress-role` and the ClusterRole `nginx-ingress-clusterrole`.

## Namespace created in this example

The `Namespace` named `nginx-ingress` is defined in this example.  The
namespace name can be changed arbitrarily as long as all of the references
change as well.


## Usage

1.  Create the `Namespace`, `Service Account`, `ClusterRole`, `Role`,
`ClusterRoleBinding`, and `RoleBinding`.

```sh
kubectl create -f ./nginx-ingress-controller-rbac.yml
```

2.  Create the nginx-ingress-controller

For this example to work, the Service must be in the nginx-ingress namespace:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress
  namespace: nginx-ingress #match namespace of service account and role
spec:
  type: LoadBalancer
  ports:
    - port: 80
      name: http
    - port: 443
      name: https
  selector:
    k8s-app: nginx-ingress-lb
```

The serviceAccountName associated with the containers in the deployment must
match the serviceAccount from nginx-ingress-controller-rbac.yml  The namespace
references in the Deployment metadata, container arguments, and POD_NAMESPACE
should be in the nginx-ingress namespace.

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  #match namespace of service account and role
  namespace: nginx-ingress 
spec:
  replicas: 2
  template:
    metadata:
      labels:
        k8s-app: nginx-ingress-lb
    spec:
      #match name of service account
      serviceAccountName: nginx-ingress-serviceaccount
      containers:
        - name: nginx-ingress-controller
          image: gcr.io/google_containers/nginx-ingress-controller:version
          #namespace matching is required in some arguments
           args:
            - /nginx-ingress-controller
            - --default-backend-service=default/default-http-backend
            - --default-ssl-certificate=$(POD_NAMESPACE)/tls-certificate
         env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            #match namespace of service account and role
            - name: POD_NAMESPACE 
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace	  
          ports:
            - containerPort: 80
            - containerPort: 443

```
