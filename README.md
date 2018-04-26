## Help us to improve the NGINX Ingress controller [completing the survey](https://docs.google.com/forms/d/15ULTOvYDsV920V0GWrspew4yyjEmTAi740Wr34UgKwA/viewform)

---

# NGINX Ingress Controller

[![Build Status](https://travis-ci.org/kubernetes/ingress-nginx.svg?branch=master)](https://travis-ci.org/kubernetes/ingress-nginx)
[![Coverage Status](https://codecov.io/gh/kubernetes/ingress-nginx/branch/master/graph/badge.svg)](https://codecov.io/gh/kubernetes/ingress-nginx)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/ingress-nginx)](https://goreportcard.com/report/github.com/kubernetes/ingress-nginx)

## Description

This repository contains the NGINX controller built around the [Kubernetes Ingress resource](http://kubernetes.io/docs/user-guide/ingress/) that uses [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#understanding-configmaps-and-pods) to store the NGINX configuration.

Learn more about using Ingress on [k8s.io](http://kubernetes.io/docs/user-guide/ingress/)

### What is an Ingress Controller?

Configuring a webserver or loadbalancer is harder than it should be. Most webserver configuration files are very similar. There are some applications that have weird little quirks that tend to throw a wrench in things, but for the most part you can apply the same logic to them and achieve a desired result.

The Ingress resource embodies this idea, and an Ingress controller is meant to handle all the quirks associated with a specific "class" of Ingress.

An Ingress Controller is a daemon, deployed as a Kubernetes Pod, that watches the apiserver's `/ingresses` endpoint for updates to the [Ingress resource](https://kubernetes.io/docs/concepts/services-networking/ingress/). Its job is to satisfy requests for Ingresses.

## Documentation

See [docs/index.md](docs/index.md) for detailed documentation.
