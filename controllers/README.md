# Ingress controllers

This directory contains ingress controllers.
=======
# Ingress Controllers

Configuring a webserver or loadbalancer is harder than it should be. Most webserver configuration files are very similar. There are some applications that have weird little quirks that tend to throw a wrench in things, but for the most part you can apply the same logic to them and achieve a desired result. The Ingress resource embodies this idea, and an Ingress controller is meant to handle all the quirks associated with a specific "class" of Ingress (be it a single instance of a loadbalancer, or a more complicated setup of frontends that provide GSLB, DDoS protection etc).

## What is an Ingress Controller?

An Ingress Controller is a daemon, deployed as a Kubernetes Pod, that watches the apiserver's `/ingresses` endpoint for updates to the [Ingress resource](https://kubernetes.io/docs/concepts/services-networking/ingress/). Its job is to satisfy requests for ingress.
