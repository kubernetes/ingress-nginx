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

## Contents

- [Conventions](#conventions)
- [Requirements](#requirements)
- [Deployment](deploy/README.md)
- [Command line arguments](docs/user-guide/cli-arguments.md)
- [Contribute](CONTRIBUTING.md)
- [TLS](docs/user-guide/tls.md)
- [Annotation ingress.class](#annotation-ingressclass)
- [Customizing NGINX](#customizing-nginx)
  - [Custom NGINX configuration](docs/user-guide/configmap.md)
  - [Annotations](docs/user-guide/annotations.md)
- [Source IP address](#source-ip-address)
- [Exposing TCP and UDP Services](docs/user-guide/exposing-tcp-udp-services.md)
- [Proxy Protocol](#proxy-protocol)
- [ModSecurity Web Application Firewall](docs/user-guide/modsecurity.md)
- [OpenTracing](docs/user-guide/opentracing.md)
- [VTS and Prometheus metrics](docs/examples/customization/custom-vts-metrics-prometheus/README.md)
- [Custom errors](docs/user-guide/custom-errors.md)
- [NGINX status page](docs/user-guide/nginx-status-page.md)
- [Running multiple ingress controllers](#running-multiple-ingress-controllers)
- [Disabling NGINX ingress controller](#disabling-nginx-ingress-controller)
- [Retries in non-idempotent methods](#retries-in-non-idempotent-methods)
- [Log format](docs/user-guide/log-format.md)
- [Websockets](#websockets)
- [Optimizing TLS Time To First Byte (TTTFB)](#optimizing-tls-time-to-first-byte-tttfb)
- [Debug & Troubleshooting](docs/troubleshooting.md)
- [Limitations](#limitations)
- [Why endpoints and not services?](#why-endpoints-and-not-services)
- [External Articles](docs/user-guide/external-articles.md)

## Conventions

Anytime we reference a tls secret, we mean (x509, pem encoded, RSA 2048, etc). You can generate such a certificate with:
`openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${KEY_FILE} -out ${CERT_FILE} -subj "/CN=${HOST}/O=${HOST}"`
and create the secret via `kubectl create secret tls ${CERT_NAME} --key ${KEY_FILE} --cert ${CERT_FILE}`

## Requirements

The default backend is a service which handles all url paths and hosts the nginx controller doesn't understand (i.e., all the requests that are not mapped with an Ingress).
Basically a default backend exposes two URLs:

- `/healthz` that returns 200
- `/` that returns 404

The sub-directory [`/images/404-server`](https://github.com/kubernetes/ingress-nginx/tree/master/images/404-server) provides a service which satisfies the requirements for a default backend.  The sub-directory [`/images/custom-error-pages`](https://github.com/kubernetes/ingress-nginx/tree/master/images/custom-error-pages) provides an additional service for the purpose of customizing the error pages served via the default backend.

## Annotation ingress.class

If you have multiple Ingress controllers in a single cluster, you can pick one by specifying the `ingress.class` 
annotation, eg creating an Ingress with an annotation like

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "gce"
```

will target the GCE controller, forcing the nginx controller to ignore it, while an annotation like

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "nginx"
```

will target the nginx controller, forcing the GCE controller to ignore it.

__Note__: Deploying multiple ingress controller and not specifying the annotation will result in both controllers fighting to satisfy the Ingress.

### Customizing NGINX

There are three  ways to customize NGINX:

1. [ConfigMap](docs/user-guide/configmap.md): using a Configmap to set global configurations in NGINX.
2. [Annotations](docs/user-guide/annotations.md): use this if you want a specific configuration for a particular Ingress rule.
3. [Custom template](docs/user-guide/custom-template.md): when more specific settings are required, like [open_file_cache](http://nginx.org/en/docs/http/ngx_http_core_module.html#open_file_cache), adjust [listen](http://nginx.org/en/docs/http/ngx_http_core_module.html#listen) options as `rcvbuf` or when is not possible to change the configuration through the ConfigMap.

## Source IP address

By default NGINX uses the content of the header `X-Forwarded-For` as the source of truth to get information about the client IP address. This works without issues in L7 **if we configure the setting `proxy-real-ip-cidr`** with the correct information of the IP/network address of trusted external load balancer.

If the ingress controller is running in AWS we need to use the VPC IPv4 CIDR.

Another option is to enable proxy protocol using `use-proxy-protocol: "true"`.

In this mode NGINX does not use the content of the header to get the source IP address of the connection.

## Proxy Protocol

If you are using a L4 proxy to forward the traffic to the NGINX pods and terminate HTTP/HTTPS there, you will lose the remote endpoint's IP address. To prevent this you could use the [Proxy Protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) for forwarding traffic, this will send the connection details before forwarding the actual TCP connection itself.

Amongst others [ELBs in AWS](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) and [HAProxy](http://www.haproxy.org/) support Proxy Protocol.

### Running multiple ingress controllers

If you're running multiple ingress controllers, or running on a cloud provider that natively handles ingress, you need to specify the annotation `kubernetes.io/ingress.class: "nginx"` in all ingresses that you would like this controller to claim.  This mechanism also provides users the ability to run _multiple_ NGINX ingress controllers (e.g. one which serves public traffic, one which serves "internal" traffic).  When utilizing this functionality the option `--ingress-class` should be changed to a value unique for the cluster within the definition of the replication controller. Here is a partial example:

```
spec:
  template:
     spec:
       containers:
         - name: nginx-ingress-internal-controller
           args:
             - /nginx-ingress-controller
             - '--default-backend-service=ingress/nginx-ingress-default-backend'
             - '--election-id=ingress-controller-leader-internal'
             - '--ingress-class=nginx-internal'
             - '--configmap=ingress/nginx-ingress-internal-controller'
```

Not specifying the annotation will lead to multiple ingress controllers claiming the same ingress. Specifying a value which does not match the class of any existing ingress controllers will result in all ingress controllers ignoring the ingress.

The use of multiple ingress controllers in a single cluster is supported in Kubernetes versions >= 1.3.

### Websockets

Support for websockets is provided by NGINX out of the box. No special configuration required.

The only requirement to avoid the close of connections is the increase of the values of `proxy-read-timeout` and `proxy-send-timeout`.

The default value of this settings is `60 seconds`.

A more adequate value to support websockets is a value higher than one hour (`3600`).

**Important:** If the NGINX ingress controller is exposed with a service `type=LoadBalancer` make sure the protocol between the loadbalancer and NGINX is TCP.

### Optimizing TLS Time To First Byte (TTTFB)

NGINX provides the configuration option [ssl_buffer_size](http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) to allow the optimization of the TLS record size.

This improves the [TLS Time To First Byte](https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/) (TTTFB).
The default value in the Ingress controller is `4k` (NGINX default is `16k`).

### Retries in non-idempotent methods

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error.
The previous behavior can be restored using `retry-non-idempotent=true` in the configuration ConfigMap.

### Disabling NGINX ingress controller

Setting the annotation `kubernetes.io/ingress.class` to any other value  which does not match a valid ingress class will force the NGINX Ingress controller to ignore your Ingress.  If you are only running a single NGINX ingress controller, this can be achieved by setting this to any value except "nginx" or an empty string.

Do this if you wish to use one of the other Ingress controllers at the same time as the NGINX controller.

### Limitations

- Ingress rules for TLS require the definition of the field `host`

### Why endpoints and not services

The NGINX ingress controller does not use [Services](http://kubernetes.io/docs/user-guide/services) to route traffic to the pods. Instead it uses the Endpoints API in order to bypass [kube-proxy](http://kubernetes.io/docs/admin/kube-proxy/) to allow NGINX features like session affinity and custom load balancing algorithms. It also removes some overhead, such as conntrack entries for iptables DNAT.
