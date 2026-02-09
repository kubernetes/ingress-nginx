# Retirement

[What You Need to Know about Ingress NGINX Retirement](https://www.kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/):

* Best-effort maintenance will continue until March 2026.
* Afterward, there will be no further releases, no bugfixes, and no updates to resolve any security vulnerabilities that may be discovered.
* Existing deployments of Ingress NGINX will not be broken.
  * Existing project artifacts such as Helm charts and container images will remain available.

You can still find the documentation for the Ingress NGINX Controller on this page.

# Overview

The Ingress NGINX Controller is built around the [Kubernetes Ingress resource](https://kubernetes.io/docs/concepts/services-networking/ingress/), using a [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) to store the controller configuration.

You can learn more about using [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) in the official [Kubernetes documentation](https://docs.k8s.io).

# Getting Started

See [Deployment](./deploy/index.md) for a whirlwind tour that will get you started.

