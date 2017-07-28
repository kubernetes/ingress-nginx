# Ingress examples

This directory contains a catalog of examples on how to run, configure and
scale Ingress. Please review the [prerequisities](PREREQUISITES.md) before
trying them.

## Basic cross platform

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Deployment | basic deployment of controllers | * | Beginner
TLS termination | terminate TLS at the ingress controller | * | Beginner
Name based virtual hosting | `Host` header routing | * | Beginner
Path routing | URL regex routing | * | Beginner
Health checking | configure/optimize health checks | * | Intermediate
Pipeline | pipeline cloud and nginx | * | Advanced

## AWS

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
AWS | basic deployment | nginx | Intermediate


## TLS

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
LetsEncrypt | acquire certs via ACME protocol | * | Intermediate
Intermediate certs | terminate TLS with intermediate certs | * | Advanced
Client certs | client cert authentication | nginx | Advanced
Re-encrypty | terminate, apply routing rules, re-encrypt | nginx | Advanced

## Scaling

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Daemonset | run multiple controllers in a daemonset | nginx/haproxy | Intermediate
Deployment | run multiple controllers as a deployment | nginx/haproxy | Intermediate
Multi-zone | bridge different zones in a single cluster | gce | Intermediate
Static-ip | a single ingress gets a single static ip | * | Intermediate
Geo-routing | route to geographically closest endpoint | nginx | Advanced
Multi-cluster | bridge Kubernetes clusters with Ingress | gce | Advanced

## Algorithms

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Session stickyness | route requests consistently to the same endpoint | nginx | Advanced
Least connections | route requests based on least connections | on-prem | Advanced
Weights | route requests to backends based on weights | nginx | Advanced

## Routing

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Redirects | send a 301 re-direct | nginx | Intermediate
URL-rewriting | re-write path | nginx | Intermediate
SNI + HTTP | HTTP routing based on SNI hostname | nginx | Advanced
SNI + TCP | TLS routing based on SNI hostname | nginx | Advanced

## Auth

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Basic auth | password protect your website | nginx | Intermediate
[External auth plugin](external-auth/nginx/README.md) | defer to an external auth service | nginx | Intermediate

## Protocols

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
TCP  | TCP loadbalancing | nginx | Intermediate
UDP | UDP loadbalancing | nginx | Intermediate
Websockets | websockets loadbalancing | nginx | Intermediate
HTTP/2 | HTTP/2 loadbalancing | * | Intermediate
Proxy protocol | leverage the proxy protocol for source IP | nginx | Advanced

## Custom controllers

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
Dummy  | A simple dummy controller that logs updates | * | Advanced

## Customization

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
custom-headers  | set custom headers before send traffic to backends  | nginx | Advanced
configuration-snippets | customize nginx location configuration using annotations | nginx | Advanced

## RBAC

Name | Description | Platform   | Complexity Level
-----| ----------- | ---------- | ----------------
rbac | Configuring Role Base Access Control | nginx | intermediate
