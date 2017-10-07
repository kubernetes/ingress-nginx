# Ingress examples

This directory contains a catalog of examples on how to run, configure and
scale Ingress. Please review the [prerequisites](PREREQUISITES.md) before
trying them.

## Basic cross platform

Name | Description | Complexity Level
-----| ----------- | ----------------
Deployment | basic deployment of controllers | Beginner
TLS termination | terminate TLS at the ingress controller | Beginner
Name based virtual hosting | `Host` header routing | Beginner
Path routing | URL regex routing | Beginner
Health checking | configure/optimize health checks | Intermediate
Pipeline | pipeline cloud and nginx | Advanced

## AWS

Name | Description | Complexity Level
-----| ----------- | ----------------
AWS | basic deployment | Intermediate


## TLS

Name | Description | Complexity Level
-----| ----------- | ----------------
LetsEncrypt | acquire certs via ACME protocol | Intermediate
Intermediate certs | terminate TLS with intermediate certs | Advanced
Client certs | client cert authentication |  Advanced
Re-encrypty | terminate, apply routing rules, re-encrypt |  Advanced

## Scaling

Name | Description | Complexity Level
-----| ----------- | ----------------
Daemonset | run multiple controllers in a daemonset | Intermediate
Deployment | run multiple controllers as a deployment | Intermediate
Static-ip | a single ingress gets a single static ip |  Intermediate
Geo-routing | route to geographically closest endpoint  | Advanced

## Algorithms

Name | Description | Complexity Level
-----| ----------- | ----------------
Session stickyness | route requests consistently to the same endpoint | Advanced
Least connections | route requests based on least connections | Advanced
Weights | route requests to backends based on weights | Advanced

## Routing

Name | Description | Complexity Level
-----| ----------- | ----------------
Redirects | send a 301 re-direct | Intermediate
URL-rewriting | re-write path | Intermediate
SNI + HTTP | HTTP routing based on SNI hostname | Advanced
SNI + TCP | TLS routing based on SNI hostname | Advanced

## Auth

Name | Description | Complexity Level
-----| ----------- | ----------------
Basic auth | password protect your website | nginx | Intermediate
[External auth plugin](external-auth/README.md) | defer to an external auth service | Intermediate

## Protocols

Name | Description | Complexity Level
-----| ----------- | ----------------
TCP  | TCP loadbalancing | Intermediate
UDP | UDP loadbalancing | Intermediate
Websockets | websockets loadbalancing | Intermediate
HTTP/2 | HTTP/2 loadbalancing | Intermediate
Proxy protocol | leverage the proxy protocol for source IP | Advanced

## Custom controllers

Name | Description | Complexity Level
-----| ----------- | ----------------
Dummy  | A simple dummy controller that logs updates | Advanced

## Customization

Name | Description | Complexity Level
-----| ----------- | ----------------
custom-headers  | set custom headers before send traffic to backends  | Advanced
configuration-snippets | customize nginx location configuration using annotations | Advanced

## RBAC

Name | Description | Complexity Level
-----| ----------- | ----------------
rbac | Configuring Role Base Access Control | intermediate
