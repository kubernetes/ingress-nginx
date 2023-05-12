# Overview

This is the documentation for the Ingress NGINX Controller.

It is built around the [Kubernetes Ingress resource](https://kubernetes.io/docs/concepts/services-networking/ingress/), using a [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) to store the controller configuration.

You can learn more about using [Ingress](http://kubernetes.io/docs/user-guide/ingress/) in the official [Kubernetes documentation](https://docs.k8s.io).

## Getting Started

See [Deployment](./deploy/) for a whirlwind tour that will get you started.


# FAQ - Kubernetes 1.22 Migration

If you are using Ingress objects in your cluster (running Kubernetes older than v1.22),
and you plan to upgrade to Kubernetes v1.22, please read [the migration guide here](./user-guide/k8s-122-migration.md).
