# Ingress examples

This directory contains a catalog of examples on how to run, configure and
scale Ingress. Please review the [prerequisites](PREREQUISITES.md) before
trying them.

## Scaling

Name | Description | Complexity Level
-----| ----------- | ----------------
Static-ip | a single ingress gets a single static ip |  Intermediate

## Algorithms

Name | Description | Complexity Level
-----| ----------- | ----------------
Session stickyness | route requests consistently to the same endpoint | Advanced

## Auth

Name | Description | Complexity Level
-----| ----------- | ----------------
Basic auth | password protect your website | nginx | Intermediate
[External auth plugin](external-auth/README.md) | defer to an external auth service | Intermediate

## Customization

Name | Description | Complexity Level
-----| ----------- | ----------------
configuration-snippets | customize nginx location configuration using annotations | Advanced
custom-headers  | set custom headers before send traffic to backends  | Advanced
